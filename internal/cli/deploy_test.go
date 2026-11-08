package cli

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	osexec "os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/permit"
	"github.com/glassops-platform/glassops/internal/core/run"
	"github.com/glassops-platform/glassops/internal/providers"
	execprovider "github.com/glassops-platform/glassops/internal/providers/exec"
)

const (
	deployCommitSHA = "5555555555555555555555555555555555555555"
	deployValidated = "0Af000000000004AAA"
	deployNewID     = "0Af000000000005AAA"
	deployTagCmd    = "git tag -f glassops/deployed/qa " + deployCommitSHA
	deployPushCmd   = "git push -f origin refs/tags/glassops/deployed/qa"
	deployStatusCmd = "git status --porcelain --untracked-files=no"
)

// deployKey is a synthetic 32 byte key, base64 as GLASSOPS_PERMIT_KEY holds it.
var deployKey = base64.StdEncoding.EncodeToString([]byte("synthetic-test-key-32-bytes-long"))

// deployFakeSigner is a synthetic Signer for these tests. It stands in for the HMAC signer until
// the permit package can sign.
type deployFakeSigner struct{ key []byte }

func (s deployFakeSigner) Alg() string { return "FAKE-HMAC-SHA256" }
func (s deployFakeSigner) Sign(c []byte) (string, error) {
	m := hmac.New(sha256.New, s.key)
	m.Write(c)
	return base64.StdEncoding.EncodeToString(m.Sum(nil)), nil
}
func (s deployFakeSigner) Verify(c []byte, v string) error {
	want, _ := s.Sign(c)
	if !hmac.Equal([]byte(want), []byte(v)) {
		return permit.ErrSignature
	}
	return nil
}

// deployFakeCanonical is the synthetic canonical form: the PERMIT as JSON with the signature value
// emptied.
func deployFakeCanonical(p permit.Permit) []byte {
	p.Signature.Value = ""
	b, _ := json.Marshal(p)
	return b
}

// deployFakeVerify checks a PERMIT the way the permit package is specified to, in the same
// order, returning the same sentinel errors. Like the real Verify, it compares the PERMIT's
// target Id in canonical form, here the fake provider's 15 character form.
func deployFakeVerify(p permit.Permit, s permit.Signer, want permit.Expect) error {
	if err := s.Verify(deployFakeCanonical(p), p.Signature.Value); err != nil {
		return fmt.Errorf("%w: %v", permit.ErrSignature, err)
	}
	switch {
	case !p.Allow:
		return permit.ErrDenied
	case !want.Now.Before(p.ExpiresAt):
		return permit.ErrExpired
	case p.Target.Name != want.TargetName, (&fakeProvider{}).CanonicalID(p.Target.ID) != want.TargetID, p.Commit != want.Commit,
		p.PlanSHA256 != want.PlanSHA256, p.ConfigSHA256 != want.ConfigSHA256, p.ChangeSetSHA256 != want.ChangeSetSHA256:
		return permit.ErrMismatch
	}
	return nil
}

// useDeployPermitFakes swaps in synthetic PERMIT functions for one test. When the permit package
// can sign and verify, the read and verify fakes can go and the real ones be used.
func useDeployPermitFakes(t *testing.T) {
	t.Helper()
	saved := permitFuncs
	permitFuncs.read = func(path string) (permit.Permit, error) {
		var p permit.Permit
		data, err := os.ReadFile(path)
		if err != nil {
			return p, err
		}
		return p, json.Unmarshal(data, &p)
	}
	permitFuncs.verify = deployFakeVerify
	permitFuncs.signer = func(key []byte) permit.Signer { return deployFakeSigner{key: key} }
	t.Cleanup(func() { permitFuncs = saved })
}

const deployConfig = `version: 1
targets:
  qa:
    id: 00D000000000002AAA
    tier: qa
    auth: { method: sfdx-url, url: $SF_QA_AUTH_URL }
salesforce:
  test_level: { qa: RunLocalTests }
`

type deployHarness struct {
	*harness
	fp        *fakeProvider
	authCalls int
	opts      providers.DeployOptions
}

// newDeployHarness sets up a project with a plan, a signed PERMIT that allows it, and a fake
// provider whose deploy passes as a full deploy.
func newDeployHarness(t *testing.T, cfg string, edit func(*permit.Permit)) *deployHarness {
	t.Helper()
	useDeployPermitFakes(t)
	h := &deployHarness{harness: newHarness(t)}
	h.write(t, "glassops.yaml", cfg)
	planSHA := h.writeDeployPlan(t, false)
	h.env["GITHUB_ACTIONS"] = "true"
	h.env["GITHUB_SHA"] = deployCommitSHA
	h.env[permit.KeyEnv] = deployKey
	h.tools[deployTagCmd] = ""
	h.tools[deployPushCmd] = ""
	// A clean checkout of the PERMIT's commit.
	h.tools["git rev-parse HEAD"] = deployCommitSHA + "\n"
	h.tools[deployStatusCmd] = ""

	h.fp = &fakeProvider{
		detect: true,
		auth: func(providers.Target) error {
			h.authCalls++
			return nil
		},
		describe: reportsAs(providers.KindSandbox),
		deploy: func(_ providers.Session, _ providers.ChangeSet, o providers.DeployOptions) (providers.DeployResult, error) {
			h.opts = o
			quick := o.ValidatedID != ""
			return providers.DeployResult{
				Passed: true, ID: deployNewID, Quick: quick, ComponentsDeployed: 4,
				Facts: map[string]any{"salesforce.deploy_id": deployNewID, "salesforce.quick": quick},
			}, nil
		},
	}
	h.app.Providers = providers.Registry{h.fp}

	loaded, err := config.Load(h.dir, "", "salesforce")
	if err != nil {
		t.Fatal(err)
	}
	cs, err := providers.ReadChangeSet(filepath.Join(h.dir, ".glassops", "plan.json"))
	if err != nil {
		t.Fatal(err)
	}
	p := permit.Permit{
		Schema: permit.Schema, ID: "a1b2c3d4", Allow: true,
		IssuedAt:     time.Date(2026, 10, 26, 11, 0, 0, 0, time.UTC),
		ExpiresAt:    time.Date(2026, 10, 26, 13, 0, 0, 0, time.UTC),
		Target:       permit.Target{Name: "qa", Tier: "qa", ID: "00D000000000002AAA", Kind: "sandbox"},
		Commit:       deployCommitSHA,
		PlanSHA256:   planSHA,
		ConfigSHA256: loaded.SHA256(),
		Validation:   &permit.Validation{DeployID: deployValidated, TestLevel: "RunLocalTests", Passed: true},
	}
	p.ChangeSetSHA256 = permit.ChangeSetDigest(cs)
	if edit != nil {
		edit(&p)
	}
	h.writePermit(t, p, true)
	return h
}

