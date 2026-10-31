package cli

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/permit"
	"github.com/glassops-platform/glassops/internal/core/run"
	"github.com/glassops-platform/glassops/internal/providers"
)

const (
	deployCommitSHA = "5555555555555555555555555555555555555555"
	deployValidated = "0Af000000000004AAA"
	deployNewID     = "0Af000000000005AAA"
	deployTagCmd    = "git tag -f glassops/deployed/qa " + deployCommitSHA
	deployPushCmd   = "git push -f origin refs/tags/glassops/deployed/qa"
)

// deployKey is a synthetic 32 byte key, base64 as GLASSOPS_PERMIT_KEY holds it.
var deployKey = base64.StdEncoding.EncodeToString([]byte("synthetic-test-key-32-bytes-long"))

// fakeSigner is a synthetic Signer for these tests. It stands in for the HMAC signer until
// the permit package can sign.
type fakeSigner struct{ key []byte }

func (s fakeSigner) Alg() string { return "FAKE-HMAC-SHA256" }
func (s fakeSigner) Sign(c []byte) (string, error) {
	m := hmac.New(sha256.New, s.key)
	m.Write(c)
	return base64.StdEncoding.EncodeToString(m.Sum(nil)), nil
}
func (s fakeSigner) Verify(c []byte, v string) error {
	want, _ := s.Sign(c)
	if !hmac.Equal([]byte(want), []byte(v)) {
		return permit.ErrSignature
	}
	return nil
}

// fakeCanonical is the synthetic canonical form: the PERMIT as JSON with the signature value
// emptied.
func fakeCanonical(p permit.Permit) []byte {
	p.Signature.Value = ""
	b, _ := json.Marshal(p)
	return b
}

// fakeVerify checks a PERMIT the way the permit package is specified to, in the same order,
// returning the same sentinel errors.
func fakeVerify(p permit.Permit, s permit.Signer, want permit.Expect) error {
	if err := s.Verify(fakeCanonical(p), p.Signature.Value); err != nil {
		return fmt.Errorf("%w: %v", permit.ErrSignature, err)
	}
	switch {
	case !p.Allow:
		return permit.ErrDenied
	case !want.Now.Before(p.ExpiresAt):
		return permit.ErrExpired
	case p.Target.Name != want.TargetName, p.Target.ID != want.TargetID, p.Commit != want.Commit,
		p.PlanSHA256 != want.PlanSHA256, p.ConfigSHA256 != want.ConfigSHA256:
		return permit.ErrMismatch
	}
	return nil
}

// usePermitFakes swaps in synthetic PERMIT functions for one test. When the permit package
// can sign and verify, the read and verify fakes can go and the real ones be used.
func usePermitFakes(t *testing.T) {
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
	permitFuncs.verify = fakeVerify
	permitFuncs.signer = func(key []byte) permit.Signer { return fakeSigner{key: key} }
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
	usePermitFakes(t)
	h := &deployHarness{harness: newHarness(t)}
	h.write(t, "glassops.yaml", cfg)
	writePlan(t, h.harness, false)
	h.env["GITHUB_ACTIONS"] = "true"
	h.env["GITHUB_SHA"] = deployCommitSHA
	h.env[permit.KeyEnv] = deployKey
	h.tools[deployTagCmd] = ""
	h.tools[deployPushCmd] = ""

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
	p := permit.Permit{
		Schema: permit.Schema, ID: "a1b2c3d4", Allow: true,
		IssuedAt:     time.Date(2026, 10, 26, 11, 0, 0, 0, time.UTC),
		ExpiresAt:    time.Date(2026, 10, 26, 13, 0, 0, 0, time.UTC),
		Target:       permit.Target{Name: "qa", Tier: "qa", ID: "00D000000000002", Kind: "sandbox"},
		Commit:       deployCommitSHA,
		PlanSHA256:   "",
		ConfigSHA256: loaded.SHA256(),
		Validation:   &permit.Validation{DeployID: deployValidated, TestLevel: "RunLocalTests", Passed: true},
	}
	if edit != nil {
		edit(&p)
	}
	h.writePermit(t, p, true)
	return h
}

// writePermit writes p, signed with the test key when sign is true.
func (h *deployHarness) writePermit(t *testing.T, p permit.Permit, sign bool) {
	t.Helper()
	if sign {
		key, _ := base64.StdEncoding.DecodeString(deployKey)
		p.Signature = permit.Signature{Alg: "FAKE-HMAC-SHA256"}
		p.Signature.Value, _ = fakeSigner{key: key}.Sign(fakeCanonical(p))
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

func TestDeployResumePassesJobID(t *testing.T) {
	h := newDeployHarness(t, deployConfig+"  deploy: { quick: require }\n", nil)
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
	writePlan(t, h.harness, true)
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
