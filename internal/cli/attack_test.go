package cli

// Red team tests for the deploy path. Each test drives the real gate, deploy and permit verify
// commands, with the real HMAC signer and the real PERMIT checks, against the fake provider.
// A test here passes when the attack is refused. An attack that succeeds goes in
// attack_open_test.go, behind the redteam_open build tag, until it is fixed. None is open now.

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	osexec "os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/permit"
	"github.com/glassops-platform/glassops/internal/core/policy"
	"github.com/glassops-platform/glassops/internal/core/run"
	"github.com/glassops-platform/glassops/internal/providers"
)

const (
	rtCommit      = "6666666666666666666666666666666666666666"
	rtOtherCommit = "7777777777777777777777777777777777777777"
	rtQAID        = "00D000000000002AAA"
	rtUATID       = "00D000000000004AAA"
	rtProdID      = "00D000000000001AAA"
	rtManifest    = "<Package><types><members>Foo</members><name>ApexClass</name></types></Package>\n"
	rtSource      = "public class Foo {}\n"
	rtSourcePath  = "force-app/main/default/classes/Foo.cls"
	rtPlanRel     = ".glassops/plan/package/package.xml"
)

// rtKey is a synthetic 32 byte key in base64.
var rtKey = base64.StdEncoding.EncodeToString([]byte("rt-synthetic-key-0123456789abcde"))

// rtOtherKey is a second synthetic 32 byte key.
var rtOtherKey = base64.StdEncoding.EncodeToString([]byte("rt-other-synthetic-key-012345678"))

// rtGateTime is the harness clock, a Monday.
var rtGateTime = time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)

const rtConfig = `version: 1
targets:
  qa:
    id: 00D000000000002AAA
    tier: qa
    auth: { method: sfdx-url, url: $SF_QA_AUTH_URL }
  uat:
    id: 00D000000000004AAA
    tier: uat
    auth: { method: sfdx-url, url: $SF_UAT_AUTH_URL }
  prod:
    id: 00D000000000001AAA
    tier: prod
    auth: { method: sfdx-url, url: $SF_PROD_AUTH_URL }
policy:
  rules:
    separation_of_duties:
      tiers: [prod]
`

// rtRule is a synthetic policy rule with a fixed outcome.
type rtRule struct {
	name    string
	outcome policy.Outcome
}

func (r rtRule) Name() string { return r.name }
func (r rtRule) Evaluate(policy.Input) []policy.Result {
	return []policy.Result{{Rule: r.name, Outcome: r.outcome, Detail: "rt synthetic " + string(r.outcome)}}
}

// rtUseRules makes gate evaluate rules and keeps the real signer and the real PERMIT checks
// in every seam, restoring the seams after t.
func rtUseRules(t *testing.T, rules func() []policy.Rule) {
	t.Helper()
	oldRules, oldSigner, oldSign, oldWrite, oldFuncs := gateRules, gateSigner, gateSign, gateWrite, permitFuncs
	t.Cleanup(func() {
		gateRules, gateSigner, gateSign, gateWrite, permitFuncs = oldRules, oldSigner, oldSign, oldWrite, oldFuncs
	})
	gateRules = rules
	gateSigner, gateSign, gateWrite = permit.NewHMAC, permit.Sign, permit.Write
	permitFuncs.read, permitFuncs.verify, permitFuncs.signer = permit.Read, permit.Verify, permit.NewHMAC
}

func rtFixed(o policy.Outcome) func() []policy.Rule {
	return func() []policy.Rule { return []policy.Rule{rtRule{name: "rt_rule", outcome: o}} }
}

// rtEnv is one project run end to end: gate, then deploy or permit verify.
type rtEnv struct {
	*harness
	fp        *fakeProvider
	authCalls int
	deployed  []providers.ChangeSet
	opts      []providers.DeployOptions
	// report is what the target says about itself after login.
	report func(providers.Target) (providers.TargetInfo, error)
}

// rtNew sets up a GitHub push run with cfg, a plan, a source file, the key and a fake
// provider whose targets report their configured Id and the kind their tier needs. Gate
// evaluates one passing synthetic rule.
func rtNew(t *testing.T, cfg string) *rtEnv {
	t.Helper()
	e := &rtEnv{harness: newHarness(t)}
	e.app.Now = func() time.Time { return rtGateTime }
	e.write(t, "glassops.yaml", cfg)
	e.write(t, rtSourcePath, rtSource)
	rtWritePlan(t, e.harness, ".glassops/plan", rtManifest, nil)
	rtGitTree(t, e)
	for k, v := range map[string]string{
		"GITHUB_ACTIONS":    "true",
		"GITHUB_REPOSITORY": "glassops-platform/glassops",
		"GITHUB_SHA":        rtCommit,
		"GITHUB_REF":        "refs/heads/main",
		"GITHUB_EVENT_NAME": "push",
		"GITHUB_ACTOR":      "deployer",
		"GITHUB_RUN_ID":     "1",
		permit.KeyEnv:       rtKey,
	} {
		e.env[k] = v
	}
	e.report = func(tg providers.Target) (providers.TargetInfo, error) {
		k := providers.KindSandbox
		switch tg.Tier {
		case "prod", "devhub":
			k = providers.KindProduction
		case "ephemeral":
			k = providers.KindScratch
		}
		return providers.TargetInfo{ID: tg.ID, Kind: k, Facts: map[string]any{}}, nil
	}
	e.fp = &fakeProvider{
		detect: true,
		auth: func(providers.Target) error {
			e.authCalls++
			return nil
		},
		describe: func(tg providers.Target) (providers.TargetInfo, error) { return e.report(tg) },
		deploy: func(_ providers.Session, cs providers.ChangeSet, o providers.DeployOptions) (providers.DeployResult, error) {
			e.deployed = append(e.deployed, cs)
			e.opts = append(e.opts, o)
			return providers.DeployResult{Passed: true, ID: "0Af00000000RT01AAA", ComponentsDeployed: 1, Facts: map[string]any{}}, nil
		},
	}
	e.app.Providers = providers.Registry{e.fp}
	rtUseRules(t, rtFixed(policy.Pass))
	return e
}

