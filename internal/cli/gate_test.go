package cli

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/facts"
	"github.com/glassops-platform/glassops/internal/core/permit"
	"github.com/glassops-platform/glassops/internal/core/policy"
	"github.com/glassops-platform/glassops/internal/core/run"
	"github.com/glassops-platform/glassops/internal/providers"
)

// The rules and the HMAC signer are built on other branches. Until they merge, these tests
// stand in for them with fake rules and a fake signer, swapped into gate's seams.

// ruleFunc is a fake policy rule.
type ruleFunc struct {
	name string
	fn   func(policy.Input) []policy.Result
	// calls counts evaluations.
	calls *int
}

func (r ruleFunc) Name() string { return r.name }
func (r ruleFunc) Evaluate(in policy.Input) []policy.Result {
	if r.calls != nil {
		*r.calls++
	}
	return r.fn(in)
}

func fixedRule(name string, o policy.Outcome) ruleFunc {
	return ruleFunc{name: name, fn: func(policy.Input) []policy.Result {
		return []policy.Result{{Rule: name, Outcome: o, Detail: "synthetic " + string(o)}}
	}}
}

// fakeSigner signs with SHA-256 over the key and the bytes. It is not an HMAC, only a stand in
// that a test can check.
type fakeSigner struct{ key []byte }

func (s fakeSigner) Alg() string { return "TEST-SHA256" }
func (s fakeSigner) Sign(b []byte) (string, error) {
	sum := sha256.Sum256(append(append([]byte{}, s.key...), b...))
	return hex.EncodeToString(sum[:]), nil
}
func (s fakeSigner) Verify(b []byte, v string) error {
	want, _ := s.Sign(b)
	if want != v {
		return errors.New("fake: bad signature")
	}
	return nil
}

// fakeCanonical is the PERMIT with an empty signature value, as plain JSON.
func fakeCanonical(p permit.Permit) []byte {
	p.Signature.Value = ""
	data, _ := json.Marshal(p)
	return data
}

// installGateFakes swaps fake rules and a fake signer into gate and restores them after t.
func installGateFakes(t *testing.T, rules ...policy.Rule) {
	t.Helper()
	oldRules, oldSigner, oldSign, oldWrite := gateRules, gateSigner, gateSign, gateWrite
	t.Cleanup(func() { gateRules, gateSigner, gateSign, gateWrite = oldRules, oldSigner, oldSign, oldWrite })
	gateRules = func() []policy.Rule { return rules }
	gateSigner = func(key []byte) permit.Signer { return fakeSigner{key: key} }
	gateSign = func(p *permit.Permit, s permit.Signer) error {
		// The algorithm is part of the signed bytes, and only the value is left out.
		p.Signature = permit.Signature{Alg: s.Alg()}
		v, err := s.Sign(fakeCanonical(*p))
		if err != nil {
			return err
		}
		p.Signature.Value = v
		return nil
	}
	gateWrite = func(path string, p permit.Permit) error {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		data, err := json.MarshalIndent(p, "", "  ")
		if err != nil {
			return err
		}
		return os.WriteFile(path, data, 0o644)
	}
}

const (
	gateRepo   = "glassops-platform/glassops"
	gateSHA    = "20459158b4f16da4cddc816ac7c1d695177af196" // the merge commit in commit-pulls.json
	labelSHA   = "24cb560cd6f9d354f4d2a592e36e4a66db8fe55b" // the commit in commit-pulls-labelled.json
	unknownSHA = "0000000000000000000000000000000000000000"
	gateRunID  = "41165934934"
	// syntheticToken is a made-up token. Tests check it never reaches output.
	syntheticToken = "ghs_syntheticTokenValueForTests"
)

// syntheticKey is 32 made-up bytes in base64.
var syntheticKey = base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef"))

const gateConfig = `version: 1
targets:
  prod:
    id: 00D000000000001AAA
    tier: prod
    auth: { method: sfdx-url, url: $SF_PROD_AUTH_URL }
  qa:
    id: 00D000000000002AAA
    tier: qa
    auth: { method: sfdx-url, url: $SF_QA_AUTH_URL }
  scratch:
    id: 00D000000000003AAA
    tier: ephemeral
    auth: { method: sfdx-url, url: $SF_SCRATCH_AUTH_URL }
policy:
  permit_ttl: 90m
salesforce:
  test_level: { prod: RunLocalTests }
`

