package salesforce

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/providers"
)

func TestVersionLess(t *testing.T) {
	cases := []struct {
		a, b string
		less bool
	}{
		{"8.20.1", "8.20.1", false},
		{"8.20.0", "8.20.1", true},
		{"8.20.1", "8.20.0", false},
		{"8.9.9", "8.20.0", true},
		{"7.99.99", "8", true},
		{"8", "8.0.0", false},
		{"8.20", "8.20.1", true},
		{"v8.21.0", "8.20.1", false},
		{"8.20.1-beta.1", "8.20.1", true},
		{"8.20.1", "8.20.1-beta.1", false},
		{"8.20.1-alpha", "8.20.1-beta", true},
		{"8.20.1+build.5", "8.20.1", false},
	}
	for _, tc := range cases {
		got, err := versionLess(tc.a, tc.b)
		if err != nil {
			t.Fatalf("versionLess(%q, %q): %v", tc.a, tc.b, err)
		}
		if got != tc.less {
			t.Errorf("versionLess(%q, %q) = %v, want %v", tc.a, tc.b, got, tc.less)
		}
	}
	for _, bad := range []string{"", "latest", "8.x", "1.2.3.4", "-1.0", "8..1"} {
		if _, err := versionLess(bad, "8.20.1"); err == nil {
			t.Errorf("versionLess(%q) accepted a non-version", bad)
		}
	}
}

// hardisFixture returns testdata/hardis/<name>, real sfdx-hardis 8.20.1 output.
func hardisFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "hardis", name))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// hardisPluginsWith is the real plugins capture with another sfdx-hardis version. Synthetic.
func hardisPluginsWith(t *testing.T, version string) []byte {
	t.Helper()
	return []byte(strings.Replace(string(hardisFixture(t, "sf-plugins-hardis.json")), `"version": "8.20.1"`, `"version": "`+version+`"`, 1))
}

func TestCheckHardisPlugins(t *testing.T) {
	got, err := CheckHardisPlugins(hardisFixture(t, "sf-plugins-hardis.json"), "")
	if err != nil || got.Version != "8.20.1" || got.Warning != "" {
		t.Fatalf("tested version: %+v, %v", got, err)
	}

	got, err = CheckHardisPlugins(hardisPluginsWith(t, "8.23.0"), "8.20")
	if err != nil || !strings.Contains(got.Warning, "newer than the tested version") {
		t.Fatalf("newer: %+v, %v", got, err)
	}

	_, err = CheckHardisPlugins(hardisFixture(t, "sf-plugins-hardis.json"), "8.21.0")
	if !errors.Is(err, ErrHardisTooOld) || !strings.Contains(err.Error(), HardisInstallSteps) {
		t.Fatalf("too old: %v", err)
	}

	wrapped := `{"status":0,"result":[{"name":"@salesforce/cli","version":"2.155.7"},{"name":"sfdx-git-delta","version":"6.45.1"}]}`
	_, err = CheckHardisPlugins([]byte(wrapped), "")
	if !errors.Is(err, ErrHardisMissing) || !strings.Contains(err.Error(), "sf plugins install sfdx-hardis@"+HardisTestedVersion) {
		t.Fatalf("missing: %v", err)
	}

	if _, err := CheckHardisPlugins([]byte("not json"), ""); err == nil || errors.Is(err, ErrHardisMissing) {
		t.Fatalf("unreadable output: %v", err)
	}
}