// rtGitTree makes the project a real git repository with everything but .glassops committed,
// so deploy's tree check runs the real git status. HEAD reads as rtCommit, the commit the
// harness CI reports, because the real commit Id changes with every run.
func rtGitTree(t *testing.T, e *rtEnv) {
	t.Helper()
	e.write(t, ".gitignore", ".glassops/\n")
	gitIn(t, e.dir, "init", "-q", "-b", "main")
	gitIn(t, e.dir, "config", "user.email", "test@example.invalid")
	gitIn(t, e.dir, "config", "user.name", "test")
	gitIn(t, e.dir, "config", "core.autocrlf", "false")
	e.commitTree(t)
	e.tools["git rev-parse HEAD"] = rtCommit + "\n"
	fake := e.app.Run
	e.app.Run = func(dir, name string, args ...string) ([]byte, error) {
		if name == "git" && len(args) > 0 && args[0] == "status" {
			cmd := osexec.Command("git", args...)
			cmd.Dir = dir
			return cmd.Output()
		}
		return fake(dir, name, args...)
	}
}

// commitTree commits every change in the project, such as a config a test wrote.
func (e *rtEnv) commitTree(t *testing.T) {
	t.Helper()
	gitIn(t, e.dir, "add", "-A")
	gitIn(t, e.dir, "commit", "-q", "--allow-empty", "-m", "rt")
}

// rtWritePlan writes manifest, and each destructive manifest, under dir and the change set
// for them at <dir>.json, so .glassops/plan gives .glassops/plan.json. The change set sits
// outside dir, so its digest covers the plan files only. It returns the change set and its
// path.
func rtWritePlan(t *testing.T, h *harness, dir, manifest string, destructive map[string]string) (providers.ChangeSet, string) {
	t.Helper()
	cs := providers.ChangeSet{
		Schema: providers.PlanSchema, Provider: "salesforce",
		Dir: dir, Manifest: dir + "/package/package.xml",
		Destructive: []string{}, Facts: map[string]any{},
	}
	h.write(t, filepath.FromSlash(cs.Manifest), manifest)
	for name, body := range destructive {
		rel := dir + "/destructive/" + name
		h.write(t, filepath.FromSlash(rel), body)
		cs.Destructive = append(cs.Destructive, rel)
	}
	sum, err := providers.PlanDigest(h.dir, filepath.Join(h.dir, filepath.FromSlash(dir)))
	if err != nil {
		t.Fatal(err)
	}
	cs.SHA256 = sum
	path := filepath.Join(h.dir, filepath.FromSlash(dir)+".json")
	if err := providers.WriteChangeSet(path, cs); err != nil {
		t.Fatal(err)
	}
	return cs, path
}

func (e *rtEnv) cmd(args ...string) int {
	return Execute(e.app, append(args, "--dir", e.dir, "--ci", "github"))
}

// gate runs gate for target and fails t unless it exits want.
func (e *rtEnv) gate(t *testing.T, want int, args ...string) {
	t.Helper()
	if code := e.cmd(append([]string{"gate"}, args...)...); code != want {
		t.Fatalf("gate %v exit %d, want %d. output:\n%s", args, code, want, e.out)
	}
}

// refuse runs deploy with args and fails t unless it exits non-zero without deploying. With
// noLogin it also fails t when deploy logged in. It returns the exit code.
func (e *rtEnv) refuse(t *testing.T, noLogin bool, args ...string) int {
	t.Helper()
	before := e.authCalls
	code := e.cmd(append([]string{"deploy"}, args...)...)
	if code == exitcode.OK {
		t.Fatalf("deploy %v exit 0, the attack was not refused. output:\n%s", args, e.out)
	}
	if len(e.deployed) != 0 {
		t.Fatalf("deploy %v reached the provider's Deploy %d times", args, len(e.deployed))
	}
	if noLogin && e.authCalls != before {
		t.Fatalf("deploy %v logged in to the target before refusing", args)
	}
	t.Logf("refused, exit %d: %s", code, e.refusal(t))
	return code
}

// refusal returns the first check of the deploy result that did not pass, as name: detail.
func (e *rtEnv) refusal(t *testing.T) string {
	t.Helper()
	r := e.result(t, "deploy")
	for _, c := range r.Checks {
		if c.Status != run.StatusPass {
			return c.Name + ": " + c.Detail
		}
	}
	return "no failing check"
}

// verifyRefused runs permit verify and fails t when it exits 0.
func (e *rtEnv) verifyRefused(t *testing.T, args ...string) {
	t.Helper()
	if code := e.cmd(append([]string{"permit", "verify"}, args...)...); code == exitcode.OK {
		t.Fatalf("permit verify %v exit 0, the attack was not refused. output:\n%s", args, e.out)
	}
}

func (e *rtEnv) permitPath() string { return filepath.Join(e.dir, ".glassops", "permit.json") }

