package salesforce

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/providers"
)

// The verify tests answer every sf call with a step 0.2 or 0.3 capture from testdata/sf. The
// project, its plan manifest and its Flow source are written by each test.

const (
	vfyDeployPass     = "0Af000000000003AAA" // deploy-report-flow-pass.json
	vfyDeployDraft    = "0Af000000000005AAA" // deploy-start-flow-draft.json
	vfyDeployCovFail  = "0Af000000000002AAA" // deploy-report-coverage-fail.json
	vfyDeployNoTests  = "0Af000000000007AAA" // deploy-start-notestrun.json
	vfyDeployCheckOnl = "0Af000000000008AAA" // deploy-validate-for-checkonly.json
)

const (
	vfyKeyFlowDef = "sf data query --use-tooling-api --query SELECT Id, DeveloperName, ActiveVersionId"
	vfyKeyFlow    = "sf data query --use-tooling-api --query SELECT Id, VersionNumber"
	vfyKeyWindow  = "sf data query --use-tooling-api --query SELECT " + windowFields
)

func vfyKeyRequest(id string) string {
	return "sf data query --use-tooling-api --query " + deployRequestQuery(id)
}

func vfyKeyReport(id string) string {
	return "sf project deploy report --job-id " + id
}

// vfyTime parses a capture timestamp.
func vfyTime(t *testing.T, s string) time.Time {
	t.Helper()
	v, ok := parseSFTime(s)
	if !ok {
		t.Fatalf("bad time %q", s)
	}
	return v
}

type vfyMember struct{ typ, name string }

