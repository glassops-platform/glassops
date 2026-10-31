package cli

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/run"
)

const analyzerFixtures = "../providers/salesforce/testdata/codeanalyzer"

// analyzerRunner stands in for "sf code-analyzer run". It copies the captured JSON and SARIF
// to the --output-file paths it is given and records the arguments.
func analyzerRunner(t *testing.T, h *harness, args *[]string, fail error, write bool) {
	t.Helper()
	jsonData, err := os.ReadFile(filepath.Join(analyzerFixtures, "code-analyzer.json"))
	if err != nil {
		t.Fatal(err)
	}
	sarifData, err := os.ReadFile(filepath.Join(analyzerFixtures, "code-analyzer.sarif"))
	if err != nil {
		t.Fatal(err)
	}
	h.app.Run = func(_ string, name string, a ...string) ([]byte, error) {
		if name != "sf" || len(a) < 2 || a[0] != "code-analyzer" {
			return h.tools.run(name, a...)
		}
		*args = a
		if fail != nil {
			return nil, fail
		}
		if write {
			for i, v := range a {
				if v != "--output-file" {
					continue
				}
				p := a[i+1]
				data := jsonData
				if strings.HasSuffix(p, ".sarif") {
					data = sarifData
				}
				if err := os.WriteFile(p, data, 0o644); err != nil {
					t.Fatal(err)
				}
			}
		}
		return nil, nil
	}
}

func flagValues(args []string, flag string) []string {
	var out []string
	for i, a := range args {
		if a == flag && i+1 < len(args) {
			out = append(out, args[i+1])
		}
	}
	return out
}

func TestAnalyzeRecordsCountsAndSarif(t *testing.T) {
	h := newHarness(t)
	h.write(t, "sfdx-project.json", "{}")
	h.write(t, "glassops.yaml", analyzeOn)
	var args []string
	analyzerRunner(t, h, &args, nil, true)
	if code := h.exec("analyze"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if got := flagValues(args, "--rule-selector"); len(got) != 1 || got[0] != "Recommended" {
		t.Fatalf("rule selector %v", got)
	}
	if strings.Contains(strings.Join(args, " "), "severity-threshold") {
		t.Fatalf("passed a severity threshold: %v", args)
	}
	if got := flagValues(args, "--workspace"); len(got) != 1 || got[0] != "." {
		t.Fatalf("workspace %v", got)
	}
	if outs := flagValues(args, "--output-file"); len(outs) != 2 {
		t.Fatalf("output files %v", outs)
	}
	r := h.result(t, "analyze")
	// The fixture is a real run: 3 violations, all severity 4.
	want := map[string]float64{"analyze.total": 3, "analyze.sev1": 0, "analyze.sev2": 0, "analyze.sev3": 0, "analyze.sev4": 3, "analyze.sev5": 0}
	for k, v := range want {
		if r.Facts[k] != v {
			t.Fatalf("fact %s = %v, want %v (facts %v)", k, r.Facts[k], v, r.Facts)
		}
	}
	sarif := filepath.Join(h.dir, ".glassops", "results", "analyze.sarif")
	relSarif := ".glassops/results/analyze.sarif"
	if !contains(r.Artifacts, relSarif) {
		t.Fatalf("artifacts %v, want %s", r.Artifacts, relSarif)
	}
	if _, err := os.Stat(sarif); err != nil {
		t.Fatal(err)
	}
	if r.Status != run.StatusPass {
		t.Fatalf("status %s", r.Status)
	}
}

func TestAnalyzeUsesConfigSelectorAndFlagOverride(t *testing.T) {
	h := newHarness(t)
	h.write(t, "sfdx-project.json", "{}")
	h.write(t, "glassops.yaml", "version: 1\nanalyze:\n  enabled: true\n  rule_selector: Security\n")
	var args []string
	analyzerRunner(t, h, &args, nil, true)
	if code := h.exec("analyze"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if got := flagValues(args, "--rule-selector"); got[0] != "Security" {
		t.Fatalf("selector %v", got)
	}
	custom := filepath.Join(h.dir, "out", "my.sarif")
	if code := h.exec("analyze", "--rule-selector", "pmd:Performance", "--sarif", custom); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if got := flagValues(args, "--rule-selector"); got[0] != "pmd:Performance" {
		t.Fatalf("selector %v", got)
	}
	if _, err := os.Stat(custom); err != nil {
		t.Fatal(err)
	}
}

func TestAnalyzeErrorsExit2(t *testing.T) {
	cases := []struct {
		name  string
		fail  error
		write bool
	}{
		{"sf fails", errors.New("exit status 1"), false},
		{"no output written", nil, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := newHarness(t)
			h.write(t, "sfdx-project.json", "{}")
			h.write(t, "glassops.yaml", analyzeOn)
			var args []string
			analyzerRunner(t, h, &args, tc.fail, tc.write)
			if code := h.exec("analyze"); code != exitcode.Error {
				t.Fatalf("exit %d, output:\n%s", code, h.out)
			}
			if r := h.result(t, "analyze"); r.Status != run.StatusError {
				t.Fatalf("status %s", r.Status)
			}
		})
	}
}

func TestAnalyzeRejectsOutputWithoutCounts(t *testing.T) {
	h := newHarness(t)
	h.write(t, "sfdx-project.json", "{}")
	h.write(t, "glassops.yaml", analyzeOn)
	h.app.Run = func(_ string, name string, a ...string) ([]byte, error) {
		if name == "sf" && a[0] == "code-analyzer" {
			for i, v := range a {
				if v == "--output-file" {
					_ = os.WriteFile(a[i+1], []byte(`{"violations":[]}`), 0o644)
				}
			}
			return nil, nil
		}
		return h.tools.run(name, a...)
	}
	if code := h.exec("analyze"); code != exitcode.Error {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
}

func TestAnalyzeDisabledIsSkipped(t *testing.T) {
	h := newHarness(t)
	h.write(t, "sfdx-project.json", "{}")
	h.write(t, "glassops.yaml", "version: 1\nanalyze:\n  enabled: false\n")
	var args []string
	analyzerRunner(t, h, &args, nil, true)
	if code := h.exec("analyze"); code != exitcode.OK || args != nil {
		t.Fatalf("exit %d, args %v", code, args)
	}
}

func TestParseViolationCounts(t *testing.T) {
	if _, err := parseViolationCounts([]byte(`not json`)); err == nil {
		t.Fatal("bad json accepted")
	}
	if _, err := parseViolationCounts([]byte(`{}`)); err == nil {
		t.Fatal("missing counts accepted")
	}
	c, err := parseViolationCounts([]byte(`{"violationCounts":{"total":7,"sev1":1,"sev2":2,"sev3":3,"sev4":0,"sev5":1}}`))
	if err != nil || c.Total != 7 || c.Sev1 != 1 || c.Sev2 != 2 || c.Sev3 != 3 || c.Sev5 != 1 {
		t.Fatalf("%+v %v", c, err)
	}
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func TestAnalyzeUnsetIsSkippedAndSaysHow(t *testing.T) {
	h := newHarness(t)
	h.write(t, "sfdx-project.json", "{}")
	var args []string
	analyzerRunner(t, h, &args, nil, true)
	if code := h.exec("analyze"); code != exitcode.OK || args != nil {
		t.Fatalf("exit %d, args %v", code, args)
	}
	r := h.result(t, "analyze")
	if r.Status != run.StatusEmpty || !strings.Contains(r.Checks[0].Detail, "analyze.enabled: true") || !strings.Contains(r.Checks[0].Detail, "glassops-analyzer") {
		t.Fatalf("result %+v", r)
	}
}
