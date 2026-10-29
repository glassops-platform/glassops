package cli

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/run"
	"github.com/glassops-platform/glassops/internal/providers"
)

const (
	hubID     = "00D000000000001AAA"
	scratchID = "00D000000000002AAA"
)

const twoTargetConfig = `version: 1
provider: salesforce
targets:
  devhub:
    id: 00D000000000001AAA
    tier: devhub
    auth: { method: sfdx-url, url: $SF_HUB_URL }
  scratch:
    id: 00D000000000002AAA
    tier: ephemeral
    auth: { method: sfdx-url, url: $SF_SCRATCH_URL }
`

// useFake swaps the harness's provider for f and writes the two-target config.
func useFake(t *testing.T, h *harness, f *fakeProvider) {
	t.Helper()
	f.detect = true
	h.app.Providers = providers.Registry{f}
	h.write(t, "sfdx-project.json", "{}")
	h.write(t, "glassops.yaml", twoTargetConfig)
	h.env["SF_HUB_URL"] = "force://hub"
	h.env["SF_SCRATCH_URL"] = "force://scratch"
}

func okAuth(providers.Target) error { return nil }

func TestDescribeWritesFactsAndTarget(t *testing.T) {
	h := newHarness(t)
	f := &fakeProvider{auth: okAuth, describe: func(tg providers.Target) (providers.TargetInfo, error) {
		return providers.TargetInfo{ID: tg.ID, Kind: providers.KindScratch, Facts: map[string]any{"salesforce.org_type": "Developer Edition"}}, nil
	}}
	useFake(t, h, f)
	if code := h.exec("describe", "--target", "scratch"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	r := h.result(t, "describe")
	if r.Status != run.StatusPass || r.Target != "scratch" || r.Provider != "salesforce" {
		t.Fatalf("result: %+v", r)
	}
	if r.Facts["target.id"] != scratchID || r.Facts["target.kind"] != "scratch" || r.Facts["salesforce.org_type"] != "Developer Edition" {
		t.Fatalf("facts: %v", r.Facts)
	}
	if !f.allClosed() || len(f.sessions) != 1 {
		t.Fatalf("sessions: %d, closed %v", len(f.sessions), f.allClosed())
	}
}

func TestDescribeStopsWithExitTwo(t *testing.T) {
	cases := []struct {
		name    string
		auth    func(providers.Target) error
		desc    func(providers.Target) (providers.TargetInfo, error)
		wantErr string
	}{
		{"guard id mismatch", okAuth, func(providers.Target) (providers.TargetInfo, error) {
			return providers.TargetInfo{ID: "00D000000000009AAA", Kind: providers.KindProduction}, nil
		}, "guard:"},
		{"guard tier mismatch", okAuth, reportsAs(providers.KindSandbox), "guard:"},
		{"login failure", func(providers.Target) error { return errors.New("invalid_grant") }, nil, "login failed"},
		{"describe failure", okAuth, func(providers.Target) (providers.TargetInfo, error) {
			return providers.TargetInfo{}, errors.New("unreadable")
		}, "unreadable"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := newHarness(t)
			f := &fakeProvider{auth: tc.auth, describe: tc.desc}
			useFake(t, h, f)
			if code := h.exec("describe", "--target", "devhub"); code != exitcode.Error {
				t.Fatalf("exit %d, want 2. output:\n%s", code, h.out)
			}
			r := h.result(t, "describe")
			if r.Status != run.StatusError || len(r.Errors) != 1 || !strings.Contains(r.Errors[0], tc.wantErr) {
				t.Fatalf("result: %+v", r)
			}
			if !strings.Contains(h.out.String(), tc.wantErr) {
				t.Fatalf("the reason must be printed, output:\n%s", h.out)
			}
			if !f.allClosed() {
				t.Fatal("a session was left open")
			}
		})
	}
}

func TestDescribeUnknownTargetAndMissingFlag(t *testing.T) {
	h := newHarness(t)
	useFake(t, h, &fakeProvider{auth: okAuth, describe: reportsAs(providers.KindProduction)})
	if code := h.exec("describe", "--target", "nope"); code != exitcode.Error {
		t.Fatalf("unknown target exit %d", code)
	}
	if code := h.exec("describe"); code != exitcode.Error {
		t.Fatalf("missing --target exit %d", code)
	}
}

func TestDescribeUnsetIDVariable(t *testing.T) {
	h := newHarness(t)
	useFake(t, h, &fakeProvider{auth: okAuth, describe: reportsAs(providers.KindProduction)})
	h.write(t, "glassops.yaml", "version: 1\nprovider: salesforce\ntargets:\n  hub:\n    id: $HUB_ID\n    tier: devhub\n    auth: { method: sfdx-url, url: $SF_HUB_URL }\n")
	if code := h.exec("describe", "--target", "hub"); code != exitcode.Error {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if _, err := os.Stat(filepath.Join(h.dir, ".glassops", "results", "describe.json")); err == nil {
		t.Fatal("a result was written for a target that could not be resolved")
	}
}
