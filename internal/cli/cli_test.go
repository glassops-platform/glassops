package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/run"
	"github.com/glassops-platform/glassops/internal/providers"
	"github.com/glassops-platform/glassops/internal/providers/salesforce"
)

// fakeTools answers the commands doctor runs. A missing key means the command fails.
type fakeTools map[string]string

func (f fakeTools) run(name string, args ...string) ([]byte, error) {
	key := strings.Join(append([]string{name}, args...), " ")
	if out, ok := f[key]; ok {
		return []byte(out), nil
	}
	return nil, errors.New("exit status 1")
}

func healthyTools() fakeTools {
	return fakeTools{
		"git --version":                         "git version 2.45.0\n",
		"git rev-parse --is-shallow-repository": "false\n",
		"sf --version":                          "@salesforce/cli/2.155.7 linux-x64 node-v22.23.4\n",
		"sf plugins --json":                     `[{"name":"@salesforce/cli","version":"2.155.7"},{"name":"sfdx-git-delta","version":"6.1.0"}]`,
	}
}

type harness struct {
	app   *App
	dir   string
	out   *bytes.Buffer
	env   map[string]string
	tools fakeTools
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	h := &harness{dir: t.TempDir(), out: &bytes.Buffer{}, env: map[string]string{}, tools: healthyTools()}
	h.app = &App{
		Stdout: h.out,
		Stderr: h.out,
		Getenv: func(k string) string { return h.env[k] },
		LookupEnv: func(k string) (string, bool) {
			v, ok := h.env[k]
			return v, ok
		},
		Now:       func() time.Time { return time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC) },
		Run:       func(_ string, name string, args ...string) ([]byte, error) { return h.tools.run(name, args...) },
		Providers: providers.Registry{salesforce.NewWithRunner(func(name string, args ...string) ([]byte, error) { return h.tools.run(name, args...) })},
		Version:   "test",
	}
	return h
}