// writeDeployPlan writes a synthetic manifest and plan.json with a matching plan_sha256, so the
// plan still verifies when deploy recomputes the digest from the files. It returns the digest.
func (h *deployHarness) writeDeployPlan(t *testing.T, empty bool) string {
	t.Helper()
	const rel = ".glassops/plan/package/package.xml"
	content := "<Package/>\n"
	h.write(t, filepath.FromSlash(rel), content)
	// The plan digest formula: each file under the plan dir, sorted by path, as path, newline,
	// content. One file here.
	sum := sha256.Sum256([]byte(rel + "\n" + content))
	digest := hex.EncodeToString(sum[:])
	cs := providers.ChangeSet{
		Schema: providers.PlanSchema, Provider: "salesforce",
		Dir: ".glassops/plan", Manifest: rel,
		Destructive: []string{}, Empty: empty, SHA256: digest, Facts: map[string]any{},
	}
	if err := providers.WriteChangeSet(filepath.Join(h.dir, ".glassops", "plan.json"), cs); err != nil {
		t.Fatal(err)
	}
	return digest
}

// writePermit writes p, signed with the test key when sign is true.
func (h *deployHarness) writePermit(t *testing.T, p permit.Permit, sign bool) {
	t.Helper()
	if sign {
		key, _ := base64.StdEncoding.DecodeString(deployKey)
		p.Signature = permit.Signature{Alg: "FAKE-HMAC-SHA256"}
		p.Signature.Value, _ = deployFakeSigner{key: key}.Sign(deployFakeCanonical(p))
	}
	data, _ := json.MarshalIndent(p, "", "  ")
	h.write(t, filepath.Join(".glassops", "permit.json"), string(data))
}

func (h *deployHarness) run(args ...string) int {
	return Execute(h.app, append(append([]string{"deploy", "--target", "qa"}, args...), "--dir", h.dir, "--ci", "github"))
}

func TestDeployPass(t *testing.T) {
	h := newDeployHarness(t, deployConfig, nil)
	if code := h.run(); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if h.authCalls != 1 || h.fp.deployCalls != 1 || !h.fp.allClosed() {
		t.Fatalf("auth %d, deploy %d, closed %v", h.authCalls, h.fp.deployCalls, h.fp.allClosed())
	}
	if h.opts.ValidatedID != deployValidated || h.opts.QuickOnly || h.opts.TestLevel != "RunLocalTests" || h.opts.Dir != h.dir {
		t.Fatalf("deploy options = %+v", h.opts)
	}
	r := h.result(t, "deploy")
	if r.Status != run.StatusPass || r.Target != "qa" {
		t.Fatalf("result = %+v", r)
	}
	want := map[string]any{
		"deploy.id": deployNewID, "deploy.permit_id": "a1b2c3d4", "deploy.quick": true,
		"deploy.tag_pushed": true, "deploy.commit": deployCommitSHA, "deploy.quick_mode": "prefer",
		"salesforce.deploy_id": deployNewID,
	}
	for k, v := range want {
		if r.Facts[k] != v {
			t.Errorf("fact %s = %v, want %v", k, r.Facts[k], v)
		}
	}
	if checkStatus(r, "permit") != run.StatusPass || checkStatus(r, "tag") != run.StatusPass {
		t.Fatalf("checks = %+v", r.Checks)
	}
	if strings.Contains(h.out.String(), deployKey) {
		t.Fatal("the PERMIT key reached the output")
	}
}

func TestDeployPermitProblemsAreExit2WithoutLogin(t *testing.T) {
	cases := []struct {
		name  string
		setup func(t *testing.T, h *deployHarness)
		want  string
	}{
		{"no PERMIT", func(t *testing.T, h *deployHarness) {
			if err := os.Remove(filepath.Join(h.dir, ".glassops", "permit.json")); err != nil {
				t.Fatal(err)
			}
		}, "read PERMIT"},
		{"no key", func(_ *testing.T, h *deployHarness) { delete(h.env, permit.KeyEnv) }, "GLASSOPS_PERMIT_KEY is not set"},
		{"key not base64", func(_ *testing.T, h *deployHarness) { h.env[permit.KeyEnv] = "not base64!" }, "not valid base64"},
		{"key too short", func(_ *testing.T, h *deployHarness) {
			h.env[permit.KeyEnv] = base64.StdEncoding.EncodeToString([]byte("short"))
		}, "at least 32"},
		{"wrong key", func(_ *testing.T, h *deployHarness) {
			h.env[permit.KeyEnv] = base64.StdEncoding.EncodeToString([]byte("another-synthetic-key-of-32-byte"))
		}, "signature"},
		{"tampered after signing", func(t *testing.T, h *deployHarness) {
			path := filepath.Join(h.dir, ".glassops", "permit.json")
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var p permit.Permit
			if err := json.Unmarshal(data, &p); err != nil {
				t.Fatal(err)
			}
			p.Validation.DeployID = "0Af000000000009AAA"
			h.writePermit(t, p, false)
		}, "signature"},
		{"denied", func(t *testing.T, h *deployHarness) {
			h.rewritePermit(t, func(p *permit.Permit) { p.Allow = false })
		}, "denies"},
		{"expired", func(t *testing.T, h *deployHarness) {
			h.rewritePermit(t, func(p *permit.Permit) { p.ExpiresAt = time.Date(2026, 10, 26, 11, 30, 0, 0, time.UTC) })
		}, "expired"},
		{"other commit", func(_ *testing.T, h *deployHarness) { h.env["GITHUB_SHA"] = planHead }, "does not match"},
		{"other target", func(t *testing.T, h *deployHarness) {
			h.rewritePermit(t, func(p *permit.Permit) { p.Target.Name = "prod" })
		}, "does not match"},
		{"config changed", func(t *testing.T, h *deployHarness) {
			h.write(t, "glassops.yaml", deployConfig+"  deploy: { quick: off }\n")
		}, "does not match"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := newDeployHarness(t, deployConfig, nil)
			tc.setup(t, h)
			if code := h.run(); code != exitcode.Error {
				t.Fatalf("exit %d, want 2. output:\n%s", code, h.out)
			}
			if h.authCalls != 0 || h.fp.deployCalls != 0 {
				t.Fatalf("logged in %d times and deployed %d times with a bad PERMIT", h.authCalls, h.fp.deployCalls)
			}
			r := h.result(t, "deploy")
			if r.Status != run.StatusError || checkStatus(r, "permit") != run.StatusError {
				t.Fatalf("result = %+v", r)
			}
			if !strings.Contains(h.out.String(), tc.want) {
				t.Fatalf("output lacks %q:\n%s", tc.want, h.out)
			}
			if strings.Contains(h.out.String(), deployKey) {
				t.Fatal("the PERMIT key reached the output")
			}
		})
	}
}

