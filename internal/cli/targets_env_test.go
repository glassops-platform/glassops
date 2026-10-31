package cli

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/run"
	"github.com/glassops-platform/glassops/internal/providers"
)

const envTargetConfig = `version: 1
provider: salesforce
targets:
  devhub:
    id: 00D000000000001AAA
    tier: devhub
    auth: { method: sfdx-url, url: $SF_HUB_URL }
  qa:
    id: 00D000000000004AAA
    tier: qa
    auth: { method: sfdx-url, url: $SF_QA_AUTH_URL }
  ephemeral:
    id: $SF_EPHEMERAL_ORG_ID
    tier: ephemeral
    auth: { method: sfdx-url, url: $SF_EPHEMERAL_AUTH_URL }
environments:
  ephemeral:
    devhub: devhub
  bare:
    devhub: devhub
`

// envTargetHarness has a fake provider that records every target it is asked to log in to,
// and reports each target as the kind its tier needs.
func envTargetHarness(t *testing.T) (*harness, *fakeProvider, *[]providers.Target) {
	t.Helper()
	h := newHarness(t)
	var logins []providers.Target
	fp := &fakeProvider{
		detect: true,
		auth:   func(tg providers.Target) error { logins = append(logins, tg); return nil },
		describe: func(tg providers.Target) (providers.TargetInfo, error) {
			k := providers.KindSandbox
			switch tg.Tier {
			case "ephemeral":
				k = providers.KindScratch
			case "devhub":
				k = providers.KindProduction
			}
			return providers.TargetInfo{ID: tg.ID, Kind: k, Facts: map[string]any{}}, nil
		},
		validateFn: func(providers.Target, providers.ChangeSet, providers.ValidateOptions) (providers.ValidateResult, error) {
			return passingResult(), nil
		},
	}
	h.app.Providers = providers.Registry{fp}
	h.write(t, "sfdx-project.json", "{}")
	h.write(t, "glassops.yaml", envTargetConfig)
	h.env["SF_HUB_URL"] = "force://hub"
	h.env["SF_QA_AUTH_URL"] = "force://qa"
	writePlan(t, h, false)
	return h, fp, &logins
}

// synthAuthFile writes a synthetic 0600 auth file where env create's provider would.
func synthAuthFile(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "glassops-env-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	path := filepath.Join(dir, "sfdx-url.txt")
	if err := os.WriteFile(path, []byte("force://PlatformCLI::SYNTHETIC@example.my.salesforce.com\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func envFileFor(h *harness, target string) string {
	return filepath.Join(h.dir, ".glassops", "env", target, "env.json")
}

func TestEnvCreateThenValidateLogsInToTheEnv(t *testing.T) {
	h, fp, logins := envTargetHarness(t)
	// Config values for the target are set, and must be ignored while env.json exists.
	h.env["SF_EPHEMERAL_ORG_ID"] = "00D000000000009AAA"
	h.env["SF_EPHEMERAL_AUTH_URL"] = "force://from-config"
	authFile := synthAuthFile(t)
	fp.createEnv = func(providers.Session, providers.EnvRequest) (providers.Env, error) {
		return providers.Env{ID: "00D000000000002AAA", RecordID: "2SR000000000001AAA", Name: "glassops-x-1-1",
			ExpiresAt: time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC), AuthFile: authFile}, nil
	}
	fp.deleteEnv = func(providers.Session, providers.Env) error { return nil }

	if code := h.exec("env", "create"); code != exitcode.OK {
		t.Fatalf("env create exit %d, output:\n%s", code, h.out)
	}
	*logins = nil
	if code := h.exec("validate", "--target", "ephemeral"); code != exitcode.OK {
		t.Fatalf("validate exit %d, output:\n%s", code, h.out)
	}
	if len(*logins) != 1 {
		t.Fatalf("logins %+v", *logins)
	}
	got := (*logins)[0]
	if got.Name != "ephemeral" || got.ID != "00D000000000002AAA" || got.AuthFile != authFile || got.Tier != "ephemeral" ||
		got.Auth.Method != "sfdx-url" || got.Auth.URL != "" {
		t.Fatalf("validate logged in to %+v", got)
	}
	if r := h.result(t, "validate"); r.Target != "ephemeral" || r.Facts["validate.fallback_used"] != false {
		t.Fatalf("validate result %+v", r)
	}

	// env delete removes env.json and the auth file, and the target goes back to the config.
	if code := h.exec("env", "delete"); code != exitcode.OK {
		t.Fatalf("env delete exit %d, output:\n%s", code, h.out)
	}
	for _, p := range []string{envFileFor(h, "ephemeral"), authFile} {
		if _, err := os.Stat(p); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("%s kept after env delete: %v", p, err)
		}
	}
	*logins = nil
	if code := h.exec("validate", "--target", "ephemeral"); code != exitcode.OK {
		t.Fatalf("validate after delete exit %d, output:\n%s", code, h.out)
	}
	if got := (*logins)[0]; got.ID != "00D000000000009AAA" || got.AuthFile != "" || got.Auth.URL != "$SF_EPHEMERAL_AUTH_URL" {
		t.Fatalf("after delete logged in to %+v", got)
	}
}