func (e *rtEnv) readPermit(t *testing.T) permit.Permit {
	t.Helper()
	p, err := permit.Read(e.permitPath())
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// editPermitJSON rewrites permit.json through a generic map, without signing again.
func (e *rtEnv) editPermitJSON(t *testing.T, edit func(m map[string]any)) {
	t.Helper()
	data, err := os.ReadFile(e.permitPath())
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	edit(m)
	out, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(e.permitPath(), out, 0o644); err != nil {
		t.Fatal(err)
	}
}

// editPermitRaw rewrites permit.json as text, without signing again.
func (e *rtEnv) editPermitRaw(t *testing.T, edit func(string) string) {
	t.Helper()
	data, err := os.ReadFile(e.permitPath())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(e.permitPath(), []byte(edit(string(data))), 0o644); err != nil {
		t.Fatal(err)
	}
}

func rtSub(m map[string]any, key string) map[string]any {
	v, _ := m[key].(map[string]any)
	return v
}

// TestRTControlGateThenDeployPasses proves the harness: an untouched PERMIT deploys, so every
// refusal below is caused by the attack and not by the setup.
func TestRTControlGateThenDeployPasses(t *testing.T) {
	e := rtNew(t, rtConfig)
	e.gate(t, exitcode.OK, "--target", "qa")
	p := e.readPermit(t)
	if !p.Allow || p.Signature.Alg != permit.Algorithm || p.Commit != rtCommit {
		t.Fatalf("permit %+v", p)
	}
	// A rewrite through a generic map that changes nothing still verifies.
	e.editPermitJSON(t, func(map[string]any) {})
	if code := e.cmd("permit", "verify", "--target", "qa"); code != exitcode.OK {
		t.Fatalf("permit verify exit %d. output:\n%s", code, e.out)
	}
	if code := e.cmd("deploy", "--target", "qa"); code != exitcode.OK {
		t.Fatalf("deploy exit %d. output:\n%s", code, e.out)
	}
	if len(e.deployed) != 1 || e.deployed[0].Manifest != rtPlanRel {
		t.Fatalf("deployed %+v", e.deployed)
	}
}

// Attack 1: replay an allow PERMIT against another target, org Id or tier.

func TestRTReplayToOtherTargetName(t *testing.T) {
	e := rtNew(t, rtConfig)
	e.gate(t, exitcode.OK, "--target", "qa")
	e.refuse(t, true, "--target", "uat")
	e.verifyRefused(t, "--target", "uat")
}

func TestRTReplayProdPermitToQA(t *testing.T) {
	e := rtNew(t, rtConfig)
	e.gate(t, exitcode.OK, "--target", "prod", "--no-connect")
	e.refuse(t, true, "--target", "qa")
}

func TestRTReplayToOtherOrgIDBehindSameName(t *testing.T) {
	// The config holds $RT_QA_ORG_ID, so its hash is the same in both runs and only the org
	// Id the variable names changes.
	cfg := strings.Replace(rtConfig, "id: "+rtQAID, "id: $RT_QA_ORG_ID", 1)
	e := rtNew(t, cfg)
	e.env["RT_QA_ORG_ID"] = rtQAID
	e.gate(t, exitcode.OK, "--target", "qa")
	e.env["RT_QA_ORG_ID"] = "00D000000000009AAA"
	e.refuse(t, true, "--target", "qa")
	e.verifyRefused(t, "--target", "qa")
}

func TestRTReplayToOtherTier(t *testing.T) {
	e := rtNew(t, rtConfig)
	e.gate(t, exitcode.OK, "--target", "qa")
	e.write(t, "glassops.yaml", strings.Replace(rtConfig, "tier: qa", "tier: dev", 1))
	e.refuse(t, true, "--target", "qa")
}

// Attack 2: tamper with any PERMIT field after gate signed it.

func TestRTTamperPermitFields(t *testing.T) {
	cases := []struct {
		name string
		edit func(m map[string]any)
	}{
		{"id", func(m map[string]any) { m["id"] = "ffffffffffffffffffffffffffffffff" }},
		{"schema", func(m map[string]any) { m["schema"] = "glassops.permit/v2" }},
		{"issued_at", func(m map[string]any) { m["issued_at"] = "2026-10-26T11:00:00Z" }},
		{"expires_at extended", func(m map[string]any) { m["expires_at"] = "2027-10-26T12:00:00Z" }},
		{"target.name", func(m map[string]any) { rtSub(m, "target")["name"] = "uat" }},
		{"target.tier", func(m map[string]any) { rtSub(m, "target")["tier"] = "dev" }},
		{"target.id", func(m map[string]any) { rtSub(m, "target")["id"] = rtUATID }},
		{"target.kind", func(m map[string]any) { rtSub(m, "target")["kind"] = "scratch" }},
		{"commit", func(m map[string]any) { m["commit"] = rtOtherCommit }},
		{"plan_sha256", func(m map[string]any) { m["plan_sha256"] = strings.Repeat("0", 64) }},
		{"config_sha256", func(m map[string]any) { m["config_sha256"] = strings.Repeat("0", 64) }},
		{"validation injected", func(m map[string]any) {
			m["validation"] = map[string]any{"deploy_id": "0Af00000000RT09AAA", "passed": true}
		}},
		{"separation_of_duties", func(m map[string]any) { rtSub(m, "separation_of_duties")["mode"] = "waived_by_config" }},
		{"checks emptied", func(m map[string]any) { m["checks"] = []any{} }},
		{"signature.alg none", func(m map[string]any) { rtSub(m, "signature")["alg"] = "none" }},
		{"signature.alg empty", func(m map[string]any) { rtSub(m, "signature")["alg"] = "" }},
		{"signature.alg lower case", func(m map[string]any) { rtSub(m, "signature")["alg"] = "hmac-sha256" }},
		{"signature.key_id", func(m map[string]any) { rtSub(m, "signature")["key_id"] = "00000000" }},
		{"signature.key_id removed", func(m map[string]any) { delete(rtSub(m, "signature"), "key_id") }},
		{"signature.value empty", func(m map[string]any) { rtSub(m, "signature")["value"] = "" }},
		{"signature.value other", func(m map[string]any) {
			rtSub(m, "signature")["value"] = base64.StdEncoding.EncodeToString(make([]byte, 32))
		}},
		{"signature removed", func(m map[string]any) { delete(m, "signature") }},
		{"unknown field", func(m map[string]any) { m["override"] = true }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := rtNew(t, rtConfig)
			e.gate(t, exitcode.OK, "--target", "qa")
			e.editPermitJSON(t, tc.edit)
			e.refuse(t, true, "--target", "qa")
			e.verifyRefused(t, "--target", "qa")
		})
	}
}

