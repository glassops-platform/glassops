package cli

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/junit"
	"github.com/glassops-platform/glassops/internal/core/run"
	"github.com/glassops-platform/glassops/internal/providers"
)

// testConfig is a synthetic config. The coverage minimums are added per test.
const testConfigHead = `version: 1
targets:
  qa:
    id: 00D000000000002AAA
    tier: qa
    auth: { method: sfdx-url, url: $SF_QA_AUTH_URL }
test:
  level: { default: RunLocalTests, qa: RunAllTestsInOrg }
`

func testHarness(t *testing.T, coverage string) (*harness, *fakeProvider) {
	t.Helper()
	h := newHarness(t)
	h.write(t, "glassops.yaml", testConfigHead+coverage)
	fp := &fakeProvider{
		detect:   true,
		auth:     func(providers.Target) error { return nil },
		describe: reportsAs(providers.KindSandbox),
		runTests: func(providers.Session, providers.TestRequest) (providers.TestResult, error) {
			return goodTestResult(), nil
		},
	}
	h.app.Providers = providers.Registry{fp}
	return h, fp
}

func goodTestResult() providers.TestResult {
	return providers.TestResult{
		Passed: true, Finished: true, RunID: "707000000000001", Ran: 4, OrgCoverage: 90,
		Classes: []providers.ClassCoverage{
			{Name: "Greeter", Lines: 10, Covered: 9, Percent: 90},
			{Name: "Mailer", Lines: 10, Covered: 8, Percent: 80},
		},
		Facts: map[string]any{},
	}
}

func TestTestRequiresTarget(t *testing.T) {
	h, fp := testHarness(t, "")
	if code := h.exec("test"); code != exitcode.Error {
		t.Fatalf("exit %d", code)
	}
	if len(fp.sessions) != 0 {
		t.Fatal("logged in without a target")
	}
}

func TestTestPassWritesFactsAndJUnit(t *testing.T) {
	h, fp := testHarness(t, "  coverage: { org_min: 75, class_min: 70 }\n")
	var got providers.TestRequest
	fp.runTests = func(_ providers.Session, r providers.TestRequest) (providers.TestResult, error) {
		got = r
		return goodTestResult(), nil
	}
	if code := h.exec("test", "--target", "qa"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	r := h.result(t, "test")
	if r.Status != run.StatusPass || r.Target != "qa" {
		t.Fatalf("result: %+v", r)
	}
	want := map[string]any{"test.ran": float64(4), "test.failed": float64(0), "test.org_coverage": float64(90),
		"test.coverage.Greeter.percent": float64(90), "test.coverage.Mailer.percent": float64(80)}
	for k, v := range want {
		if r.Facts[k] != v {
			t.Errorf("fact %s = %v, want %v", k, r.Facts[k], v)
		}
	}
	if got.TestLevel != "RunAllTestsInOrg" || got.Dir != h.dir {
		t.Fatalf("request: %+v", got)
	}
	if !fp.allClosed() || len(fp.sessions) != 1 {
		t.Fatalf("sessions: %d", len(fp.sessions))
	}
	data, err := os.ReadFile(filepath.Join(h.dir, ".glassops", "results", "test.junit.xml"))
	if err != nil {
		t.Fatal(err)
	}
	var s junit.Suites
	if err := xml.Unmarshal(data, &s); err != nil || s.Tests != 4 {
		t.Fatalf("junit: %v %+v", err, s)
	}
	if len(r.Artifacts) != 1 || !strings.HasSuffix(filepath.ToSlash(r.Artifacts[0]), "results/test.junit.xml") {
		t.Fatalf("artifacts: %v", r.Artifacts)
	}
}

func TestTestLevelFlagAndTierDefault(t *testing.T) {
	h, fp := testHarness(t, "")
	var levels []string
	fp.runTests = func(_ providers.Session, r providers.TestRequest) (providers.TestResult, error) {
		levels = append(levels, r.TestLevel)
		return goodTestResult(), nil
	}
	h.exec("test", "--target", "qa", "--level", "RunSpecifiedTests", "--tests", "A,B")
	if len(levels) != 1 || levels[0] != "RunSpecifiedTests" {
		t.Fatalf("levels: %v", levels)
	}
}

func TestTestCustomJUnitPath(t *testing.T) {
	h, _ := testHarness(t, "")
	path := filepath.Join(h.dir, "out", "x.xml")
	if code := h.exec("test", "--target", "qa", "--junit", path); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
}

func TestTestGateFailures(t *testing.T) {
	cases := []struct {
		name     string
		coverage string
		mutate   func(*providers.TestResult)
		check    string
	}{
		{"failing test", "", func(r *providers.TestResult) {
			r.Passed, r.Failed, r.Failures = false, 1, []string{"FailingTest.fails: boom"}
		}, "tests"},
		{"unfinished run", "", func(r *providers.TestResult) { r.Passed, r.Finished = false, false }, "run"},
		{"org below minimum", "  coverage: { org_min: 95 }\n", func(r *providers.TestResult) {}, "coverage.org"},
		{"org coverage unknown", "  coverage: { org_min: 50 }\n", func(r *providers.TestResult) { r.OrgCoverage = -1 }, "coverage.org"},
		{"class below minimum", "  coverage: { class_min: 85 }\n", func(r *providers.TestResult) {}, "coverage.class.Mailer"},
		{"no class coverage", "  coverage: { class_min: 50 }\n", func(r *providers.TestResult) { r.Classes = nil }, "coverage.class"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h, fp := testHarness(t, tc.coverage)
			fp.runTests = func(providers.Session, providers.TestRequest) (providers.TestResult, error) {
				r := goodTestResult()
				tc.mutate(&r)
				return r, nil
			}
			if code := h.exec("test", "--target", "qa"); code != exitcode.Fail {
				t.Fatalf("exit %d, want 1. output:\n%s", code, h.out)
			}
			r := h.result(t, "test")
			if r.Status != run.StatusFail || checkStatus(r, tc.check) != run.StatusFail {
				t.Fatalf("status %s, check %s = %s", r.Status, tc.check, checkStatus(r, tc.check))
			}
			if _, err := os.Stat(filepath.Join(h.dir, ".glassops", "results", "test.junit.xml")); err != nil {
				t.Fatalf("junit is still written on a failure: %v", err)
			}
		})
	}
}

