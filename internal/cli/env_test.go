package cli

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/glassops-platform/glassops/internal/ci"
	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/run"
	"github.com/glassops-platform/glassops/internal/providers"
)

const envConfig = `version: 1
provider: salesforce
targets:
  devhub:
    id: 00D000000000001AAA
    tier: devhub
    auth: { method: sfdx-url, url: $SF_HUB_URL }
  other:
    id: 00D000000000003AAA
    tier: devhub
    auth: { method: sfdx-url, url: $SF_HUB_URL }
environments:
  ephemeral:
    devhub: devhub
    definition: config/project-scratch-def.json
    ttl_days: 2
    shape: $SHAPE_ORG_ID
    prefix: gx
`

// synthEnv is a created environment as a provider reports it. Synthetic values.
func synthEnv() providers.Env {
	return providers.Env{
		ID: "00D000000000002AAA", RecordID: "2SR000000000001AAA", Name: "gx-repo-12-345",
		ExpiresAt: time.Date(2026, 10, 28, 0, 0, 0, 0, time.UTC),
		AuthFile:  filepath.Join(os.TempDir(), "glassops-env-1", "sfdx-url.txt"),
		Facts:     map[string]any{"salesforce.signup_attempts": 1},
	}
}

func envHarness(t *testing.T) (*harness, *fakeProvider) {
	t.Helper()
	h := newHarness(t)
	fp := &fakeProvider{detect: true, auth: okAuth, describe: reportsAs(providers.KindProduction)}
	h.app.Providers = providers.Registry{fp}
	h.write(t, "sfdx-project.json", "{}")
	h.write(t, "glassops.yaml", envConfig)
	h.env["SF_HUB_URL"] = "force://hub"
	h.env["SHAPE_ORG_ID"] = "00D000000000009AAA"
	return h, fp
}

func (h *harness) envFile() string {
	return filepath.Join(h.dir, ".glassops", "env", "ephemeral", "env.json")
}

