package cli

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/run"
	"github.com/glassops-platform/glassops/internal/providers"
	"github.com/glassops-platform/glassops/internal/providers/salesforce"
)

const vfyConfig = `version: 1
targets:
  prod:
    id: 00D000000000001AAA
    tier: prod
    auth: { method: sfdx-url, url: $SF_PROD_AUTH_URL }
salesforce:
  test_level: { prod: RunLocalTests }
`

const (
	vfyDeployID = "0Af000000000003AAA"
	vfyCommit   = "6666666666666666666666666666666666666666"
)

// vfyDeployStart is when the synthetic deploy result says the deploy step started. The
// harness clock reads 2026-10-26 12:00 UTC.
var vfyDeployStart = time.Date(2026, 10, 26, 11, 50, 0, 0, time.UTC)

type vfyHarness struct {
	*harness
	fp        *fakeProvider
	authCalls int
	req       providers.VerifyRequest
	calls     int
	cs        providers.ChangeSet
}

// newVerifyHarness writes a config, a plan whose digest matches its files and a passed
// deploy result, and installs a fake provider whose verify returns out.
func newVerifyHarness(t *testing.T, cfg string, out providers.VerifyResult) *vfyHarness {
	t.Helper()
	h := &vfyHarness{harness: newHarness(t)}
	h.write(t, "glassops.yaml", cfg)
	h.write(t, filepath.FromSlash(".glassops/plan/package/package.xml"), "<Package/>\n")
	sum, err := providers.PlanDigest(h.dir, filepath.Join(h.dir, ".glassops", "plan"))
	if err != nil {
		t.Fatal(err)
	}
	h.cs = providers.ChangeSet{Schema: providers.PlanSchema, Provider: "salesforce", Dir: ".glassops/plan",
		Manifest: ".glassops/plan/package/package.xml", SHA256: sum, Facts: map[string]any{}}
	if err := providers.WriteChangeSet(filepath.Join(h.dir, ".glassops", "plan.json"), h.cs); err != nil {
		t.Fatal(err)
	}
	h.writeDeployResult(t, nil)
	h.fp = &fakeProvider{
		detect: true,
		auth: func(providers.Target) error {
			h.authCalls++
			return nil
		},
		describe: reportsAs(providers.KindProduction),
		verifyFn: func(_ providers.Session, req providers.VerifyRequest) (providers.VerifyResult, error) {
			h.calls++
			h.req = req
			return out, nil
		},
	}
	h.app.Providers = providers.Registry{h.fp}
	return h
}