// rewritePermit edits the PERMIT on disk and signs it again, so only the edited field is wrong.
func (h *deployHarness) rewritePermit(t *testing.T, edit func(*permit.Permit)) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(h.dir, ".glassops", "permit.json"))
	if err != nil {
		t.Fatal(err)
	}
	var p permit.Permit
	if err := json.Unmarshal(data, &p); err != nil {
		t.Fatal(err)
	}
	edit(&p)
	h.writePermit(t, p, true)
}

func TestDeployFailIsExit1(t *testing.T) {
	h := newDeployHarness(t, deployConfig, nil)
	h.fp.deploy = func(providers.Session, providers.ChangeSet, providers.DeployOptions) (providers.DeployResult, error) {
		return providers.DeployResult{
			ID: "0Af000000000002AAA", ComponentsDeployed: 2,
			Problems: []string{"ApexClass Greeter: Missing '<EOF>' at 'public' (line 7)"},
			Facts:    map[string]any{"salesforce.deploy_status": "Failed"},
		}, nil
	}
	if code := h.run(); code != exitcode.Fail {
		t.Fatalf("exit %d, want 1. output:\n%s", code, h.out)
	}
	r := h.result(t, "deploy")
	if r.Status != run.StatusFail || checkStatus(r, "problem.1") != run.StatusFail || r.Facts["deploy.id"] != "0Af000000000002AAA" {
		t.Fatalf("result = %+v", r)
	}
	if _, ok := r.Facts["deploy.tag_pushed"]; ok {
		t.Fatal("the tag was touched after a failed deploy")
	}
	if !h.fp.allClosed() {
		t.Fatal("session left open")
	}
}

func TestDeployProviderErrorIsExit2(t *testing.T) {
	h := newDeployHarness(t, deployConfig, nil)
	h.fp.deploy = func(providers.Session, providers.ChangeSet, providers.DeployOptions) (providers.DeployResult, error) {
		return providers.DeployResult{}, errors.New("sf project deploy quick: sf:UNKNOWN_EXCEPTION: invalid parameter value")
	}
	if code := h.run(); code != exitcode.Error {
		t.Fatalf("exit %d, want 2", code)
	}
	if r := h.result(t, "deploy"); checkStatus(r, "deploy") != run.StatusError {
		t.Fatalf("result = %+v", r)
	}
}

func TestDeployTimeoutKeepsJobID(t *testing.T) {
	h := newDeployHarness(t, deployConfig, nil)
	h.fp.deploy = func(_ providers.Session, _ providers.ChangeSet, o providers.DeployOptions) (providers.DeployResult, error) {
		h.opts = o
		return providers.DeployResult{ID: "0Af000000000003AAA", TimedOut: true, Facts: map[string]any{}}, nil
	}
	if code := h.run("--wait", "1m"); code != exitcode.Error {
		t.Fatalf("exit %d, want 2", code)
	}
	r := h.result(t, "deploy")
	if r.Facts["deploy.id"] != "0Af000000000003AAA" || r.Facts["deploy.timed_out"] != true {
		t.Fatalf("facts = %v", r.Facts)
	}
	if !strings.Contains(h.out.String(), "--resume 0Af000000000003AAA") {
		t.Fatalf("output gives no resume command:\n%s", h.out)
	}
	if h.opts.Wait != time.Minute {
		t.Fatalf("wait = %s", h.opts.Wait)
	}
}

// writeTimedOutDeploy writes the deploy result a run that timed out on job id under the
// harness PERMIT leaves behind.
func (h *deployHarness) writeTimedOutDeploy(t *testing.T, target, id, permitID string, timedOut bool) {
	t.Helper()
	r := run.New("deploy", h.app.Now)
	r.Target = target
	r.Facts["deploy.id"] = id
	r.Facts["deploy.permit_id"] = permitID
	r.Facts["deploy.timed_out"] = timedOut
	r.Degrade(run.StatusError)
	if _, err := r.Write(filepath.Join(h.dir, ".glassops", "results"), h.app.Now); err != nil {
		t.Fatal(err)
	}
}