// TestReadHardisOutput reads every captured sfdx-hardis output shape.
func TestReadHardisOutput(t *testing.T) {
	cases := []struct {
		file     string
		exit     int
		bodyID   string // the sf result read straight from outputstring
		errName  string // the *SFError name, when the run failed
		dataID   string // the deploy Id the failure carries for deploy report
		contains string // in the error message
	}{
		{file: "hardis-start-pass.json", bodyID: "0Af000000000002AAA"},
		{file: "hardis-start-destructive.json", bodyID: "0Af000000000001AAA"},
		{file: "hardis-start-with-commands.json", bodyID: "0Af000000000005AAA"},
		{file: "hardis-validate-pass.json", bodyID: "0Af000000000007AAA"},
		{file: "hardis-start-testfail.json", exit: 1, errName: "HardisDeployFailed", dataID: "0Af000000000003AAA", contains: "Apex tests: 0 completed, 1 failed"},
		{file: "hardis-validate-fail.json", exit: 1, errName: "HardisDeployFailed", dataID: "0Af000000000006AAA", contains: "Components: 0 deployed, 5 failed"},
		{file: "hardis-start-timeout.json", exit: 1, errName: "ClientTimeoutError", dataID: "0Af000000000004AAA", contains: "timed out"},
		{file: "hardis-agent-flag-passed-through.json", exit: 2, errName: "Error", contains: "Nonexistent flag: --agent"},
		{file: "hardis-quick-pass.json", exit: 2, errName: "Error", contains: "Nonexistent flag: -i"},
		{file: "hardis-quick-jobid-flag.json", exit: 2, errName: "Error", contains: "Nonexistent flag: --job-id"},
	}
	for _, tc := range cases {
		t.Run(tc.file, func(t *testing.T) {
			// sfdx-hardis prints a line before its JSON. The reader must skip it.
			out := append([]byte("WS Client started\n"), hardisFixture(t, tc.file)...)
			env, err := readHardisOutput(hardisDeployCommand, out, tc.exit)
			if tc.errName == "" {
				if err != nil {
					t.Fatal(err)
				}
				body, ok := parseDeployBody(env.Result)
				if !ok || body.ID != tc.bodyID || body.Status != "Succeeded" {
					t.Fatalf("body %+v, ok %v", body, ok)
				}
				return
			}
			var sfErr *SFError
			if !errors.As(err, &sfErr) {
				t.Fatalf("want *SFError, got %v", err)
			}
			if sfErr.Name != tc.errName || sfErr.Status == 0 || !strings.Contains(sfErr.Error(), tc.contains) {
				t.Fatalf("error %+v", sfErr)
			}
			if got := failureDeployID(err); got != tc.dataID {
				t.Fatalf("deploy Id %q, want %q", got, tc.dataID)
			}
			if strings.Contains(sfErr.Message, "at SfCommandError") {
				t.Fatalf("message carries the stack: %q", sfErr.Message)
			}
		})
	}
}

func TestReadHardisOutputRefusesOddShapes(t *testing.T) {
	for name, out := range map[string]string{
		"no JSON":                     "hardis nosuch is not a sf command.",
		"not JSON":                    "{not json",
		"no outputstring, status 0":   `{"status":0,"result":{}}`,
		"failure around a pass":       `{"status":1,"result":{"outputstring":{"status":0,"result":{"id":"0Af000000000001AAA","status":"Succeeded"}}}}`,
		"outputstring a number":       `{"status":1,"result":{"outputstring":42}}`,
		"text with no Id and no JSON": `{"status":1,"result":{"outputstring":"No deployment result available to summarize\n"}}`,
	} {
		t.Run(name, func(t *testing.T) {
			env, err := readHardisOutput(hardisDeployCommand, []byte(out), 1)
			if err == nil {
				t.Fatalf("accepted %q as %+v", out, env)
			}
			if failureDeployID(err) != "" {
				t.Fatalf("an Id came from %q", out)
			}
		})
	}
}

const (
	hardisStartPrefix   = "sf " + hardisDeployCommand
	hardisPluginsPrefix = "sf plugins --json"
)

// hardisSetup is deploySetup with the hardis engine selected and sfdx-hardis 8.20.1 listed.
func hardisSetup(t *testing.T, replies map[string]fakeReply, minVersion string) (Engine, *fakeExec, *session, string, providers.ChangeSet) {
	t.Helper()
	p, f, s, dir, cs := deploySetup(t, nil)
	if _, ok := replies[hardisPluginsPrefix]; !ok {
		replies[hardisPluginsPrefix] = fakeReply{Stdout: hardisFixture(t, "sf-plugins-hardis.json")}
	}
	f.replies = replies
	if err := p.Configure(&config.Config{Salesforce: config.Salesforce{Engine: config.EngineHardis, Hardis: config.Hardis{MinVersion: minVersion}}}); err != nil {
		t.Fatal(err)
	}
	if p.Engine().Name() != config.EngineHardis {
		t.Fatalf("engine %s", p.Engine().Name())
	}
	return p.Engine(), f, s, dir, cs
}

