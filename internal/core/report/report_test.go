package report

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/permit"
	"github.com/glassops-platform/glassops/internal/core/policy"
	"github.com/glassops-platform/glassops/internal/core/run"
	"github.com/glassops-platform/glassops/internal/core/sarif"
)

var update = flag.Bool("update", false, "rewrite the golden files")

// All inputs below are synthetic Go literals made for this test.

var t0 = time.Date(2026, 10, 30, 9, 0, 0, 0, time.UTC)

func res(step string, offset int, st run.Status, target string, checks ...run.Check) run.Result {
	r := run.Result{
		Schema: run.Schema, Step: step, Status: st, Target: target,
		Started: t0.Add(time.Duration(offset) * time.Minute), Ended: t0.Add(time.Duration(offset)*time.Minute + time.Second),
		Checks: checks, Facts: map[string]any{}, Artifacts: []string{}, Errors: []string{},
	}
	return r
}

func chk(name string, st run.Status, detail string) run.Check {
	return run.Check{Name: name, Status: st, Detail: detail}
}

// writeAll writes every step type into dir, plus a PERMIT and a foreign SARIF file.
func writeAll(t *testing.T, dir string, st map[string]run.Status) {
	t.Helper()
	status := func(step string, def run.Status) run.Status {
		if s, ok := st[step]; ok {
			return s
		}
		return def
	}
	results := []run.Result{
		res("doctor", 0, status("doctor", run.StatusPass), "", chk("tool.git", run.StatusPass, "2.45.0"), chk("tool.sf", run.StatusPass, "2.155.7")),
		res("plan", 1, status("plan", run.StatusPass), "", chk("plan", run.StatusPass, "3 components")),
		res("validate", 2, status("validate", run.StatusFail), "qa",
			chk("problem.1", run.StatusFail, "ApexClass AccountService: Variable does not exist: acct (line 14)"),
			chk("problem.2", run.StatusFail, "force-app/main/default/classes/Util.cls:27 Unexpected token | here"),
			chk("problem.3", run.StatusFail, `force-app\main\default\triggers\T.trigger: missing semicolon`),
		),
		res("check", 3, status("check", run.StatusPass), "", chk("analyze", run.StatusPass, "no violations")),
		res("test", 4, status("test", run.StatusSkipped), "qa"),
		res("gate", 5, status("gate", run.StatusFail), "prod",
			chk("window.prod", run.StatusFail, "outside the deploy window"),
			chk("plan_not_empty", run.StatusPass, "3 components")),
		res("deploy", 6, status("deploy", run.StatusError), "prod",
			chk("deploy", run.StatusError, "the target stopped answering for u.ser@example.com")),
	}
	results[6].Facts["salesforce.deploy_id"] = "0AfXX0000000001AAA"
	results[6].Errors = []string{"poll timed out"}
	for _, r := range results {
		if _, err := r.Write(dir, func() time.Time { return r.Ended }); err != nil {
			t.Fatal(err)
		}
	}
	p := permit.Permit{
		Schema: permit.Schema, ID: "abc", Allow: false,
		IssuedAt: t0.Add(5 * time.Minute), ExpiresAt: t0.Add(2*time.Hour + 5*time.Minute),
		Target: permit.Target{Name: "prod", Tier: "production", ID: "00D000000000001AAA", Kind: "salesforce"},
		Checks: []policy.Result{
			{Rule: "plan_not_empty", Outcome: policy.Pass, Detail: "3 components"},
			{Rule: "window.prod", Outcome: policy.Fail, Detail: "outside the deploy window"},
			{Rule: "separation_of_duties", Outcome: policy.SatisfiedByApproval, Detail: "approved", Mechanism: "environment", Approver: "reviewer1"},
		},
		SeparationOfDuties: &permit.SoD{Mode: "satisfied_by_approval", Mechanism: "environment", Approver: "reviewer1", Author: "dev1", Deployer: "dev1"},
	}
	writeJSON(t, filepath.Join(dir, "permit.json"), p)
	other := `{"$schema":"https://json.schemastore.org/sarif-2.1.0.json","version":"2.1.0","runs":[{"tool":{"driver":{"name":"Code Analyzer"}},"results":[{"ruleId":"pmd:X","level":"warning","message":{"text":"m"}}]}]}`
	if err := os.WriteFile(filepath.Join(dir, "analyzer.sarif"), []byte(other), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeJSON(t *testing.T, path string, v any) {
	t.Helper()
	data, err := jsonIndent(v)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func golden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name)
	if *update {
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != string(want) {
		t.Fatalf("%s differs from the golden file (run with -update to refresh)\n--- got ---\n%s\n--- want ---\n%s", name, got, want)
	}
}

func TestSummaryGolden(t *testing.T) {
	dir := t.TempDir()
	writeAll(t, dir, nil)
	b, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	golden(t, "summary-all-steps.md", Summary(b))
}

func TestSummaryGoldenAllPass(t *testing.T) {
	dir := t.TempDir()
	r := res("doctor", 0, run.StatusPass, "", chk("tool.git", run.StatusPass, "2.45.0"))
	if _, err := r.Write(dir, func() time.Time { return t0 }); err != nil {
		t.Fatal(err)
	}
	b, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	golden(t, "summary-pass.md", Summary(b))
	if b.ExitCode() != exitcode.OK {
		t.Fatalf("exit %d, want 0", b.ExitCode())
	}
}

func TestSummaryHidesIdentity(t *testing.T) {
	dir := t.TempDir()
	writeAll(t, dir, nil)
	b, _ := Load(dir)
	out := Summary(b)
	if strings.Contains(out, "u.ser@example.com") || !strings.Contains(out, "u***@example.com") {
		t.Fatalf("an email reached the summary unmasked:\n%s", out)
	}
	if strings.Contains(out, "00D000000000001AAA") {
		t.Fatalf("the org Id from the PERMIT reached the summary:\n%s", out)
	}
}

func TestSARIFGolden(t *testing.T) {
	dir := t.TempDir()
	writeAll(t, dir, nil)
	b, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	data, err := sarif.Marshal(SARIF(b, "test"))
	if err != nil {
		t.Fatal(err)
	}
	if err := sarif.Validate(data); err != nil {
		t.Fatalf("report SARIF is invalid: %v", err)
	}
	golden(t, "report.sarif", string(data))
}

func TestSARIFFindings(t *testing.T) {
	dir := t.TempDir()
	writeAll(t, dir, nil)
	b, _ := Load(dir)
	log := SARIF(b, "test")
	if len(log.Runs) != 2 || log.Runs[1].Tool.Driver.Name != "Code Analyzer" {
		t.Fatalf("the analyzer run should be merged after glassops's: %+v", log.Runs)
	}
	got := map[string]sarif.Result{}
	for _, r := range log.Runs[0].Results {
		got[r.RuleID+"|"+r.Message.Text] = r
	}
	var withLine, withFile, noLoc, gateDenials int
	for _, r := range log.Runs[0].Results {
		if r.Level != sarif.LevelError {
			t.Errorf("%s has level %s, want error", r.RuleID, r.Level)
		}
		if r.RuleID == "glassops.gate.window.prod" {
			gateDenials++
		}
		switch {
		case len(r.Locations) == 0:
			noLoc++
		case r.Locations[0].PhysicalLocation.Region != nil:
			withLine++
		default:
			withFile++
		}
	}
	if gateDenials != 1 {
		t.Errorf("gate denials = %d, want 1 (only the failed rule)", gateDenials)
	}
	if withLine != 1 || withFile != 1 {
		t.Errorf("located results: line %d, file only %d, want 1 and 1", withLine, withFile)
	}
	if noLoc < 2 {
		t.Errorf("results without a location = %d, want at least 2", noLoc)
	}
}

func TestLocate(t *testing.T) {
	cases := []struct {
		in   string
		file string
		line int
	}{
		{"ApexClass Foo: bad (line 12)", "", 0},
		{"force-app/main/default/classes/Foo.cls:27 bad", "force-app/main/default/classes/Foo.cls", 27},
		{`force-app\a\Foo.cls: bad (line 3)`, "force-app/a/Foo.cls", 3},
		{"Foo.cls bad", "Foo.cls", 0},
		{"Foo.testMethod: assertion failed", "", 0},
	}
	for _, tc := range cases {
		f, l := locate(tc.in)
		if f != tc.file || l != tc.line {
			t.Errorf("locate(%q) = %q, %d; want %q, %d", tc.in, f, l, tc.file, tc.line)
		}
	}
}

func TestExitCodeIsWorstStatus(t *testing.T) {
	cases := []struct {
		name string
		st   map[string]run.Status
		want int
	}{
		{"error beats fail", nil, exitcode.Error},
		{"fail without error", map[string]run.Status{"deploy": run.StatusPass}, exitcode.Fail},
		{"pass", map[string]run.Status{"deploy": run.StatusPass, "validate": run.StatusPass, "gate": run.StatusPass}, exitcode.OK},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			writeAll(t, dir, tc.st)
			b, err := Load(dir)
			if err != nil {
				t.Fatal(err)
			}
			if b.ExitCode() != tc.want {
				t.Fatalf("exit %d, want %d", b.ExitCode(), tc.want)
			}
		})
	}
}