func TestDeployResumePassesJobID(t *testing.T) {
	h := newDeployHarness(t, deployConfig+"  deploy: { quick: require }\n", nil)
	h.writeTimedOutDeploy(t, "qa", "0Af000000000003AAA", "a1b2c3d4", true)
	if code := h.run("--resume", "0Af000000000003AAA"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if h.opts.ResumeID != "0Af000000000003AAA" || h.opts.ValidatedID != "" {
		t.Fatalf("deploy options = %+v", h.opts)
	}
}

// The provider falls back to a full deploy when the org refuses the quick deploy. The CLI
// passes the validated Id without QuickOnly and reports what the provider did.
func TestDeployQuickRefusedFallsBack(t *testing.T) {
	h := newDeployHarness(t, deployConfig, nil)
	h.fp.deploy = func(_ providers.Session, _ providers.ChangeSet, o providers.DeployOptions) (providers.DeployResult, error) {
		h.opts = o
		return providers.DeployResult{
			Passed: true, ID: "0Af000000000001AAA", ComponentsDeployed: 4,
			Facts: map[string]any{"salesforce.quick": false, "salesforce.quick_refused": true},
		}, nil
	}
	if code := h.run(); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if h.opts.ValidatedID != deployValidated || h.opts.QuickOnly {
		t.Fatalf("deploy options = %+v", h.opts)
	}
	r := h.result(t, "deploy")
	if r.Facts["deploy.quick"] != false || r.Facts["salesforce.quick_refused"] != true || r.Facts["deploy.tag_pushed"] != true {
		t.Fatalf("facts = %v", r.Facts)
	}
}

func TestDeployQuickRequire(t *testing.T) {
	cfg := deployConfig + "  deploy: { quick: require }\n"

	// require sets QuickOnly, so a refusal is the provider's failure and exits 1.
	h := newDeployHarness(t, cfg, nil)
	h.fp.deploy = func(_ providers.Session, _ providers.ChangeSet, o providers.DeployOptions) (providers.DeployResult, error) {
		h.opts = o
		return providers.DeployResult{
			Problems: []string{"quick deploy of validation " + deployValidated + " refused. A quick deploy is required, so no full deploy was run."},
			Facts:    map[string]any{"salesforce.quick_refused": true},
		}, nil
	}
	if code := h.run(); code != exitcode.Fail {
		t.Fatalf("exit %d, want 1. output:\n%s", code, h.out)
	}
	if !h.opts.QuickOnly || h.opts.ValidatedID != deployValidated {
		t.Fatalf("deploy options = %+v", h.opts)
	}

	// require with no passed validation in the PERMIT fails before any login.
	h = newDeployHarness(t, cfg, func(p *permit.Permit) { p.Validation = nil })
	if code := h.run(); code != exitcode.Fail {
		t.Fatalf("exit %d, want 1. output:\n%s", code, h.out)
	}
	if h.authCalls != 0 || h.fp.deployCalls != 0 {
		t.Fatal("require without a validation still logged in")
	}
	if r := h.result(t, "deploy"); checkStatus(r, "quick") != run.StatusFail {
		t.Fatalf("checks = %+v", r.Checks)
	}
}

func TestDeployQuickOffIgnoresValidation(t *testing.T) {
	h := newDeployHarness(t, deployConfig+"  deploy: { quick: off }\n", nil)
	if code := h.run(); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if h.opts.ValidatedID != "" {
		t.Fatalf("quick off passed a validated Id: %+v", h.opts)
	}
}

func TestDeployFailedValidationIsNotQuickDeployed(t *testing.T) {
	h := newDeployHarness(t, deployConfig, func(p *permit.Permit) { p.Validation.Passed = false })
	if code := h.run(); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if h.opts.ValidatedID != "" {
		t.Fatalf("a failed validation was quick deployed: %+v", h.opts)
	}
}

func TestDeployTagPushFailureWarns(t *testing.T) {
	h := newDeployHarness(t, deployConfig, nil)
	delete(h.tools, deployPushCmd)
	if code := h.run(); code != exitcode.OK {
		t.Fatalf("exit %d, want 0 after a deploy whose tag push failed. output:\n%s", code, h.out)
	}
	r := h.result(t, "deploy")
	if r.Status != run.StatusPass || r.Facts["deploy.tag_pushed"] != false || checkStatus(r, "tag") != run.StatusSkipped {
		t.Fatalf("result = %+v", r)
	}
	if !strings.Contains(h.out.String(), "::warning") || !strings.Contains(h.out.String(), "tag was not updated") {
		t.Fatalf("no warning in output:\n%s", h.out)
	}
}

func TestDeployGuardMismatchIsExit2(t *testing.T) {
	h := newDeployHarness(t, deployConfig, nil)
	h.fp.describe = func(providers.Target) (providers.TargetInfo, error) {
		return providers.TargetInfo{ID: "00D000000000009AAA", Kind: providers.KindSandbox, Facts: map[string]any{}}, nil
	}
	if code := h.run(); code != exitcode.Error {
		t.Fatalf("exit %d, want 2", code)
	}
	if h.fp.deployCalls != 0 || !h.fp.allClosed() {
		t.Fatal("deployed after a guard mismatch, or left a session open")
	}
}

func TestDeployEmptyPlanSkipsLogin(t *testing.T) {
	h := newDeployHarness(t, deployConfig, nil)
	h.writeDeployPlan(t, true)
	// The PERMIT binds the empty flag, so it is issued for the empty plan.
	cs, err := providers.ReadChangeSet(filepath.Join(h.dir, ".glassops", "plan.json"))
	if err != nil {
		t.Fatal(err)
	}
	h.rewritePermit(t, func(p *permit.Permit) { p.ChangeSetSHA256 = permit.ChangeSetDigest(cs) })
	if code := h.run(); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	if h.authCalls != 0 {
		t.Fatal("an empty plan logged in")
	}
	if r := h.result(t, "deploy"); r.Status != run.StatusEmpty {
		t.Fatalf("status = %s", r.Status)
	}
}

// The exec provider through the real commands, the real git, the real PERMIT signer and the
// real verifier. Only the exec commands themselves are faked, by execShell.

// execShell stands in for sh -c. It records each command with its environment and runs the
// Go function registered for it. An unknown command cannot start.
type execShell struct {
	cmds map[string]func(env map[string]string) int
	runs []execRun
}

type execRun struct {
	command string
	dir     string
	env     map[string]string
}

func (s *execShell) run(_ context.Context, dir string, env []string, command string, stdout, _ io.Writer) (int, error) {
	m := map[string]string{}
	for _, kv := range env {
		k, v, _ := strings.Cut(kv, "=")
		m[k] = v
	}
	s.runs = append(s.runs, execRun{command: command, dir: dir, env: m})
	c, ok := s.cmds[command]
	if !ok {
		return -1, errors.New("cannot start")
	}
	if command == "./describe.sh" {
		_, _ = io.WriteString(stdout, m["DESCRIBE_OUT"])
	}
	return c(m), nil
}

func (s *execShell) ran(command string) bool {
	for _, r := range s.runs {
		if r.command == command {
			return true
		}
	}
	return false
}

const execConfig = `version: 1
provider: exec
exec:
  paths: [app]
  deploy: ./deploy.sh
targets:
  qa:
    id: $QA_TARGET_ID
    tier: qa
  prod:
    id: $PROD_TARGET_ID
    tier: prod
`

// execKey is a synthetic 32 byte key, base64 as GLASSOPS_PERMIT_KEY holds it.
var execKey = base64.StdEncoding.EncodeToString([]byte("synthetic-exec-key-32-bytes-long"))

type execProject struct {
	*harness
	sh   *execShell
	base string
	head string
}

func gitIn(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := osexec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v: %s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out))
}