func TestRTTamperDenyPermitIntoAllow(t *testing.T) {
	cases := []struct {
		name string
		edit func(e *rtEnv, t *testing.T)
	}{
		{"untouched deny", func(*rtEnv, *testing.T) {}},
		{"allow set true", func(e *rtEnv, t *testing.T) {
			e.editPermitJSON(t, func(m map[string]any) { m["allow"] = true })
		}},
		{"duplicate allow key, last wins", func(e *rtEnv, t *testing.T) {
			e.editPermitRaw(t, func(s string) string {
				i := strings.LastIndex(s, "}")
				return s[:i] + `, "allow": true` + "\n}\n"
			})
		}},
		{"case variant ALLOW key", func(e *rtEnv, t *testing.T) {
			e.editPermitRaw(t, func(s string) string {
				i := strings.LastIndex(s, "}")
				return s[:i] + `, "ALLOW": true` + "\n}\n"
			})
		}},
		{"second document appended", func(e *rtEnv, t *testing.T) {
			p := e.readPermit(t)
			p.Allow = true
			data, _ := json.Marshal(p)
			e.editPermitRaw(t, func(s string) string { return s + string(data) })
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := rtNew(t, rtConfig)
			rtUseRules(t, rtFixed(policy.Fail))
			e.gate(t, exitcode.Fail, "--target", "qa")
			if e.readPermit(t).Allow {
				t.Fatal("gate wrote an allow PERMIT for a failing rule")
			}
			tc.edit(e, t)
			e.refuse(t, true, "--target", "qa")
			e.verifyRefused(t, "--target", "qa")
		})
	}
}

// Attack 2 and 3: tamper with or swap the plan, the plan digest or the config between gate and
// deploy.

func TestRTSwapPlanAfterGate(t *testing.T) {
	other := "<Package><types><members>*</members><name>ApexClass</name></types></Package>\n"
	cases := []struct {
		name string
		edit func(e *rtEnv, t *testing.T) []string
	}{
		{"a different plan, consistent with its own files", func(e *rtEnv, t *testing.T) []string {
			rtWritePlan(t, e.harness, ".glassops/plan", other, nil)
			return nil
		}},
		{"manifest edited, plan.json kept", func(e *rtEnv, t *testing.T) []string {
			e.write(t, filepath.FromSlash(rtPlanRel), other)
			return nil
		}},
		{"plan_sha256 edited to the new files' digest", func(e *rtEnv, t *testing.T) []string {
			e.write(t, filepath.FromSlash(rtPlanRel), other)
			cs, err := providers.ReadChangeSet(filepath.Join(e.dir, ".glassops", "plan.json"))
			if err != nil {
				t.Fatal(err)
			}
			cs.SHA256, _ = providers.PlanDigest(e.dir, filepath.Join(e.dir, ".glassops", "plan"))
			if err := providers.WriteChangeSet(filepath.Join(e.dir, ".glassops", "plan.json"), cs); err != nil {
				t.Fatal(err)
			}
			return nil
		}},
		{"an extra file in the plan dir", func(e *rtEnv, t *testing.T) []string {
			e.write(t, ".glassops/plan/destructive/destructiveChanges.xml", other)
			return nil
		}},
		{"another plan through --plan", func(e *rtEnv, t *testing.T) []string {
			_, path := rtWritePlan(t, e.harness, ".glassops/other", other, nil)
			return []string{"--plan", path}
		}},
		{"plan dir pointed at another dir", func(e *rtEnv, t *testing.T) []string {
			rtWritePlan(t, e.harness, ".glassops/other", other, nil)
			cs, err := providers.ReadChangeSet(filepath.Join(e.dir, ".glassops", "plan.json"))
			if err != nil {
				t.Fatal(err)
			}
			cs.Dir, cs.Manifest = ".glassops/other", ".glassops/other/package/package.xml"
			if err := providers.WriteChangeSet(filepath.Join(e.dir, ".glassops", "plan.json"), cs); err != nil {
				t.Fatal(err)
			}
			return nil
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := rtNew(t, rtConfig)
			e.gate(t, exitcode.OK, "--target", "qa")
			extra := tc.edit(e, t)
			e.refuse(t, true, append([]string{"--target", "qa"}, extra...)...)
			e.verifyRefused(t, append([]string{"--target", "qa"}, extra...)...)
		})
	}
}

func TestRTSwapConfigAfterGate(t *testing.T) {
	cases := []struct {
		name string
		edit func(e *rtEnv, t *testing.T) []string
	}{
		{"a rule removed", func(e *rtEnv, t *testing.T) []string {
			e.write(t, "glassops.yaml", strings.Split(rtConfig, "policy:")[0])
			return nil
		}},
		{"quick deploy turned off", func(e *rtEnv, t *testing.T) []string {
			e.write(t, "glassops.yaml", rtConfig+"salesforce:\n  deploy: { quick: off }\n")
			return nil
		}},
		{"a longer permit_ttl", func(e *rtEnv, t *testing.T) []string {
			e.write(t, "glassops.yaml", strings.Replace(rtConfig, "policy:\n", "policy:\n  permit_ttl: 24h\n", 1))
			return nil
		}},
		{"a comment only", func(e *rtEnv, t *testing.T) []string {
			e.write(t, "glassops.yaml", rtConfig+"# rt\n")
			return nil
		}},
		{"the config deleted, defaults used", func(e *rtEnv, t *testing.T) []string {
			if err := os.Remove(filepath.Join(e.dir, "glassops.yaml")); err != nil {
				t.Fatal(err)
			}
			return nil
		}},
		{"a weaker file through --config", func(e *rtEnv, t *testing.T) []string {
			e.write(t, "weaker.yaml", strings.Split(rtConfig, "policy:")[0])
			return []string{"--config", filepath.Join(e.dir, "weaker.yaml")}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := rtNew(t, rtConfig)
			e.gate(t, exitcode.OK, "--target", "qa")
			extra := tc.edit(e, t)
			e.refuse(t, true, append([]string{"--target", "qa"}, extra...)...)
			e.verifyRefused(t, append([]string{"--target", "qa"}, extra...)...)
		})
	}
}