func (h *harness) readEnv(t *testing.T) EnvState {
	t.Helper()
	s, err := readEnvState(h.envFile())
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestEnvCreateWritesEnvJSON(t *testing.T) {
	h, fp := envHarness(t)
	var got providers.EnvRequest
	var hubName string
	fp.createEnv = func(hub providers.Session, req providers.EnvRequest) (providers.Env, error) {
		got, hubName = req, hub.Target().Name
		return synthEnv(), nil
	}
	if code := h.exec("env", "create"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if hubName != "devhub" || got.DefinitionFile != "config/project-scratch-def.json" || got.Dir != h.dir ||
		got.TTL != 48*time.Hour || got.Shape != "00D000000000009AAA" {
		t.Fatalf("request %+v on %s", got, hubName)
	}
	wantName := "gx-" + nameJunk.ReplaceAllString(strings.ToLower(filepath.Base(h.dir)), "-")
	if !strings.HasPrefix(got.Name, strings.TrimRight(wantName, "-")) || !strings.HasSuffix(got.Name, "-20261026t120000") {
		t.Fatalf("name %q", got.Name)
	}
	s := h.readEnv(t)
	want := synthEnv()
	if s.Target != "ephemeral" || s.DevHub != "devhub" || s.ID != want.ID || s.RecordID != want.RecordID || s.Name != want.Name ||
		s.AuthFile != want.AuthFile || !s.Ready || s.ExpiresAt == nil || !s.ExpiresAt.Equal(want.ExpiresAt) || s.Provider != "salesforce" {
		t.Fatalf("env.json %+v", s)
	}
	r := h.result(t, "env-create")
	if r.Status != run.StatusPass || r.Target != "ephemeral" || r.Facts["env.id"] != want.ID || r.Facts["env.name"] != want.Name ||
		r.Facts["env.expires_at"] != "2026-10-28T00:00:00Z" || r.Facts["salesforce.signup_attempts"] != float64(1) {
		t.Fatalf("result %+v", r)
	}
	if !strings.Contains(h.out.String(), "output env_file="+h.envFile()) || !strings.Contains(h.out.String(), "output env_id=00D000000000002AAA") {
		t.Fatalf("outputs:\n%s", h.out)
	}
	if !fp.allClosed() {
		t.Fatal("hub session left open")
	}
}

func TestEnvCreateFlagsOverrideConfig(t *testing.T) {
	h, fp := envHarness(t)
	var got providers.EnvRequest
	fp.createEnv = func(_ providers.Session, req providers.EnvRequest) (providers.Env, error) {
		got = req
		return synthEnv(), nil
	}
	envFile := filepath.Join(t.TempDir(), "state", "env.json")
	if code := h.exec("env", "create", "--name", "mine-1", "--ttl-days", "5", "--shape", "00D000000000008", "--env-file", envFile); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if got.Name != "mine-1" || got.TTL != 5*24*time.Hour || got.Shape != "00D000000000008" {
		t.Fatalf("request %+v", got)
	}
	if _, err := readEnvState(envFile); err != nil {
		t.Fatal(err)
	}
}

func TestEnvCreateFailureAfterRequestStillRecords(t *testing.T) {
	h, fp := envHarness(t)
	fp.createEnv = func(providers.Session, providers.EnvRequest) (providers.Env, error) {
		return providers.Env{RecordID: "2SR000000000001AAA", Name: "gx-repo-1-1"}, errors.New("salesforce: the scratch org was not active after 15m0s")
	}
	if code := h.exec("env", "create"); code != exitcode.Error {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	s := h.readEnv(t)
	if s.RecordID != "2SR000000000001AAA" || s.Ready || s.ID != "" {
		t.Fatalf("env.json %+v", s)
	}
	r := h.result(t, "env-create")
	if checkStatus(r, "create") != run.StatusError || !strings.Contains(h.out.String(), "env delete can remove it") {
		t.Fatalf("result %+v\n%s", r, h.out)
	}
}

func TestEnvCreateFailureBeforeRequestWritesNothing(t *testing.T) {
	h, fp := envHarness(t)
	fp.createEnv = func(providers.Session, providers.EnvRequest) (providers.Env, error) {
		return providers.Env{}, errors.New("salesforce: env create cannot apply definition keys settings")
	}
	if code := h.exec("env", "create"); code != exitcode.Error {
		t.Fatalf("exit %d", code)
	}
	if _, err := os.Stat(h.envFile()); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("env.json written: %v", err)
	}
}

func TestEnvCreateRefusals(t *testing.T) {
	cases := []struct {
		name  string
		setup func(t *testing.T, h *harness, fp *fakeProvider)
		args  []string
		want  string
	}{
		{"env.json already exists", func(t *testing.T, h *harness, _ *fakeProvider) {
			if err := writeEnvState(h.envFile(), EnvState{Schema: EnvSchema, Target: "ephemeral", Name: "gx-old"}); err != nil {
				t.Fatal(err)
			}
		}, nil, `already records environment "gx-old"`},
		{"environment not configured", nil, []string{"--target", "nope"}, "environments.nope is not configured"},
		{"shape variable not set", func(_ *testing.T, h *harness, _ *fakeProvider) { delete(h.env, "SHAPE_ORG_ID") }, nil, "$SHAPE_ORG_ID, which is not set"},
		{"ttl below one", nil, []string{"--ttl-days", "-1"}, "--ttl-days must be 1 or more"},
		{"hub guard mismatch", func(_ *testing.T, _ *harness, fp *fakeProvider) {
			fp.describe = func(providers.Target) (providers.TargetInfo, error) {
				return providers.TargetInfo{ID: "00D000000000009AAA", Kind: providers.KindProduction}, nil
			}
		}, nil, "Id does not match"},
		{"hub login fails", func(_ *testing.T, _ *harness, fp *fakeProvider) {
			fp.auth = func(providers.Target) error { return errors.New("bad url") }
		}, nil, "login failed"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h, fp := envHarness(t)
			called := false
			fp.createEnv = func(providers.Session, providers.EnvRequest) (providers.Env, error) {
				called = true
				return synthEnv(), nil
			}
			if tc.setup != nil {
				tc.setup(t, h, fp)
			}
			if code := h.exec(append([]string{"env", "create"}, tc.args...)...); code != exitcode.Error {
				t.Fatalf("exit %d, output:\n%s", code, h.out)
			}
			if called || !strings.Contains(h.out.String(), tc.want) {
				t.Fatalf("called %v, output:\n%s", called, h.out)
			}
			if !fp.allClosed() {
				t.Fatal("session left open")
			}
		})
	}
}

func TestEnvDeleteReadsEnvJSON(t *testing.T) {
	h, fp := envHarness(t)
	want := synthEnv()
	if err := writeEnvState(h.envFile(), EnvState{Schema: EnvSchema, Target: "ephemeral", DevHub: "devhub",
		ID: want.ID, RecordID: want.RecordID, Name: want.Name, AuthFile: want.AuthFile}); err != nil {
		t.Fatal(err)
	}
	var got providers.Env
	fp.deleteEnv = func(hub providers.Session, env providers.Env) error {
		if hub.Target().Name != "devhub" {
			t.Errorf("hub %s", hub.Target().Name)
		}
		got = env
		return nil
	}
	if code := h.exec("env", "delete"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if got.ID != want.ID || got.RecordID != want.RecordID || got.Name != want.Name || got.AuthFile != want.AuthFile {
		t.Fatalf("delete %+v", got)
	}
	if _, err := os.Stat(h.envFile()); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("env.json kept: %v", err)
	}
	if r := h.result(t, "env-delete"); r.Status != run.StatusPass || checkStatus(r, "delete") != run.StatusPass {
		t.Fatalf("result %+v", r)
	}
	// A second delete finds nothing recorded and passes without logging in.
	before := len(fp.sessions)
	if code := h.exec("env", "delete"); code != exitcode.OK || len(fp.sessions) != before {
		t.Fatalf("second delete exit %d, sessions %d", code, len(fp.sessions)-before)
	}
	if r := h.result(t, "env-delete"); checkStatus(r, "delete") != run.StatusSkipped {
		t.Fatalf("result %+v", r)
	}
}

func TestEnvDeleteByName(t *testing.T) {
	cases := []struct {
		name      string
		withState bool
	}{{"no env.json", false}, {"env.json names another environment", true}}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h, fp := envHarness(t)
			if tc.withState {
				if err := writeEnvState(h.envFile(), EnvState{Schema: EnvSchema, Target: "ephemeral", DevHub: "devhub", ID: "00D000000000002AAA", Name: "gx-recorded"}); err != nil {
					t.Fatal(err)
				}
			}
			var got providers.Env
			fp.deleteEnv = func(_ providers.Session, env providers.Env) error { got = env; return nil }
			if code := h.exec("env", "delete", "--name", "gx-other"); code != exitcode.OK {
				t.Fatalf("exit %d, output:\n%s", code, h.out)
			}
			if got.Name != "gx-other" || got.ID != "" || got.RecordID != "" {
				t.Fatalf("delete %+v", got)
			}
			if _, err := os.Stat(h.envFile()); tc.withState && err != nil {
				t.Fatalf("env.json for another environment was removed: %v", err)
			}
		})
	}
}