// newExecProject makes a git repository with two commits. The second changes app/a.txt,
// adds app/b.txt, deletes app/old.txt and changes other/c.txt, which is outside exec.paths.
func newExecProject(t *testing.T, cfg string) *execProject {
	t.Helper()
	if _, err := osexec.LookPath("git"); err != nil {
		t.Skip("git is not on the PATH")
	}
	h := newHarness(t)
	p := &execProject{harness: h, sh: &execShell{cmds: map[string]func(map[string]string) int{}}}
	gitIn(t, h.dir, "init", "-q", "-b", "main")
	gitIn(t, h.dir, "config", "user.email", "test@example.invalid")
	gitIn(t, h.dir, "config", "user.name", "test")
	gitIn(t, h.dir, "config", "core.autocrlf", "false")
	h.write(t, "app/a.txt", "one\n")
	h.write(t, "app/old.txt", "old\n")
	h.write(t, "other/c.txt", "one\n")
	h.write(t, ".gitignore", ".glassops/\n")
	gitIn(t, h.dir, "add", "-A")
	gitIn(t, h.dir, "commit", "-q", "-m", "one")
	p.base = gitIn(t, h.dir, "rev-parse", "HEAD")
	h.write(t, "app/a.txt", "two\n")
	h.write(t, "app/b.txt", "new\n")
	h.write(t, "other/c.txt", "two\n")
	if err := os.Remove(filepath.Join(h.dir, "app", "old.txt")); err != nil {
		t.Fatal(err)
	}
	h.write(t, "glassops.yaml", cfg)
	gitIn(t, h.dir, "add", "-A")
	gitIn(t, h.dir, "commit", "-q", "-m", "two")
	p.head = gitIn(t, h.dir, "rev-parse", "HEAD")

	h.app.Run = func(dir, name string, args ...string) ([]byte, error) {
		// The deployed tag push has no remote to go to, which deploy reports as a warning.
		cmd := osexec.Command(name, args...)
		cmd.Dir = dir
		return cmd.Output()
	}
	realGit := func(ctx context.Context, dir string, args ...string) ([]byte, error) {
		cmd := osexec.CommandContext(ctx, "git", args...)
		cmd.Dir = dir
		return cmd.Output()
	}
	base := []string{"PATH=/bin", permit.KeyEnv + "=" + execKey, "TOOL_TOKEN=tool"}
	h.app.Providers = providers.Registry{execprovider.NewWithRunners(p.sh.run, realGit, base)}
	h.env[permit.KeyEnv] = execKey
	// schemas/config.json takes only a Salesforce org Id or a $VAR as a target id, so the
	// exec targets' own Ids come in through the environment.
	h.env["QA_TARGET_ID"] = "acct-qa"
	h.env["PROD_TARGET_ID"] = "acct-prod"
	return p
}

// execDash runs glassops with the harness flags before args, so a command after -- gets
// nothing appended to it.
func (h *harness) execDash(args ...string) int {
	return Execute(h.app, append([]string{args[0], "--dir", h.dir, "--ci", "local"}, args[1:]...))
}

func (p *execProject) plan(t *testing.T) providers.ChangeSet {
	t.Helper()
	if code := p.exec("plan", "--from", p.base, "--to", p.head); code != exitcode.OK {
		t.Fatalf("plan exit %d: %s", code, p.out.String())
	}
	cs, err := providers.ReadChangeSet(filepath.Join(p.dir, ".glassops", "plan.json"))
	if err != nil {
		t.Fatal(err)
	}
	return cs
}

// signPermit writes a PERMIT that allows cs to target, signed with the real HMAC signer.
func (p *execProject) signPermit(t *testing.T, cs providers.ChangeSet, name, id, tier string) permit.Permit {
	t.Helper()
	loaded, err := config.Load(p.dir, "", "exec")
	if err != nil {
		t.Fatal(err)
	}
	key, err := permit.ParseKey(execKey)
	if err != nil {
		t.Fatal(err)
	}
	now := p.app.Now()
	pm := permit.Permit{
		Schema: permit.Schema, ID: "synthetic-permit", Allow: true,
		IssuedAt: now, ExpiresAt: now.Add(time.Hour),
		Target: permit.Target{Name: name, ID: id, Tier: tier, Kind: string(providers.KindProduction)},
		Commit: p.head, PlanSHA256: cs.SHA256, ConfigSHA256: loaded.SHA256(),
		ChangeSetSHA256: permit.ChangeSetDigest(cs),
	}
	if err := permit.Sign(&pm, permit.NewHMAC(key)); err != nil {
		t.Fatal(err)
	}
	if err := permit.Write(filepath.Join(p.dir, ".glassops", "permit.json"), pm); err != nil {
		t.Fatal(err)
	}
	return pm
}