// noHardisWrappers fails when the engine sent sf a hardis command other than deploy start.
// The 8.20.1 validate wrapper deploys for real and the quick wrapper does not work.
func noHardisWrappers(t *testing.T, f *fakeExec) {
	t.Helper()
	for _, c := range f.commands() {
		for _, never := range []string{"hardis:project:deploy:validate", "hardis:project:deploy:quick", "hardis:project:deploy:smart", "hardis:deploy", " --agent", "--dry-run"} {
			if strings.Contains(c, never) {
				t.Fatalf("engine sent %q: %s", never, c)
			}
		}
	}
}

func TestHardisDeployPass(t *testing.T) {
	e, f, s, dir, cs := hardisSetup(t, map[string]fakeReply{
		queryPrefix:       sandboxOrg(t),
		hardisStartPrefix: {Stdout: hardisFixture(t, "hardis-start-pass.json")},
	}, "")
	t.Setenv("GITHUB_TOKEN", "ghs_should_not_reach_hardis")
	res, err := e.Deploy(context.Background(), s, cs, providers.DeployOptions{Dir: dir, TestLevel: "RunSpecifiedTests", Tests: []string{"GreeterTest"}})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Passed || res.Quick || res.TimedOut || res.ID != "0Af000000000002AAA" || res.ComponentsDeployed != 2 || len(res.Problems) != 0 {
		t.Fatalf("result: %+v", res)
	}
	for k, v := range map[string]any{
		"salesforce.engine":              "hardis",
		"salesforce.hardis.version":      "8.20.1",
		"salesforce.hardis.deploy_path":  "hardis",
		"salesforce.deploy_status":       "Succeeded",
		"salesforce.check_only":          false,
		"salesforce.components_deployed": 2,
	} {
		if res.Facts[k] != v {
			t.Errorf("fact %s = %v, want %v", k, res.Facts[k], v)
		}
	}

	cmds := f.commands()
	if len(cmds) != 3 || cmds[0] != hardisPluginsPrefix || !strings.HasPrefix(cmds[1], queryPrefix) || !strings.HasPrefix(cmds[2], hardisStartPrefix) {
		t.Fatalf("commands: %v", cmds)
	}
	manifest := filepath.Join(dir, ".glassops", "plan", "package", "package.xml")
	destr := filepath.Join(dir, ".glassops", "plan", "destructiveChanges", "destructiveChanges.xml")
	for _, frag := range []string{
		"--manifest " + manifest, "--post-destructive-changes " + destr, "--target-org glassops-target",
		"--wait 60", "--test-level RunSpecifiedTests", "--tests GreeterTest", "--json",
	} {
		if !strings.Contains(cmds[2], frag) {
			t.Errorf("hardis deploy lacks %q: %s", frag, cmds[2])
		}
	}
	for _, never := range []string{"--ignore-errors", "--ignore-warnings", "--ignore-conflicts", "--delta", "--packagexml"} {
		if strings.Contains(cmds[2], never) {
			t.Errorf("hardis deploy has %s: %s", never, cmds[2])
		}
	}
	noHardisWrappers(t, f)

	// The pull request tokens are blank for the hardis process, whatever the process env holds.
	start := f.calls[2]
	merged := mergeEnv(os.Environ(), start.Env)
	for _, k := range hardisBlankedEnv {
		found := false
		for _, kv := range merged {
			name, v, _ := strings.Cut(kv, "=")
			if strings.EqualFold(name, k) {
				if v != "" {
					t.Errorf("hardis gets %s with a value", k)
				}
				found = true
			}
		}
		if !found {
			t.Errorf("%s is not blanked for hardis", k)
		}
	}
	// Plain sf calls keep the process env as before.
	for _, kv := range f.calls[1].Env {
		if strings.HasPrefix(kv, "GITHUB_TOKEN=") {
			t.Errorf("the org query got %s", kv)
		}
	}
}

func TestHardisDeployDestructiveAndScratch(t *testing.T) {
	e, f, s, dir, cs := hardisSetup(t, map[string]fakeReply{
		queryPrefix:       {Stdout: fixture(t, "query-organization-scratch.json")},
		hardisStartPrefix: {Stdout: hardisFixture(t, "hardis-start-destructive.json")},
	}, "")
	cs.Destructive = []string{".glassops/plan/destructiveChanges/destructiveChangesPre.xml"}
	pre := filepath.Join(dir, ".glassops", "plan", "destructiveChanges", "destructiveChangesPre.xml")
	if err := os.WriteFile(pre, hardisFixtureOrEmpty(t), 0o644); err != nil {
		t.Fatal(err)
	}
	sum, err := planDigest(dir, filepath.Join(dir, ".glassops", "plan"))
	if err != nil {
		t.Fatal(err)
	}
	cs.SHA256 = sum
	res, err := e.Deploy(context.Background(), s, cs, providers.DeployOptions{Dir: dir})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Passed || res.ComponentsDeployed != 3 {
		t.Fatalf("result: %+v", res)
	}
	start := f.commands()[2]
	if !strings.Contains(start, "--pre-destructive-changes "+pre) || !strings.Contains(start, "--ignore-conflicts") {
		t.Fatalf("hardis deploy: %s", start)
	}
	noHardisWrappers(t, f)
}