// gateHarness returns a harness running as a GitHub push to main, with the config, a plan,
// a permit key and a fake provider whose describe reports each target as the guard expects.
func gateHarness(t *testing.T, cfg string) (*harness, *fakeProvider) {
	t.Helper()
	h := newHarness(t)
	h.write(t, "glassops.yaml", cfg)
	writeGatePlan(t, h, providers.ChangeSet{Destructive: []string{}, Facts: map[string]any{}})
	for k, v := range map[string]string{
		"GITHUB_ACTIONS":      "true",
		"GITHUB_REPOSITORY":   gateRepo,
		"GITHUB_SHA":          gateSHA,
		"GITHUB_REF":          "refs/heads/main",
		"GITHUB_EVENT_NAME":   "push",
		"GITHUB_ACTOR":        "deployer",
		"GITHUB_RUN_ID":       gateRunID,
		"GLASSOPS_PERMIT_KEY": syntheticKey,
	} {
		h.env[k] = v
	}
	fp := &fakeProvider{
		detect: true,
		auth:   func(providers.Target) error { return nil },
		describe: func(tg providers.Target) (providers.TargetInfo, error) {
			k := providers.KindSandbox
			switch tg.Tier {
			case "prod":
				k = providers.KindProduction
			case "ephemeral":
				k = providers.KindScratch
			}
			return providers.TargetInfo{ID: tg.ID, Kind: k, Facts: map[string]any{"salesforce.org_type": "synthetic"}}, nil
		},
	}
	h.app.Providers = providers.Registry{fp}
	return h, fp
}

// writeGatePlan writes cs as the plan, filling the fields every plan has.
func writeGatePlan(t *testing.T, h *harness, cs providers.ChangeSet) {
	t.Helper()
	cs.Schema, cs.Provider = providers.PlanSchema, "salesforce"
	cs.Dir, cs.Manifest = ".glassops/plan", ".glassops/plan/package/package.xml"
	cs.SHA256 = "5e1f0c3a9b7d2e4f6a8c0b1d3e5f7a9c2b4d6e8f0a1c3e5b7d9f1a3c5e7b9d1f" // synthetic
	if err := providers.WriteChangeSet(filepath.Join(h.dir, ".glassops", "plan.json"), cs); err != nil {
		t.Fatal(err)
	}
}

// githubServer serves the GitHub fixtures by path and counts requests.
type githubServer struct {
	*httptest.Server
	mu   sync.Mutex
	hits []string
}

func (s *githubServer) count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.hits)
}