func TestEnvDeleteRefusals(t *testing.T) {
	cases := []struct {
		name  string
		state *EnvState
		raw   string
		fail  error
		want  string
	}{
		{"recorded for another target", &EnvState{Schema: EnvSchema, Target: "other", DevHub: "devhub", ID: "00D000000000002AAA"}, "", nil, "is for target other"},
		{"recorded from another hub", &EnvState{Schema: EnvSchema, Target: "ephemeral", DevHub: "other", ID: "00D000000000002AAA"}, "", nil, "created from hub other"},
		{"unreadable env.json", nil, "{not json", nil, "parse"},
		{"wrong schema", nil, `{"schema":"glassops.env/v0"}`, nil, "schema is"},
		{"provider delete fails", &EnvState{Schema: EnvSchema, Target: "ephemeral", DevHub: "devhub", ID: "00D000000000002AAA"}, "", errors.New("salesforce: delete ActiveScratchOrg 2AS000000000001AAA: insufficient access"), "insufficient access"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h, fp := envHarness(t)
			if tc.state != nil {
				if err := writeEnvState(h.envFile(), *tc.state); err != nil {
					t.Fatal(err)
				}
			} else {
				h.write(t, ".glassops/env/ephemeral/env.json", tc.raw)
			}
			called := false
			fp.deleteEnv = func(providers.Session, providers.Env) error { called = true; return tc.fail }
			if code := h.exec("env", "delete"); code != exitcode.Error {
				t.Fatalf("exit %d, output:\n%s", code, h.out)
			}
			if !strings.Contains(h.out.String(), tc.want) || called != (tc.fail != nil) {
				t.Fatalf("called %v, output:\n%s", called, h.out)
			}
			if _, err := os.Stat(h.envFile()); err != nil {
				t.Fatalf("env.json removed after a refusal: %v", err)
			}
		})
	}
}