// hardisFixtureOrEmpty is a destructive manifest body for the pre-destructive test.
func hardisFixtureOrEmpty(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "sgd", "delta", "destructiveChanges", "destructiveChanges.xml"))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestHardisDeployFailureReadsReport(t *testing.T) {
	cases := []struct {
		name, out, report, id string
	}{
		// The compile failure comes from the validate wrapper, which ran deploy start underneath
		// (checkOnly false), so it is the deploy start failure shape too.
		{"compile failure", "hardis-validate-fail.json", "hardis-validate-fail-report.json", "0Af000000000006AAA"},
		{"test failure", "hardis-start-testfail.json", "hardis-validate-fail-report.json", "0Af000000000003AAA"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e, f, s, dir, cs := hardisSetup(t, map[string]fakeReply{
				queryPrefix:       sandboxOrg(t),
				hardisStartPrefix: {Stdout: hardisFixture(t, tc.out), ExitCode: 1},
				reportPrefix:      {Stdout: hardisFixture(t, tc.report)},
			}, "")
			res, err := e.Deploy(context.Background(), s, cs, providers.DeployOptions{Dir: dir})
			if err != nil {
				t.Fatal(err)
			}
			if res.Passed || len(res.Problems) == 0 {
				t.Fatalf("result: %+v", res)
			}
			cmds := f.commands()
			if last := cmds[len(cmds)-1]; !strings.HasPrefix(last, reportPrefix+" --job-id "+tc.id) {
				t.Fatalf("no deploy report on %s: %v", tc.id, cmds)
			}
			noHardisWrappers(t, f)
		})
	}

	// The compile failure maps its component failures to problem lines.
	e, _, s, dir, cs := hardisSetup(t, map[string]fakeReply{
		queryPrefix:       sandboxOrg(t),
		hardisStartPrefix: {Stdout: hardisFixture(t, "hardis-validate-fail.json"), ExitCode: 1},
		reportPrefix:      {Stdout: hardisFixture(t, "hardis-validate-fail-report.json")},
	}, "")
	res, _ := e.Deploy(context.Background(), s, cs, providers.DeployOptions{Dir: dir})
	if !strings.Contains(strings.Join(res.Problems, "\n"), "ApexClass Greeter") || res.Facts["salesforce.deploy_status"] != "Failed" {
		t.Fatalf("problems: %v facts: %v", res.Problems, res.Facts)
	}
}

// syntheticInProgressReport is a deploy report body still running. Synthetic, in the shape of
// the step 0.5 deploy-report-inprogress capture.
const syntheticInProgressReport = `{"status":0,"result":{"id":"0Af000000000004AAA","status":"InProgress","done":false,"checkOnly":false,"numberComponentsTotal":2,"numberComponentsDeployed":2,"details":{}}}`

func TestHardisDeployTimeout(t *testing.T) {
	e, f, s, dir, cs := hardisSetup(t, map[string]fakeReply{
		queryPrefix:       sandboxOrg(t),
		hardisStartPrefix: {Stdout: hardisFixture(t, "hardis-start-timeout.json"), ExitCode: 1},
		reportPrefix:      {Stdout: []byte(syntheticInProgressReport)},
	}, "")
	res, err := e.Deploy(context.Background(), s, cs, providers.DeployOptions{Dir: dir})
	if err != nil {
		t.Fatal(err)
	}
	if res.Passed || !res.TimedOut || res.ID != "0Af000000000004AAA" {
		t.Fatalf("result: %+v", res)
	}
	cmds := f.commands()
	if !strings.HasPrefix(cmds[len(cmds)-1], reportPrefix+" --job-id 0Af000000000004AAA") {
		t.Fatalf("commands: %v", cmds)
	}
}