func (h *harness) write(t *testing.T, rel, content string) {
	t.Helper()
	path := filepath.Join(h.dir, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func (h *harness) exec(args ...string) int {
	return Execute(h.app, append(args, "--dir", h.dir, "--ci", "local"))
}

func (h *harness) result(t *testing.T, step string) run.Result {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(h.dir, ".glassops", "results", step+".json"))
	if err != nil {
		t.Fatal(err)
	}
	var r run.Result
	if err := json.Unmarshal(data, &r); err != nil {
		t.Fatal(err)
	}
	return r
}

func checkStatus(r run.Result, name string) run.Status {
	for _, c := range r.Checks {
		if c.Name == name {
			return c.Status
		}
	}
	return ""
}

func TestDoctorPassesOnHealthySalesforceProject(t *testing.T) {
	h := newHarness(t)
	h.write(t, "sfdx-project.json", "{}")
	if code := h.exec("doctor"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	r := h.result(t, "doctor")
	if r.Status != run.StatusPass || r.Provider != "salesforce" {
		t.Fatalf("result: %+v", r)
	}
	if r.Facts["tools.sf"] != "2.155.7" || r.Facts["tools.sfdx-git-delta"] != "6.1.0" {
		t.Fatalf("facts: %v", r.Facts)
	}
}

func TestDoctorFailures(t *testing.T) {
	cases := []struct {
		name  string
		setup func(t *testing.T, h *harness)
		check string
	}{
		{"shallow clone", func(t *testing.T, h *harness) {
			h.tools["git rev-parse --is-shallow-repository"] = "true\n"
		}, "git.checkout"},
		{"not a repo", func(t *testing.T, h *harness) {
			delete(h.tools, "git rev-parse --is-shallow-repository")
		}, "git.checkout"},
		{"sgd missing", func(t *testing.T, h *harness) {
			h.tools["sf plugins --json"] = `[{"name":"@salesforce/cli","version":"2.155.7"}]`
		}, "tool.sfdx-git-delta"},
		{"sf missing", func(t *testing.T, h *harness) {
			delete(h.tools, "sf --version")
		}, "tool.sf"},
		{"invalid config", func(t *testing.T, h *harness) {
			h.write(t, "glassops.yaml", "version: 1\nfoo: bar\n")
		}, "config"},
		{"env var missing", func(t *testing.T, h *harness) {
			h.write(t, "glassops.yaml", "version: 1\ntargets:\n  qa:\n    auth: { method: sfdx-url, url: $SF_QA_AUTH_URL }\n")
		}, "env"},
		{"config provider without project", func(t *testing.T, h *harness) {
			os.Remove(filepath.Join(h.dir, "sfdx-project.json"))
			h.write(t, "glassops.yaml", "version: 1\nprovider: salesforce\n")
		}, "provider"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := newHarness(t)
			h.write(t, "sfdx-project.json", "{}")
			tc.setup(t, h)
			if code := h.exec("doctor"); code != exitcode.Fail {
				t.Fatalf("exit %d, want %d. output:\n%s", code, exitcode.Fail, h.out)
			}
			r := h.result(t, "doctor")
			if got := checkStatus(r, tc.check); got != run.StatusFail {
				t.Fatalf("check %s = %q, want fail. checks: %+v", tc.check, got, r.Checks)
			}
		})
	}
}

func TestDoctorNeverPrintsEnvValues(t *testing.T) {
	h := newHarness(t)
	h.write(t, "sfdx-project.json", "{}")
	h.write(t, "glassops.yaml", "version: 1\ntargets:\n  qa:\n    auth: { method: sfdx-url, url: $SF_QA_AUTH_URL }\n")
	h.env["SF_QA_AUTH_URL"] = "force://super-secret-value"
	if code := h.exec("doctor"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	data, _ := os.ReadFile(filepath.Join(h.dir, ".glassops", "results", "doctor.json"))
	if strings.Contains(h.out.String(), "super-secret") || strings.Contains(string(data), "super-secret") {
		t.Fatal("a secret value leaked into output or the result")
	}
}

func TestInitWritesValidConfigAndWorkflow(t *testing.T) {
	h := newHarness(t)
	h.write(t, "sfdx-project.json", "{}")
	h.write(t, ".github/CODEOWNERS", "")
	if code := h.exec("init"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	data, err := os.ReadFile(filepath.Join(h.dir, "glassops.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Parse(data)
	if err != nil {
		t.Fatalf("generated config does not parse: %v", err)
	}
	if cfg.Provider != "salesforce" || len(cfg.Pipelines["pr"]) != 8 || cfg.Pipelines["pr"][7].When != "always" {
		t.Fatalf("generated config: %+v", cfg)
	}
	wf, err := os.ReadFile(filepath.Join(h.dir, ".github", "workflows", "glassops.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(wf), "fetch-depth: 0") || !strings.Contains(string(wf), "glassops doctor") {
		t.Fatalf("workflow:\n%s", wf)
	}
	if r := h.result(t, "init"); len(r.Artifacts) != 2 {
		t.Fatalf("artifacts: %v", r.Artifacts)
	}
}

func TestInitRefusesToOverwrite(t *testing.T) {
	h := newHarness(t)
	h.write(t, "glassops.yaml", "version: 1\n# mine\n")
	if code := h.exec("init"); code != exitcode.Fail {
		t.Fatalf("exit %d, want %d", code, exitcode.Fail)
	}
	data, _ := os.ReadFile(filepath.Join(h.dir, "glassops.yaml"))
	if !strings.Contains(string(data), "# mine") {
		t.Fatal("init overwrote an existing file without --force")
	}
	if code := h.exec("init", "--force"); code != exitcode.OK {
		t.Fatalf("--force exit %d, output:\n%s", code, h.out)
	}
	data, _ = os.ReadFile(filepath.Join(h.dir, "glassops.yaml"))
	if strings.Contains(string(data), "# mine") {
		t.Fatal("--force did not overwrite")
	}
}

func TestInitWithoutProviderOrCI(t *testing.T) {
	h := newHarness(t)
	if code := h.exec("init"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	data, _ := os.ReadFile(filepath.Join(h.dir, "glassops.yaml"))
	if strings.Contains(string(data), "provider:") {
		t.Fatalf("provider written without detection:\n%s", data)
	}
	if _, err := os.Stat(filepath.Join(h.dir, ".github")); !os.IsNotExist(err) {
		t.Fatal("a CI file was written with no CI detected")
	}
}

func TestUsageErrorsExitTwo(t *testing.T) {
	h := newHarness(t)
	if code := h.exec("doctor", "--no-such-flag"); code != exitcode.Error {
		t.Fatalf("exit %d, want %d", code, exitcode.Error)
	}
	if code := Execute(h.app, []string{"doctor", "--ci", "jenkins", "--dir", h.dir}); code != exitcode.Error {
		t.Fatalf("unknown CI exit %d, want %d", code, exitcode.Error)
	}
}

func TestGitHubOutputs(t *testing.T) {
	h := newHarness(t)
	h.write(t, "sfdx-project.json", "{}")
	outFile := filepath.Join(h.dir, "gh-output")
	sumFile := filepath.Join(h.dir, "gh-summary")
	h.env["GITHUB_ACTIONS"] = "true"
	h.env["GITHUB_OUTPUT"] = outFile
	h.env["GITHUB_STEP_SUMMARY"] = sumFile
	if code := Execute(h.app, []string{"doctor", "--dir", h.dir}); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	out, _ := os.ReadFile(outFile)
	if !strings.Contains(string(out), "status=pass") {
		t.Fatalf("GITHUB_OUTPUT = %q", out)
	}
	sum, _ := os.ReadFile(sumFile)
	if !strings.Contains(string(sum), "| tool.sf | pass |") {
		t.Fatalf("summary = %q", sum)
	}
}