func TestExecGatePermitVerifyAndDeployWithTheRealSigner(t *testing.T) {
	p := newExecProject(t, execConfig)
	p.sh.cmds["./other.sh 'two words' plain"] = func(env map[string]string) int {
		body := `{"schema":"glassops.engine-result/v1","step":"deploy","passed":true,"id":"dep-1","components_deployed":2,"facts":{"exec.tool":"other"}}`
		_ = os.WriteFile(filepath.Join(env[execprovider.EnvResultsDir], "exec-deploy.json"), []byte(body), 0o644)
		return 0
	}
	cs := p.plan(t)

	code := p.exec("gate", "--target", "qa", "--no-connect", "--deploy-command", `["./other.sh","two words","plain"]`)
	pm, err := permit.Read(filepath.Join(p.dir, ".glassops", "permit.json"))
	if err != nil {
		t.Fatalf("gate exit %d, no PERMIT: %v: %s", code, err, p.out.String())
	}
	if pm.PlanSHA256 != cs.SHA256 || pm.Commit != p.head || pm.Target.ID != "acct-qa" || pm.Signature.Alg != permit.Algorithm {
		t.Fatalf("PERMIT = %+v", pm)
	}
	if code != exitcode.OK || !pm.Allow {
		t.Fatalf("gate exit %d, allow %v: %s", code, pm.Allow, p.out.String())
	}
	if code := p.exec("permit", "verify", "--target", "qa"); code != exitcode.OK {
		t.Fatalf("permit verify exit %d: %s", code, p.out.String())
	}

	if code := p.execDash("deploy", "--target", "qa", "--", "./other.sh", "two words", "plain"); code != exitcode.OK {
		t.Fatalf("deploy exit %d: %s", code, p.out.String())
	}
	if p.sh.ran("./deploy.sh") || !p.sh.ran("./other.sh 'two words' plain") {
		t.Fatalf("runs = %+v", p.sh.runs)
	}
	env := p.sh.runs[len(p.sh.runs)-1].env
	abs := func(parts ...string) string {
		a, _ := filepath.Abs(filepath.Join(append([]string{p.dir}, parts...)...))
		return a
	}
	for k, v := range map[string]string{
		execprovider.EnvPermit:     abs(".glassops", "permit.json"),
		execprovider.EnvPlan:       abs(".glassops", "plan.json"),
		execprovider.EnvPlanDir:    abs(".glassops", "plan"),
		execprovider.EnvTarget:     "qa",
		execprovider.EnvTargetID:   "acct-qa",
		execprovider.EnvResultsDir: abs(".glassops", "results"),
		"TOOL_TOKEN":               "tool",
	} {
		if env[k] != v {
			t.Errorf("%s = %q, want %q", k, env[k], v)
		}
	}
	if _, ok := env[permit.KeyEnv]; ok {
		t.Fatal("the PERMIT key reached the deploy command")
	}
	r := p.result(t, "deploy")
	if r.Facts["deploy.id"] != "dep-1" || r.Facts["exec.result_file"] != true || r.Facts["exec.deploy_override"] != true ||
		r.Facts["exec.tool"] != "other" || r.Facts["deploy.guard"] != "skipped" || r.Facts["deploy.permit_id"] != pm.ID {
		t.Fatalf("facts = %v", r.Facts)
	}
	if checkStatus(r, "guard") != run.StatusSkipped || checkStatus(r, "deploy") != run.StatusPass {
		t.Fatalf("checks = %+v", r.Checks)
	}
}

func TestExecDeployWithoutResultFileUsesExitCode(t *testing.T) {
	for _, c := range []struct {
		exit int
		want int
	}{{0, exitcode.OK}, {5, exitcode.Fail}} {
		t.Run(fmt.Sprint(c.exit), func(t *testing.T) {
			p := newExecProject(t, execConfig)
			p.sh.cmds["./deploy.sh"] = func(map[string]string) int { return c.exit }
			p.signPermit(t, p.plan(t), "qa", "acct-qa", "qa")
			if code := p.exec("deploy", "--target", "qa"); code != c.want {
				t.Fatalf("deploy exit %d, want %d: %s", code, c.want, p.out.String())
			}
			r := p.result(t, "deploy")
			if r.Facts["exec.result_file"] != false || r.Facts["exec.exit_code"] != float64(c.exit) || r.Facts["exec.deploy_override"] != false {
				t.Fatalf("facts = %v", r.Facts)
			}
		})
	}
}

func TestExecProdGuard(t *testing.T) {
	withDescribe := strings.Replace(execConfig, "  deploy: ./deploy.sh\n", "  deploy: ./deploy.sh\n  describe: ./describe.sh\n", 1)
	cases := []struct {
		name     string
		cfg      string
		describe string
		want     int
		detail   string
	}{
		{"prod without describe refuses", execConfig, "", exitcode.Error, "exec.describe is required for a prod tier target"},
		{"wrong describe id refuses", withDescribe, `{"id":"acct-other","kind":"production"}`, exitcode.Error, "Id does not match"},
		{"prod on a sandbox refuses", withDescribe, `{"id":"acct-prod","kind":"sandbox"}`, exitcode.Error, "needs a production target"},
		{"right describe deploys", withDescribe, `{"id":"acct-prod","kind":"production"}`, exitcode.OK, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := newExecProject(t, c.cfg)
			p.sh.cmds["./deploy.sh"] = func(map[string]string) int { return 0 }
			p.sh.cmds["./describe.sh"] = func(map[string]string) int { return 0 }
			p.signPermit(t, p.plan(t), "prod", "acct-prod", "prod")
			// The fake shell prints DESCRIBE_OUT for ./describe.sh. It comes from the
			// provider's base environment, so it is set through a provider rebuilt here.
			p.setDescribe(c.describe)
			code := p.exec("deploy", "--target", "prod")
			if code != c.want {
				t.Fatalf("deploy exit %d, want %d: %s", code, c.want, p.out.String())
			}
			if c.want != exitcode.OK {
				if p.sh.ran("./deploy.sh") {
					t.Fatal("the deploy command ran on a refused target")
				}
				if !strings.Contains(p.out.String(), c.detail) {
					t.Fatalf("output lacks %q: %s", c.detail, p.out.String())
				}
				return
			}
			if !p.sh.ran("./deploy.sh") {
				t.Fatal("the deploy command did not run")
			}
			if checkStatus(p.result(t, "deploy"), "guard") != "" {
				t.Fatal("the guard ran, so it must not be recorded as skipped")
			}
		})
	}
}

