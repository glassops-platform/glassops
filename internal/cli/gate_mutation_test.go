package cli

import (
	"strings"
	"testing"

	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/facts"
	"github.com/glassops-platform/glassops/internal/core/policy"
	"github.com/glassops-platform/glassops/internal/core/run"
)

// gate.go resultFacts: the validation.bound detail names the read problem only for the validate
// result. The mutant `step != "validate"` swaps the two messages.
func TestMutGateUnreadableResultBoundDetail(t *testing.T) {
	for _, tc := range []struct {
		step     string
		contains string
		absent   string
	}{
		{"validate", "the validate result cannot be read", ""},
		{"plan", "no validate result", "cannot be read"},
	} {
		t.Run(tc.step, func(t *testing.T) {
			installGateFakes(t, fixedRule("plan_not_empty", policy.Pass))
			h, _ := gateHarness(t, gateConfig)
			h.write(t, ".glassops/results/"+tc.step+".json", "{not json")
			if code := h.gate("--target", "prod"); code != exitcode.Error {
				t.Fatalf("exit %d, want 2. output:\n%s", code, h.out)
			}
			p := readPermit(t, h)
			if p.Allow || p.Validation != nil {
				t.Fatalf("an unreadable %s result must deny and bind nothing: allow=%v validation=%+v", tc.step, p.Allow, p.Validation)
			}
			if c, ok := findCheck(p.Checks, "evidence.results."+tc.step); !ok || c.Outcome != policy.Fail {
				t.Fatalf("no evidence failure for the %s result: %+v", tc.step, p.Checks)
			}
			c, ok := findCheck(p.Checks, "validation.bound")
			if !ok || c.Outcome != policy.Skipped || !strings.Contains(c.Detail, tc.contains) {
				t.Fatalf("validation.bound %+v, want skipped containing %q", c, tc.contains)
			}
			if tc.absent != "" && strings.Contains(c.Detail, tc.absent) {
				t.Fatalf("validation.bound %+v must not contain %q", c, tc.absent)
			}
		})
	}
}

// gate.go resultFacts: an unreadable plan result must not stop the validate result from being
// read and bound.
func TestMutGateUnreadablePlanResultStillBindsValidate(t *testing.T) {
	installGateFakes(t, fixedRule("plan_not_empty", policy.Pass))
	h, _ := gateHarness(t, gateConfig)
	h.write(t, ".glassops/results/plan.json", "{not json")
	r := run.New("validate", nil)
	r.Target = "prod"
	r.Status = run.StatusPass
	r.Facts["salesforce.deploy_id"] = "0Af000000000001AAA"
	r.Facts["salesforce.check_only"] = true
	r.Facts["validate.test_level"] = "RunLocalTests"
	r.Facts["validate.plan_sha256"] = planHash(t, h)
	r.Facts["validate.change_set_sha256"] = changeSetHash(t, h)
	writeResult(t, h, r)
	if code := h.gate("--target", "prod"); code != exitcode.Error {
		t.Fatalf("exit %d, want 2. output:\n%s", code, h.out)
	}
	p := readPermit(t, h)
	if p.Allow {
		t.Fatal("allowed with an unreadable plan result")
	}
	if p.Validation == nil || p.Validation.DeployID != "0Af000000000001AAA" {
		t.Fatalf("the validate result was not bound: %+v", p.Validation)
	}
}

// gate.go ciFacts: the configured approval label counts when it is not the first label on the
// pull request.
func TestMutGateLabelApprovalWhenLabelIsNotFirst(t *testing.T) {
	var seen []facts.Approval
	rule := ruleFunc{name: "separation_of_duties", fn: func(in policy.Input) []policy.Result {
		seen = in.Facts.Approvals()
		return []policy.Result{{Rule: "separation_of_duties", Outcome: policy.Fail, Detail: "synthetic"}}
	}}
	installGateFakes(t, rule)
	// The pull request carries dependencies, then github_actions.
	cfg := strings.Replace(sodConfig, "{ environment: glassops-deploy }", "{ label: github_actions, approvers: [rdbumstead] }", 1)
	h, _ := gateHarness(t, cfg)
	h.env["GITHUB_SHA"] = labelSHA
	delete(h.env, "GITHUB_RUN_ID")
	h.useAPI(newGitHubServer(t, defaultRoutes()))
	h.gate("--target", "prod")
	if len(seen) != 1 || seen[0].Mechanism != "label" || seen[0].Label != "github_actions" {
		t.Fatalf("approvals: %+v", seen)
	}
}

// gate.go ciFacts: a pull request ref that is not a number is an evidence failure, and the
// API is not asked about it.
func TestMutGateNonNumericPullRequestRefDenies(t *testing.T) {
	installGateFakes(t, fixedRule("plan_not_empty", policy.Pass))
	h, _ := gateHarness(t, gateConfig)
	h.env["GITHUB_EVENT_NAME"] = "pull_request"
	h.env["GITHUB_REF"] = "refs/pull/abc/merge"
	delete(h.env, "GITHUB_RUN_ID")
	gs := newGitHubServer(t, defaultRoutes())
	h.useAPI(gs)
	if code := h.gate("--target", "prod"); code != exitcode.Error {
		t.Fatalf("exit %d, want 2. output:\n%s", code, h.out)
	}
	p := readPermit(t, h)
	c, ok := findCheck(p.Checks, "evidence.ci.pull_request")
	if p.Allow || !ok || c.Outcome != policy.Fail || !strings.Contains(c.Detail, `"abc" is not a number`) {
		t.Fatalf("allow=%v check=%+v", p.Allow, c)
	}
	if gs.count() != 0 {
		t.Fatalf("the API was called for a bad number: %v", gs.hits)
	}
}