func TestRTSwapCommitAfterGate(t *testing.T) {
	e := rtNew(t, rtConfig)
	e.gate(t, exitcode.OK, "--target", "qa")
	e.env["GITHUB_SHA"] = rtOtherCommit
	e.refuse(t, true, "--target", "qa")
	e.verifyRefused(t, "--target", "qa")
	// With no CI commit, deploy falls back to git rev-parse HEAD, which must match too.
	e.env["GITHUB_SHA"] = ""
	e.tools["git rev-parse HEAD"] = rtOtherCommit + "\n"
	e.refuse(t, true, "--target", "qa")
}

// Attack 4: key confusion.

func TestRTKeyGateRefusesWeakOrMissingKeys(t *testing.T) {
	cases := []struct {
		name  string
		setup func(e *rtEnv, t *testing.T)
	}{
		{"unset", func(e *rtEnv, _ *testing.T) { delete(e.env, permit.KeyEnv) }},
		{"empty", func(e *rtEnv, _ *testing.T) { e.env[permit.KeyEnv] = "" }},
		{"whitespace", func(e *rtEnv, _ *testing.T) { e.env[permit.KeyEnv] = " \n\t " }},
		{"31 bytes", func(e *rtEnv, _ *testing.T) {
			e.env[permit.KeyEnv] = base64.StdEncoding.EncodeToString(make([]byte, 31))
		}},
		{"not base64", func(e *rtEnv, _ *testing.T) { e.env[permit.KeyEnv] = "this is not a key!" }},
		{"only in files the repo controls", func(e *rtEnv, t *testing.T) {
			delete(e.env, permit.KeyEnv)
			e.write(t, ".env", permit.KeyEnv+"="+rtKey+"\n")
			e.write(t, ".glassops/permit.key", rtKey+"\n")
			e.write(t, ".glassops/.env", permit.KeyEnv+"="+rtKey+"\n")
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := rtNew(t, rtConfig)
			tc.setup(e, t)
			e.gate(t, exitcode.Error, "--target", "qa")
			if _, err := os.Stat(e.permitPath()); !errors.Is(err, fs.ErrNotExist) {
				t.Fatalf("gate wrote a PERMIT without a usable key: %v", err)
			}
			if e.authCalls != 0 {
				t.Fatal("gate logged in without a usable key")
			}
			if strings.Contains(e.out.String(), rtKey) {
				t.Fatal("the key reached the output")
			}
		})
	}
}

func TestRTKeyDeployRefusesWrongOrWeakKeys(t *testing.T) {
	cases := []struct {
		name string
		key  *string
	}{
		{"unset", nil},
		{"empty", rtStr("")},
		{"another valid key", &rtOtherKey},
		{"31 bytes", rtStr(base64.StdEncoding.EncodeToString(make([]byte, 31)))},
		{"the right key truncated", rtStr(rtKey[:len(rtKey)-8])},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := rtNew(t, rtConfig)
			e.gate(t, exitcode.OK, "--target", "qa")
			if tc.key == nil {
				delete(e.env, permit.KeyEnv)
			} else {
				e.env[permit.KeyEnv] = *tc.key
			}
			e.refuse(t, true, "--target", "qa")
			e.verifyRefused(t, "--target", "qa")
			if strings.Contains(e.out.String(), rtKey) || strings.Contains(e.out.String(), rtOtherKey) {
				t.Fatal("a key reached the output")
			}
		})
	}
}

func rtStr(s string) *string { return &s }

const rtSigstorePolicy = "  permit:\n    signing: sigstore\n    identity: https://github.com/glassops-platform/glassops/.github/workflows/glassops-deploy.yml@refs/heads/main\n    issuer: https://token.actions.githubusercontent.com\n"

func TestRTKeyHMACPermitWhereConfigAsksSigstore(t *testing.T) {
	// Only the selection logic is on this base: the sigstore verifier itself is a stub.
	sigstoreCfg := strings.Replace(rtConfig, "policy:\n", "policy:\n"+rtSigstorePolicy, 1)
	t.Run("deploy and permit verify", func(t *testing.T) {
		e := rtNew(t, rtConfig)
		e.gate(t, exitcode.OK, "--target", "qa")
		e.write(t, "glassops.yaml", sigstoreCfg)
		if code := e.refuse(t, true, "--target", "qa"); code != exitcode.Error {
			t.Fatalf("exit %d, want 2", code)
		}
		e.verifyRefused(t, "--target", "qa")
	})
	t.Run("gate signs nothing", func(t *testing.T) {
		e := rtNew(t, sigstoreCfg)
		e.gate(t, exitcode.Error, "--target", "qa")
		if _, err := os.Stat(e.permitPath()); !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("gate wrote a PERMIT with no sigstore signer: %v", err)
		}
	})
	t.Run("permit verify with a config that does not load", func(t *testing.T) {
		// peekConfig falls back to hmac when the config does not load. The command must
		// still refuse, because it loads the config again for real.
		e := rtNew(t, rtConfig)
		e.gate(t, exitcode.OK, "--target", "qa")
		e.write(t, "glassops.yaml", sigstoreCfg+"  not_a_field: true\n")
		if code := e.cmd("permit", "verify", "--target", "qa"); code != exitcode.Error {
			t.Fatalf("exit %d, want 2. output:\n%s", code, e.out)
		}
	})
}

// Attack 5: expiry and clock.

func TestRTExpiryAndClock(t *testing.T) {
	cases := []struct {
		name string
		now  time.Time
	}{
		{"exactly at expiry", rtGateTime.Add(2 * time.Hour)},
		{"after expiry", rtGateTime.Add(3 * time.Hour)},
		{"a year later", rtGateTime.AddDate(1, 0, 0)},
		{"no clock", time.Time{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := rtNew(t, rtConfig)
			e.gate(t, exitcode.OK, "--target", "qa")
			e.app.Now = func() time.Time { return tc.now }
			e.refuse(t, true, "--target", "qa")
			e.verifyRefused(t, "--target", "qa")
		})
	}
}