func TestHardisDeployFlagRefusalIsError(t *testing.T) {
	e, _, s, dir, cs := hardisSetup(t, map[string]fakeReply{
		queryPrefix:       sandboxOrg(t),
		hardisStartPrefix: {Stdout: hardisFixture(t, "hardis-agent-flag-passed-through.json"), ExitCode: 2},
	}, "")
	if _, err := e.Deploy(context.Background(), s, cs, providers.DeployOptions{Dir: dir}); err == nil || !strings.Contains(err.Error(), "Nonexistent flag") {
		t.Fatalf("err = %v", err)
	}
}

func TestHardisQuickIsNative(t *testing.T) {
	e, f, s, dir, cs := hardisSetup(t, map[string]fakeReply{
		quickPrefix: {Stdout: fixture(t, "deploy-quick-pass.json")},
	}, "")
	res, err := e.Deploy(context.Background(), s, cs, providers.DeployOptions{Dir: dir, ValidatedID: validatedID})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Passed || !res.Quick || res.Facts["salesforce.hardis.quick_path"] != "native" || res.Facts["salesforce.engine"] != "hardis" {
		t.Fatalf("result: %+v", res)
	}
	cmds := f.commands()
	if len(cmds) != 2 || !strings.HasPrefix(cmds[1], quickPrefix+" --job-id "+validatedID) {
		t.Fatalf("commands: %v", cmds)
	}
	noHardisWrappers(t, f)
}

func TestHardisQuickRefusedFallsBackToHardis(t *testing.T) {
	e, f, s, dir, cs := hardisSetup(t, map[string]fakeReply{
		quickPrefix:       {Stdout: fixture(t, "deploy-quick-already-used.json"), ExitCode: 1},
		queryPrefix:       sandboxOrg(t),
		hardisStartPrefix: {Stdout: hardisFixture(t, "hardis-start-pass.json")},
	}, "")
	res, err := e.Deploy(context.Background(), s, cs, providers.DeployOptions{Dir: dir, ValidatedID: validatedID})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Passed || res.Quick || res.Facts["salesforce.quick_refused"] != true ||
		res.Facts["salesforce.hardis.fallback_path"] != "hardis" || res.Facts["salesforce.hardis.deploy_path"] != "hardis" {
		t.Fatalf("result: %+v", res)
	}
	reason, _ := res.Facts["salesforce.quick_fallback"].(string)
	if !strings.Contains(reason, "quick deploy of validation "+validatedID+" refused") || strings.Contains(reason, "no full deploy was run") {
		t.Fatalf("fallback reason %q", reason)
	}
	cmds := f.commands()
	if !strings.HasPrefix(cmds[len(cmds)-1], hardisStartPrefix) {
		t.Fatalf("commands: %v", cmds)
	}
	noHardisWrappers(t, f)
}

