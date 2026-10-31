package cli

import (
	"strings"
	"testing"

	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/run"
)

const analyzeOn = "version: 1\nanalyze:\n  enabled: true\n"

const pluginsWithAnalyzer = `[{"name":"@salesforce/cli","version":"2.155.7"},{"name":"sfdx-git-delta","version":"6.45.1"},{"name":"@salesforce/plugin-code-analyzer","version":"5.16.0"}]`

func TestDoctorAnalyzeJava(t *testing.T) {
	cases := []struct {
		name    string
		tools   map[string]string
		want    run.Status
		contain string
	}{
		{"java 17", map[string]string{"java --version": "openjdk 17.0.9 2023-10-17\nOpenJDK Runtime Environment\n"}, run.StatusPass, "openjdk 17.0.9"},
		{"java 11", map[string]string{"java --version": "openjdk 11.0.2 2019-01-15\n"}, run.StatusPass, ""},
		{"oracle java 21", map[string]string{"java --version": "java 21.0.1 2023-10-17 LTS\n"}, run.StatusPass, ""},
		{"java 10", map[string]string{"java --version": "openjdk 10.0.2 2018-07-17\n"}, run.StatusFail, "Java 10.0.2 found"},
		{"java 8", map[string]string{"java -version": ""}, run.StatusFail, "Java 8 or older"},
		{"no java", nil, run.StatusFail, "java not found"},
		{"unreadable", map[string]string{"java --version": "garbage\n"}, run.StatusFail, "could not read"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := newHarness(t)
			h.write(t, "sfdx-project.json", "{}")
			h.write(t, "glassops.yaml", analyzeOn)
			h.tools["sf plugins --json"] = pluginsWithAnalyzer
			for k, v := range tc.tools {
				h.tools[k] = v
			}
			code := h.exec("doctor")
			r := h.result(t, "doctor")
			var got run.Check
			for _, c := range r.Checks {
				if c.Name == "tool.java" {
					got = c
				}
			}
			if got.Status != tc.want || !strings.Contains(got.Detail, tc.contain) {
				t.Fatalf("java check %+v, want %s containing %q", got, tc.want, tc.contain)
			}
			if (tc.want == run.StatusFail) != (code == exitcode.Fail) {
				t.Fatalf("exit %d for %s", code, tc.want)
			}
		})
	}
}

func TestDoctorAnalyzePlugin(t *testing.T) {
	h := newHarness(t)
	h.write(t, "sfdx-project.json", "{}")
	h.write(t, "glassops.yaml", analyzeOn)
	h.tools["java --version"] = "openjdk 17.0.9 2023-10-17\n"

	// Not installed: healthyTools lists only the CLI and sgd.
	if code := h.exec("doctor"); code != exitcode.Fail {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if got := checkStatus(h.result(t, "doctor"), "tool.code-analyzer"); got != run.StatusFail {
		t.Fatalf("plugin check %s", got)
	}

	h.tools["sf plugins --json"] = pluginsWithAnalyzer
	if code := h.exec("doctor"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	r := h.result(t, "doctor")
	for _, c := range r.Checks {
		if c.Name == "tool.code-analyzer" && c.Detail != "5.16.0" {
			t.Fatalf("detail %q", c.Detail)
		}
	}
}

func TestDoctorAnalyzeSilentUnlessEnabled(t *testing.T) {
	for _, cfg := range []string{"", "version: 1\nanalyze:\n  enabled: false\n"} {
		h := newHarness(t)
		h.write(t, "sfdx-project.json", "{}")
		if cfg != "" {
			h.write(t, "glassops.yaml", cfg)
		}
		if code := h.exec("doctor"); code != exitcode.OK {
			t.Fatalf("exit %d, output:\n%s", code, h.out)
		}
		r := h.result(t, "doctor")
		for _, c := range r.Checks {
			if c.Name == "tool.java" || c.Name == "tool.code-analyzer" {
				t.Fatalf("unexpected check %+v", c)
			}
		}
	}
}