func TestRTGateWithoutAClockNeverDeploys(t *testing.T) {
	e := rtNew(t, rtConfig)
	e.app.Now = func() time.Time { return time.Time{} }
	e.cmd("gate", "--target", "qa")
	e.refuse(t, true, "--target", "qa")
	e.app.Now = func() time.Time { return rtGateTime }
	e.refuse(t, true, "--target", "qa")
}

// Attack 6: injection. The annotation channel must stay one line whatever the input holds.
// The stdout table is covered by TestRTWorkflowCommandInjectionThroughStdout.

func TestRTInjectionAnnotationsStayOneLine(t *testing.T) {
	e := rtNew(t, rtConfig)
	e.gate(t, exitcode.OK, "--target", "qa")
	name := "qa\n::warning::rt-forged\r\n%0A"
	e.refuse(t, true, "--target", name)
	for _, line := range strings.Split(e.out.String(), "\n") {
		if strings.HasPrefix(line, "::error::") && strings.Contains(line, "rt-forged") {
			// finish writes a line break as the two characters \n before the CI layer sees it,
			// and the GitHub layer still escapes a literal %.
			if !strings.Contains(line, `\n::warning::rt-forged`) || !strings.Contains(line, "%250A") {
				t.Fatalf("annotation not escaped: %q", line)
			}
			r := e.result(t, "deploy")
			if r.Target != name {
				t.Fatalf("result target %q, want the name as given", r.Target)
			}
			return
		}
	}
	t.Fatalf("no error annotation carried the target name. output:\n%s", e.out)
}

// Attack 7: the org guard.

func TestRTGuardRefusesTheWrongOrg(t *testing.T) {
	cases := []struct {
		name   string
		target string
		info   providers.TargetInfo
		gate   []string
	}{
		{"another org Id", "qa", providers.TargetInfo{ID: rtUATID, Kind: providers.KindSandbox}, nil},
		{"no org Id", "qa", providers.TargetInfo{Kind: providers.KindSandbox}, nil},
		{"the same Id in another case", "qa", providers.TargetInfo{ID: strings.ToLower(rtQAID), Kind: providers.KindSandbox}, nil},
		{"control: the 15 character form of the same Id", "qa", providers.TargetInfo{ID: "00D000000000002", Kind: providers.KindSandbox}, nil},
		{"prod Id reporting a sandbox", "prod", providers.TargetInfo{ID: rtProdID, Kind: providers.KindSandbox}, []string{"--no-connect"}},
		{"prod Id reporting a scratch org", "prod", providers.TargetInfo{ID: rtProdID, Kind: providers.KindScratch}, []string{"--no-connect"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := rtNew(t, rtConfig)
			e.gate(t, exitcode.OK, append([]string{"--target", tc.target}, tc.gate...)...)
			e.report = func(providers.Target) (providers.TargetInfo, error) { return tc.info, nil }
			if strings.HasPrefix(tc.name, "control:") {
				// The 15 character form of the configured Id is the same org, so this one
				// must deploy. It is the control for the case-sensitive compare above.
				if code := e.cmd("deploy", "--target", tc.target); code != exitcode.OK {
					t.Fatalf("exit %d. output:\n%s", code, e.out)
				}
				return
			}
			if code := e.refuse(t, false, "--target", tc.target); code != exitcode.Error {
				t.Fatalf("exit %d, want 2", code)
			}
			if !e.fp.allClosed() {
				t.Fatal("a session was left open")
			}
		})
	}
}

func TestRTGuardInGateDeniesAndGivesNoEphemeralPermit(t *testing.T) {
	// A prod target whose login reports a scratch org must not get the automatic ephemeral
	// PERMIT, which skips every rule.
	// The synthetic rule passes, so only the guard can deny here.
	e := rtNew(t, rtConfig)
	e.report = func(providers.Target) (providers.TargetInfo, error) {
		return providers.TargetInfo{ID: rtProdID, Kind: providers.KindScratch}, nil
	}
	e.gate(t, exitcode.Error, "--target", "prod")
	p := e.readPermit(t)
	if p.Allow {
		t.Fatal("allowed a prod target that reported a scratch org")
	}
	for _, c := range p.Checks {
		if c.Rule == "ephemeral" {
			t.Fatal("an automatic ephemeral PERMIT was issued for a prod target")
		}
	}
	e.refuse(t, true, "--target", "prod")
}

func TestRTGuardTargetWithoutID(t *testing.T) {
	cfg := strings.Replace(rtConfig, "    id: "+rtQAID+"\n", "", 1)
	e := rtNew(t, cfg)
	e.gate(t, exitcode.OK, "--target", "qa", "--no-connect")
	if id := e.readPermit(t).Target.ID; id != "" {
		t.Fatalf("permit target id %q", id)
	}
	e.refuse(t, true, "--target", "qa")
	// With a login, the guard itself refuses a target with no configured Id.
	e2 := rtNew(t, cfg)
	e2.gate(t, exitcode.Error, "--target", "qa")
	if e2.readPermit(t).Allow {
		t.Fatal("gate allowed a target with no configured Id")
	}
}

// Attack 8: other paths found reading the code.

func TestRTStalePermitIsRemovedByAFailedGate(t *testing.T) {
	e := rtNew(t, rtConfig)
	e.gate(t, exitcode.OK, "--target", "qa")
	// The next gate stops before it can write a PERMIT. The earlier allow must not survive.
	if err := os.Remove(filepath.Join(e.dir, ".glassops", "plan.json")); err != nil {
		t.Fatal(err)
	}
	e.gate(t, exitcode.Error, "--target", "qa")
	if _, err := os.Stat(e.permitPath()); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("the earlier PERMIT survived a failed gate: %v", err)
	}
	rtWritePlan(t, e.harness, ".glassops/plan", rtManifest, nil)
	e.refuse(t, true, "--target", "qa")
}