// gate never binds a workspace, so exec.describe must still find the project dir through the
// config. The guard runs in the project dir, never the process's working dir, and the PERMIT
// records what the target reported.
func TestExecGateRunsDescribeInTheProjectDir(t *testing.T) {
	withDescribe := strings.Replace(execConfig, "  deploy: ./deploy.sh\n", "  deploy: ./deploy.sh\n  describe: ./describe.sh\n", 1)
	for _, c := range []struct {
		name     string
		describe string
		allow    bool
	}{
		{"right id", `{"id":"acct-prod","kind":"production"}`, true},
		{"wrong id", `{"id":"acct-other","kind":"production"}`, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			p := newExecProject(t, withDescribe)
			p.sh.cmds["./describe.sh"] = func(map[string]string) int { return 0 }
			p.plan(t)
			p.setDescribe(c.describe)
			code := p.exec("gate", "--target", "prod")
			pm, err := permit.Read(filepath.Join(p.dir, ".glassops", "permit.json"))
			if err != nil {
				t.Fatalf("gate exit %d, no PERMIT: %v: %s", code, err, p.out.String())
			}
			if pm.Allow != c.allow {
				t.Fatalf("gate exit %d, allow %v, want %v: %s", code, pm.Allow, c.allow, p.out.String())
			}
			if !p.sh.ran("./describe.sh") {
				t.Fatal("gate did not run exec.describe")
			}
			for _, r := range p.sh.runs {
				if r.command == "./describe.sh" && r.dir != p.dir {
					t.Fatalf("exec.describe ran in %q, want the project dir %q", r.dir, p.dir)
				}
			}
			if c.allow && (pm.Target.ID != "acct-prod" || pm.Target.Kind != string(providers.KindProduction)) {
				t.Fatalf("PERMIT target = %+v", pm.Target)
			}
		})
	}
}

// setDescribe rebuilds the exec provider with DESCRIBE_OUT in its base environment.
func (p *execProject) setDescribe(out string) {
	realGit := func(ctx context.Context, dir string, args ...string) ([]byte, error) {
		cmd := osexec.CommandContext(ctx, "git", args...)
		cmd.Dir = dir
		return cmd.Output()
	}
	base := []string{"PATH=/bin", "DESCRIBE_OUT=" + out}
	p.app.Providers = providers.Registry{execprovider.NewWithRunners(p.sh.run, realGit, base)}
}

