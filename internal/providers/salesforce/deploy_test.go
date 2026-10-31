package salesforce

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/glassops-platform/glassops/internal/providers"
)

const (
	startPrefix  = "sf project deploy start"
	quickPrefix  = "sf project deploy quick"
	resumePrefix = "sf project deploy resume"
	queryPrefix  = "sf data query"
	validatedID  = "0Af000000000004AAA"
)

// deployProject copies the sgd delta manifests into a temp project's plan dir and returns the
// project dir and a change set whose plan_sha256 matches them.
func deployProject(t *testing.T) (string, providers.ChangeSet) {
	t.Helper()
	dir := t.TempDir()
	for _, rel := range []string{"package/package.xml", "destructiveChanges/destructiveChanges.xml", "destructiveChanges/package.xml"} {
		data, err := os.ReadFile(filepath.Join("testdata", "sgd", "delta", filepath.FromSlash(rel)))
		if err != nil {
			t.Fatal(err)
		}
		dst := filepath.Join(dir, ".glassops", "plan", filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(dst, data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	sum, err := planDigest(dir, filepath.Join(dir, ".glassops", "plan"))
	if err != nil {
		t.Fatal(err)
	}
	cs := testChangeSet()
	cs.Dir = ".glassops/plan"
	cs.SHA256 = sum
	return dir, cs
}

func deploySetup(t *testing.T, replies map[string]fakeReply) (*Provider, *fakeExec, *session, string, providers.ChangeSet) {
	t.Helper()
	p, f, s := validateSetup(t, replies)
	dir, cs := deployProject(t)
	return p, f, s, dir, cs
}

// sandboxOrg answers the organization query as a Developer Edition org that is not a
// scratch org, so deploy start gets no --ignore-conflicts.
func sandboxOrg(t *testing.T) fakeReply {
	return fakeReply{Stdout: fixture(t, "query-organization-devhub.json")}
}

func TestDeployStartPass(t *testing.T) {
	p, f, s, dir, cs := deploySetup(t, nil)
	f.replies = map[string]fakeReply{
		queryPrefix: sandboxOrg(t),
		startPrefix: {Stdout: fixture(t, "deploy-start-pass.json")},
	}
	res, err := p.Deploy(context.Background(), s, cs, providers.DeployOptions{Dir: dir, TestLevel: "RunLocalTests"})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Passed || res.Quick || res.TimedOut || res.ID != "0Af000000000001AAA" || res.ComponentsDeployed != 4 || len(res.Problems) != 0 {
		t.Fatalf("result: %+v", res)
	}
	want := map[string]any{
		"salesforce.deploy_id":           "0Af000000000001AAA",
		"salesforce.deploy_status":       "Succeeded",
		"salesforce.check_only":          false,
		"salesforce.quick":               false,
		"salesforce.components_deployed": 4,
		"salesforce.tests_total":         1,
		"salesforce.tests_failed":        0,
	}
	for k, v := range want {
		if res.Facts[k] != v {
			t.Errorf("fact %s = %v, want %v", k, res.Facts[k], v)
		}
	}
	cmds := f.commands()
	if len(cmds) != 2 || !strings.HasPrefix(cmds[1], startPrefix) {
		t.Fatalf("commands: %v", cmds)
	}
	start := cmds[1]
	manifest := filepath.Join(dir, ".glassops", "plan", "package", "package.xml")
	destr := filepath.Join(dir, ".glassops", "plan", "destructiveChanges", "destructiveChanges.xml")
	for _, frag := range []string{
		"--manifest " + manifest, "--target-org glassops-target", "--wait 60",
		"--post-destructive-changes " + destr, "--test-level RunLocalTests", "--json",
	} {
		if !strings.Contains(start, frag) {
			t.Errorf("deploy start lacks %q: %s", frag, start)
		}
	}
	for _, never := range []string{"--ignore-errors", "--ignore-warnings", "--ignore-conflicts", "--dry-run"} {
		if strings.Contains(start, never) {
			t.Errorf("deploy start has %s: %s", never, start)
		}
	}
}

func TestDeployScratchIgnoresConflicts(t *testing.T) {
	p, f, s, dir, cs := deploySetup(t, nil)
	f.replies = map[string]fakeReply{
		queryPrefix: {Stdout: fixture(t, "query-organization-scratch.json")},
		startPrefix: {Stdout: fixture(t, "deploy-start-pass.json")},
	}
	if _, err := p.Deploy(context.Background(), s, cs, providers.DeployOptions{Dir: dir}); err != nil {
		t.Fatal(err)
	}
	cmds := f.commands()
	if !strings.Contains(cmds[len(cmds)-1], "--ignore-conflicts") {
		t.Fatalf("scratch deploy start has no --ignore-conflicts: %v", cmds)
	}
}

func TestDeployOrgQueryFailureIsError(t *testing.T) {
	p, f, s, dir, cs := deploySetup(t, nil)
	f.replies = map[string]fakeReply{
		queryPrefix: {Stdout: []byte(`{"status":1,"name":"INVALID_SESSION_ID","message":"Session expired"}`), ExitCode: 1},
		startPrefix: {Stdout: fixture(t, "deploy-start-pass.json")},
	}
	if _, err := p.Deploy(context.Background(), s, cs, providers.DeployOptions{Dir: dir}); err == nil {
		t.Fatal("deploy ran without knowing the org kind")
	}
	for _, c := range f.commands() {
		if strings.HasPrefix(c, startPrefix) {
			t.Fatal("deploy start ran after the org query failed")
		}
	}
}

func TestDeployResultBodies(t *testing.T) {
	cases := []struct {
		name      string
		fixture   string
		passed    bool
		timedOut  bool
		id        string
		deployed  int
		problems  []string
		wantFacts map[string]any
	}{
		{
			name: "compile failure has a full body", fixture: "deploy-start-compile-fail.json",
			id: "0Af000000000002AAA", deployed: 2,
			problems:  []string{"ApexClass Greeter: Missing '<EOF>' at 'public' (line 7)", "ApexClass GreeterTest: Dependent class"},
			wantFacts: map[string]any{"salesforce.deploy_status": "Failed", "salesforce.components_failed": 2},
		},
		{
			name: "SucceededPartial fails", fixture: "deploy-start-succeeded-partial.json",
			id: "0Af000000000006AAA", deployed: 2,
			problems:  []string{"ApexClass Greeter:", "SucceededPartial"},
			wantFacts: map[string]any{"salesforce.deploy_status": "SucceededPartial"},
		},
		{
			name: "async start is not finished", fixture: "deploy-start-async.json",
			timedOut: true, id: "0Af000000000003AAA",
			problems:  []string{"has not finished, status Queued"},
			wantFacts: map[string]any{"salesforce.timed_out": true},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p, f, s, dir, cs := deploySetup(t, nil)
			out := fixture(t, tc.fixture)
			f.replies = map[string]fakeReply{queryPrefix: sandboxOrg(t), startPrefix: {Stdout: out, ExitCode: 1}}
			res, err := p.Deploy(context.Background(), s, cs, providers.DeployOptions{Dir: dir})
			if err != nil {
				t.Fatal(err)
			}
			if res.Passed != tc.passed || res.TimedOut != tc.timedOut || res.ID != tc.id || res.ComponentsDeployed != tc.deployed {
				t.Fatalf("result: %+v", res)
			}
			joined := strings.Join(res.Problems, "\n")
			for _, want := range tc.problems {
				if !strings.Contains(joined, want) {
					t.Errorf("problems lack %q:\n%s", want, joined)
				}
			}
			for k, v := range tc.wantFacts {
				if res.Facts[k] != v {
					t.Errorf("fact %s = %v, want %v", k, res.Facts[k], v)
				}
			}
		})
	}
}

// A Succeeded body that still lists a component failure is a failure. The destructive "not
// found" fixture shows sf can report a failure with numberComponentErrors at 0.
func TestDeploySucceededWithComponentFailureFails(t *testing.T) {
	synthetic := strings.Replace(string(fixture(t, "deploy-start-pass.json")),
		`"componentFailures": []`,
		`"componentFailures": {"componentType": "", "fullName": "destructiveChanges.xml", "problem": "No CustomField named: Account.Tier__c found", "problemType": "Warning"}`, 1)
	if synthetic == string(fixture(t, "deploy-start-pass.json")) {
		t.Fatal("the synthetic body was not built")
	}
	p, f, s, dir, cs := deploySetup(t, nil)
	f.replies = map[string]fakeReply{queryPrefix: sandboxOrg(t), startPrefix: {Stdout: []byte(synthetic)}}
	res, err := p.Deploy(context.Background(), s, cs, providers.DeployOptions{Dir: dir})
	if err != nil {
		t.Fatal(err)
	}
	if res.Passed || !strings.Contains(strings.Join(res.Problems, " "), "No CustomField named") {
		t.Fatalf("result: %+v", res)
	}
}

func TestDeployZeroComponentsFails(t *testing.T) {
	synthetic := strings.Replace(string(fixture(t, "deploy-start-pass.json")), `"numberComponentsDeployed": 4`, `"numberComponentsDeployed": 0`, 1)
	p, f, s, dir, cs := deploySetup(t, nil)
	f.replies = map[string]fakeReply{queryPrefix: sandboxOrg(t), startPrefix: {Stdout: []byte(synthetic)}}
	res, err := p.Deploy(context.Background(), s, cs, providers.DeployOptions{Dir: dir})
	if err != nil {
		t.Fatal(err)
	}
	if res.Passed || !strings.Contains(strings.Join(res.Problems, " "), "no components") {
		t.Fatalf("result: %+v", res)
	}
}

func TestDeployCheckOnlyBodyIsError(t *testing.T) {
	p, f, s, dir, cs := deploySetup(t, nil)
	f.replies = map[string]fakeReply{queryPrefix: sandboxOrg(t), startPrefix: {Stdout: fixture(t, "deploy-validate-pass.json")}}
	if _, err := p.Deploy(context.Background(), s, cs, providers.DeployOptions{Dir: dir}); !errors.Is(err, ErrCheckOnlyDeploy) {
		t.Fatalf("err = %v, want ErrCheckOnlyDeploy", err)
	}
}

// A failure envelope with only a deploy Id in data, the shape deploy validate uses, is read
// again with deploy report.
func TestDeployThinFailureReadsReport(t *testing.T) {
	p, f, s, dir, cs := deploySetup(t, nil)
	f.replies = map[string]fakeReply{
		queryPrefix:  sandboxOrg(t),
		startPrefix:  {Stdout: []byte(`{"status":1,"name":"SyntheticThinError","message":"synthetic","data":{"deployId":"0Af000000000003AAA"}}`), ExitCode: 1},
		reportPrefix: {Stdout: fixture(t, "deploy-report-pass.json")},
	}
	res, err := p.Deploy(context.Background(), s, cs, providers.DeployOptions{Dir: dir})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Passed || res.ID != "0Af000000000003AAA" {
		t.Fatalf("result: %+v", res)
	}
	cmds := f.commands()
	if !strings.Contains(cmds[len(cmds)-1], "deploy report --job-id 0Af000000000003AAA") {
		t.Fatalf("commands: %v", cmds)
	}
}

func TestDeployStartErrorWithoutBodyIsError(t *testing.T) {
	p, f, s, dir, cs := deploySetup(t, nil)
	f.replies = map[string]fakeReply{
		queryPrefix: sandboxOrg(t),
		startPrefix: {Stdout: []byte(`{"status":1,"name":"SourceConflictError","message":"synthetic conflict"}`), ExitCode: 1},
	}
	_, err := p.Deploy(context.Background(), s, cs, providers.DeployOptions{Dir: dir})
	var sfErr *SFError
	if !errors.As(err, &sfErr) || sfErr.Name != "SourceConflictError" {
		t.Fatalf("err = %v", err)
	}
}

// A resume waits with deploy report, because deploy resume takes no --target-org and finds
// the org only in the local deploy cache of the run that started the deploy. Both commands
// return the same body, so both fixtures are read the same way.
func TestDeployResume(t *testing.T) {
	for _, fx := range []string{"deploy-report-pass.json", "deploy-resume-pass.json"} {
		t.Run(fx, func(t *testing.T) {
			p, f, s, dir, cs := deploySetup(t, nil)
			f.replies = map[string]fakeReply{reportPrefix: {Stdout: fixture(t, fx)}}
			res, err := p.Deploy(context.Background(), s, cs, providers.DeployOptions{Dir: dir, ResumeID: "0Af000000000003AAA", Wait: 90 * time.Second})
			if err != nil {
				t.Fatal(err)
			}
			if !res.Passed || res.ID != "0Af000000000003AAA" || res.ComponentsDeployed != 4 {
				t.Fatalf("result: %+v", res)
			}
			cmds := f.commands()
			if len(cmds) != 1 || !strings.Contains(cmds[0], "deploy report --job-id 0Af000000000003AAA --wait 2 --target-org glassops-target") {
				t.Fatalf("commands: %v", cmds)
			}
			if strings.Contains(cmds[0], resumePrefix) {
				t.Fatalf("deploy resume used: %v", cmds)
			}
		})
	}
}

func TestDeployQuickPass(t *testing.T) {
	p, f, s, dir, cs := deploySetup(t, nil)
	f.replies = map[string]fakeReply{quickPrefix: {Stdout: fixture(t, "deploy-quick-pass.json")}}
	res, err := p.Deploy(context.Background(), s, cs, providers.DeployOptions{Dir: dir, ValidatedID: validatedID, TestLevel: "RunLocalTests"})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Passed || !res.Quick || res.ID != "0Af000000000005AAA" || res.Facts["salesforce.quick"] != true {
		t.Fatalf("result: %+v", res)
	}
	cmds := f.commands()
	if len(cmds) != 1 || !strings.Contains(cmds[0], "deploy quick --job-id "+validatedID+" --wait 60 --target-org glassops-target") {
		t.Fatalf("commands: %v", cmds)
	}
	if strings.Contains(cmds[0], "--test-level") {
		t.Fatalf("quick deploy got a test level: %s", cmds[0])
	}
}

func TestDeployQuickRefusedFallsBack(t *testing.T) {
	p, f, s, dir, cs := deploySetup(t, nil)
	f.replies = map[string]fakeReply{
		quickPrefix: {Stdout: fixture(t, "deploy-quick-already-used.json"), ExitCode: 1},
		queryPrefix: sandboxOrg(t),
		startPrefix: {Stdout: fixture(t, "deploy-start-pass.json")},
	}
	res, err := p.Deploy(context.Background(), s, cs, providers.DeployOptions{Dir: dir, ValidatedID: validatedID})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Passed || res.Quick || res.Facts["salesforce.quick"] != false || res.Facts["salesforce.quick_refused"] != true {
		t.Fatalf("result: %+v", res)
	}
	if reason, _ := res.Facts["salesforce.quick_fallback"].(string); !strings.Contains(reason, "can't be used for quick deployment") {
		t.Fatalf("fallback reason = %q", reason)
	}
	cmds := f.commands()
	if len(cmds) != 3 || !strings.HasPrefix(cmds[0], quickPrefix) || !strings.HasPrefix(cmds[2], startPrefix) {
		t.Fatalf("commands: %v", cmds)
	}
}

func TestDeployQuickRequiredRefusesFallback(t *testing.T) {
	p, f, s, dir, cs := deploySetup(t, nil)
	f.replies = map[string]fakeReply{
		quickPrefix: {Stdout: fixture(t, "deploy-quick-already-used.json"), ExitCode: 1},
		queryPrefix: sandboxOrg(t),
		startPrefix: {Stdout: fixture(t, "deploy-start-pass.json")},
	}
	res, err := p.Deploy(context.Background(), s, cs, providers.DeployOptions{Dir: dir, ValidatedID: validatedID, QuickOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	if res.Passed || len(res.Problems) != 1 || !strings.Contains(res.Problems[0], "required") || res.Facts["salesforce.quick_refused"] != true {
		t.Fatalf("result: %+v", res)
	}
	if cmds := f.commands(); len(cmds) != 1 {
		t.Fatalf("a full deploy ran after a required quick deploy was refused: %v", cmds)
	}
}

func TestDeployQuickOtherErrorIsError(t *testing.T) {
	p, f, s, dir, cs := deploySetup(t, nil)
	f.replies = map[string]fakeReply{
		quickPrefix: {Stdout: fixture(t, "deploy-quick-invalid-id.json"), ExitCode: 1},
		startPrefix: {Stdout: fixture(t, "deploy-start-pass.json")},
	}
	_, err := p.Deploy(context.Background(), s, cs, providers.DeployOptions{Dir: dir, ValidatedID: "0Af000000000009AAA"})
	if err == nil || !strings.Contains(err.Error(), "invalid parameter value") {
		t.Fatalf("err = %v", err)
	}
	if cmds := f.commands(); len(cmds) != 1 {
		t.Fatalf("a full deploy ran after an unknown quick deploy error: %v", cmds)
	}
}

func TestDeployPreDestructiveManifest(t *testing.T) {
	p, f, s, dir, cs := deploySetup(t, nil)
	pre := filepath.Join(dir, ".glassops", "plan", "destructiveChanges", "destructiveChangesPre.xml")
	if err := os.WriteFile(pre, []byte("<Package/>\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	sum, err := planDigest(dir, filepath.Join(dir, ".glassops", "plan"))
	if err != nil {
		t.Fatal(err)
	}
	cs.SHA256 = sum
	cs.Destructive = append(cs.Destructive, ".glassops/plan/destructiveChanges/destructiveChangesPre.xml")
	f.replies = map[string]fakeReply{queryPrefix: sandboxOrg(t), startPrefix: {Stdout: fixture(t, "deploy-start-pass.json")}}
	if _, err := p.Deploy(context.Background(), s, cs, providers.DeployOptions{Dir: dir}); err != nil {
		t.Fatal(err)
	}
	cmds := f.commands()
	if !strings.Contains(cmds[len(cmds)-1], "--pre-destructive-changes "+pre) {
		t.Fatalf("deploy start: %s", cmds[len(cmds)-1])
	}
}

func TestDeployRefusesChangedPlan(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(t *testing.T, dir string, cs *providers.ChangeSet)
	}{
		{"manifest edited after the plan", func(t *testing.T, dir string, _ *providers.ChangeSet) {
			if err := os.WriteFile(filepath.Join(dir, ".glassops", "plan", "package", "package.xml"), []byte("<Package/>\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		}},
		{"file added to the plan dir", func(t *testing.T, dir string, _ *providers.ChangeSet) {
			if err := os.WriteFile(filepath.Join(dir, ".glassops", "plan", "extra.xml"), []byte("x"), 0o644); err != nil {
				t.Fatal(err)
			}
		}},
		{"manifest outside the plan dir", func(_ *testing.T, _ string, cs *providers.ChangeSet) {
			cs.Manifest = "elsewhere/package.xml"
		}},
		{"plan dir outside the project", func(_ *testing.T, _ string, cs *providers.ChangeSet) {
			cs.Dir = "../plan"
		}},
		{"no plan hash", func(_ *testing.T, _ string, cs *providers.ChangeSet) {
			cs.SHA256 = ""
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p, f, s, dir, cs := deploySetup(t, nil)
			f.replies = map[string]fakeReply{queryPrefix: sandboxOrg(t), startPrefix: {Stdout: fixture(t, "deploy-start-pass.json")}}
			tc.mutate(t, dir, &cs)
			if _, err := p.Deploy(context.Background(), s, cs, providers.DeployOptions{Dir: dir}); err == nil {
				t.Fatal("deploy accepted a changed plan")
			}
			if cmds := f.commands(); len(cmds) != 0 {
				t.Fatalf("sf ran for a changed plan: %v", cmds)
			}
		})
	}
}

func TestWaitMinutes(t *testing.T) {
	cases := map[time.Duration]string{0: "60", time.Second: "1", time.Minute: "1", 61 * time.Second: "2", 30 * time.Minute: "30"}
	for d, want := range cases {
		if got := waitMinutes(d); got != want {
			t.Errorf("waitMinutes(%s) = %s, want %s", d, got, want)
		}
	}
}

func TestDeployNeedsOwnSession(t *testing.T) {
	p := newTestProvider(t, newFakeExec(nil), nil)
	if _, err := p.Deploy(context.Background(), nil, providers.ChangeSet{}, providers.DeployOptions{}); err == nil {
		t.Fatal("deploy accepted a session it did not make")
	}
}