func TestRTGateWithEditedPlanFilesDenies(t *testing.T) {
	// The plan files changed after plan wrote them and before gate: gate must deny.
	e := rtNew(t, rtConfig)
	e.write(t, filepath.FromSlash(rtPlanRel), rtManifest+"<!-- rt -->\n")
	e.gate(t, exitcode.Error, "--target", "qa")
	if e.readPermit(t).Allow {
		t.Fatal("gate allowed plan files edited after plan")
	}
	e.refuse(t, true, "--target", "qa")
}

func TestRTQuickDeployUsesOnlyTheSignedValidation(t *testing.T) {
	// A validate result for another plan must not be bound, and a bound one is the only Id
	// deploy may quick deploy.
	e := rtNew(t, rtConfig)
	r := run.New("validate", nil)
	r.Target = "qa"
	r.Facts["salesforce.deploy_id"] = "0Af00000000RT02AAA"
	r.Facts["salesforce.check_only"] = true
	r.Facts["validate.plan_sha256"] = strings.Repeat("a", 64)
	writeResult(t, e.harness, r)
	e.gate(t, exitcode.OK, "--target", "qa")
	if v := e.readPermit(t).Validation; v != nil {
		t.Fatalf("a validation of another plan was bound: %+v", v)
	}
	if code := e.cmd("deploy", "--target", "qa"); code != exitcode.OK {
		t.Fatalf("exit %d. output:\n%s", code, e.out)
	}
	if got := e.opts[0].ValidatedID; got != "" {
		t.Fatalf("quick deployed %q with no bound validation", got)
	}
}

// Findings fixed in the hardening round. Each test was an open attack before the fix.

// rtEditPlan rewrites .glassops/plan.json without touching the plan files, so plan_sha256
// and the files still agree.
func rtEditPlan(t *testing.T, e *rtEnv, edit func(cs *providers.ChangeSet)) {
	t.Helper()
	path := filepath.Join(e.dir, ".glassops", "plan.json")
	cs, err := providers.ReadChangeSet(path)
	if err != nil {
		t.Fatal(err)
	}
	edit(&cs)
	if err := providers.WriteChangeSet(path, cs); err != nil {
		t.Fatal(err)
	}
}

// Finding F1, now refused: plan.json fields outside the plan files were not bound by
// plan_sha256, so the destructive list could be changed after gate under a PERMIT that still
// verified. Here the
// deploy manifest itself is listed as a post-destructive manifest.
func TestRTDestructiveAddedToPlanAfterGate(t *testing.T) {
	e := rtNew(t, rtConfig)
	e.gate(t, exitcode.OK, "--target", "qa")
	rtEditPlan(t, e, func(cs *providers.ChangeSet) { cs.Destructive = []string{rtPlanRel} })
	e.refuse(t, true, "--target", "qa")
}

// Finding F2, now refused (owner decision: clean tree at the commit): the source the
// manifest names was not bound by the PERMIT. Only the manifests are hashed, and the commit is
// read from the CI, not from the tree, so a source file changed between gate and deploy was
// deployed.
func TestRTSourceChangedAfterGate(t *testing.T) {
	e := rtNew(t, rtConfig)
	e.gate(t, exitcode.OK, "--target", "qa")
	e.write(t, rtSourcePath, "public class Foo { /* rt: changed after gate */ }\n")
	e.refuse(t, true, "--target", "qa")
}

const rtSoDConfig = `version: 1
targets:
  qa:
    id: 00D000000000002AAA
    tier: qa
    auth: { method: sfdx-url, url: $SF_QA_AUTH_URL }
policy:
  rules:
    separation_of_duties:
      tiers: [qa]
`

// Finding F3, now refused: gate copied facts from results/plan.json, results/validate.json and the facts in
// plan.json, then sets the ci facts only when it has a value for them. A ci fact it has no
// value for keeps whatever an unsigned file said. On a push with no pull request found, or
// with no API (Bitbucket, GitLab, no token), an injected ci.pr_author or ci.approvals
// satisfies separation of duties for an author deploying their own change.
func TestRTInjectedCIFactsDoNotSatisfySoD(t *testing.T) {
	cases := []struct {
		name   string
		cfg    string
		setup  func(t *testing.T, e *rtEnv)
		inject func(t *testing.T, e *rtEnv)
	}{
		{"ci.pr_author in results/plan.json", rtSoDConfig, nil, func(t *testing.T, e *rtEnv) {
			r := run.New("plan", nil)
			r.Facts["ci.pr_author"] = "someone-else"
			writeResult(t, e.harness, r)
		}},
		{"ci.pr_author in results/plan.json, GitHub API on, a push with no pull request", rtSoDConfig,
			func(t *testing.T, e *rtEnv) {
				e.env["GITHUB_SHA"] = unknownSHA
				delete(e.env, "GITHUB_RUN_ID")
				e.useAPI(newGitHubServer(t, defaultRoutes()))
			},
			func(t *testing.T, e *rtEnv) {
				r := run.New("plan", nil)
				r.Facts["ci.pr_author"] = "someone-else"
				writeResult(t, e.harness, r)
			}},
		{"ci.pr_author in the facts of plan.json", rtSoDConfig, nil, func(t *testing.T, e *rtEnv) {
			rtEditPlan(t, e, func(cs *providers.ChangeSet) { cs.Facts["ci.pr_author"] = "someone-else" })
		}},
		{"ci.approvals in results/validate.json", strings.Replace(rtSoDConfig, "tiers: [qa]", "tiers: [qa]\n      approval: { environment: glassops-deploy }", 1),
			nil, func(t *testing.T, e *rtEnv) {
				r := run.New("validate", nil)
				r.Status = run.StatusFail // not bound for a quick deploy, its facts are still read
				r.Facts["ci.approvals"] = []any{map[string]any{
					"mechanism": "environment", "environment": "glassops-deploy", "state": "approved", "approver": "someone-else",
				}}
				writeResult(t, e.harness, r)
			}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := rtNew(t, tc.cfg)
			rtUseRules(t, policy.Default)
			if tc.setup != nil {
				tc.setup(t, e)
			}
			// The control: without the injected fact the author deploying their own push is
			// denied.
			e.gate(t, exitcode.Fail, "--target", "qa")
			tc.inject(t, e)
			code := e.cmd("gate", "--target", "qa")
			if p := e.readPermit(t); p.Allow || code == exitcode.OK {
				t.Fatalf("gate exit %d, allow %v: an unsigned file supplied separation of duties evidence. output:\n%s", code, p.Allow, e.out)
			}
		})
	}
}