func TestEnvSweep(t *testing.T) {
	cases := []struct {
		name   string
		args   []string
		maxAge time.Duration
	}{
		{"default is ttl days", nil, 48 * time.Hour},
		{"flag", []string{"--max-age", "3h"}, 3 * time.Hour},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h, fp := envHarness(t)
			var gotPrefix string
			var gotAge time.Duration
			fp.sweepEnvs = func(_ providers.Session, prefix string, maxAge time.Duration) ([]providers.Env, error) {
				gotPrefix, gotAge = prefix, maxAge
				return []providers.Env{synthEnv()}, nil
			}
			if code := h.exec(append([]string{"env", "sweep"}, tc.args...)...); code != exitcode.OK {
				t.Fatalf("exit %d, output:\n%s", code, h.out)
			}
			if gotPrefix != "gx" || gotAge != tc.maxAge {
				t.Fatalf("prefix %q, max age %s", gotPrefix, gotAge)
			}
			r := h.result(t, "env-sweep")
			if r.Facts["env.deleted_count"] != float64(1) || checkStatus(r, "deleted.1") != run.StatusPass {
				t.Fatalf("result %+v", r)
			}
			list, _ := json.Marshal(r.Facts["env.deleted"])
			if !strings.Contains(string(list), `"name":"gx-repo-12-345"`) || strings.Contains(string(list), "sfdx-url") {
				t.Fatalf("deleted %s", list)
			}
		})
	}
}

func TestEnvSweepReportsPartialFailure(t *testing.T) {
	h, fp := envHarness(t)
	fp.sweepEnvs = func(providers.Session, string, time.Duration) ([]providers.Env, error) {
		return []providers.Env{synthEnv()}, errors.New("salesforce: gx-bad has a CreatedDate that cannot be read, so it was left alone")
	}
	if code := h.exec("env", "sweep"); code != exitcode.Error {
		t.Fatalf("exit %d", code)
	}
	r := h.result(t, "env-sweep")
	if checkStatus(r, "sweep") != run.StatusError || checkStatus(r, "deleted.1") != run.StatusPass {
		t.Fatalf("result %+v", r)
	}
}

func TestEnvCommandsNeedEnvironments(t *testing.T) {
	h := newHarness(t)
	fp := &fakeProvider{detect: true, auth: okAuth, describe: reportsAs(providers.KindProduction)}
	h.app.Providers = providers.Registry{fp}
	h.write(t, "sfdx-project.json", "{}")
	h.write(t, "glassops.yaml", twoTargetConfig)
	for _, sub := range []string{"create", "delete", "sweep"} {
		if code := h.exec("env", sub); code != exitcode.Error {
			t.Fatalf("env %s exit %d", sub, code)
		}
		if !strings.Contains(h.out.String(), "environments.ephemeral is not configured") || len(fp.sessions) != 0 {
			t.Fatalf("env %s output:\n%s", sub, h.out)
		}
	}
}

func TestEnvName(t *testing.T) {
	now := time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		name string
		ctx  ci.Context
		want string
	}{
		{"pull request", ci.Context{Repo: "glassops-platform/GlassOps", PRNumber: "47", Ref: "refs/pull/47/merge", RunID: "123456"}, "glassops-glassops-47-123456"},
		{"branch", ci.Context{Repo: "o/r", Ref: "refs/heads/feat/Env_Create", RunID: "9"}, "glassops-r-feat-env-create-9"},
		{"local", ci.Context{Ref: "main"}, "glassops-proj-main-20261026t120000"},
		{"long parts are capped", ci.Context{Repo: "o/" + strings.Repeat("r", 50), Ref: strings.Repeat("b", 50), RunID: "1"},
			"glassops-" + strings.Repeat("r", 30) + "-" + strings.Repeat("b", 30) + "-1"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := envName(tc.ctx, "glassops", filepath.Join(t.TempDir(), "Proj"), now); got != tc.want {
				t.Fatalf("envName = %q, want %q", got, tc.want)
			}
		})
	}
}