func TestTestMinimumsNotSetMeansNoCoverageGate(t *testing.T) {
	h, fp := testHarness(t, "")
	fp.runTests = func(providers.Session, providers.TestRequest) (providers.TestResult, error) {
		r := goodTestResult()
		r.OrgCoverage, r.Classes = 1, []providers.ClassCoverage{{Name: "Low", Percent: 1}}
		return r, nil
	}
	if code := h.exec("test", "--target", "qa"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
}

func TestTestProviderErrorIsExit2(t *testing.T) {
	h, fp := testHarness(t, "")
	fp.runTests = func(providers.Session, providers.TestRequest) (providers.TestResult, error) {
		return providers.TestResult{}, os.ErrDeadlineExceeded
	}
	if code := h.exec("test", "--target", "qa"); code != exitcode.Error {
		t.Fatalf("exit %d", code)
	}
	if !fp.allClosed() {
		t.Fatal("session left open")
	}
}

func TestTestGuardMismatchStopsBeforeTests(t *testing.T) {
	h, fp := testHarness(t, "")
	fp.describe = func(tg providers.Target) (providers.TargetInfo, error) {
		return providers.TargetInfo{ID: "00D000000000009AAA", Kind: providers.KindSandbox, Facts: map[string]any{}}, nil
	}
	called := false
	fp.runTests = func(providers.Session, providers.TestRequest) (providers.TestResult, error) {
		called = true
		return goodTestResult(), nil
	}
	if code := h.exec("test", "--target", "qa"); code != exitcode.Error || called {
		t.Fatalf("exit %d, tests ran: %v", code, called)
	}
}

func TestTestUnconfiguredTargetIsExit2(t *testing.T) {
	h, _ := testHarness(t, "")
	if code := h.exec("test", "--target", "nope"); code != exitcode.Error {
		t.Fatalf("exit %d", code)
	}
}

func noneFoundResult() providers.TestResult {
	return providers.TestResult{Finished: true, OrgCoverage: -1, Facts: map[string]any{"test.none_found": true}}
}

func TestTestNoneFoundWithMinimumFails(t *testing.T) {
	h, fp := testHarness(t, "  coverage: { org_min: 75 }\n")
	fp.runTests = func(providers.Session, providers.TestRequest) (providers.TestResult, error) {
		return noneFoundResult(), nil
	}
	if code := h.exec("test", "--target", "qa"); code != exitcode.Fail {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	r := h.result(t, "test")
	if r.Status != run.StatusFail || r.Facts["test.none_found"] != true || !strings.Contains(h.out.String(), "no Apex tests ran in qa, so coverage cannot be checked") {
		t.Fatalf("result %+v\n%s", r, h.out)
	}
}

func TestTestNoneFoundWithoutMinimumIsEmpty(t *testing.T) {
	h, fp := testHarness(t, "")
	fp.runTests = func(providers.Session, providers.TestRequest) (providers.TestResult, error) {
		return noneFoundResult(), nil
	}
	if code := h.exec("test", "--target", "qa"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	r := h.result(t, "test")
	if r.Status != run.StatusEmpty || checkStatus(r, "tests") != run.StatusSkipped || r.Facts["test.none_found"] != true {
		t.Fatalf("result %+v", r)
	}
}