const rtDestructiveConfig = `version: 1
targets:
  qa:
    id: 00D000000000002AAA
    tier: qa
    auth: { method: sfdx-url, url: $SF_QA_AUTH_URL }
policy:
  rules:
    destructive_needs_label:
      label: destructive-ok
      tiers: [qa]
`

// Finding F4, now refused: change.destructive_count came from the provider's count in plan.json facts, which
// plan_sha256 does not cover, and those facts are merged over the facts gate derives. A plan
// that sends a destructive manifest can tell gate it has none.
func TestRTDestructiveCountNotOverriddenByPlanFacts(t *testing.T) {
	for _, key := range []string{"salesforce.destructive_count", "change.destructive_count"} {
		t.Run(key, func(t *testing.T) {
			e := rtNew(t, rtDestructiveConfig)
			rtUseRules(t, policy.Default)
			rtWritePlan(t, e.harness, ".glassops/plan", rtManifest, map[string]string{
				"destructiveChanges.xml": "<Package><types><members>Bar</members><name>ApexClass</name></types></Package>\n",
			})
			// The control: a destructive plan with no pull request label is denied.
			e.gate(t, exitcode.Fail, "--target", "qa")
			rtEditPlan(t, e, func(cs *providers.ChangeSet) { cs.Facts[key] = 0 })
			code := e.cmd("gate", "--target", "qa")
			if p := e.readPermit(t); p.Allow || code == exitcode.OK {
				t.Fatalf("gate exit %d, allow %v for a plan with a destructive manifest and no label. output:\n%s", code, p.Allow, e.out)
			}
		})
	}
}

// Finding F5, now refused: finish printed every check detail to stdout as it is, and GitHub reads any stdout
// line that starts with :: as a workflow command. A target name, or any provider message,
// holding a newline injects commands such as ::warning::, ::add-mask:: or ::stop-commands::.
// The step summary gets the same text and its table is broken the same way.
func TestRTWorkflowCommandInjectionThroughStdout(t *testing.T) {
	e := rtNew(t, rtConfig)
	summary := filepath.Join(e.dir, "rt-summary.md")
	e.env["GITHUB_STEP_SUMMARY"] = summary
	e.gate(t, exitcode.OK, "--target", "qa")
	e.refuse(t, true, "--target", "qa\n::warning title=rt::rt-forged\n# rt-forged heading")
	for _, line := range strings.Split(e.out.String(), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "::warning title=rt::") {
			t.Fatalf("an input became a workflow command on its own stdout line: %q", line)
		}
	}
	data, _ := os.ReadFile(summary)
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "# rt-forged") {
			t.Fatalf("an input broke out of the summary table: %q", line)
		}
	}
}

// Finding F6, now refused: --resume took any job Id. The PERMIT verifies for this plan, then deploy reports
// on a job the PERMIT never named, and when that job succeeded it moves the deployed tag to
// this commit. The next plan then starts from a commit whose changes may never have been
// deployed.
func TestRTResumeOfAJobThePermitNeverNamed(t *testing.T) {
	e := rtNew(t, rtConfig)
	e.gate(t, exitcode.OK, "--target", "qa")
	e.refuse(t, true, "--target", "qa", "--resume", "0Af00000000RT99AAA")
}

const rtWindowConfig = `version: 1
targets:
  qa:
    id: 00D000000000002AAA
    tier: qa
    auth: { method: sfdx-url, url: $SF_QA_AUTH_URL }
policy:
  windows:
    qa:
      - { days: [mon], from: "09:00", to: "12:30", tz: UTC }
`

// Finding F7, now refused (owner decision: gate caps expires_at at the window end): deploy
// honoured a PERMIT for its whole lifetime and never looks at the clock rules
// again. A PERMIT issued inside the deploy window deploys after the window closes, for up to
// policy.permit_ttl (24 hours at most).
func TestRTDeployAfterTheWindowClosed(t *testing.T) {
	e := rtNew(t, rtWindowConfig)
	rtUseRules(t, policy.Default)
	e.gate(t, exitcode.OK, "--target", "qa") // Monday 12:00 UTC, inside the window
	e.app.Now = func() time.Time { return rtGateTime.Add(time.Hour) }
	// The control: gate itself refuses at 13:00.
	saved, err := os.ReadFile(e.permitPath())
	if err != nil {
		t.Fatal(err)
	}
	e.gate(t, exitcode.Fail, "--target", "qa")
	if err := os.WriteFile(e.permitPath(), saved, 0o644); err != nil {
		t.Fatal(err)
	}
	e.refuse(t, true, "--target", "qa")
}

// Finding F8, now refused: a PERMIT from the future verified. Verify checks expires_at only, so a PERMIT whose
// issued_at is after the deploy clock, from a gate with a skewed clock, is honoured for its
// whole lifetime plus the skew.
func TestRTPermitFromTheFuture(t *testing.T) {
	e := rtNew(t, rtConfig)
	e.app.Now = func() time.Time { return rtGateTime.Add(6 * time.Hour) }
	e.gate(t, exitcode.OK, "--target", "qa")
	e.app.Now = func() time.Time { return rtGateTime }
	e.refuse(t, true, "--target", "qa")
}