func TestHardisQuickOnlyRefusedRunsNoDeploy(t *testing.T) {
	e, f, s, dir, cs := hardisSetup(t, map[string]fakeReply{
		quickPrefix: {Stdout: fixture(t, "deploy-quick-already-used.json"), ExitCode: 1},
	}, "")
	res, err := e.Deploy(context.Background(), s, cs, providers.DeployOptions{Dir: dir, ValidatedID: validatedID, QuickOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	if res.Passed || len(res.Problems) != 1 || !strings.Contains(res.Problems[0], "no full deploy was run") {
		t.Fatalf("result: %+v", res)
	}
	for _, c := range f.commands() {
		if strings.HasPrefix(c, hardisStartPrefix) || strings.HasPrefix(c, startPrefix) {
			t.Fatalf("a deploy ran: %s", c)
		}
	}
}

func TestHardisResumeIsNative(t *testing.T) {
	e, f, s, dir, cs := hardisSetup(t, map[string]fakeReply{
		reportPrefix: {Stdout: hardisFixture(t, "hardis-start-pass-report.json")},
	}, "")
	res, err := e.Deploy(context.Background(), s, cs, providers.DeployOptions{Dir: dir, ResumeID: "0Af000000000002AAA"})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Passed || res.Facts["salesforce.hardis.resume_path"] != "native" {
		t.Fatalf("result: %+v", res)
	}
	if cmds := f.commands(); len(cmds) != 2 || !strings.HasPrefix(cmds[1], reportPrefix+" --job-id 0Af000000000002AAA") {
		t.Fatalf("commands: %v", cmds)
	}
}

func TestHardisRefusesChangedPlan(t *testing.T) {
	e, f, s, dir, cs := hardisSetup(t, map[string]fakeReply{}, "")
	cs.SHA256 = strings.Repeat("0", 64)
	if _, err := e.Deploy(context.Background(), s, cs, providers.DeployOptions{Dir: dir}); !errors.Is(err, ErrPlanChanged) {
		t.Fatalf("err = %v", err)
	}
	if len(f.commands()) != 0 {
		t.Fatalf("commands ran: %v", f.commands())
	}
}

func TestHardisMissingOrOldFails(t *testing.T) {
	cases := []struct {
		name    string
		plugins fakeReply
		min     string
		want    error
	}{
		{"missing", fakeReply{Stdout: []byte(`[{"name":"@salesforce/cli","version":"2.155.7"},{"name":"sfdx-git-delta","version":"6.45.1"}]`)}, "", ErrHardisMissing},
		{"older than min_version", fakeReply{Stdout: hardisFixture(t, "sf-plugins-hardis.json")}, "8.22.0", ErrHardisTooOld},
		{"older than the tested version", fakeReply{Stdout: hardisPluginsWith(t, "8.19.2")}, "", ErrHardisTooOld},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e, f, s, dir, cs := hardisSetup(t, map[string]fakeReply{hardisPluginsPrefix: tc.plugins}, tc.min)
			_, err := e.Deploy(context.Background(), s, cs, providers.DeployOptions{Dir: dir})
			if !errors.Is(err, tc.want) || !strings.Contains(err.Error(), "sf plugins install sfdx-hardis@8.20.1") {
				t.Fatalf("deploy err = %v", err)
			}
			_, err = e.Validate(context.Background(), s, cs, providers.ValidateOptions{Dir: dir})
			if !errors.Is(err, tc.want) {
				t.Fatalf("validate err = %v", err)
			}
			for _, c := range f.commands() {
				if c != hardisPluginsPrefix {
					t.Fatalf("ran %s after the version check failed", c)
				}
			}
		})
	}
	e, _, s, dir, cs := hardisSetup(t, map[string]fakeReply{hardisPluginsPrefix: {Stdout: []byte("oops"), ExitCode: 1}}, "")
	if _, err := e.Deploy(context.Background(), s, cs, providers.DeployOptions{Dir: dir}); err == nil || !strings.Contains(err.Error(), "could not list sf plugins") {
		t.Fatalf("err = %v", err)
	}
}

