package cli

import (
	"errors"
	"strings"
	"testing"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/run"
)

// pluginsWithHardis keeps the name, version and type of each entry "sf plugins --json" printed
// in glassops-hardis:local (sf 2.155.7, sfdx-hardis 8.20.1). Every other field is left out.
const pluginsWithHardis = `[
 {"name":"@salesforce/cli","version":"2.155.7","type":"core"},
 {"name":"sfdx-git-delta","version":"6.45.1","type":"user"},
 {"name":"sfdx-hardis","version":"8.20.1","type":"user"}
]`

// pluginsWithHardisVersion is pluginsWithHardis with another sfdx-hardis version. Synthetic.
func pluginsWithHardisVersion(v string) string {
	return strings.Replace(pluginsWithHardis, `"version":"8.20.1"`, `"version":"`+v+`"`, 1)
}

// errNoSf stands in for sf not being on PATH.
var errNoSf = errors.New("exec: sf: not found")

// hardisCheckOf returns the tool.sfdx-hardis check of r.
func hardisCheckOf(t *testing.T, r run.Result) run.Check {
	t.Helper()
	for _, c := range r.Checks {
		if c.Name == "tool.sfdx-hardis" {
			return c
		}
	}
	t.Fatalf("no tool.sfdx-hardis check in %+v", r.Checks)
	return run.Check{}
}

const hardisConfig = "version: 1\nsalesforce:\n  engine: hardis\n"

func TestDoctorHardis(t *testing.T) {
	cases := []struct {
		name    string
		config  string
		plugins string
		want    run.Status
		code    int
		contain []string
	}{
		{"tested version", hardisConfig, pluginsWithHardis, run.StatusPass, exitcode.OK, []string{"8.20.1"}},
		{"newer warns", hardisConfig, pluginsWithHardisVersion("8.23.0"), run.StatusPass, exitcode.OK,
			[]string{"warning:", "8.23.0", "newer than the tested version 8.20.1"}},
		{"older than the default minimum", hardisConfig, pluginsWithHardisVersion("8.19.0"), run.StatusFail, exitcode.Fail,
			[]string{"older than the minimum", "8.19.0", "8.20.1 or later", "sf plugins install sfdx-hardis@8.20.1", "never installs"}},
		{"prerelease of the minimum is older", hardisConfig, pluginsWithHardisVersion("8.20.1-beta202610271200.0"), run.StatusFail, exitcode.Fail,
			[]string{"older than the minimum"}},
		{"configured minimum", hardisConfig + "  hardis: { min_version: \"8.22\" }\n", pluginsWithHardisVersion("8.21.0"), run.StatusFail, exitcode.Fail,
			[]string{"8.21.0 is installed, 8.22 or later is needed"}},
		{"configured minimum met", hardisConfig + "  hardis: { min_version: \"8.21\" }\n", pluginsWithHardisVersion("8.21.0"), run.StatusPass, exitcode.OK,
			[]string{"warning:", "8.21.0"}},
		{"missing", hardisConfig, `[{"name":"@salesforce/cli","version":"2.155.7"},{"name":"sfdx-git-delta","version":"6.45.1"}]`, run.StatusFail, exitcode.Fail,
			[]string{"sfdx-hardis is not installed", "examples/hardis/Dockerfile", "never installs"}},
		{"unreadable version", hardisConfig, pluginsWithHardisVersion("next"), run.StatusFail, exitcode.Fail,
			[]string{"not a version"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := newHarness(t)
			h.write(t, "sfdx-project.json", "{}")
			h.write(t, "glassops.yaml", tc.config)
			h.tools["sf plugins --json"] = tc.plugins
			if code := h.exec("doctor"); code != tc.code {
				t.Fatalf("exit %d, want %d, output:\n%s", code, tc.code, h.out)
			}
			c := hardisCheckOf(t, h.result(t, "doctor"))
			if c.Status != tc.want {
				t.Fatalf("status %s, want %s: %s", c.Status, tc.want, c.Detail)
			}
			for _, s := range tc.contain {
				if !strings.Contains(c.Detail, s) {
					t.Errorf("detail %q lacks %q", c.Detail, s)
				}
			}
		})
	}
}

// TestDoctorHardisNoSf covers sf itself failing: the check fails with the install steps.
func TestDoctorHardisNoSf(t *testing.T) {
	cfg := &config.Config{Salesforce: config.Salesforce{Engine: config.EngineHardis}}
	rememberProjectDir(cfg, t.TempDir())
	app := &App{Run: func(string, string, ...string) ([]byte, error) { return nil, errNoSf }}
	checks := doctorHardis(app, cfg)
	if len(checks) != 2 || checks[0].Name != "tool.sfdx-hardis" || checks[0].Status != run.StatusFail ||
		!strings.Contains(checks[0].Detail, "sf plugins install sfdx-hardis@") {
		t.Fatalf("checks = %+v", checks)
	}
}

// TestDoctorHardisCommands: doctor fails when the sfdx-hardis config defines deploy commands,
// which engine hardis refuses, and passes without them.
func TestDoctorHardisCommands(t *testing.T) {
	h := newHarness(t)
	h.write(t, "sfdx-project.json", "{}")
	h.write(t, "glassops.yaml", hardisConfig)
	h.tools["sf plugins --json"] = pluginsWithHardis
	if code := h.exec("doctor"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if got := checkStatus(h.result(t, "doctor"), "salesforce.hardis.commands"); got != run.StatusPass {
		t.Fatalf("salesforce.hardis.commands = %s", got)
	}

	h.write(t, "config/.sfdx-hardis.yml", "commandsPostDeploy:\n  - id: x\n    command: echo hi\n")
	if code := h.exec("doctor"); code != exitcode.Fail {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	r := h.result(t, "doctor")
	if checkStatus(r, "salesforce.hardis.commands") != run.StatusFail ||
		!strings.Contains(detailOf(r, "salesforce.hardis.commands"), "config/.sfdx-hardis.yml defines commandsPostDeploy") {
		t.Fatalf("checks: %+v", r.Checks)
	}
}

// TestDoctorHardisSilent: the native engine, the default and another provider run no hardis
// check, so doctor on a native project never lists sf plugins for it.
func TestDoctorHardisSilent(t *testing.T) {
	called := false
	app := &App{Run: func(string, string, ...string) ([]byte, error) { called = true; return nil, errNoSf }}
	for _, cfg := range []*config.Config{
		{},
		{Salesforce: config.Salesforce{Engine: config.EngineNative}},
		{Provider: "exec", Salesforce: config.Salesforce{Engine: config.EngineHardis}},
	} {
		if checks := doctorHardis(app, cfg); checks != nil {
			t.Fatalf("checks = %+v for %+v", checks, cfg)
		}
	}
	if called {
		t.Fatal("doctorHardis ran a program without engine hardis")
	}

	h := newHarness(t)
	h.write(t, "sfdx-project.json", "{}")
	if code := h.exec("doctor"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if got := checkStatus(h.result(t, "doctor"), "tool.sfdx-hardis"); got != "" {
		t.Fatalf("tool.sfdx-hardis ran on the native engine: %s", got)
	}
}
