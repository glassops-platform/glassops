package salesforce

import (
	"context"
	"encoding/xml"
	"strings"
	"testing"
	"time"

	"github.com/glassops-platform/glassops/internal/core/junit"
	"github.com/glassops-platform/glassops/internal/providers"
)

const (
	runTestPrefix = "sf apex run test"
	getTestPrefix = "sf apex get test"
)

func testerSetup(t *testing.T, replies map[string]fakeReply) (*Provider, *fakeExec, *session) {
	t.Helper()
	f := newFakeExec(replies)
	p := newTestProvider(t, f, nil)
	return p, f, newTestSession(t, providers.Target{Name: "qa", ID: "00D000000000001AAA", Tier: "qa"})
}

func TestRunTestsPass(t *testing.T) {
	p, f, s := testerSetup(t, map[string]fakeReply{runTestPrefix: {Stdout: fixture(t, "apex-test-pass.json")}})
	res, err := p.RunTests(context.Background(), s, providers.TestRequest{
		Dir: t.TempDir(), TestLevel: "RunLocalTests", Wait: 90 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Passed || !res.Finished || res.Ran != 1 || res.Failed != 0 || res.OrgCoverage != 100 || res.RunID != "707000000000001" {
		t.Fatalf("result: %+v", res)
	}
	if len(res.Classes) != 1 || res.Classes[0] != (providers.ClassCoverage{Name: "Greeter", Lines: 2, Covered: 2, Percent: 100}) {
		t.Fatalf("classes: %+v", res.Classes)
	}
	cmds := f.commands()
	if len(cmds) != 1 {
		t.Fatalf("commands: %v", cmds)
	}
	for _, frag := range []string{"apex run test", "--test-level RunLocalTests", "--code-coverage",
		"--result-format json", "--target-org glassops-target", "--wait 2", "--json"} {
		if !strings.Contains(cmds[0], frag) {
			t.Errorf("command lacks %q: %s", frag, cmds[0])
		}
	}
	if strings.Contains(cmds[0], "--tests") {
		t.Errorf("no tests were named: %s", cmds[0])
	}
}

func TestRunTestsNamesTests(t *testing.T) {
	p, f, s := testerSetup(t, map[string]fakeReply{runTestPrefix: {Stdout: fixture(t, "apex-test-pass.json")}})
	_, err := p.RunTests(context.Background(), s, providers.TestRequest{
		Dir: t.TempDir(), TestLevel: "RunSpecifiedTests", Tests: []string{"GreeterTest", "OtherTest.one"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if c := f.commands()[0]; !strings.Contains(c, "--tests GreeterTest,OtherTest.one") || !strings.Contains(c, "--wait 60") {
		t.Fatalf("command: %s", c)
	}
}

func TestRunTestsFailureReadsBodyOnExit100(t *testing.T) {
	p, _, s := testerSetup(t, map[string]fakeReply{runTestPrefix: {Stdout: fixture(t, "apex-test-fail.json"), ExitCode: 100}})
	res, err := p.RunTests(context.Background(), s, providers.TestRequest{Dir: t.TempDir()})
	if err != nil {
		t.Fatalf("a failing test is a result, not an error: %v", err)
	}
	if res.Passed || !res.Finished || res.Ran != 2 || res.Failed != 1 || res.OrgCoverage != 100 {
		t.Fatalf("result: %+v", res)
	}
	if len(res.Failures) != 1 || !strings.HasPrefix(res.Failures[0], "FailingTest.fails: System.AssertException") {
		t.Fatalf("failures: %v", res.Failures)
	}
	var back junit.Suites
	if err := xml.Unmarshal(res.JUnit, &back); err != nil {
		t.Fatalf("junit does not parse: %v", err)
	}
	if back.Tests != 2 || back.Failures != 1 {
		t.Fatalf("junit totals: %+v", back)
	}
	if strings.Contains(string(res.JUnit), "user@example.com") {
		t.Fatalf("junit holds a username: %s", res.JUnit)
	}
}

func TestRunTestsUnfinishedThenResumed(t *testing.T) {
	p, f, s := testerSetup(t, map[string]fakeReply{
		runTestPrefix: {Stdout: fixture(t, "apex-test-async.json")},
		getTestPrefix: {Stdout: fixture(t, "apex-get-test-fail.json"), ExitCode: 100},
	})
	res, err := p.RunTests(context.Background(), s, providers.TestRequest{Dir: t.TempDir(), TestLevel: "RunLocalTests"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Passed || !res.Finished || res.Failed != 1 || res.RunID != "707000000000005" {
		t.Fatalf("result: %+v", res)
	}
	cmds := f.commands()
	if len(cmds) != 2 || !strings.Contains(cmds[1], "apex get test --test-run-id 707000000000005") ||
		!strings.Contains(cmds[1], "--code-coverage") {
		t.Fatalf("commands: %v", cmds)
	}
}

func TestRunTestsStaysUnfinished(t *testing.T) {
	async := fixture(t, "apex-test-async.json")
	p, _, s := testerSetup(t, map[string]fakeReply{runTestPrefix: {Stdout: async}, getTestPrefix: {Stdout: async}})
	res, err := p.RunTests(context.Background(), s, providers.TestRequest{Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if res.Finished || res.Passed || res.RunID != "707000000000005" {
		t.Fatalf("result: %+v", res)
	}
}

func TestRunTestsResume(t *testing.T) {
	p, f, s := testerSetup(t, map[string]fakeReply{getTestPrefix: {Stdout: fixture(t, "apex-get-test-fail.json"), ExitCode: 100}})
	res, err := p.RunTests(context.Background(), s, providers.TestRequest{Dir: t.TempDir(), ResumeID: "707000000000005"})
	if err != nil || res.Failed != 1 {
		t.Fatalf("%v %+v", err, res)
	}
	if cmds := f.commands(); len(cmds) != 1 || strings.Contains(cmds[0], "apex run test") {
		t.Fatalf("commands: %v", cmds)
	}
}

func TestRunTestsRefusesNoTestRun(t *testing.T) {
	p, f, s := testerSetup(t, nil)
	if _, err := p.RunTests(context.Background(), s, providers.TestRequest{Dir: t.TempDir(), TestLevel: "NoTestRun"}); err == nil {
		t.Fatal("NoTestRun ran")
	}
	if len(f.commands()) != 0 {
		t.Fatalf("sf was called: %v", f.commands())
	}
}

func TestRunTestsToolErrors(t *testing.T) {
	p, _, s := testerSetup(t, map[string]fakeReply{runTestPrefix: {
		Stdout: []byte(`{"status":1,"name":"NamedOrgNotFound","message":"no org"}`), ExitCode: 1}})
	if _, err := p.RunTests(context.Background(), s, providers.TestRequest{Dir: t.TempDir()}); err == nil {
		t.Fatal("an sf error came back as a result")
	}
}

func TestApexBodyNeitherSummaryNorRunID(t *testing.T) {
	if _, err := (apexBody{}).toResult(""); err == nil {
		t.Fatal("empty body accepted")
	}
}

func TestParsePercent(t *testing.T) {
	cases := map[string]struct {
		n  int
		ok bool
	}{"100%": {100, true}, " 87% ": {87, true}, "74.9%": {74, true}, "": {0, false}, "n/a": {0, false}, "120%": {0, false}}
	for in, want := range cases {
		n, ok := parsePercent(in)
		if n != want.n || ok != want.ok {
			t.Errorf("%q: %d %v", in, n, ok)
		}
	}
}