func TestHardisRefusesDeployCommands(t *testing.T) {
	cases := []struct {
		name, file, body string
		want             []string
	}{
		{"pre deploy", "config/.sfdx-hardis.yml", "projectName: probe\ncommandsPreDeploy:\n  - id: a\n    command: echo hi\n",
			[]string{"config/.sfdx-hardis.yml defines commandsPreDeploy"}},
		{"both", "config/.sfdx-hardis.yml", "commandsPreDeploy: []\ncommandsPostDeploy:\n  - id: b\n    command: echo hi\n",
			[]string{"defines commandsPreDeploy", "defines commandsPostDeploy"}},
		{"branch file", "config/branches/.sfdx-hardis-uat.yml", "commandsPostDeploy:\n  - id: b\n    command: echo hi\n",
			[]string{"config/branches/.sfdx-hardis-uat.yml defines commandsPostDeploy"}},
		{"unreadable YAML", "config/.sfdx-hardis.yml", "commandsPreDeploy: [\n", []string{"could not read the sfdx-hardis config as YAML"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e, f, s, dir, cs := hardisSetup(t, map[string]fakeReply{}, "")
			path := filepath.Join(dir, filepath.FromSlash(tc.file))
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(tc.body), 0o644); err != nil {
				t.Fatal(err)
			}
			dres, err := e.Deploy(context.Background(), s, cs, providers.DeployOptions{Dir: dir})
			if err != nil {
				t.Fatal(err)
			}
			vres, err := e.Validate(context.Background(), s, cs, providers.ValidateOptions{Dir: dir})
			if err != nil {
				t.Fatal(err)
			}
			for _, problems := range [][]string{dres.Problems, vres.Problems} {
				joined := strings.Join(problems, "\n")
				for _, w := range tc.want {
					if !strings.Contains(joined, w) {
						t.Errorf("problems %q lack %q", joined, w)
					}
				}
			}
			if dres.Passed || vres.Passed || dres.Facts["salesforce.hardis.commands_refused"] != true {
				t.Fatalf("deploy %+v validate %+v", dres, vres)
			}
			if len(f.commands()) != 0 {
				t.Fatalf("sf ran before the refusal: %v", f.commands())
			}
		})
	}

	// A hardis config without deploy commands does not stop the deploy.
	e, _, s, dir, cs := hardisSetup(t, map[string]fakeReply{
		queryPrefix:       sandboxOrg(t),
		hardisStartPrefix: {Stdout: hardisFixture(t, "hardis-start-pass.json")},
	}, "")
	if err := os.MkdirAll(filepath.Join(dir, "config"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config", ".sfdx-hardis.yml"), []byte("projectName: probe\nuseDeltaDeployment: false\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := e.Deploy(context.Background(), s, cs, providers.DeployOptions{Dir: dir})
	if err != nil || !res.Passed {
		t.Fatalf("result %+v, err %v", res, err)
	}
}

func TestHardisValidateIsNative(t *testing.T) {
	pass := fixture(t, "deploy-validate-pass.json")
	e, f, s, dir, cs := hardisSetup(t, map[string]fakeReply{
		validatePrefix: {Stdout: pass},
		reportPrefix:   {Stdout: pass},
	}, "")
	res, err := e.Validate(context.Background(), s, cs, providers.ValidateOptions{Dir: dir, TestLevel: "RunLocalTests"})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Passed || res.Facts["salesforce.hardis.validate_path"] != "native" || res.Facts["salesforce.engine"] != "hardis" ||
		res.Facts["salesforce.hardis.version"] != "8.20.1" || res.Facts["salesforce.check_only"] != true {
		t.Fatalf("result: %+v", res)
	}
	cmds := f.commands()
	if len(cmds) != 3 || cmds[0] != hardisPluginsPrefix || !strings.HasPrefix(cmds[1], validatePrefix) || !strings.HasPrefix(cmds[2], reportPrefix) {
		t.Fatalf("commands: %v", cmds)
	}
	noHardisWrappers(t, f)
}

// TestHardisValidateNeverTrustsTheWrapper: the captured validate wrapper output is a real
// deploy, checkOnly false. If such a body ever reached validate, it is refused.
func TestHardisValidateNeverTrustsTheWrapper(t *testing.T) {
	var body struct {
		Result struct {
			OutputString json.RawMessage `json:"outputstring"`
		} `json:"result"`
	}
	if err := json.Unmarshal(hardisFixture(t, "hardis-validate-pass.json"), &body); err != nil {
		t.Fatal(err)
	}
	e, _, s, dir, cs := hardisSetup(t, map[string]fakeReply{
		validatePrefix: {Stdout: body.Result.OutputString},
		reportPrefix:   {Stdout: hardisFixture(t, "hardis-validate-pass-report.json")},
	}, "")
	if _, err := e.Validate(context.Background(), s, cs, providers.ValidateOptions{Dir: dir}); !errors.Is(err, ErrNotCheckOnly) {
		t.Fatalf("err = %v", err)
	}
}

// TestHardisPlanIsNative: the hardis engine plans exactly as the native engine does.
func TestHardisPlanIsNative(t *testing.T) {
	plan := func(engine string) providers.ChangeSet {
		dir := planProject(t)
		exec := func(_ context.Context, c Cmd) (Output, error) {
			out := argValue(c.Args, "--output-dir")
			data, err := os.ReadFile(filepath.Join("testdata", "sgd", "delta", "package", "package.xml"))
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(out, "package.xml"), data, 0o644); err != nil {
				t.Fatal(err)
			}
			return Output{Stdout: []byte(`{"status":0,"result":{"path":"x/package.xml","name":"package.xml"},"warnings":[]}`)}, nil
		}
		p := planProvider(exec)
		if err := p.Configure(&config.Config{Salesforce: config.Salesforce{Engine: engine}}); err != nil {
			t.Fatal(err)
		}
		req := planReq(dir)
		req.Full, req.From, req.To = true, "", ""
		cs, err := p.Plan(context.Background(), req)
		if err != nil {
			t.Fatal(err)
		}
		return cs
	}
	native, hardis := plan(config.EngineNative), plan(config.EngineHardis)
	if native.SHA256 == "" || !reflect.DeepEqual(native, hardis) {
		t.Fatalf("native %+v\nhardis %+v", native, hardis)
	}
}