// vfyProject writes an sfdx project with the plan manifest listing members, an optional
// destructive manifest, and a Flow source file with the given status for each name in
// flows. It returns the dir and the change set plan would write.
func vfyProject(t *testing.T, members []vfyMember, destructive []vfyMember, flows map[string]string) (string, providers.ChangeSet) {
	t.Helper()
	dir := t.TempDir()
	write := func(rel, content string) {
		t.Helper()
		p := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(ProjectFile, `{"packageDirectories":[{"path":"force-app","default":true}]}`)
	for name, status := range flows {
		write("force-app/main/default/flows/"+name+".flow-meta.xml",
			`<?xml version="1.0" encoding="UTF-8"?>
<Flow xmlns="http://soap.sforce.com/2006/04/metadata">
    <label>`+name+`</label>
    <processType>AutoLaunchedFlow</processType>
    <status>`+status+`</status>
</Flow>
`)
	}
	pkgXML := func(ms []vfyMember) string {
		var b strings.Builder
		b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n" + `<Package xmlns="http://soap.sforce.com/2006/04/metadata">` + "\n")
		for _, m := range ms {
			b.WriteString("  <types><members>" + m.name + "</members><name>" + m.typ + "</name></types>\n")
		}
		b.WriteString("  <version>62.0</version>\n</Package>\n")
		return b.String()
	}
	write(".glassops/plan/package/package.xml", pkgXML(members))
	cs := providers.ChangeSet{Schema: providers.PlanSchema, Provider: "salesforce", Dir: ".glassops/plan", Manifest: ".glassops/plan/package/package.xml"}
	if destructive != nil {
		write(".glassops/plan/destructiveChanges/destructiveChanges.xml", pkgXML(destructive))
		cs.Destructive = []string{".glassops/plan/destructiveChanges/destructiveChanges.xml"}
	}
	sum, err := providers.PlanDigest(dir, filepath.Join(dir, ".glassops", "plan"))
	if err != nil {
		t.Fatal(err)
	}
	cs.SHA256 = sum
	return dir, cs
}

// vfyCase is one verify run against captured replies.
type vfyCase struct {
	id        string
	testLevel string
	quick     bool
	from, to  string
	flows     string
	foreign   string
	replies   map[string]fakeReply
}

func vfyRun(t *testing.T, dir string, cs providers.ChangeSet, c vfyCase) (providers.VerifyResult, *fakeExec, error) {
	t.Helper()
	f := newFakeExec(c.replies)
	p := newTestProvider(t, f, nil)
	sess := newTestSession(t, providers.Target{Name: "prod", ID: "00D000000000001AAA", Tier: "prod"})
	from := vfyTime(t, c.from)
	req := providers.VerifyRequest{
		Dir:  dir,
		Plan: cs,
		Deploy: providers.DeployRecord{
			ID: c.id, Quick: c.quick, TestLevel: c.testLevel,
			Started: from, Ended: from.Add(5 * time.Second),
		},
		WindowStart:    from,
		WindowEnd:      vfyTime(t, c.to),
		ForeignDeploys: c.foreign,
		Flows:          c.flows,
	}
	res, err := p.Verify(context.Background(), sess, req)
	return res, f, err
}

func vfyReply(t *testing.T, name string) fakeReply {
	return fakeReply{Stdout: fixture(t, name)}
}

// vfyPassReplies answer every call for the deploy that put Greeter, GreeterTest and the
// Active Flow Glassops_Probe in the org.
func vfyPassReplies(t *testing.T) map[string]fakeReply {
	return map[string]fakeReply{
		vfyKeyRequest(vfyDeployPass): vfyReply(t, "tooling-deployrequest-pass.json"),
		vfyKeyReport(vfyDeployPass):  vfyReply(t, "deploy-report-flow-pass.json"),
		vfyKeyFlowDef:                vfyReply(t, "tooling-flowdefinition-active.json"),
		vfyKeyFlow:                   vfyReply(t, "tooling-flow-active.json"),
		vfyKeyWindow:                 vfyReply(t, "tooling-deployrequest-window.json"),
	}
}

var vfyPassMembers = []vfyMember{{"ApexClass", "Greeter"}, {"ApexClass", "GreeterTest"}, {"Flow", "Glassops_Probe"}}

// vfyPassCase is the deploy 0Af...3 with the window closing before the next deploy started.
func vfyPassCase(t *testing.T) vfyCase {
	return vfyCase{
		id: vfyDeployPass, testLevel: "RunLocalTests",
		from: "2026-11-08T18:33:11.000+0000", to: "2026-11-08T18:33:30.000+0000",
		replies: vfyPassReplies(t),
	}
}

func vfyStatus(res providers.VerifyResult, name string) (providers.VerifyStatus, string) {
	for _, c := range res.Checks {
		if c.Name == name {
			return c.Status, c.Detail
		}
	}
	return "", ""
}

func vfyWant(t *testing.T, res providers.VerifyResult, name string, want providers.VerifyStatus, detail ...string) {
	t.Helper()
	got, d := vfyStatus(res, name)
	if got != want {
		t.Fatalf("%s = %s (%s), want %s. All checks: %+v", name, got, d, want, res.Checks)
	}
	for _, s := range detail {
		if !strings.Contains(d, s) {
			t.Fatalf("%s detail %q does not contain %q", name, d, s)
		}
	}
}

func TestVerifyPassesOnCapturedDeploy(t *testing.T) {
	dir, cs := vfyProject(t, vfyPassMembers, nil, map[string]string{"Glassops_Probe": "Active"})
	res, f, err := vfyRun(t, dir, cs, vfyPassCase(t))
	if err != nil {
		t.Fatal(err)
	}
	if !res.Passed {
		t.Fatalf("verify failed: %+v", res.Checks)
	}
	vfyWant(t, res, CheckVerifyStatus, providers.VerifyPass)
	vfyWant(t, res, CheckVerifyCheckOnly, providers.VerifyPass)
	vfyWant(t, res, CheckVerifyMembers, providers.VerifyPass, "all 3 plan members")
	vfyWant(t, res, CheckVerifyTests, providers.VerifyPass, "1 of 1 tests ran at test level RunLocalTests")
	vfyWant(t, res, CheckVerifyFlows, providers.VerifyPass, "Glassops_Probe")
	vfyWant(t, res, CheckVerifyForeign, providers.VerifyPass)
	if res.Facts["verify.components_deployed"] != 3 || res.Facts["verify.plan_members"] != 3 || res.Facts["verify.component_successes"] != 3 {
		t.Fatalf("count facts: %v", res.Facts)
	}
	// verify only reads: no deploy start, quick or validate.
	for _, c := range f.commands() {
		if strings.Contains(c, "deploy start") || strings.Contains(c, "deploy quick") || strings.Contains(c, "deploy validate") {
			t.Fatalf("verify ran %q", c)
		}
	}
	// The window query is bounded on both ends and leaves check-only deploys out.
	var window string
	for _, c := range f.commands() {
		if strings.HasPrefix(c, vfyKeyWindow) {
			window = c
		}
	}
	for _, s := range []string{"CheckOnly = false", "CreatedDate >= 2026-11-08T18:33:11Z", "CreatedDate <= 2026-11-08T18:33:31Z"} {
		if !strings.Contains(window, s) {
			t.Fatalf("window query %q lacks %q", window, s)
		}
	}
}

func TestVerifyDraftFlowFails(t *testing.T) {
	// The second deploy sent the Flow as Draft. The org kept version 1 active and version 2
	// is the Draft latest. The source here says Active, as a project that expects the Flow
	// live would.
	draft := func(t *testing.T) vfyCase {
		return vfyCase{
			id: vfyDeployDraft, testLevel: "NoTestRun",
			from: "2026-11-08T18:34:01.000+0000", to: "2026-11-08T18:34:10.000+0000",
			replies: map[string]fakeReply{
				// No DeployRequest by Id was captured for this deploy. The window capture
				// holds its row, and verify picks the row whose Id matches.
				vfyKeyRequest(vfyDeployDraft): vfyReply(t, "tooling-deployrequest-window.json"),
				vfyKeyReport(vfyDeployDraft):  vfyReply(t, "deploy-start-flow-draft.json"),
				vfyKeyFlowDef:                 vfyReply(t, "tooling-flowdefinition-draft.json"),
				vfyKeyFlow:                    vfyReply(t, "tooling-flow-draft.json"),
				vfyKeyWindow:                  vfyReply(t, "tooling-deployrequest-window.json"),
			},
		}
	}
	members := []vfyMember{{"Flow", "Glassops_Probe"}}
	dir, cs := vfyProject(t, members, nil, map[string]string{"Glassops_Probe": "Active"})

	res, _, err := vfyRun(t, dir, cs, draft(t))
	if err != nil {
		t.Fatal(err)
	}
	if res.Passed {
		t.Fatalf("verify passed with a Draft latest version: %+v", res.Checks)
	}
	vfyWant(t, res, CheckVerifyFlows, providers.VerifyFail, "runs version 1", "latest version 2 is Draft")
	vfyWant(t, res, CheckVerifyStatus, providers.VerifyPass)
	vfyWant(t, res, CheckVerifyMembers, providers.VerifyPass)
	vfyWant(t, res, CheckVerifyTests, providers.VerifySkipped, "NoTestRun")
	if got := res.Facts["verify.flows_not_live"].([]string); len(got) != 1 || got[0] != "Glassops_Probe" {
		t.Fatalf("flows_not_live = %v", got)
	}

	t.Run("warn", func(t *testing.T) {
		c := draft(t)
		c.flows = config.VerifyWarn
		res, _, err := vfyRun(t, dir, cs, c)
		if err != nil {
			t.Fatal(err)
		}
		vfyWant(t, res, CheckVerifyFlows, providers.VerifyWarn, "latest version 2 is Draft")
		if !res.Passed {
			t.Fatalf("a warning failed verify: %+v", res.Checks)
		}
	})
	t.Run("off", func(t *testing.T) {
		c := draft(t)
		c.flows = config.VerifyOff
		res, f, err := vfyRun(t, dir, cs, c)
		if err != nil {
			t.Fatal(err)
		}
		vfyWant(t, res, CheckVerifyFlows, providers.VerifySkipped)
		for _, cmd := range f.commands() {
			if strings.HasPrefix(cmd, vfyKeyFlowDef) {
				t.Fatal("verify.flows off still queried FlowDefinition")
			}
		}
	})
	t.Run("draft in source is not checked", func(t *testing.T) {
		dir, cs := vfyProject(t, members, nil, map[string]string{"Glassops_Probe": "Draft"})
		res, _, err := vfyRun(t, dir, cs, draft(t))
		if err != nil {
			t.Fatal(err)
		}
		vfyWant(t, res, CheckVerifyFlows, providers.VerifySkipped, "no Flow in the plan is Active in source")
		if !res.Passed {
			t.Fatalf("verify failed: %+v", res.Checks)
		}
	})
}

func TestVerifyActiveFlowModifiedBeforeDeployFails(t *testing.T) {
	// The active capture's rows, read as if the deploy had started after the Flow's last
	// change, which is the Draft deploy's start: a changed Flow whose active version predates
	// the deploy did not take.
	act := flowDefinitionRow{ID: "300000000000001AAA", ActiveVersionID: "301000000000001AAA", LatestVersionID: "301000000000001AAA"}
	versions := map[string]flowVersionRow{"301000000000001": {ID: "301000000000001AAA", VersionNumber: 1, LastModifiedDate: "2026-11-08T18:33:12.000+0000"}}
	if msg := flowProblem("Glassops_Probe", act, true, versions, true, vfyTime(t, "2026-11-08T18:34:01.000+0000")); !strings.Contains(msg, "before the deploy started") {
		t.Fatalf("problem = %q", msg)
	}
	if msg := flowProblem("Glassops_Probe", act, true, versions, false, vfyTime(t, "2026-11-08T18:34:01.000+0000")); msg != "" {
		t.Fatalf("an unchanged Flow failed: %q", msg)
	}
	if msg := flowProblem("Glassops_Probe", act, false, versions, true, time.Time{}); !strings.Contains(msg, "not in the target") {
		t.Fatalf("problem = %q", msg)
	}
	noActive := act
	noActive.ActiveVersionID = ""
	if msg := flowProblem("Glassops_Probe", noActive, true, versions, true, time.Time{}); !strings.Contains(msg, "no active version") {
		t.Fatalf("problem = %q", msg)
	}
}

func TestVerifyMissingPlanMemberFails(t *testing.T) {
	members := append(append([]vfyMember{}, vfyPassMembers...), vfyMember{"ApexClass", "Orphan"})
	dir, cs := vfyProject(t, members, nil, map[string]string{"Glassops_Probe": "Active"})
	res, _, err := vfyRun(t, dir, cs, vfyPassCase(t))
	if err != nil {
		t.Fatal(err)
	}
	if res.Passed {
		t.Fatal("verify passed with a plan member missing from the deploy")
	}
	vfyWant(t, res, CheckVerifyMembers, providers.VerifyFail, "1 of 4 plan members", "ApexClass:Orphan")
	if got := res.Facts["verify.members_missing"].([]string); len(got) != 1 || got[0] != "ApexClass:Orphan" {
		t.Fatalf("members_missing = %v", got)
	}
}

func TestVerifyCheckOnlyFails(t *testing.T) {
	dir, cs := vfyProject(t, []vfyMember{{"ApexClass", "Greeter"}, {"ApexClass", "GreeterTest"}}, nil, nil)
	res, _, err := vfyRun(t, dir, cs, vfyCase{
		id: vfyDeployCheckOnl, testLevel: "RunLocalTests",
		from: "2026-11-08T18:41:05.000+0000", to: "2026-11-08T18:41:20.000+0000",
		replies: map[string]fakeReply{
			vfyKeyRequest(vfyDeployCheckOnl): vfyReply(t, "tooling-deployrequest-checkonly.json"),
			vfyKeyReport(vfyDeployCheckOnl):  vfyReply(t, "deploy-validate-for-checkonly.json"),
			vfyKeyWindow:                     vfyReply(t, "tooling-deployrequest-window.json"),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Passed {
		t.Fatal("verify passed a check-only deploy")
	}
	vfyWant(t, res, CheckVerifyCheckOnly, providers.VerifyFail, "was check-only")
	if res.Facts["verify.check_only"] != true {
		t.Fatalf("check_only fact = %v", res.Facts["verify.check_only"])
	}
}

func TestVerifyTestsNotRunWhenRequiredFails(t *testing.T) {
	dir, cs := vfyProject(t, []vfyMember{{"ApexClass", "Greeter"}}, nil, nil)
	c := vfyCase{
		id:   vfyDeployNoTests,
		from: "2026-11-08T18:42:42.000+0000", to: "2026-11-08T18:42:50.000+0000",
		replies: map[string]fakeReply{
			vfyKeyRequest(vfyDeployNoTests): vfyReply(t, "tooling-deployrequest-notestrun.json"),
			vfyKeyReport(vfyDeployNoTests):  vfyReply(t, "deploy-start-notestrun.json"),
			vfyKeyWindow:                    vfyReply(t, "tooling-deployrequest-window.json"),
		},
	}
	for _, tc := range []struct {
		level  string
		quick  bool
		want   providers.VerifyStatus
		passed bool
	}{
		{"RunLocalTests", false, providers.VerifyFail, false},
		{"RunSpecifiedTests", false, providers.VerifyFail, false},
		{"NoTestRun", false, providers.VerifySkipped, true},
		{"", false, providers.VerifySkipped, true},
		{"RunLocalTests", true, providers.VerifySkipped, true},
	} {
		t.Run(tc.level, func(t *testing.T) {
			c := c
			c.testLevel, c.quick = tc.level, tc.quick
			res, _, err := vfyRun(t, dir, cs, c)
			if err != nil {
				t.Fatal(err)
			}
			vfyWant(t, res, CheckVerifyTests, tc.want)
			if res.Passed != tc.passed {
				t.Fatalf("passed = %v, checks %+v", res.Passed, res.Checks)
			}
			if tc.want == providers.VerifyFail {
				vfyWant(t, res, CheckVerifyTests, tc.want, "tests did not run")
			}
		})
	}
}

func TestVerifyForeignDeployInWindow(t *testing.T) {
	// The window runs on past the two later deploys in the capture: the Draft Flow deploy
	// and the Orphan deploy.
	dir, cs := vfyProject(t, vfyPassMembers, nil, map[string]string{"Glassops_Probe": "Active"})
	for _, tc := range []struct {
		mode   string
		want   providers.VerifyStatus
		passed bool
	}{
		{"", providers.VerifyWarn, true},
		{config.VerifyWarn, providers.VerifyWarn, true},
		{config.VerifyFail, providers.VerifyFail, false},
		{config.VerifyOff, providers.VerifySkipped, true},
	} {
		t.Run("mode "+tc.mode, func(t *testing.T) {
			c := vfyPassCase(t)
			c.to = "2026-11-08T18:35:00.000+0000"
			c.foreign = tc.mode
			res, f, err := vfyRun(t, dir, cs, c)
			if err != nil {
				t.Fatal(err)
			}
			vfyWant(t, res, CheckVerifyForeign, tc.want)
			if res.Passed != tc.passed {
				t.Fatalf("passed = %v, checks %+v", res.Passed, res.Checks)
			}
			if tc.mode == config.VerifyOff {
				for _, cmd := range f.commands() {
					if strings.HasPrefix(cmd, vfyKeyWindow) {
						t.Fatal("verify.foreign_deploys off still ran the window query")
					}
				}
				return
			}
			vfyWant(t, res, CheckVerifyForeign, tc.want, "2 other deploys", "0Af000000000005AAA", "0Af000000000006AAA")
			// The capture's first row was created before the window and is not counted.
			_, d := vfyStatus(res, CheckVerifyForeign)
			if strings.Contains(d, "0Af000000000002AAA") || strings.Contains(d, vfyDeployPass) {
				t.Fatalf("detail counts a deploy outside the window or our own: %s", d)
			}
			if res.Facts["verify.foreign_deploys_count"] != 2 {
				t.Fatalf("count = %v", res.Facts["verify.foreign_deploys_count"])
			}
		})
	}
}

func TestVerifyFailedDeployWithAllSuccessesFails(t *testing.T) {
	// Every component is in componentSuccesses and numberComponentErrors is 0, but the deploy
	// is Failed on org-wide coverage and nothing was deployed.
	dir, cs := vfyProject(t, []vfyMember{{"ApexClass", "Greeter"}, {"ApexClass", "GreeterTest"}, {"ApexClass", "Orphan"}}, nil, nil)
	res, _, err := vfyRun(t, dir, cs, vfyCase{
		id: vfyDeployCovFail, testLevel: "RunLocalTests",
		from: "2026-11-08T18:32:26.000+0000", to: "2026-11-08T18:32:40.000+0000",
		replies: map[string]fakeReply{
			vfyKeyRequest(vfyDeployCovFail): vfyReply(t, "tooling-deployrequest-coverage-fail.json"),
			vfyKeyReport(vfyDeployCovFail):  vfyReply(t, "deploy-report-coverage-fail.json"),
			vfyKeyWindow:                    vfyReply(t, "tooling-deployrequest-window.json"),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Passed {
		t.Fatal("verify passed a Failed deploy")
	}
	vfyWant(t, res, CheckVerifyStatus, providers.VerifyFail, "is Failed in the target", "50%")
	// The listing alone would have passed. The status check is what catches it.
	vfyWant(t, res, CheckVerifyMembers, providers.VerifyPass)
}

func TestVerifyDestructiveMembersOnQuickReport(t *testing.T) {
	env, err := parseEnvelope(fixture(t, "deploy-report-quick.json"))
	if err != nil {
		t.Fatal(err)
	}
	var rep reportBody
	if err := json.Unmarshal(env.Result, &rep); err != nil {
		t.Fatal(err)
	}
	pkg := manifest{Types: []manifestType{{Name: "ApexClass", Members: []string{"Greeter", "GreeterTest"}}, {Name: "PermissionSet", Members: []string{"Greeter_User"}}}}
	for _, tc := range []struct {
		name string
		dest []manifest
		pkg  manifest
		want providers.VerifyStatus
	}{
		{"deleted member found", []manifest{{Types: []manifestType{{Name: "CustomField", Members: []string{"Account.Tier__c"}}}}}, pkg, providers.VerifyPass},
		{"destructive member not deleted", []manifest{{Types: []manifestType{{Name: "CustomField", Members: []string{"Account.Other__c"}}}}}, pkg, providers.VerifyFail},
		{"deleted is not deployed", nil, manifest{Types: []manifestType{{Name: "CustomField", Members: []string{"Account.Tier__c"}}}}, providers.VerifyFail},
		{"wildcard matched by type", nil, manifest{Types: []manifestType{{Name: "ApexClass", Members: []string{"*"}}}}, providers.VerifyPass},
		{"wildcard with no component", nil, manifest{Types: []manifestType{{Name: "Flow", Members: []string{"*"}}}}, providers.VerifyFail},
		{"names match without case", nil, manifest{Types: []manifestType{{Name: "ApexClass", Members: []string{"greeter"}}}}, providers.VerifyPass},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v := &verifyRun{res: providers.VerifyResult{Facts: map[string]any{}}}
			v.membersCheck(deployRequestRow{}, rep, tc.pkg, tc.dest)
			vfyWant(t, v.res, CheckVerifyMembers, tc.want)
		})
	}
}

func TestVerifyErrors(t *testing.T) {
	dir, cs := vfyProject(t, vfyPassMembers, nil, map[string]string{"Glassops_Probe": "Active"})
	cases := []struct {
		name   string
		edit   func(c *vfyCase, cs *providers.ChangeSet, dir string)
		want   string
		noCall bool
	}{
		{"deploy Id that is not one", func(c *vfyCase, _ *providers.ChangeSet, _ string) { c.id = "0Af' OR Id != '" }, "is not a deploy Id", true},
		{"empty deploy Id", func(c *vfyCase, _ *providers.ChangeSet, _ string) { c.id = "" }, "is not a deploy Id", true},
		{"plan files changed", func(_ *vfyCase, _ *providers.ChangeSet, dir string) {
			_ = os.WriteFile(filepath.Join(dir, ".glassops", "plan", "package", "package.xml"), []byte("<Package/>"), 0o644)
		}, ErrPlanChanged.Error(), true},
		{"DeployRequest query fails", func(c *vfyCase, _ *providers.ChangeSet, _ string) {
			c.replies[vfyKeyRequest(vfyDeployPass)] = fakeReply{Stdout: []byte(`{"status":1,"name":"INVALID_FIELD","message":"No such column"}`), ExitCode: 1}
		}, "read DeployRequest", false},
		{"no DeployRequest with our Id", func(c *vfyCase, _ *providers.ChangeSet, _ string) {
			c.replies[vfyKeyRequest(vfyDeployPass)] = vfyReply(t, "tooling-deployrequest-notestrun.json")
		}, "has no DeployRequest", false},
		{"report for another deploy", func(c *vfyCase, _ *providers.ChangeSet, _ string) {
			c.replies[vfyKeyReport(vfyDeployPass)] = vfyReply(t, "deploy-report-coverage-fail.json")
		}, "answered for", false},
		{"report with no body", func(c *vfyCase, _ *providers.ChangeSet, _ string) {
			c.replies[vfyKeyReport(vfyDeployPass)] = fakeReply{Stdout: []byte(`{"status":1,"name":"NotFound","message":"no deploy"}`), ExitCode: 1}
		}, "read deploy report", false},
		{"window query fails", func(c *vfyCase, _ *providers.ChangeSet, _ string) {
			c.replies[vfyKeyWindow] = fakeReply{Stdout: []byte(`{"status":1,"name":"MALFORMED_QUERY","message":"bad"}`), ExitCode: 1}
		}, "deploys in the window", false},
		{"window ending before it starts", func(c *vfyCase, _ *providers.ChangeSet, _ string) { c.to = "2026-11-08T18:00:00.000+0000" }, "deploy window", false},
		{"flow query fails", func(c *vfyCase, _ *providers.ChangeSet, _ string) {
			c.replies[vfyKeyFlowDef] = fakeReply{Stdout: []byte(`{"status":1,"name":"INVALID_TYPE","message":"bad"}`), ExitCode: 1}
		}, "read FlowDefinition", false},
		{"flow source missing", func(_ *vfyCase, _ *providers.ChangeSet, dir string) {
			_ = os.Remove(filepath.Join(dir, "force-app", "main", "default", "flows", "Glassops_Probe.flow-meta.xml"))
		}, "no Glassops_Probe.flow-meta.xml", false},
		{"partial query result", func(c *vfyCase, _ *providers.ChangeSet, _ string) {
			c.replies[vfyKeyWindow] = fakeReply{Stdout: []byte(`{"status":0,"result":{"records":[],"totalSize":3000,"done":false}}`)}
		}, "only part of its records", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir, cs := dir, cs
			if tc.name == "plan files changed" || tc.name == "flow source missing" {
				dir, cs = vfyProject(t, vfyPassMembers, nil, map[string]string{"Glassops_Probe": "Active"})
			}
			c := vfyPassCase(t)
			tc.edit(&c, &cs, dir)
			res, f, err := vfyRun(t, dir, cs, c)
			if err == nil {
				t.Fatalf("no error, result %+v", res)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error %q does not contain %q", err, tc.want)
			}
			if tc.name == "plan files changed" && !errors.Is(err, ErrPlanChanged) {
				t.Fatalf("error %v is not ErrPlanChanged", err)
			}
			if tc.noCall && len(f.commands()) > 0 {
				t.Fatalf("verify called sf: %v", f.commands())
			}
		})
	}
}

func TestVerifyNeedsOwnSession(t *testing.T) {
	p := newTestProvider(t, newFakeExec(nil), nil)
	if _, err := p.Verify(context.Background(), vfyOtherSession{}, providers.VerifyRequest{}); err == nil {
		t.Fatal("verify accepted a session from another provider")
	}
}

type vfyOtherSession struct{}

func (vfyOtherSession) Target() providers.Target { return providers.Target{} }
func (vfyOtherSession) Close() error             { return nil }

func TestVerifyQueriesQuoteOnlyCheckedValues(t *testing.T) {
	for _, bad := range []string{"0Af00000000000'", "0Af000000000003AAA OR", "001000000000003AAA", "0Af0000000000"} {
		if fullDeployIDPattern.MatchString(bad) {
			t.Errorf("%q passed as a deploy Id", bad)
		}
	}
	for _, bad := range []string{"a'b", "Flow Name", "1Flow", ""} {
		if flowNamePattern.MatchString(bad) {
			t.Errorf("%q passed as a Flow name", bad)
		}
	}
	if q := flowDefinitionQuery([]string{"A", "B_c"}); !strings.HasSuffix(q, "WHERE DeveloperName IN ('A', 'B_c')") {
		t.Fatalf("query = %s", q)
	}
}

func TestVerifyWindowSkipsCheckOnlyRows(t *testing.T) {
	// The window query leaves check-only deploys out. If one came back anyway, here the
	// captured validate 0Af...8, it is still not counted.
	dir, cs := vfyProject(t, vfyPassMembers, nil, map[string]string{"Glassops_Probe": "Active"})
	c := vfyPassCase(t)
	c.replies[vfyKeyWindow] = vfyReply(t, "tooling-deployrequest-checkonly.json")
	res, _, err := vfyRun(t, dir, cs, c)
	if err != nil {
		t.Fatal(err)
	}
	vfyWant(t, res, CheckVerifyForeign, providers.VerifyPass)
}

func TestVerifyTestsEnabledButNoneRan(t *testing.T) {
	// Synthetic: no capture has tests enabled with a total of 0.
	yes := true
	zero := flexInt(0)
	synthetic := deployRequestRow{ID: vfyDeployPass, RunTestsEnabled: &yes, NumberTestsTotal: &zero, NumberTestsCompleted: &zero, NumberTestErrors: &zero}
	v := &verifyRun{res: providers.VerifyResult{Facts: map[string]any{}}}
	v.testsCheck(synthetic, providers.DeployRecord{TestLevel: "RunLocalTests"}, true)
	vfyWant(t, v.res, CheckVerifyTests, providers.VerifyFail, "no tests ran")
}

// vfyFolderReport reads the real full-source deploy report and manifest, the two fixtures that
// show folder and Settings naming.
func vfyFolderReport(t *testing.T) (reportBody, manifest) {
	t.Helper()
	env, err := parseEnvelope(fixture(t, "deploy-report-folders-settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	var rep reportBody
	if err := json.Unmarshal(env.Result, &rep); err != nil {
		t.Fatal(err)
	}
	pkg, err := readManifest(filepath.Join("testdata", "sf", "package-folders-settings.xml"))
	if err != nil {
		t.Fatal(err)
	}
	return rep, pkg
}

func TestVerifyMembersMatchRealFolderAndSettingsNaming(t *testing.T) {
	rep, pkg := vfyFolderReport(t)
	n := 0
	for _, ty := range pkg.Types {
		n += len(ty.Members)
	}
	if n != 18 { // the 19 successes include package.xml itself
		t.Fatalf("manifest has %d members, the capture has 18", n)
	}
	v := &verifyRun{res: providers.VerifyResult{Facts: map[string]any{}}}
	v.membersCheck(deployRequestRow{}, rep, pkg, nil)
	vfyWant(t, v.res, CheckVerifyMembers, providers.VerifyPass, "all 18 plan members")
}

func TestVerifyMembersRealNamingNegatives(t *testing.T) {
	rep, _ := vfyFolderReport(t)
	// without returns the successes minus the ones whose type and name match.
	without := func(typ, name string) reportBody {
		out := rep
		out.Details.ComponentSuccesses = nil
		for _, c := range rep.Details.ComponentSuccesses {
			if c.ComponentType == typ && c.FullName == name {
				continue
			}
			out.Details.ComponentSuccesses = append(out.Details.ComponentSuccesses, c)
		}
		if len(out.Details.ComponentSuccesses) == len(rep.Details.ComponentSuccesses) {
			t.Fatalf("no success %s %s to remove", typ, name)
		}
		return out
	}
	one := func(typ string, members ...string) manifest {
		return manifest{Types: []manifestType{{Name: typ, Members: members}}}
	}
	for _, tc := range []struct {
		name string
		rep  reportBody
		pkg  manifest
		want providers.VerifyStatus
	}{
		{"report folder deployed", rep, one("Report", "Portfolio_Owner_Reports/"), providers.VerifyPass},
		{"report folder missing", without("ReportFolder", "Portfolio_Owner_Reports"), one("Report", "Portfolio_Owner_Reports/"), providers.VerifyFail},
		{"dashboard folder missing", without("DashboardFolder", "Portfolio_Owner_Dashboards"), one("Dashboard", "Portfolio_Owner_Dashboards/"), providers.VerifyFail},
		{"email folder missing", without("EmailFolder", "Portfolio_Contact"), one("EmailTemplate", "Portfolio_Contact"), providers.VerifyFail},
		{"email template named like a folder that was not deployed", rep, one("EmailTemplate", "Portfolio_Nothing"), providers.VerifyFail},
		{"settings with no matching success", without("EmailAdministrationSettings", "EmailAdministration"), one("Settings", "EmailAdministration"), providers.VerifyFail},
		{"settings that was never deployed", rep, one("Settings", "Mobile"), providers.VerifyFail},
		{"folder name with the wrong type", rep, one("Dashboard", "Portfolio_Owner_Reports/"), providers.VerifyFail},
		{"slash-less report is not its folder", rep, one("Report", "Portfolio_Owner_Reports"), providers.VerifyFail},
		{"slash-less dashboard is not its folder", rep, one("Dashboard", "Portfolio_Owner_Dashboards"), providers.VerifyFail},
		{"email folder name under another type", rep, one("GenAiPlugin", "Portfolio_Testimonial"), providers.VerifyFail},
		{"settings name matches only the Settings suffix type", rep, one("EmailAdministrationSettings", "EmailAdministration"), providers.VerifyPass},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v := &verifyRun{res: providers.VerifyResult{Facts: map[string]any{}}}
			v.membersCheck(deployRequestRow{}, tc.rep, tc.pkg, nil)
			vfyWant(t, v.res, CheckVerifyMembers, tc.want)
		})
	}
}

func TestReportedAs(t *testing.T) {
	for _, tc := range []struct{ typ, name, wantTyp, wantName string }{
		{"Dashboard", "F/", "DashboardFolder", "F"},
		{"Report", "F/", "ReportFolder", "F"},
		{"Document", "F/", "DocumentFolder", "F"}, // [unverified]
		{"Dashboard", "F/D", "Dashboard", "F/D"},
		{"Dashboard", "/", "Dashboard", "/"},
		{"EmailTemplate", "F", "EmailFolder", "F"},
		{"EmailTemplate", "F/T", "EmailTemplate", "F/T"},
		{"EmailTemplate", "*", "EmailTemplate", "*"},
		{"Settings", "Mobile", "MobileSettings", "Mobile"},
		{"ApexClass", "A", "ApexClass", "A"},
	} {
		gt, gn := reportedAs(tc.typ, tc.name)
		if gt != tc.wantTyp || gn != tc.wantName {
			t.Errorf("reportedAs(%s, %s) = %s, %s, want %s, %s", tc.typ, tc.name, gt, gn, tc.wantTyp, tc.wantName)
		}
	}
}

func TestVerifyRelevantTestsWithNoApex(t *testing.T) {
	// Synthetic: no RunRelevantTests deploy was captured.
	yes, no := true, false
	zero, three, one := flexInt(0), flexInt(3), flexInt(1)
	for _, tc := range []struct {
		name    string
		row     deployRequestRow
		hasApex bool
		want    providers.VerifyStatus
		detail  string
	}{
		{"no apex, none ran", deployRequestRow{RunTestsEnabled: &yes, NumberTestsTotal: &zero, NumberTestsCompleted: &zero, NumberTestErrors: &zero}, false, providers.VerifySkipped, "plan has no Apex"},
		{"no apex, tests off", deployRequestRow{RunTestsEnabled: &no, NumberTestsTotal: &zero, NumberTestsCompleted: &zero, NumberTestErrors: &zero}, false, providers.VerifySkipped, "plan has no Apex"},
		{"apex, none ran", deployRequestRow{RunTestsEnabled: &yes, NumberTestsTotal: &zero, NumberTestsCompleted: &zero, NumberTestErrors: &zero}, true, providers.VerifyFail, "no tests ran"},
		{"apex, tests off", deployRequestRow{RunTestsEnabled: &no, NumberTestsTotal: &zero, NumberTestsCompleted: &zero, NumberTestErrors: &zero}, true, providers.VerifyFail, "tests did not run"},
		{"no apex, critical test failed", deployRequestRow{RunTestsEnabled: &yes, NumberTestsTotal: &three, NumberTestsCompleted: &three, NumberTestErrors: &one}, false, providers.VerifyFail, "1 of 3 tests failed"},
		{"no apex, counts unknown", deployRequestRow{RunTestsEnabled: &yes}, false, providers.VerifyFail, "does not say"},
		{"apex, tests passed", deployRequestRow{RunTestsEnabled: &yes, NumberTestsTotal: &three, NumberTestsCompleted: &three, NumberTestErrors: &zero}, true, providers.VerifyPass, "3 of 3 tests ran"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.row.ID = vfyDeployPass
			v := &verifyRun{res: providers.VerifyResult{Facts: map[string]any{}}}
			v.testsCheck(tc.row, providers.DeployRecord{TestLevel: "RunRelevantTests"}, tc.hasApex)
			vfyWant(t, v.res, CheckVerifyTests, tc.want, tc.detail)
		})
	}
	// Other levels keep requiring tests on a plan with no Apex.
	v := &verifyRun{res: providers.VerifyResult{Facts: map[string]any{}}}
	v.testsCheck(deployRequestRow{ID: vfyDeployPass, RunTestsEnabled: &yes, NumberTestsTotal: &zero, NumberTestsCompleted: &zero, NumberTestErrors: &zero}, providers.DeployRecord{TestLevel: "RunLocalTests"}, false)
	vfyWant(t, v.res, CheckVerifyTests, providers.VerifyFail, "no tests ran")
}

func TestHasApex(t *testing.T) {
	cls := manifest{Types: []manifestType{{Name: "ApexClass", Members: []string{"Greeter"}}}}
	trg := manifest{Types: []manifestType{{Name: "ApexTrigger", Members: []string{"AccountTrigger"}}}}
	flow := manifest{Types: []manifestType{{Name: "Flow", Members: []string{"Probe"}}}}
	if hasApex(flow, nil) {
		t.Fatal("a Flow only plan has no Apex")
	}
	if !hasApex(cls, nil) || !hasApex(flow, []manifest{trg}) {
		t.Fatal("an Apex class in the plan or a trigger in a destructive manifest counts as Apex")
	}
}

// vfyRequestRow reads the one DeployRequest row of a Tooling capture.
func vfyRequestRow(t *testing.T, name string) deployRequestRow {
	t.Helper()
	env, err := parseEnvelope(fixture(t, name))
	if err != nil {
		t.Fatal(err)
	}
	var q struct {
		Records []deployRequestRow `json:"records"`
	}
	if err := json.Unmarshal(env.Result, &q); err != nil || len(q.Records) != 1 {
		t.Fatalf("%s: %v, %d records", name, err, len(q.Records))
	}
	return q.Records[0]
}

func TestVerifyTestsOnProductionCaptures(t *testing.T) {
	noApex := vfyRequestRow(t, "tooling-deployrequest-relevant-noapex.json")
	apex := vfyRequestRow(t, "tooling-deployrequest-relevant-apex.json")
	quick := vfyRequestRow(t, "tooling-deployrequest-quick.json")
	for _, tc := range []struct {
		name    string
		row     deployRequestRow
		deploy  providers.DeployRecord
		hasApex bool
		want    providers.VerifyStatus
		detail  string
	}{
		// Production reports RunTestsEnabled true and 0 tests when the change has no Apex.
		{"relevant, no Apex", noApex, providers.DeployRecord{TestLevel: "RunRelevantTests"}, false, providers.VerifySkipped, "plan has no Apex"},
		{"relevant, same row with Apex in the plan", noApex, providers.DeployRecord{TestLevel: "RunRelevantTests"}, true, providers.VerifyFail, "no tests ran"},
		{"local, same row", noApex, providers.DeployRecord{TestLevel: "RunLocalTests"}, false, providers.VerifyFail, "no tests ran"},
		{"relevant, Apex", apex, providers.DeployRecord{TestLevel: "RunRelevantTests"}, true, providers.VerifyPass, "272 of 272 tests ran"},
		// A quick deploy reports RunTestsEnabled false and 0 tests, so only the quick skip keeps it from failing.
		{"quick deploy", quick, providers.DeployRecord{TestLevel: "RunRelevantTests", Quick: true}, true, providers.VerifySkipped, "reuses the tests"},
		{"quick deploy read as a full deploy", quick, providers.DeployRecord{TestLevel: "RunRelevantTests"}, true, providers.VerifyFail, "tests did not run"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v := &verifyRun{res: providers.VerifyResult{Facts: map[string]any{}}}
			v.testsCheck(tc.row, tc.deploy, tc.hasApex)
			vfyWant(t, v.res, CheckVerifyTests, tc.want, tc.detail)
		})
	}
}