func TestExecTamperedPermitIsRefused(t *testing.T) {
	cases := []struct {
		name string
		edit func(t *testing.T, path string)
	}{
		{"target edited", func(t *testing.T, path string) { editPermitJSON(t, path, `"acct-qa"`, `"acct-prod"`) }},
		{"allow flipped", func(t *testing.T, path string) { editPermitJSON(t, path, `"allow": true`, `"allow": false`) }},
		{"plan edited after gate", func(t *testing.T, path string) {
			changes := filepath.Join(filepath.Dir(path), "plan", "changes.txt")
			if err := os.WriteFile(changes, []byte("app/a.txt\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := newExecProject(t, execConfig)
			p.sh.cmds["./deploy.sh"] = func(map[string]string) int { return 0 }
			p.signPermit(t, p.plan(t), "qa", "acct-qa", "qa")
			c.edit(t, filepath.Join(p.dir, ".glassops", "permit.json"))
			if code := p.exec("permit", "verify", "--target", "qa"); code == exitcode.OK {
				t.Fatalf("permit verify passed a tampered PERMIT: %s", p.out.String())
			}
			if code := p.exec("deploy", "--target", "qa"); code != exitcode.Error {
				t.Fatalf("deploy exit %d, want 2: %s", code, p.out.String())
			}
			if p.sh.ran("./deploy.sh") {
				t.Fatal("the deploy command ran under a tampered PERMIT")
			}
		})
	}
}

func editPermitJSON(t *testing.T, path, from, to string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), from) {
		t.Fatalf("PERMIT has no %s: %s", from, data)
	}
	if err := os.WriteFile(path, []byte(strings.Replace(string(data), from, to, 1)), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestDeployCommandNeedsExecAndADash(t *testing.T) {
	h := newDeployHarness(t, deployConfig, nil)
	if code := h.execDash("deploy", "--target", "qa", "--", "./x.sh"); code != exitcode.Error || h.authCalls != 0 {
		t.Fatalf("exit %d, logins %d: %s", code, h.authCalls, h.out.String())
	}
	if !strings.Contains(h.out.String(), "needs provider exec") {
		t.Fatalf("output: %s", h.out.String())
	}
	if code := h.run("deploy", "--target", "qa", "./x.sh"); code != exitcode.Error || h.authCalls != 0 {
		t.Fatalf("an argument without -- must be refused: exit %d", code)
	}
}

func TestShellQuote(t *testing.T) {
	cases := map[string]string{
		"./deploy.sh": "./deploy.sh",
		"--env=prod":  "--env=prod",
		"two words":   "'two words'",
		"it's":        `'it'"'"'s'`,
		"":            "''",
		"$HOME":       "'$HOME'",
		"a;b":         "'a;b'",
	}
	for in, want := range cases {
		if got := shellQuote(in); got != want {
			t.Errorf("shellQuote(%q) = %q, want %q", in, got, want)
		}
	}
}

// deploy.json records the plan and the test level the deploy used, so verify can tie the
// deploy to the plan.
func TestDeployRecordsPlanAndTestLevel(t *testing.T) {
	for _, c := range []struct {
		name  string
		cfg   string
		level string
	}{
		{"full deploy at the configured level", deployConfig + "  deploy: { quick: off }\n", "RunLocalTests"},
		{"quick deploy at the validation's level", strings.Replace(deployConfig, "{ qa: RunLocalTests }", "{ qa: NoTestRun }", 1), "RunLocalTests"},
	} {
		t.Run(c.name, func(t *testing.T) {
			h := newDeployHarness(t, c.cfg, nil)
			if code := h.run(); code != exitcode.OK {
				t.Fatalf("exit %d, output:\n%s", code, h.out)
			}
			cs, err := providers.ReadChangeSet(filepath.Join(h.dir, ".glassops", "plan.json"))
			if err != nil {
				t.Fatal(err)
			}
			r := h.result(t, "deploy")
			if r.Facts["deploy.plan_sha256"] != cs.SHA256 || r.Facts["deploy.test_level"] != c.level {
				t.Fatalf("facts = %v", r.Facts)
			}
		})
	}
}

// A failed deploy still records the plan and the level it was sent with.
func TestDeployRecordsPlanAndTestLevelOnFailure(t *testing.T) {
	h := newDeployHarness(t, deployConfig+"  deploy: { quick: off }\n", nil)
	h.fp.deploy = func(_ providers.Session, _ providers.ChangeSet, o providers.DeployOptions) (providers.DeployResult, error) {
		return providers.DeployResult{ID: deployNewID, Problems: []string{"rt"}, Facts: map[string]any{}}, nil
	}
	if code := h.run(); code != exitcode.Fail {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	r := h.result(t, "deploy")
	if r.Facts["deploy.plan_sha256"] == nil || r.Facts["deploy.plan_sha256"] == "" || r.Facts["deploy.test_level"] != "RunLocalTests" {
		t.Fatalf("facts = %v", r.Facts)
	}
}

// The PERMIT binds the command deploy -- runs for provider exec, by its sha256 only. A command
// it does not bind, or a PERMIT that binds none, is refused before the command runs. Without
// --, exec.deploy from the signed config runs as before.
func TestExecDeployCommandIsBound(t *testing.T) {
	gate := func(t *testing.T, p *execProject, extra ...string) {
		t.Helper()
		if code := p.exec(append([]string{"gate", "--target", "qa", "--no-connect"}, extra...)...); code != exitcode.OK {
			t.Fatalf("gate exit %d: %s", code, p.out.String())
		}
	}
	setup := func(t *testing.T) *execProject {
		p := newExecProject(t, execConfig)
		for _, c := range []string{"./other.sh 'two words' plain", "./other.sh 'two words'", "./other.sh two words plain", "./deploy.sh"} {
			p.sh.cmds[c] = func(map[string]string) int { return 0 }
		}
		p.plan(t)
		return p
	}
	t.Run("control: the bound command", func(t *testing.T) {
		p := setup(t)
		gate(t, p, "--deploy-command", `["./other.sh","two words","plain"]`)
		data, _ := os.ReadFile(filepath.Join(p.dir, ".glassops", "permit.json"))
		if strings.Contains(string(data), "other.sh") || !strings.Contains(string(data), "deploy_command_sha256") {
			t.Fatalf("PERMIT: %s", data)
		}
		if code := p.execDash("deploy", "--target", "qa", "--", "./other.sh", "two words", "plain"); code != exitcode.OK {
			t.Fatalf("deploy exit %d: %s", code, p.out.String())
		}
	})
	for name, argv := range map[string][]string{
		"another argument":                 {"./other.sh", "two words", "other"},
		"one argument fewer":               {"./other.sh", "two words"},
		"the same words split another way": {"./other.sh", "two", "words", "plain"},
	} {
		t.Run(name, func(t *testing.T) {
			p := setup(t)
			gate(t, p, "--deploy-command", `["./other.sh","two words","plain"]`)
			if code := p.execDash(append([]string{"deploy", "--target", "qa", "--"}, argv...)...); code != exitcode.Error {
				t.Fatalf("deploy exit %d, want 2: %s", code, p.out.String())
			}
			if len(p.sh.runs) != 0 {
				t.Fatalf("a command ran: %+v", p.sh.runs)
			}
		})
	}
	t.Run("a PERMIT that binds no command", func(t *testing.T) {
		p := setup(t)
		gate(t, p)
		if code := p.execDash("deploy", "--target", "qa", "--", "./other.sh", "two words", "plain"); code != exitcode.Error {
			t.Fatalf("deploy exit %d, want 2: %s", code, p.out.String())
		}
		if len(p.sh.runs) != 0 || !strings.Contains(p.out.String(), "binds no deploy command") {
			t.Fatalf("runs %+v, output: %s", p.sh.runs, p.out.String())
		}
	})
	t.Run("without -- the config command runs", func(t *testing.T) {
		p := setup(t)
		gate(t, p, "--deploy-command", `["./other.sh","two words","plain"]`)
		if code := p.exec("deploy", "--target", "qa"); code != exitcode.OK || !p.sh.ran("./deploy.sh") {
			t.Fatalf("deploy exit %d, runs %+v: %s", code, p.sh.runs, p.out.String())
		}
	})
	for name, value := range map[string]string{
		"not JSON":      "./other.sh plain",
		"an empty list": "[]",
		"not strings":   `["./other.sh", 1]`,
		"trailing data": `["./other.sh"] ["x"]`,
		"empty command": `[""]`,
	} {
		t.Run("gate refuses "+name, func(t *testing.T) {
			p := setup(t)
			if code := p.exec("gate", "--target", "qa", "--no-connect", "--deploy-command", value); code != exitcode.Error {
				t.Fatalf("gate exit %d, want 2: %s", code, p.out.String())
			}
			if _, err := os.Stat(filepath.Join(p.dir, ".glassops", "permit.json")); err == nil {
				t.Fatal("gate wrote a PERMIT")
			}
		})
	}
}

func TestGateDeployCommandNeedsExec(t *testing.T) {
	h := newDeployHarness(t, deployConfig, nil)
	if code := Execute(h.app, []string{"gate", "--target", "qa", "--no-connect", "--deploy-command", `["./x.sh"]`, "--dir", h.dir, "--ci", "github"}); code != exitcode.Error {
		t.Fatalf("exit %d: %s", code, h.out.String())
	}
	if !strings.Contains(h.out.String(), "needs provider exec") {
		t.Fatalf("output: %s", h.out.String())
	}
}

// F2 with the real git and the real exec provider: a tracked file changed after gate is
// refused, and the refusal names it.
func TestExecDeployRefusesAModifiedTree(t *testing.T) {
	p := newExecProject(t, execConfig)
	p.sh.cmds["./deploy.sh"] = func(map[string]string) int { return 0 }
	p.signPermit(t, p.plan(t), "qa", "acct-qa", "qa")
	p.write(t, "other/c.txt", "changed after gate\n")
	if code := p.exec("deploy", "--target", "qa"); code != exitcode.Error {
		t.Fatalf("deploy exit %d, want 2: %s", code, p.out.String())
	}
	if p.sh.ran("./deploy.sh") || !strings.Contains(p.out.String(), "other/c.txt") || strings.Contains(p.out.String(), "changed after gate") {
		t.Fatalf("runs %+v, output: %s", p.sh.runs, p.out.String())
	}
	gitIn(t, p.dir, "checkout", "-q", "--", "other/c.txt")
	if code := p.exec("deploy", "--target", "qa"); code != exitcode.OK {
		t.Fatalf("deploy of the clean tree exit %d: %s", code, p.out.String())
	}
}