func TestEnvTargetResolution(t *testing.T) {
	ready := func(target, authFile string) EnvState {
		return EnvState{Schema: EnvSchema, Provider: "salesforce", Target: target, DevHub: "devhub",
			ID: "00D000000000002AAA", Name: "glassops-x-1-1", AuthFile: authFile, Ready: true}
	}
	cases := []struct {
		name    string
		target  string
		state   func(t *testing.T) *EnvState
		raw     string
		wantErr string
		notConf bool
		check   func(t *testing.T, tg providers.Target)
	}{
		{"env-only target gets the ephemeral tier", "bare", func(t *testing.T) *EnvState { s := ready("bare", synthAuthFile(t)); return &s }, "", "", false,
			func(t *testing.T, tg providers.Target) {
				if tg.Tier != "ephemeral" || tg.ID != "00D000000000002AAA" || tg.AuthFile == "" {
					t.Fatalf("target %+v", tg)
				}
			}},
		{"no env.json uses the config", "ephemeral", nil, "", "target not configured", true, nil},
		{"env-only target without env.json is not configured", "bare", nil, "", "target not configured", true, nil},
		{"create that failed is not configured", "ephemeral", func(t *testing.T) *EnvState {
			s := ready("ephemeral", "")
			s.Ready, s.ID = false, ""
			return &s
		}, "", "was not created", true, nil},
		{"env.json for another target", "ephemeral", func(t *testing.T) *EnvState { s := ready("other", synthAuthFile(t)); return &s }, "", "is for target other", false, nil},
		{"unreadable env.json", "ephemeral", nil, "{broken", "parse", false, nil},
		{"auth file outside temp and .glassops", "ephemeral", func(t *testing.T) *EnvState {
			p := filepath.Join(t.TempDir(), "outside", "sfdx-url.txt")
			_ = os.MkdirAll(filepath.Dir(p), 0o700)
			_ = os.WriteFile(p, []byte("force://x"), 0o600)
			for _, k := range []string{"TMPDIR", "TMP", "TEMP"} {
				t.Setenv(k, t.TempDir())
			}
			s := ready("ephemeral", p)
			return &s
		}, "", "outside the temp dir and .glassops", false, nil},
		{"auth file inside the project .glassops", "ephemeral", nil, "", "", false, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h, _, _ := envTargetHarness(t)
			proj, err := h.app.loadProject(&globals{dir: h.dir})
			if err != nil {
				t.Fatal(err)
			}
			switch {
			case tc.state != nil:
				if err := writeEnvState(envFileFor(h, tc.target), *tc.state(t)); err != nil {
					t.Fatal(err)
				}
			case tc.raw != "":
				h.write(t, ".glassops/env/"+tc.target+"/env.json", tc.raw)
			case strings.HasPrefix(tc.name, "auth file inside"):
				p := filepath.Join(h.dir, ".glassops", "env", tc.target, "sfdx-url.txt")
				h.write(t, ".glassops/env/"+tc.target+"/sfdx-url.txt", "force://x")
				if err := writeEnvState(envFileFor(h, tc.target), ready(tc.target, p)); err != nil {
					t.Fatal(err)
				}
			}
			tg, err := h.app.target(proj.cfg, tc.target)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("err %v, want %q", err, tc.wantErr)
				}
				if errors.Is(err, ErrTargetNotConfigured) != tc.notConf {
					t.Fatalf("not configured = %v, want %v", errors.Is(err, ErrTargetNotConfigured), tc.notConf)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if tc.check != nil {
				tc.check(t, tg)
			}
		})
	}
}

func TestEnvJSONIgnoredForTargetNotUnderEnvironments(t *testing.T) {
	h, _, logins := envTargetHarness(t)
	if err := writeEnvState(envFileFor(h, "qa"), EnvState{Schema: EnvSchema, Target: "qa", DevHub: "devhub",
		ID: "00D000000000002AAA", AuthFile: synthAuthFile(t), Ready: true}); err != nil {
		t.Fatal(err)
	}
	if code := h.exec("validate", "--target", "qa"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if got := (*logins)[0]; got.ID != "00D000000000004AAA" || got.AuthFile != "" {
		t.Fatalf("logged in to %+v", got)
	}
}

func TestFailedCreateFallsBackLikeAnUnsetTarget(t *testing.T) {
	h, fp, logins := envTargetHarness(t)
	fp.createEnv = func(providers.Session, providers.EnvRequest) (providers.Env, error) {
		return providers.Env{RecordID: "2SR000000000001AAA", Name: "glassops-x-1-1"}, errors.New("salesforce: the scratch org was not active after 15m0s")
	}
	if code := h.exec("env", "create"); code != exitcode.Error {
		t.Fatalf("env create exit %d", code)
	}
	*logins = nil
	if code := h.exec("validate", "--target", "ephemeral", "--fallback-target", "qa"); code != exitcode.OK {
		t.Fatalf("validate exit %d, output:\n%s", code, h.out)
	}
	r := h.result(t, "validate")
	if r.Target != "qa" || r.Facts["validate.fallback_used"] != true || len(*logins) != 1 || (*logins)[0].Name != "qa" {
		t.Fatalf("result %+v, logins %+v", r, *logins)
	}
	if checkStatus(r, "validate") != run.StatusPass {
		t.Fatalf("result %+v", r)
	}
}
