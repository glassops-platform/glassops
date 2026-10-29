package cli

import (
	"errors"
	"strings"
	"testing"

	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/run"
	"github.com/glassops-platform/glassops/internal/providers"
)

func detail(r run.Result, name string) string {
	for _, c := range r.Checks {
		if c.Name == name {
			return c.Detail
		}
	}
	return ""
}

func TestDoctorWithoutAuthNeverAuthenticates(t *testing.T) {
	h := newHarness(t)
	f := &fakeProvider{
		auth: func(providers.Target) error { t.Error("Authenticate was called without --auth"); return nil },
		describe: func(providers.Target) (providers.TargetInfo, error) {
			t.Error("Describe was called without --auth")
			return providers.TargetInfo{}, nil
		},
	}
	useFake(t, h, f)
	if code := h.exec("doctor"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	for _, c := range h.result(t, "doctor").Checks {
		if strings.HasPrefix(c.Name, "auth") {
			t.Fatalf("unexpected check %s", c.Name)
		}
	}
	if len(f.sessions) != 0 {
		t.Fatal("a session was opened")
	}
}

func TestDoctorAuthChecksEveryTargetInOrder(t *testing.T) {
	h := newHarness(t)
	var order []string
	f := &fakeProvider{
		auth:     func(tg providers.Target) error { order = append(order, tg.Name); return nil },
		describe: reportsAs(providers.KindProduction),
	}
	useFake(t, h, f)
	// devhub needs production, scratch needs scratch, so give each what it needs.
	f.describe = func(tg providers.Target) (providers.TargetInfo, error) {
		k := providers.KindProduction
		if tg.Tier == "ephemeral" {
			k = providers.KindScratch
		}
		return providers.TargetInfo{ID: tg.ID, Kind: k}, nil
	}
	if code := h.exec("doctor", "--auth"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if strings.Join(order, ",") != "devhub,scratch" {
		t.Fatalf("order: %v", order)
	}
	r := h.result(t, "doctor")
	if d := detail(r, "auth.devhub"); d != "production "+hubID {
		t.Fatalf("auth.devhub detail %q", d)
	}
	if d := detail(r, "auth.scratch"); d != "scratch "+scratchID {
		t.Fatalf("auth.scratch detail %q", d)
	}
	if !f.allClosed() || len(f.sessions) != 2 {
		t.Fatalf("sessions %d, closed %v", len(f.sessions), f.allClosed())
	}
}

func TestDoctorAuthOnlyNamedTarget(t *testing.T) {
	h := newHarness(t)
	var seen []string
	f := &fakeProvider{
		auth:     func(tg providers.Target) error { seen = append(seen, tg.Name); return nil },
		describe: reportsAs(providers.KindScratch),
	}
	useFake(t, h, f)
	if code := h.exec("doctor", "--auth", "--target", "scratch"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if len(seen) != 1 || seen[0] != "scratch" {
		t.Fatalf("authenticated: %v", seen)
	}
}

func TestDoctorAuthErrorsExitTwo(t *testing.T) {
	cases := []struct {
		name string
		auth func(providers.Target) error
		desc func(providers.Target) (providers.TargetInfo, error)
		want string
	}{
		{"login failure", func(providers.Target) error { return errors.New("invalid_grant") }, nil, "login failed"},
		{"guard mismatch", okAuth, reportsAs(providers.KindSandbox), "guard:"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := newHarness(t)
			f := &fakeProvider{auth: tc.auth, describe: tc.desc}
			useFake(t, h, f)
			if code := h.exec("doctor", "--auth", "--target", "devhub"); code != exitcode.Error {
				t.Fatalf("exit %d, want 2. output:\n%s", code, h.out)
			}
			r := h.result(t, "doctor")
			if checkStatus(r, "auth.devhub") != run.StatusError || !strings.Contains(detail(r, "auth.devhub"), tc.want) {
				t.Fatalf("checks: %+v", r.Checks)
			}
			if !f.allClosed() {
				t.Fatal("a session was left open")
			}
		})
	}
}

func TestDoctorAuthContinuesAfterOneTargetFails(t *testing.T) {
	h := newHarness(t)
	f := &fakeProvider{
		auth: func(tg providers.Target) error {
			if tg.Name == "devhub" {
				return errors.New("invalid_grant")
			}
			return nil
		},
		describe: reportsAs(providers.KindScratch),
	}
	useFake(t, h, f)
	if code := h.exec("doctor", "--auth"); code != exitcode.Error {
		t.Fatalf("exit %d", code)
	}
	r := h.result(t, "doctor")
	if checkStatus(r, "auth.devhub") != run.StatusError || checkStatus(r, "auth.scratch") != run.StatusPass {
		t.Fatalf("checks: %+v", r.Checks)
	}
}

func TestDoctorAuthUnknownTarget(t *testing.T) {
	h := newHarness(t)
	f := &fakeProvider{auth: okAuth, describe: reportsAs(providers.KindProduction)}
	useFake(t, h, f)
	if code := h.exec("doctor", "--auth", "--target", "nope"); code != exitcode.Error {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if got := checkStatus(h.result(t, "doctor"), "auth.nope"); got != run.StatusError {
		t.Fatalf("auth.nope = %q", got)
	}
	if len(f.sessions) != 0 {
		t.Fatal("logged in for an unknown target")
	}
}

func TestDoctorAuthWithNoTargetsIsSkipped(t *testing.T) {
	h := newHarness(t)
	f := &fakeProvider{auth: func(providers.Target) error { t.Error("Authenticate called"); return nil }}
	useFake(t, h, f)
	h.write(t, "glassops.yaml", "version: 1\nprovider: salesforce\n")
	if code := h.exec("doctor", "--auth"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if got := checkStatus(h.result(t, "doctor"), "auth"); got != run.StatusSkipped {
		t.Fatalf("auth = %q", got)
	}
}

func TestDoctorAuthWithoutConfigIsSkipped(t *testing.T) {
	h := newHarness(t)
	h.write(t, "sfdx-project.json", "{}")
	h.write(t, "glassops.yaml", "version: 1\nfoo: bar\n")
	if code := h.exec("doctor", "--auth"); code != exitcode.Fail {
		t.Fatalf("exit %d", code)
	}
	if got := checkStatus(h.result(t, "doctor"), "auth"); got != run.StatusSkipped {
		t.Fatalf("auth = %q", got)
	}
}

func TestDoctorAuthUnsetIDVariable(t *testing.T) {
	h := newHarness(t)
	f := &fakeProvider{auth: func(providers.Target) error { t.Error("Authenticate called"); return nil }}
	useFake(t, h, f)
	h.write(t, "glassops.yaml", "version: 1\nprovider: salesforce\ntargets:\n  hub:\n    id: $HUB_ID\n    tier: devhub\n    auth: { method: sfdx-url, url: $SF_HUB_URL }\n")
	h.exec("doctor", "--auth")
	if got := checkStatus(h.result(t, "doctor"), "auth.hub"); got != run.StatusError {
		t.Fatalf("auth.hub = %q", got)
	}
}
