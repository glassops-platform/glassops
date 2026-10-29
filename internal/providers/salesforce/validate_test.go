package salesforce

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glassops-platform/glassops/internal/providers"
)

const validatePrefix = "sf project deploy validate"
const reportPrefix = "sf project deploy report"

func validateSetup(t *testing.T, replies map[string]fakeReply) (*Provider, *fakeExec, *session) {
	t.Helper()
	f := newFakeExec(replies)
	p := newTestProvider(t, f, nil)
	s := newTestSession(t, providers.Target{Name: "qa", ID: "00D000000000001AAA", Tier: "qa"})
	return p, f, s
}

func testChangeSet() providers.ChangeSet {
	return providers.ChangeSet{
		Schema: providers.PlanSchema, Provider: "salesforce",
		Manifest:    ".glassops/plan/package/package.xml",
		Destructive: []string{".glassops/plan/destructiveChanges/destructiveChanges.xml"},
	}
}

func TestValidatePass(t *testing.T) {
	pass := fixture(t, "deploy-validate-pass.json")
	// The pass fixture holds the full deploy body, the same shape deploy report returns.
	p, f, s := validateSetup(t, map[string]fakeReply{
		validatePrefix: {Stdout: pass},
		reportPrefix:   {Stdout: pass},
	})
	dir := t.TempDir()
	res, err := p.Validate(context.Background(), s, testChangeSet(), providers.ValidateOptions{
		Dir: dir, TestLevel: "RunLocalTests",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Passed || res.ID == "" || len(res.Problems) != 0 {
		t.Fatalf("result: %+v", res)
	}
	want := map[string]any{
		"salesforce.check_only":                     true,
		"salesforce.deploy_status":                  "Succeeded",
		"salesforce.tests_total":                    1,
		"salesforce.tests_failed":                   0,
		"salesforce.coverage.Greeter.lines":         2,
		"salesforce.coverage.Greeter.percent":       100,
		"salesforce.coverage.Greeter.lines_covered": 2,
	}
	for k, v := range want {
		if res.Facts[k] != v {
			t.Errorf("fact %s = %v, want %v", k, res.Facts[k], v)
		}
	}
	if res.Facts["salesforce.deploy_id"] != res.ID {
		t.Errorf("deploy_id fact %v, id %s", res.Facts["salesforce.deploy_id"], res.ID)
	}

	cmds := f.commands()
	if len(cmds) != 2 {
		t.Fatalf("commands: %v", cmds)
	}
	v := cmds[0]
	manifest := filepath.Join(dir, ".glassops", "plan", "package", "package.xml")
	destr := filepath.Join(dir, ".glassops", "plan", "destructiveChanges", "destructiveChanges.xml")
	for _, frag := range []string{
		"project deploy validate", "--manifest " + manifest, "--target-org glassops-target",
		"--wait 60", "--post-destructive-changes " + destr, "--test-level RunLocalTests", "--json",
	} {
		if !strings.Contains(v, frag) {
			t.Errorf("validate command lacks %q: %s", frag, v)
		}
	}
	if strings.Contains(v, "--tests") {
		t.Errorf("--tests without RunSpecifiedTests: %s", v)
	}
	if !strings.Contains(cmds[1], "project deploy report --job-id "+res.ID+" --target-org glassops-target") {
		t.Errorf("report command: %s", cmds[1])
	}
	for _, c := range cmds {
		if strings.Contains(c, "deploy start") {
			t.Fatalf("deploy start was called: %s", c)
		}
	}
	if f.calls[0].Dir != dir {
		t.Errorf("run dir %q, want %q", f.calls[0].Dir, dir)
	}
}

func TestValidateTestLevelArguments(t *testing.T) {
	pass := fixture(t, "deploy-validate-pass.json")
	cases := []struct {
		name  string
		opts  providers.ValidateOptions
		has   []string
		hasNo []string
	}{
		{"unset", providers.ValidateOptions{Tests: []string{"A"}}, nil, []string{"--test-level", "--tests"}},
		{"specified", providers.ValidateOptions{TestLevel: "RunSpecifiedTests", Tests: []string{"A", "B"}},
			[]string{"--test-level RunSpecifiedTests", "--tests A", "--tests B"}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p, f, s := validateSetup(t, map[string]fakeReply{validatePrefix: {Stdout: pass}, reportPrefix: {Stdout: pass}})
			tc.opts.Dir = t.TempDir()
			if _, err := p.Validate(context.Background(), s, providers.ChangeSet{Manifest: "m.xml"}, tc.opts); err != nil {
				t.Fatal(err)
			}
			v := f.commands()[0]
			for _, h := range tc.has {
				if !strings.Contains(v, h) {
					t.Errorf("missing %q in %s", h, v)
				}
			}
			for _, h := range tc.hasNo {
				if strings.Contains(v, h) {
					t.Errorf("unexpected %q in %s", h, v)
				}
			}
			if strings.Contains(v, "--post-destructive-changes") {
				t.Errorf("no destructive entries, flag present: %s", v)
			}
		})
	}
}

func TestValidateCompileFail(t *testing.T) {
	p, f, s := validateSetup(t, map[string]fakeReply{
		validatePrefix: {Stdout: fixture(t, "deploy-validate-compile-fail.json"), ExitCode: 1},
		reportPrefix:   {Stdout: fixture(t, "deploy-report-compile-fail.json")},
	})
	res, err := p.Validate(context.Background(), s, testChangeSet(), providers.ValidateOptions{Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if res.Passed || res.ID != "0Af000000000002AAA" {
		t.Fatalf("result: %+v", res)
	}
	want := []string{
		"ApexClass Greeter: Variable does not exist: undefinedVariable (line 6)",
		"ApexClass GreeterTest: Dependent class is invalid and needs recompilation: Class Greeter : Variable does not exist: undefinedVariable (line 6)",
	}
	if len(res.Problems) != 2 || res.Problems[0] != want[0] || res.Problems[1] != want[1] {
		t.Fatalf("problems: %q", res.Problems)
	}
	if res.Facts["salesforce.components_failed"] != 2 || res.Facts["salesforce.components_total"] != 2 || res.Facts["salesforce.deploy_status"] != "Failed" {
		t.Fatalf("facts: %v", res.Facts)
	}
	if cmds := f.commands(); len(cmds) != 2 || !strings.Contains(cmds[1], "--job-id 0Af000000000002AAA") {
		t.Fatalf("commands: %v", cmds)
	}
}

func TestValidateTestFail(t *testing.T) {
	p, _, s := validateSetup(t, map[string]fakeReply{
		validatePrefix: {Stdout: fixture(t, "deploy-validate-test-fail.json"), ExitCode: 1},
		reportPrefix:   {Stdout: fixture(t, "deploy-report-test-fail.json")},
	})
	res, err := p.Validate(context.Background(), s, testChangeSet(), providers.ValidateOptions{Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if res.Passed || len(res.Problems) != 1 {
		t.Fatalf("result: %+v", res)
	}
	if res.Problems[0] != "GreeterTest.greetsByName: System.AssertException: Assertion Failed: Expected: Goodbye, Ada, Actual: Hello, Ada" {
		t.Fatalf("problem: %q", res.Problems[0])
	}
	if res.Facts["salesforce.tests_failed"] != 1 || res.Facts["salesforce.tests_total"] != 1 {
		t.Fatalf("facts: %v", res.Facts)
	}
}

// flipCheckOnly returns the fixture with result.checkOnly set to v, in memory only.
func flipCheckOnly(t *testing.T, data []byte, v bool) []byte {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	m["result"].(map[string]any)["checkOnly"] = v
	out, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestValidateRefusesNotCheckOnly(t *testing.T) {
	pass := fixture(t, "deploy-validate-pass.json")
	failReport := fixture(t, "deploy-report-compile-fail.json")
	cases := []struct {
		name     string
		validate fakeReply
		report   fakeReply
	}{
		{"validate body", fakeReply{Stdout: flipCheckOnly(t, pass, false)}, fakeReply{Stdout: pass}},
		{"report body", fakeReply{Stdout: pass}, fakeReply{Stdout: flipCheckOnly(t, pass, false)}},
		{"failed validate, report body", fakeReply{Stdout: fixture(t, "deploy-validate-compile-fail.json"), ExitCode: 1},
			fakeReply{Stdout: flipCheckOnly(t, failReport, false)}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p, _, s := validateSetup(t, map[string]fakeReply{validatePrefix: tc.validate, reportPrefix: tc.report})
			_, err := p.Validate(context.Background(), s, testChangeSet(), providers.ValidateOptions{Dir: t.TempDir()})
			if !errors.Is(err, ErrNotCheckOnly) {
				t.Fatalf("err = %v, want ErrNotCheckOnly", err)
			}
		})
	}
}

func TestValidateErrorsWithoutDeployID(t *testing.T) {
	p, f, s := validateSetup(t, map[string]fakeReply{
		validatePrefix: {Stdout: []byte(`{"status":1,"name":"SomethingElse","message":"boom"}`), ExitCode: 1},
	})
	_, err := p.Validate(context.Background(), s, testChangeSet(), providers.ValidateOptions{Dir: t.TempDir()})
	var sfErr *SFError
	if !errors.As(err, &sfErr) || sfErr.Name != "SomethingElse" {
		t.Fatalf("err = %v", err)
	}
	if len(f.commands()) != 1 {
		t.Fatalf("report ran after a failure with no Id: %v", f.commands())
	}
}

func TestValidateReportFailureIsError(t *testing.T) {
	p, _, s := validateSetup(t, map[string]fakeReply{
		validatePrefix: {Stdout: fixture(t, "deploy-validate-compile-fail.json"), ExitCode: 1},
		reportPrefix:   {Stdout: []byte(`{"status":1,"name":"X","message":"no report"}`), ExitCode: 1},
	})
	if _, err := p.Validate(context.Background(), s, testChangeSet(), providers.ValidateOptions{Dir: t.TempDir()}); err == nil {
		t.Fatal("want an error when the report cannot be read")
	}
}

func TestValidateSingleObjectsAndNeverLeaksUser(t *testing.T) {
	// The Metadata API returns one failure as a bare object, and counts as strings.
	body := `{"status":0,"result":{"checkOnly":true,"id":"0Af000000000009AAA","status":"Failed",
	"createdByName":"Some Person","numberComponentsTotal":"3","numberComponentErrors":"1",
	"details":{"componentFailures":{"componentType":"CustomField","fullName":"Account.X__c","problem":"bad\nthing","lineNumber":"4"},
	"runTestResult":{"failures":{"name":"T","methodName":"m","message":"nope"},
	"codeCoverage":{"name":"Foo","numLocations":"4","numLocationsNotCovered":"1"},"numTestsRun":1,"numFailures":1}}}}`
	p, _, s := validateSetup(t, map[string]fakeReply{
		validatePrefix: {Stdout: []byte(body)}, reportPrefix: {Stdout: []byte(body)},
	})
	res, err := p.Validate(context.Background(), s, testChangeSet(), providers.ValidateOptions{Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Problems) != 2 || res.Problems[0] != "CustomField Account.X__c: bad thing (line 4)" || res.Problems[1] != "T.m: nope" {
		t.Fatalf("problems: %q", res.Problems)
	}
	if res.Facts["salesforce.components_total"] != 3 || res.Facts["salesforce.coverage.Foo.percent"] != 75 {
		t.Fatalf("facts: %v", res.Facts)
	}
	all, _ := json.Marshal(res)
	if strings.Contains(string(all), "Some Person") || strings.Contains(strings.ToLower(string(all)), "createdby") {
		t.Fatalf("result leaks a user: %s", all)
	}
}

func TestValidateNeedsOwnSession(t *testing.T) {
	p, _, _ := validateSetup(t, nil)
	if _, err := p.Validate(context.Background(), nil, testChangeSet(), providers.ValidateOptions{}); err == nil {
		t.Fatal("want an error for a foreign session")
	}
}