func newGitHubServer(t *testing.T, routes map[string]string) *githubServer {
	t.Helper()
	gs := &githubServer{}
	gs.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gs.mu.Lock()
		gs.hits = append(gs.hits, r.URL.Path)
		gs.mu.Unlock()
		if r.Header.Get("Authorization") != "Bearer "+syntheticToken {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		spec, ok := routes[r.URL.Path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"Not Found"}`))
			return
		}
		status, file := http.StatusOK, spec
		if s, f, ok := strings.Cut(spec, " "); ok {
			if s == "422" {
				status = http.StatusUnprocessableEntity
			} else if s == "500" {
				status = http.StatusInternalServerError
			}
			file = f
		}
		data, err := os.ReadFile(filepath.Join("..", "ci", "testdata", "github", file))
		if err != nil {
			t.Errorf("fixture %s: %v", file, err)
		}
		w.WriteHeader(status)
		_, _ = w.Write(data)
	}))
	t.Cleanup(gs.Close)
	return gs
}

// defaultRoutes serves the real pull and approvals fixtures for gateSHA and gateRunID.
func defaultRoutes() map[string]string {
	return map[string]string{
		"/repos/" + gateRepo + "/commits/" + gateSHA + "/pulls":            "commit-pulls.json",
		"/repos/" + gateRepo + "/commits/" + labelSHA + "/pulls":           "commit-pulls-labelled.json",
		"/repos/" + gateRepo + "/commits/" + unknownSHA + "/pulls":         "422 commit-pulls-unknown.json",
		"/repos/" + gateRepo + "/actions/runs/" + gateRunID + "/approvals": "run-approvals.json",
	}
}

func (h *harness) useAPI(gs *githubServer) {
	h.env["GITHUB_TOKEN"] = syntheticToken
	h.env["GITHUB_API_URL"] = gs.URL
}

func (h *harness) gate(args ...string) int {
	return Execute(h.app, append(append([]string{"gate"}, args...), "--dir", h.dir, "--ci", "github"))
}

func readPermit(t *testing.T, h *harness) permit.Permit {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(h.dir, ".glassops", "permit.json"))
	if err != nil {
		t.Fatalf("PERMIT: %v\noutput:\n%s", err, h.out)
	}
	var p permit.Permit
	if err := json.Unmarshal(data, &p); err != nil {
		t.Fatal(err)
	}
	return p
}

func readFacts(t *testing.T, h *harness) *facts.Document {
	t.Helper()
	d, err := facts.Read(filepath.Join(h.dir, ".glassops", "facts.json"))
	if err != nil {
		t.Fatal(err)
	}
	return d
}

// verifySigned checks the fake signature over the PERMIT as it was written.
func verifySigned(t *testing.T, p permit.Permit) {
	t.Helper()
	key, _ := base64.StdEncoding.DecodeString(syntheticKey)
	if p.Signature.Alg != "TEST-SHA256" || p.Signature.Value == "" {
		t.Fatalf("signature: %+v", p.Signature)
	}
	if err := (fakeSigner{key: key}).Verify(fakeCanonical(p), p.Signature.Value); err != nil {
		t.Fatalf("the written PERMIT does not verify: %v", err)
	}
}

func findCheck(checks []policy.Result, rule string) (policy.Result, bool) {
	for _, c := range checks {
		if c.Rule == rule {
			return c, true
		}
	}
	return policy.Result{}, false
}

func TestGateAllowWritesSignedPermitAndFacts(t *testing.T) {
	installGateFakes(t, fixedRule("plan_not_empty", policy.Pass))
	h, fp := gateHarness(t, gateConfig)
	if code := h.gate("--target", "prod"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	p := readPermit(t, h)
	verifySigned(t, p)
	loaded, err := config.Load(h.dir, "", "salesforce")
	if err != nil {
		t.Fatal(err)
	}
	cs, err := providers.ReadChangeSet(filepath.Join(h.dir, ".glassops", "plan.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !p.Allow || p.Schema != permit.Schema || len(p.ID) != 32 || p.Commit != gateSHA ||
		p.PlanSHA256 != cs.SHA256 || p.ConfigSHA256 != loaded.SHA256() {
		t.Fatalf("PERMIT: %+v", p)
	}
	if p.Target != (permit.Target{Name: "prod", Tier: "prod", ID: "00D000000000001AAA", Kind: "production"}) {
		t.Fatalf("target: %+v", p.Target)
	}
	if !p.IssuedAt.Equal(h.app.Now()) || p.ExpiresAt.Sub(p.IssuedAt) != 90*time.Minute {
		t.Fatalf("issued %s expires %s", p.IssuedAt, p.ExpiresAt)
	}
	if g, ok := findCheck(p.Checks, "guard"); !ok || g.Outcome != policy.Pass {
		t.Fatalf("guard check: %+v", p.Checks)
	}
	if p.SeparationOfDuties == nil || p.SeparationOfDuties.Mode != "not_applicable" || p.SeparationOfDuties.Deployer != "deployer" {
		t.Fatalf("sod: %+v", p.SeparationOfDuties)
	}
	if !fp.allClosed() || len(fp.sessions) != 1 {
		t.Fatalf("sessions: %d, all closed %v", len(fp.sessions), fp.allClosed())
	}

	d := readFacts(t, h)
	for key, want := range map[string]any{
		facts.TargetName: "prod", facts.TargetTier: "prod", facts.TargetID: "00D000000000001AAA",
		facts.TargetKind: "production", facts.ChangeEmpty: false, facts.ChangePlanSHA256: cs.SHA256,
		facts.CIRepo: gateRepo, facts.CICommit: gateSHA, facts.CIActor: "deployer", facts.CIEvent: "push",
		"salesforce.org_type": "synthetic",
	} {
		if got, _ := d.Get(key); got != want {
			t.Errorf("fact %s = %v, want %v", key, got, want)
		}
	}
	if n, ok := d.Int(facts.ChangeDestructiveCount); !ok || n != 0 {
		t.Errorf("destructive count %v %v", n, ok)
	}

	r := h.result(t, "gate")
	if r.Status != run.StatusPass || r.Facts["gate.allow"] != true || checkStatus(r, "plan_not_empty") != run.StatusPass {
		t.Fatalf("result: %+v", r)
	}
}

func TestGateDenyStillWritesSignedPermit(t *testing.T) {
	installGateFakes(t, fixedRule("plan_not_empty", policy.Pass), fixedRule("window.prod", policy.Fail), fixedRule("blackout", policy.Skipped))
	h, _ := gateHarness(t, gateConfig)
	if code := h.gate("--target", "prod"); code != exitcode.Fail {
		t.Fatalf("exit %d, want 1. output:\n%s", code, h.out)
	}
	p := readPermit(t, h)
	verifySigned(t, p)
	if p.Allow {
		t.Fatal("a denied PERMIT allows")
	}
	if c, ok := findCheck(p.Checks, "window.prod"); !ok || c.Outcome != policy.Fail {
		t.Fatalf("checks: %+v", p.Checks)
	}
	if c, ok := findCheck(p.Checks, "blackout"); !ok || c.Outcome != policy.Skipped {
		t.Fatalf("every rule result is recorded, checks: %+v", p.Checks)
	}
	r := h.result(t, "gate")
	if r.Status != run.StatusFail || checkStatus(r, "window.prod") != run.StatusFail || r.Facts["gate.allow"] != false {
		t.Fatalf("result: %+v", r)
	}
}

func TestGateKeyProblemsExitTwoWithoutPermit(t *testing.T) {
	cases := []struct {
		name, key, want string
	}{
		{"missing", "", "is not set"},
		{"not base64", "not*base64!", "not valid standard base64"},
		{"too short", base64.StdEncoding.EncodeToString([]byte("short-key")), "decodes to 9 bytes"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			installGateFakes(t, fixedRule("plan_not_empty", policy.Pass))
			h, fp := gateHarness(t, gateConfig)
			if tc.key == "" {
				delete(h.env, "GLASSOPS_PERMIT_KEY")
			} else {
				h.env["GLASSOPS_PERMIT_KEY"] = tc.key
			}
			// A PERMIT from an earlier run must not survive a run that cannot sign.
			h.write(t, ".glassops/permit.json", `{"allow":true}`)
			if code := h.gate("--target", "prod"); code != exitcode.Error {
				t.Fatalf("exit %d, want 2. output:\n%s", code, h.out)
			}
			if _, err := os.Stat(filepath.Join(h.dir, ".glassops", "permit.json")); !os.IsNotExist(err) {
				t.Fatalf("a PERMIT exists after a key failure: %v", err)
			}
			if !strings.Contains(h.out.String(), tc.want) || len(fp.sessions) != 0 {
				t.Fatalf("output:\n%s", h.out)
			}
			if tc.key != "" && strings.Contains(h.out.String(), tc.key) {
				t.Fatal("the key value was printed")
			}
			if checkStatus(h.result(t, "gate"), "permit.key") != run.StatusError {
				t.Fatal("no permit.key error in the result")
			}
		})
	}
}

func TestGateMissingPlanExitsTwo(t *testing.T) {
	installGateFakes(t, fixedRule("plan_not_empty", policy.Pass))
	h, _ := gateHarness(t, gateConfig)
	if code := h.gate("--target", "prod", "--plan", filepath.Join(h.dir, "nope.json")); code != exitcode.Error {
		t.Fatalf("exit %d, want 2. output:\n%s", code, h.out)
	}
	if _, err := os.Stat(filepath.Join(h.dir, ".glassops", "permit.json")); !os.IsNotExist(err) {
		t.Fatal("a PERMIT was written without a plan")
	}
	if checkStatus(h.result(t, "gate"), "plan") != run.StatusError {
		t.Fatal("no plan error in the result")
	}
}

func TestGateUnknownTargetExitsTwo(t *testing.T) {
	installGateFakes(t, fixedRule("plan_not_empty", policy.Pass))
	h, _ := gateHarness(t, gateConfig)
	if code := h.gate("--target", "staging"); code != exitcode.Error {
		t.Fatalf("exit %d, want 2. output:\n%s", code, h.out)
	}
}

// sodFromApprovals is a fake separation of duties rule: satisfied by an approved environment
// approval for glassops-deploy, else a fail.
func sodFromApprovals() ruleFunc {
	return ruleFunc{name: "separation_of_duties", fn: func(in policy.Input) []policy.Result {
		for _, a := range in.Facts.Approvals() {
			if a.Mechanism == "environment" && a.State == "approved" && a.Environment == "glassops-deploy" {
				return []policy.Result{{Rule: "separation_of_duties", Outcome: policy.SatisfiedByApproval, Detail: "synthetic", Mechanism: "environment", Approver: a.Approver}}
			}
		}
		return []policy.Result{{Rule: "separation_of_duties", Outcome: policy.Fail, Detail: "synthetic: no approval"}}
	}}
}

const sodConfig = `version: 1
targets:
  prod:
    id: 00D000000000001AAA
    tier: prod
    auth: { method: sfdx-url, url: $SF_PROD_AUTH_URL }
policy:
  rules:
    separation_of_duties:
      tiers: [prod]
      approval: { environment: glassops-deploy }
`

func TestGateSoDSatisfiedByEnvironmentApprovalFixture(t *testing.T) {
	installGateFakes(t, sodFromApprovals())
	h, _ := gateHarness(t, sodConfig)
	gs := newGitHubServer(t, defaultRoutes())
	h.useAPI(gs)
	if code := h.gate("--target", "prod"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	p := readPermit(t, h)
	verifySigned(t, p)
	want := permit.SoD{Mode: "satisfied_by_approval", Mechanism: "environment", Approver: "rdbumstead", Author: "rdbumstead", Deployer: "deployer"}
	if !p.Allow || p.SeparationOfDuties == nil || *p.SeparationOfDuties != want {
		t.Fatalf("allow %v, sod %+v", p.Allow, p.SeparationOfDuties)
	}
	if p.ExpiresAt.Sub(p.IssuedAt) != config.DefaultPermitTTL {
		t.Fatalf("default TTL not applied: %s", p.ExpiresAt.Sub(p.IssuedAt))
	}
	d := readFacts(t, h)
	if v, _ := d.String(facts.CIPRNumber); v != "46" {
		t.Fatalf("pr number %q", v)
	}
	if v, _ := d.String(facts.CIPRAuthor); v != "rdbumstead" {
		t.Fatalf("pr author %q", v)
	}
	if l, ok := d.Strings(facts.CILabels); !ok || len(l) != 0 {
		t.Fatalf("labels %v %v", l, ok)
	}
	if gs.count() != 2 {
		t.Fatalf("API requests: %v", gs.hits)
	}
	assertNoToken(t, h)
}

// assertNoToken checks that the token reached none of the output and files gate writes.
func assertNoToken(t *testing.T, h *harness) {
	t.Helper()
	if strings.Contains(h.out.String(), syntheticToken) {
		t.Fatal("the token was printed")
	}
	for _, rel := range []string{".glassops/permit.json", ".glassops/facts.json", ".glassops/results/gate.json"} {
		data, err := os.ReadFile(filepath.Join(h.dir, rel))
		if err == nil && strings.Contains(string(data), syntheticToken) {
			t.Fatalf("the token is in %s", rel)
		}
	}
}

func TestGateLabelApprovalHasNoApprover(t *testing.T) {
	var seen []facts.Approval
	rule := ruleFunc{name: "separation_of_duties", fn: func(in policy.Input) []policy.Result {
		seen = in.Facts.Approvals()
		return []policy.Result{{Rule: "separation_of_duties", Outcome: policy.Fail, Detail: "synthetic"}}
	}}
	installGateFakes(t, rule)
	cfg := strings.Replace(sodConfig, "{ environment: glassops-deploy }", "{ label: dependencies, approvers: [rdbumstead] }", 1)
	h, _ := gateHarness(t, cfg)
	h.env["GITHUB_SHA"] = labelSHA
	delete(h.env, "GITHUB_RUN_ID")
	h.useAPI(newGitHubServer(t, defaultRoutes()))
	if code := h.gate("--target", "prod"); code != exitcode.Fail {
		t.Fatalf("exit %d, want 1. output:\n%s", code, h.out)
	}
	if len(seen) != 1 || seen[0] != (facts.Approval{Mechanism: "label", Label: "dependencies"}) {
		t.Fatalf("approvals: %+v", seen)
	}
	d := readFacts(t, h)
	if l, _ := d.Strings(facts.CILabels); len(l) != 2 {
		t.Fatalf("labels %v", l)
	}
	if a, _ := d.String(facts.CIPRAuthor); a != "dependabot[bot]" {
		t.Fatalf("author %q", a)
	}
	if p := readPermit(t, h); p.SeparationOfDuties.Mode != "enforced" {
		t.Fatalf("sod %+v", p.SeparationOfDuties)
	}
}

func TestGateUnknownCommitHasNoPullRequest(t *testing.T) {
	installGateFakes(t, fixedRule("plan_not_empty", policy.Pass))
	h, _ := gateHarness(t, gateConfig)
	h.env["GITHUB_SHA"] = unknownSHA
	delete(h.env, "GITHUB_RUN_ID")
	h.useAPI(newGitHubServer(t, defaultRoutes()))
	if code := h.gate("--target", "prod"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	d := readFacts(t, h)
	for _, k := range []string{facts.CIPRNumber, facts.CIPRAuthor, facts.CILabels} {
		if _, ok := d.Get(k); ok {
			t.Fatalf("fact %s is set for an unknown commit", k)
		}
	}
}

func TestGateNoTokenMeansNoAPICallsAndRulesFail(t *testing.T) {
	// A synthetic rule that needs the pull request author, which only the API supplies on a push.
	needsAuthor := ruleFunc{name: "separation_of_duties", fn: func(in policy.Input) []policy.Result {
		if _, ok := in.Facts.String(facts.CIPRAuthor); !ok {
			return []policy.Result{{Rule: "separation_of_duties", Outcome: policy.Fail, Detail: "synthetic: no author"}}
		}
		return []policy.Result{{Rule: "separation_of_duties", Outcome: policy.Pass, Detail: "synthetic"}}
	}}
	installGateFakes(t, needsAuthor)
	h, _ := gateHarness(t, gateConfig)
	gs := newGitHubServer(t, defaultRoutes())
	h.env["GITHUB_API_URL"] = gs.URL
	if code := h.gate("--target", "prod"); code != exitcode.Fail {
		t.Fatalf("exit %d, want 1. output:\n%s", code, h.out)
	}
	if gs.count() != 0 {
		t.Fatalf("API called without a token: %v", gs.hits)
	}
	d := readFacts(t, h)
	if _, ok := d.Get(facts.CIApprovals); ok {
		t.Fatal("approvals are set without a token")
	}
	if p := readPermit(t, h); p.Allow {
		t.Fatal("allowed without the evidence a rule needs")
	}
}

func TestGateAPIErrorDeniesAndExitsTwo(t *testing.T) {
	installGateFakes(t, fixedRule("plan_not_empty", policy.Pass))
	h, _ := gateHarness(t, gateConfig)
	routes := defaultRoutes()
	routes["/repos/"+gateRepo+"/actions/runs/"+gateRunID+"/approvals"] = "500 issue-labels.json"
	h.useAPI(newGitHubServer(t, routes))
	if code := h.gate("--target", "prod"); code != exitcode.Error {
		t.Fatalf("exit %d, want 2. output:\n%s", code, h.out)
	}
	p := readPermit(t, h)
	verifySigned(t, p)
	if c, ok := findCheck(p.Checks, "evidence.ci.approvals"); p.Allow || !ok || c.Outcome != policy.Fail {
		t.Fatalf("allow %v, checks %+v", p.Allow, p.Checks)
	}
	assertNoToken(t, h)
}

func TestGateEphemeralTargetGetsAutomaticPermit(t *testing.T) {
	calls := 0
	failing := fixedRule("window.ephemeral", policy.Fail)
	failing.calls = &calls
	installGateFakes(t, failing)
	h, fp := gateHarness(t, gateConfig)
	gs := newGitHubServer(t, defaultRoutes())
	h.useAPI(gs)
	if code := h.gate("--target", "scratch"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	p := readPermit(t, h)
	verifySigned(t, p)
	if !p.Allow || p.Target.Kind != "scratch" || len(p.Checks) != 2 || p.Checks[0].Rule != "guard" || p.Checks[1].Rule != "ephemeral" {
		t.Fatalf("PERMIT: %+v", p)
	}
	if p.SeparationOfDuties.Mode != "not_applicable" {
		t.Fatalf("sod %+v", p.SeparationOfDuties)
	}
	if calls != 0 || gs.count() != 0 {
		t.Fatalf("rules ran %d times and the API was called %d times for an ephemeral target", calls, gs.count())
	}
	if len(fp.sessions) != 1 || !fp.allClosed() {
		t.Fatal("the guard did not log in, or left a session open")
	}
	if h.result(t, "gate").Facts["gate.ephemeral"] != true {
		t.Fatal("gate.ephemeral not recorded")
	}
}

func TestGateEphemeralWithoutConnectRunsTheRules(t *testing.T) {
	installGateFakes(t, fixedRule("plan_not_empty", policy.Fail))
	h, fp := gateHarness(t, gateConfig)
	if code := h.gate("--target", "scratch", "--no-connect"); code != exitcode.Fail {
		t.Fatalf("exit %d, want 1. output:\n%s", code, h.out)
	}
	if p := readPermit(t, h); p.Allow || p.Target.Kind != "scratch" || len(fp.sessions) != 0 {
		t.Fatalf("PERMIT: %+v", p)
	}
}

func TestGateNoConnectUsesConfiguredTarget(t *testing.T) {
	cases := []struct{ target, id, kind string }{
		{"prod", "00D000000000001AAA", "production"},
		{"qa", "00D000000000002AAA", "unknown"},
	}
	for _, tc := range cases {
		t.Run(tc.target, func(t *testing.T) {
			installGateFakes(t, fixedRule("plan_not_empty", policy.Pass))
			h, fp := gateHarness(t, gateConfig)
			if code := h.gate("--target", tc.target, "--no-connect"); code != exitcode.OK {
				t.Fatalf("exit %d, output:\n%s", code, h.out)
			}
			p := readPermit(t, h)
			if p.Target.ID != tc.id || p.Target.Kind != tc.kind || len(fp.sessions) != 0 {
				t.Fatalf("target %+v, sessions %d", p.Target, len(fp.sessions))
			}
			if g, _ := findCheck(p.Checks, "guard"); g.Outcome != policy.Skipped {
				t.Fatalf("guard: %+v", g)
			}
		})
	}
}

func TestGateGuardMismatchDeniesAndExitsTwo(t *testing.T) {
	installGateFakes(t, fixedRule("plan_not_empty", policy.Pass))
	h, fp := gateHarness(t, gateConfig)
	fp.describe = func(providers.Target) (providers.TargetInfo, error) {
		return providers.TargetInfo{ID: "00D000000000009AAA", Kind: providers.KindProduction}, nil
	}
	if code := h.gate("--target", "prod"); code != exitcode.Error {
		t.Fatalf("exit %d, want 2. output:\n%s", code, h.out)
	}
	p := readPermit(t, h)
	if g, _ := findCheck(p.Checks, "guard"); p.Allow || g.Outcome != policy.Fail {
		t.Fatalf("PERMIT: %+v", p)
	}
	if !fp.allClosed() {
		t.Fatal("session left open")
	}
}

func TestGateNoRulesNeverAllows(t *testing.T) {
	installGateFakes(t)
	h, _ := gateHarness(t, gateConfig)
	if code := h.gate("--target", "prod"); code != exitcode.Error {
		t.Fatalf("exit %d, want 2. output:\n%s", code, h.out)
	}
	if p := readPermit(t, h); p.Allow {
		t.Fatal("an empty rule set allowed a deploy")
	}
}

func TestGatePlanWithoutHashDenies(t *testing.T) {
	installGateFakes(t, fixedRule("plan_not_empty", policy.Pass))
	h, _ := gateHarness(t, gateConfig)
	writePlan(t, h, false) // the shared helper leaves plan_sha256 empty
	if code := h.gate("--target", "prod"); code != exitcode.Error {
		t.Fatalf("exit %d, want 2. output:\n%s", code, h.out)
	}
	if p := readPermit(t, h); p.Allow {
		t.Fatal("allowed a plan with no hash")
	}
}

func TestGateNoCommitDenies(t *testing.T) {
	installGateFakes(t, fixedRule("plan_not_empty", policy.Pass))
	h, _ := gateHarness(t, gateConfig)
	delete(h.env, "GITHUB_SHA")
	if code := h.gate("--target", "prod"); code != exitcode.Error {
		t.Fatalf("exit %d, want 2. output:\n%s", code, h.out)
	}
	if p := readPermit(t, h); p.Allow {
		t.Fatal("allowed with no commit")
	}
}

func writeResult(t *testing.T, h *harness, r *run.Result) {
	t.Helper()
	if _, err := r.Write(filepath.Join(h.dir, ".glassops", "results"), h.app.Now); err != nil {
		t.Fatal(err)
	}
}

func TestGateBindsValidationForTheSameTarget(t *testing.T) {
	validate := func(target string, status run.Status, checkOnly bool) *run.Result {
		r := run.New("validate", nil)
		r.Target = target
		r.Status = status
		r.Facts["salesforce.deploy_id"] = "0Af000000000001AAA"
		r.Facts["salesforce.check_only"] = checkOnly
		r.Facts["salesforce.coverage.Foo.percent"] = 91
		return r
	}
	cases := []struct {
		name string
		r    *run.Result
		want *permit.Validation
	}{
		{"passed for this target", validate("prod", run.StatusPass, true), &permit.Validation{DeployID: "0Af000000000001AAA", TestLevel: "RunLocalTests", Passed: true}},
		{"fallback or other target", validate("qa", run.StatusPass, true), nil},
		{"failed", validate("prod", run.StatusFail, true), nil},
		{"not check-only", validate("prod", run.StatusPass, false), nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			installGateFakes(t, fixedRule("plan_not_empty", policy.Pass))
			h, _ := gateHarness(t, gateConfig)
			writeResult(t, h, tc.r)
			if code := h.gate("--target", "prod"); code != exitcode.OK {
				t.Fatalf("exit %d, output:\n%s", code, h.out)
			}
			p := readPermit(t, h)
			if (p.Validation == nil) != (tc.want == nil) || (tc.want != nil && *p.Validation != *tc.want) {
				t.Fatalf("validation %+v, want %+v", p.Validation, tc.want)
			}
			if n, ok := readFacts(t, h).Int("salesforce.coverage.Foo.percent"); !ok || n != 91 {
				t.Fatal("validate facts were not merged")
			}
		})
	}
}

func TestGateUnreadableResultDenies(t *testing.T) {
	installGateFakes(t, fixedRule("plan_not_empty", policy.Pass))
	h, _ := gateHarness(t, gateConfig)
	h.write(t, ".glassops/results/plan.json", "{not json")
	if code := h.gate("--target", "prod"); code != exitcode.Error {
		t.Fatalf("exit %d, want 2. output:\n%s", code, h.out)
	}
	if p := readPermit(t, h); p.Allow {
		t.Fatal("allowed with an unreadable result")
	}
}

func TestGateResultFactsCannotOverrideGateFacts(t *testing.T) {
	installGateFakes(t, fixedRule("plan_not_empty", policy.Pass))
	h, _ := gateHarness(t, gateConfig)
	r := run.New("plan", nil)
	r.Facts[facts.TargetTier] = "dev"
	r.Facts[facts.CIActor] = "someone-else"
	r.Facts["salesforce.destructive_count"] = 3
	writeResult(t, h, r)
	if code := h.gate("--target", "prod"); code != exitcode.OK {
		t.Fatalf("exit %d, output:\n%s", code, h.out)
	}
	d := readFacts(t, h)
	if v, _ := d.String(facts.TargetTier); v != "prod" {
		t.Fatalf("tier %q", v)
	}
	if v, _ := d.String(facts.CIActor); v != "deployer" {
		t.Fatalf("actor %q", v)
	}
}

func TestGateDestructiveCount(t *testing.T) {
	cases := []struct {
		name  string
		facts map[string]any
		dest  []string
		want  int
	}{
		{"provider count", map[string]any{"salesforce.destructive_count": 4}, []string{"a.xml"}, 4},
		{"manifests only", map[string]any{}, []string{"pre.xml", "post.xml"}, 2},
		{"none", map[string]any{}, []string{}, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			installGateFakes(t, fixedRule("plan_not_empty", policy.Pass))
			h, _ := gateHarness(t, gateConfig)
			writeGatePlan(t, h, providers.ChangeSet{Destructive: tc.dest, Facts: tc.facts})
			if code := h.gate("--target", "prod"); code != exitcode.OK {
				t.Fatalf("exit %d, output:\n%s", code, h.out)
			}
			if n, _ := readFacts(t, h).Int(facts.ChangeDestructiveCount); n != tc.want {
				t.Fatalf("count %d, want %d", n, tc.want)
			}
		})
	}
}