// writeDeployResult writes the deploy.json a passed full deploy of vfyDeployID writes, then
// lets edit change it.
func (h *vfyHarness) writeDeployResult(t *testing.T, edit func(*run.Result)) {
	t.Helper()
	r := run.Result{
		Schema: run.Schema, Step: "deploy", Status: run.StatusPass,
		Started: vfyDeployStart, Ended: vfyDeployStart.Add(3 * time.Minute),
		Target: "prod", Provider: "salesforce",
		Facts: map[string]any{
			"deploy.id": vfyDeployID, "deploy.quick": false, "deploy.timed_out": false,
			"deploy.commit": vfyCommit, "deploy.permit_id": "a1b2c3d4",
		},
		Artifacts: []string{}, Errors: []string{},
	}
	if edit != nil {
		edit(&r)
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	h.write(t, filepath.FromSlash(".glassops/results/deploy.json"), string(data))
}

func (h *vfyHarness) verify(args ...string) int {
	return h.exec(append([]string{"verify", "--target", "prod"}, args...)...)
}

func vfyAllPass() providers.VerifyResult {
	return providers.VerifyResult{Passed: true, Checks: []providers.VerifyCheck{
		{Name: "verify.status", Status: providers.VerifyPass, Detail: "Succeeded"},
		{Name: "verify.members", Status: providers.VerifyPass, Detail: "all 3"},
	}, Facts: map[string]any{"verify.deploy_status": "Succeeded"}}
}

// vfyShell replaces shellRunner for one test. It records the command and its environment and
// answers with code, or with err when err is set.
func vfyShell(t *testing.T, code int, err error, env *[]string, command *string) {
	t.Helper()
	saved := shellRunner
	t.Cleanup(func() { shellRunner = saved })
	shellRunner = func(_ context.Context, _ string, e []string, c string, stdout, _ io.Writer) (int, error) {
		*env, *command = e, c
		io.WriteString(stdout, "smoke output\n")
		return code, err
	}
}

func TestVerifyPassesAndPassesTheDeployRecord(t *testing.T) {
	h := newVerifyHarness(t, vfyConfig, vfyAllPass())
	if code := h.verify(); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	r := h.result(t, "verify")
	if r.Status != run.StatusPass || r.Target != "prod" || r.Provider != "salesforce" {
		t.Fatalf("result: %+v", r)
	}
	req := h.req
	if req.Deploy.ID != vfyDeployID || req.Deploy.Quick || req.Deploy.Commit != vfyCommit || req.Deploy.TestLevel != "RunLocalTests" {
		t.Fatalf("deploy record: %+v", req.Deploy)
	}
	if !req.WindowStart.Equal(vfyDeployStart) || !req.WindowEnd.Equal(h.app.Now()) {
		t.Fatalf("window %v to %v", req.WindowStart, req.WindowEnd)
	}
	if req.ForeignDeploys != "warn" || req.Flows != "fail" {
		t.Fatalf("modes %q %q, want the defaults warn and fail", req.ForeignDeploys, req.Flows)
	}
	if req.Plan.SHA256 != h.cs.SHA256 || req.Dir != h.dir {
		t.Fatalf("plan %+v dir %s", req.Plan, req.Dir)
	}
	if checkStatus(r, "verify.status") != run.StatusPass || checkStatus(r, "verify.smoke") != run.StatusSkipped {
		t.Fatalf("checks: %+v", r.Checks)
	}
	if r.Facts["verify.deploy_id"] != vfyDeployID || r.Facts["verify.permit_id"] != "a1b2c3d4" || r.Facts["verify.passed"] != true ||
		r.Facts["verify.deploy_status"] != "Succeeded" || r.Facts["verify.plan_bound_to_deploy"] != false {
		t.Fatalf("facts: %v", r.Facts)
	}
	if !h.fp.allClosed() {
		t.Fatal("a session was left open")
	}
}

func TestVerifyConfigModesReachTheProvider(t *testing.T) {
	cfg := vfyConfig + "verify:\n  foreign_deploys: fail\n  flows: warn\n"
	h := newVerifyHarness(t, cfg, vfyAllPass())
	if code := h.verify(); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if h.req.ForeignDeploys != "fail" || h.req.Flows != "warn" {
		t.Fatalf("modes %q %q", h.req.ForeignDeploys, h.req.Flows)
	}
}

func TestVerifyOutcomes(t *testing.T) {
	cases := []struct {
		name   string
		out    providers.VerifyResult
		want   int
		status run.Status
		check  string
		detail string
	}{
		{"failed check exits 1", providers.VerifyResult{Passed: false, Checks: []providers.VerifyCheck{
			{Name: "verify.flows", Status: providers.VerifyFail, Detail: "Flow X runs version 1"},
		}}, exitcode.Fail, run.StatusFail, "verify.flows", "Flow X runs version 1"},
		{"warning passes", providers.VerifyResult{Passed: true, Checks: []providers.VerifyCheck{
			{Name: "verify.foreign_deploys", Status: providers.VerifyWarn, Detail: "2 other deploys"},
		}}, exitcode.OK, run.StatusPass, "verify.foreign_deploys", "warning: 2 other deploys"},
		{"not passed without a failing check still fails", providers.VerifyResult{Passed: false, Checks: []providers.VerifyCheck{
			{Name: "verify.status", Status: providers.VerifyPass},
		}}, exitcode.Fail, run.StatusFail, "verify.status", ""},
		{"unknown check status is an error", providers.VerifyResult{Passed: true, Checks: []providers.VerifyCheck{
			{Name: "verify.status", Status: "maybe"},
		}}, exitcode.Error, run.StatusError, "verify.status", `status "maybe"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := newVerifyHarness(t, vfyConfig, tc.out)
			if code := h.verify(); code != tc.want {
				t.Fatalf("exit %d, want %d. output:\n%s", code, tc.want, h.out)
			}
			r := h.result(t, "verify")
			if r.Status != tc.status {
				t.Fatalf("status %s, want %s", r.Status, tc.status)
			}
			if !strings.Contains(detailOf(r, tc.check), tc.detail) {
				t.Fatalf("%s detail %q", tc.check, detailOf(r, tc.check))
			}
		})
	}
	t.Run("warnings are listed", func(t *testing.T) {
		h := newVerifyHarness(t, vfyConfig, cases[1].out)
		h.verify()
		got, _ := h.result(t, "verify").Facts["verify.warnings"].([]any)
		if len(got) != 1 || got[0] != "verify.foreign_deploys" {
			t.Fatalf("verify.warnings = %v", got)
		}
	})
}

func TestVerifyProviderErrorExits2(t *testing.T) {
	h := newVerifyHarness(t, vfyConfig, vfyAllPass())
	h.fp.verifyFn = func(providers.Session, providers.VerifyRequest) (providers.VerifyResult, error) {
		return providers.VerifyResult{}, errors.New("salesforce: read DeployRequest: boom")
	}
	if code := h.verify(); code != exitcode.Error {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	r := h.result(t, "verify")
	if checkStatus(r, "verify") != run.StatusError || !strings.Contains(detailOf(r, "verify"), "boom") {
		t.Fatalf("checks: %+v", r.Checks)
	}
	if !h.fp.allClosed() {
		t.Fatal("a session was left open")
	}
}

// Everything verify can check without the target is checked before it logs in.
func TestVerifyRefusesBeforeLogin(t *testing.T) {
	cases := []struct {
		name  string
		setup func(t *testing.T, h *vfyHarness)
		check string
		want  string
	}{
		{"plan digest mismatch", func(t *testing.T, h *vfyHarness) {
			h.write(t, filepath.FromSlash(".glassops/plan/package/package.xml"), "<Package><types><members>Evil</members></types></Package>\n")
		}, "plan", "changed after plan wrote them"},
		{"no plan", func(t *testing.T, h *vfyHarness) {
			_ = os.Remove(filepath.Join(h.dir, ".glassops", "plan.json"))
		}, "plan", "read plan"},
		{"no deploy result", func(t *testing.T, h *vfyHarness) {
			_ = os.Remove(filepath.Join(h.dir, ".glassops", "results", "deploy.json"))
		}, "deploy_result", "read the deploy result"},
		{"deploy result of another step", func(t *testing.T, h *vfyHarness) {
			h.writeDeployResult(t, func(r *run.Result) { r.Step = "validate" })
		}, "deploy_result", "is not a deploy result"},
		{"deploy failed", func(t *testing.T, h *vfyHarness) {
			h.writeDeployResult(t, func(r *run.Result) { r.Status = run.StatusFail })
		}, "deploy_result", "status fail"},
		{"deploy timed out", func(t *testing.T, h *vfyHarness) {
			h.writeDeployResult(t, func(r *run.Result) { r.Status = run.StatusError; r.Facts["deploy.timed_out"] = true })
		}, "deploy_result", "status error"},
		{"timed out fact on a pass", func(t *testing.T, h *vfyHarness) {
			h.writeDeployResult(t, func(r *run.Result) { r.Facts["deploy.timed_out"] = true })
		}, "deploy_result", "did not finish"},
		{"deploy for another target", func(t *testing.T, h *vfyHarness) {
			h.writeDeployResult(t, func(r *run.Result) { r.Target = "qa" })
		}, "deploy_result", `target "qa"`},
		{"deploy from another provider", func(t *testing.T, h *vfyHarness) {
			h.writeDeployResult(t, func(r *run.Result) { r.Provider = "exec" })
		}, "deploy_result", `provider "exec"`},
		{"deploy without an Id", func(t *testing.T, h *vfyHarness) {
			h.writeDeployResult(t, func(r *run.Result) { delete(r.Facts, "deploy.id") })
		}, "deploy_result", "no deploy.id"},
		{"deploy without a start", func(t *testing.T, h *vfyHarness) {
			h.writeDeployResult(t, func(r *run.Result) { r.Started = time.Time{} })
		}, "deploy_result", "no start time"},
		{"deploy sent another plan", func(t *testing.T, h *vfyHarness) {
			h.writeDeployResult(t, func(r *run.Result) { r.Facts["deploy.plan_sha256"] = strings.Repeat("a", 64) })
		}, "plan", "the deploy sent plan"},
		{"target not configured", func(t *testing.T, h *vfyHarness) {
			h.write(t, "glassops.yaml", strings.Replace(vfyConfig, "  prod:", "  qa:", 1))
		}, "target", "not configured"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := newVerifyHarness(t, vfyConfig, vfyAllPass())
			tc.setup(t, h)
			if code := h.verify(); code != exitcode.Error {
				t.Fatalf("exit %d, want 2. output:\n%s", code, h.out)
			}
			if h.authCalls != 0 || h.calls != 0 {
				t.Fatalf("verify logged in (%d) or verified (%d)", h.authCalls, h.calls)
			}
			r := h.result(t, "verify")
			if checkStatus(r, tc.check) != run.StatusError || !strings.Contains(detailOf(r, tc.check), tc.want) {
				t.Fatalf("checks: %+v", r.Checks)
			}
		})
	}
}

func TestVerifyGuardMismatchStopsBeforeVerify(t *testing.T) {
	h := newVerifyHarness(t, vfyConfig, vfyAllPass())
	h.fp.describe = func(t providers.Target) (providers.TargetInfo, error) {
		return providers.TargetInfo{ID: "00D000000000009AAA", Kind: providers.KindProduction}, nil
	}
	if code := h.verify(); code != exitcode.Error {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if h.calls != 0 {
		t.Fatal("verify ran against a target the guard refused")
	}
	if !h.fp.allClosed() {
		t.Fatal("a session was left open")
	}
}

func TestVerifyEmptyDeployIsEmpty(t *testing.T) {
	h := newVerifyHarness(t, vfyConfig, vfyAllPass())
	h.writeDeployResult(t, func(r *run.Result) { r.Status = run.StatusEmpty; delete(r.Facts, "deploy.id") })
	if code := h.verify(); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if r := h.result(t, "verify"); r.Status != run.StatusEmpty || h.authCalls != 0 {
		t.Fatalf("status %s, logins %d", r.Status, h.authCalls)
	}
}

func TestVerifyDeployFactsOverrideConfig(t *testing.T) {
	h := newVerifyHarness(t, vfyConfig, vfyAllPass())
	h.writeDeployResult(t, func(r *run.Result) {
		r.Facts["deploy.test_level"] = "NoTestRun"
		r.Facts["deploy.plan_sha256"] = h.cs.SHA256
		r.Facts["deploy.quick"] = true
	})
	if code := h.verify(); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if h.req.Deploy.TestLevel != "NoTestRun" || !h.req.Deploy.Quick {
		t.Fatalf("deploy record %+v", h.req.Deploy)
	}
	if h.result(t, "verify").Facts["verify.plan_bound_to_deploy"] != true {
		t.Fatal("plan_bound_to_deploy is not true")
	}
}

func TestVerifyDeployResultFlag(t *testing.T) {
	h := newVerifyHarness(t, vfyConfig, vfyAllPass())
	src := filepath.Join(h.dir, ".glassops", "results", "deploy.json")
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	other := filepath.Join(h.dir, "saved-deploy.json")
	if err := os.WriteFile(other, data, 0o644); err != nil {
		t.Fatal(err)
	}
	_ = os.Remove(src)
	if code := h.verify("--deploy-result", other); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
}

func TestVerifySmoke(t *testing.T) {
	cfg := vfyConfig + "verify:\n  smoke: ./scripts/smoke.sh\n"
	cases := []struct {
		name   string
		code   int
		err    error
		exit   int
		status run.Status
	}{
		{"pass", 0, nil, exitcode.OK, run.StatusPass},
		{"fail", 3, nil, exitcode.Fail, run.StatusFail},
		{"cannot start", -1, errors.New("sh: not found"), exitcode.Error, run.StatusError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := newVerifyHarness(t, cfg, vfyAllPass())
			var env []string
			var command string
			vfyShell(t, tc.code, tc.err, &env, &command)
			if code := h.verify(); code != tc.exit {
				t.Fatalf("exit %d, want %d. output:\n%s", code, tc.exit, h.out)
			}
			r := h.result(t, "verify")
			if checkStatus(r, "verify.smoke") != tc.status || command != "./scripts/smoke.sh" {
				t.Fatalf("smoke %s (%s), command %q", checkStatus(r, "verify.smoke"), detailOf(r, "verify.smoke"), command)
			}
			results, _ := filepath.Abs(filepath.Join(h.dir, ".glassops", "results"))
			for _, want := range []string{ResultsDirEnv + "=" + results, "GLASSOPS_TARGET=prod", "GLASSOPS_TARGET_ID=00D000000000001AAA"} {
				found := false
				for _, e := range env {
					found = found || e == want
				}
				if !found {
					t.Fatalf("smoke env lacks %s", want)
				}
			}
		})
	}
	t.Run("not run on a provider error", func(t *testing.T) {
		h := newVerifyHarness(t, cfg, vfyAllPass())
		h.fp.verifyFn = func(providers.Session, providers.VerifyRequest) (providers.VerifyResult, error) {
			return providers.VerifyResult{}, errors.New("boom")
		}
		var env []string
		command := ""
		vfyShell(t, 0, nil, &env, &command)
		h.verify()
		if command != "" {
			t.Fatal("smoke ran after the provider errored")
		}
	})
}

// vfyFixture reads a capture from the Salesforce provider's testdata.
func vfyFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "providers", "salesforce", "testdata", "sf", name))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// TestVerifyEndToEndOnCaptures runs verify through the real Salesforce provider. Every sf
// call is answered with a capture: the login and org from a scratch org, and the deploy, its
// DeployRequest, the Flow and the window from the step 0.2 deploy of Greeter, GreeterTest and
// the Active Flow Glassops_Probe.
func TestVerifyEndToEndOnCaptures(t *testing.T) {
	cases := []struct {
		name    string
		cfg     string
		deploy  string
		request string
		report  string
		flowDef string
		flow    string
		now     string
		want    int
		check   string
		status  run.Status
	}{
		{"pass", "", "0Af000000000003AAA", "tooling-deployrequest-pass.json", "deploy-report-flow-pass.json",
			"tooling-flowdefinition-active.json", "tooling-flow-active.json", "18:33:30", exitcode.OK, "verify.foreign_deploys", run.StatusPass},
		{"draft flow fails", "", "0Af000000000005AAA", "tooling-deployrequest-window.json", "deploy-start-flow-draft.json",
			"tooling-flowdefinition-draft.json", "tooling-flow-draft.json", "18:34:10", exitcode.Fail, "verify.flows", run.StatusFail},
		{"foreign deploy fails under fail", "verify:\n  foreign_deploys: fail\n", "0Af000000000003AAA", "tooling-deployrequest-pass.json", "deploy-report-flow-pass.json",
			"tooling-flowdefinition-active.json", "tooling-flow-active.json", "18:35:00", exitcode.Fail, "verify.foreign_deploys", run.StatusFail},
		{"foreign deploy warns by default", "", "0Af000000000003AAA", "tooling-deployrequest-pass.json", "deploy-report-flow-pass.json",
			"tooling-flowdefinition-active.json", "tooling-flow-active.json", "18:35:00", exitcode.OK, "verify.foreign_deploys", run.StatusPass},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := newHarness(t)
			h.env["SF_QA_AUTH_URL"] = "force://synthetic-client:synthetic-secret:synthetic-refresh@example.my.salesforce.com"
			h.write(t, "glassops.yaml", `version: 1
targets:
  qa:
    id: 00D000000000002AAA
    tier: qa
    auth: { method: sfdx-url, url: $SF_QA_AUTH_URL }
salesforce:
  test_level: { qa: RunLocalTests }
`+tc.cfg)
			h.write(t, "sfdx-project.json", `{"packageDirectories":[{"path":"force-app","default":true}]}`)
			h.write(t, filepath.FromSlash("force-app/main/default/flows/Glassops_Probe.flow-meta.xml"),
				"<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<Flow xmlns=\"http://soap.sforce.com/2006/04/metadata\">\n    <status>Active</status>\n</Flow>\n")
			members := "<members>Greeter</members><members>GreeterTest</members><name>ApexClass</name>"
			if tc.deploy == "0Af000000000005AAA" {
				members = ""
			}
			h.write(t, filepath.FromSlash(".glassops/plan/package/package.xml"), `<?xml version="1.0" encoding="UTF-8"?>
<Package xmlns="http://soap.sforce.com/2006/04/metadata">
  <types>`+members+`</types>
  <types><members>Glassops_Probe</members><name>Flow</name></types>
  <version>62.0</version>
</Package>
`)
			sum, err := providers.PlanDigest(h.dir, filepath.Join(h.dir, ".glassops", "plan"))
			if err != nil {
				t.Fatal(err)
			}
			cs := providers.ChangeSet{Schema: providers.PlanSchema, Provider: "salesforce", Dir: ".glassops/plan",
				Manifest: ".glassops/plan/package/package.xml", SHA256: sum, Facts: map[string]any{}}
			if err := providers.WriteChangeSet(filepath.Join(h.dir, ".glassops", "plan.json"), cs); err != nil {
				t.Fatal(err)
			}
			// The deploy step started a second before the org's StartDate. verify runs at now,
			// which is 18:35 when the window takes in the two later deploys of the capture.
			start := map[string]time.Time{
				"0Af000000000003AAA": time.Date(2026, 11, 8, 18, 33, 10, 0, time.UTC),
				"0Af000000000005AAA": time.Date(2026, 11, 8, 18, 34, 0, 0, time.UTC),
			}[tc.deploy]
			data, _ := json.Marshal(run.Result{Schema: run.Schema, Step: "deploy", Status: run.StatusPass, Target: "qa", Provider: "salesforce",
				Started: start, Ended: start.Add(5 * time.Second),
				Facts: map[string]any{"deploy.id": tc.deploy, "deploy.quick": false, "deploy.timed_out": false}})
			h.write(t, filepath.FromSlash(".glassops/results/deploy.json"), string(data))
			now, err := time.Parse(time.RFC3339, "2026-11-08T"+tc.now+"Z")
			if err != nil {
				t.Fatal(err)
			}
			h.app.Now = func() time.Time { return now }

			replies := map[string]string{
				"sf org login sfdx-url":                      "login-sfdx-url-scratch.json",
				"sf data query --query SELECT Id, IsSandbox": "query-organization-scratch.json",
				"sf org display":                             "org-display-scratch.json",
				"sf data query --use-tooling-api --query SELECT Id, Status, CheckOnly, NumberComponentsDeployed": tc.request,
				"sf project deploy report --job-id " + tc.deploy:                                                 tc.report,
				"sf data query --use-tooling-api --query SELECT Id, DeveloperName":                               tc.flowDef,
				"sf data query --use-tooling-api --query SELECT Id, VersionNumber":                               tc.flow,
				"sf data query --use-tooling-api --query SELECT Id, Status, CheckOnly, CreatedDate":              "tooling-deployrequest-window.json",
			}
			var commands []string
			exec := func(_ context.Context, c salesforce.Cmd) (salesforce.Output, error) {
				if c.Name == "icacls" {
					return salesforce.Output{}, nil
				}
				key := strings.Join(append([]string{c.Name}, c.Args...), " ")
				commands = append(commands, key)
				best := ""
				for prefix := range replies {
					if strings.HasPrefix(key, prefix) && len(prefix) > len(best) {
						best = prefix
					}
				}
				if best == "" {
					return salesforce.Output{Stdout: []byte(`{"status":1,"name":"Unexpected","message":"no reply"}`), ExitCode: 1}, nil
				}
				return salesforce.Output{Stdout: vfyFixture(t, replies[best])}, nil
			}
			h.app.Providers = providers.Registry{salesforce.NewWithOptions(salesforce.Options{
				Exec: exec, LookupEnv: h.app.LookupEnv, TempDir: t.TempDir(),
			})}

			if code := h.exec("verify", "--target", "qa"); code != tc.want {
				t.Fatalf("exit %d, want %d. output:\n%s", code, tc.want, h.out)
			}
			r := h.result(t, "verify")
			if checkStatus(r, tc.check) != tc.status {
				t.Fatalf("%s = %s (%s). checks: %+v", tc.check, checkStatus(r, tc.check), detailOf(r, tc.check), r.Checks)
			}
			warnings, _ := r.Facts["verify.warnings"].([]any)
			switch tc.name {
			case "pass":
				for _, c := range r.Checks {
					if c.Status != run.StatusPass && c.Status != run.StatusSkipped {
						t.Fatalf("check %s is %s: %s", c.Name, c.Status, c.Detail)
					}
				}
				if len(warnings) != 0 {
					t.Fatalf("warnings on a clean pass: %v", warnings)
				}
			case "foreign deploy warns by default":
				if !strings.HasPrefix(detailOf(r, tc.check), "warning: 2 other deploys") || len(warnings) != 1 {
					t.Fatalf("detail %q, warnings %v", detailOf(r, tc.check), warnings)
				}
			}
			for _, c := range commands {
				if strings.Contains(c, "deploy start") || strings.Contains(c, "deploy quick") || strings.Contains(c, "deploy validate") {
					t.Fatalf("verify ran %q", c)
				}
			}
			if strings.Contains(h.out.String(), "synthetic-secret") {
				t.Fatal("the auth URL reached the output")
			}
		})
	}
}
