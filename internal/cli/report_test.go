package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/run"
	"github.com/glassops-platform/glassops/internal/core/sarif"
)

// seedResult writes a synthetic result for a step into the harness's results dir.
func seedResult(t *testing.T, h *harness, step string, st run.Status, checks ...run.Check) {
	t.Helper()
	r := run.New(step, h.app.Now)
	r.Checks = checks
	r.Status = st
	if _, err := r.Write(filepath.Join(h.dir, ".glassops", "results"), h.app.Now); err != nil {
		t.Fatal(err)
	}
}

func TestReportExitsWithWorstStatus(t *testing.T) {
	cases := []struct {
		name string
		st   run.Status
		want int
	}{
		{"pass", run.StatusPass, exitcode.OK},
		{"fail", run.StatusFail, exitcode.Fail},
		{"error", run.StatusError, exitcode.Error},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := newHarness(t)
			seedResult(t, h, "doctor", run.StatusPass)
			seedResult(t, h, "validate", tc.st, run.Check{Name: "problem.1", Status: tc.st, Detail: "force-app/A.cls:4 bad"})
			if code := h.exec("report"); code != tc.want {
				t.Fatalf("exit %d, want %d\n%s", code, tc.want, h.out)
			}
			data, err := os.ReadFile(filepath.Join(h.dir, ".glassops", "report.sarif"))
			if err != nil {
				t.Fatalf("default SARIF path not written: %v", err)
			}
			if err := sarif.Validate(data); err != nil {
				t.Fatalf("SARIF invalid: %v", err)
			}
			if !strings.Contains(h.out.String(), "## glassops report: "+string(tc.st)) {
				t.Fatalf("the summary did not reach the CI layer:\n%s", h.out)
			}
		})
	}
}

func TestReportWritesNamedFiles(t *testing.T) {
	h := newHarness(t)
	seedResult(t, h, "doctor", run.StatusPass)
	sum := filepath.Join(h.dir, "out", "summary.md")
	sar := filepath.Join(h.dir, "out", "r.sarif")
	if code := h.exec("report", "--summary", sum, "--sarif", sar); code != exitcode.OK {
		t.Fatalf("exit %d\n%s", code, h.out)
	}
	for _, p := range []string{sum, sar} {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("not written: %v", err)
		}
	}
}

func TestReportWithNoResultsIsAnError(t *testing.T) {
	h := newHarness(t)
	if code := h.exec("report"); code != exitcode.Error {
		t.Fatalf("exit %d, want %d: no evidence must not pass", code, exitcode.Error)
	}
}

func TestReportDoesNotMergeItsOwnSARIFBack(t *testing.T) {
	h := newHarness(t)
	seedResult(t, h, "doctor", run.StatusPass)
	sar := filepath.Join(h.dir, ".glassops", "results", "report.sarif")
	for i := 0; i < 2; i++ {
		if code := h.exec("report", "--sarif", sar); code != exitcode.OK {
			t.Fatalf("run %d exit %d", i, code)
		}
	}
	l, err := sarif.Read(sar)
	if err != nil {
		t.Fatal(err)
	}
	if len(l.Runs) != 1 {
		t.Fatalf("runs = %d after two reports, want 1", len(l.Runs))
	}
}