func TestNoResultsIsAnError(t *testing.T) {
	if _, err := Load(t.TempDir()); err != ErrNoResults {
		t.Fatalf("err = %v, want ErrNoResults", err)
	}
}

func TestUnreadableEvidenceFailsClosed(t *testing.T) {
	dir := t.TempDir()
	r := res("doctor", 0, run.StatusPass, "")
	_, _ = r.Write(dir, func() time.Time { return t0 })
	_ = os.WriteFile(filepath.Join(dir, "broken.json"), []byte("{"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "other.json"), []byte(`{"schema":"something/else"}`), 0o644)
	b, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Problems) != 2 || b.ExitCode() != exitcode.Error {
		t.Fatalf("problems %v, exit %d; want 2 problems and exit 2", b.Problems, b.ExitCode())
	}
	if !strings.Contains(Summary(b), "Unreadable evidence") {
		t.Fatal("the summary should list unreadable files")
	}
}

func TestPermitPathOverridesResultsDir(t *testing.T) {
	dir := t.TempDir()
	r := res("doctor", 0, run.StatusPass, "")
	_, _ = r.Write(dir, func() time.Time { return t0 })
	p := permit.Permit{Schema: permit.Schema, Allow: true, Target: permit.Target{Name: "qa", Tier: "qa"}}
	pp := filepath.Join(t.TempDir(), "permit.json")
	writeJSON(t, pp, p)
	b, err := Load(dir, filepath.Join(t.TempDir(), "missing.json"), pp)
	if err != nil {
		t.Fatal(err)
	}
	if b.Permit == nil || !b.Permit.Allow {
		t.Fatal("the PERMIT at the given path was not read")
	}
	if !strings.Contains(Summary(b), "PERMIT: allow") {
		t.Fatal("the summary should show the PERMIT outcome")
	}
}

func jsonIndent(v any) ([]byte, error) { return json.MarshalIndent(v, "", "  ") }
