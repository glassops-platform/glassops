//go:build redteam_open

package policy

// Red team attacks on the gate decision that currently succeed. Each test asserts the
// behaviour EvaluateAll should have, so it fails until the finding is fixed. Reproduce with:
//
//	go test -tags redteam_open -run TestRTOpen ./internal/core/policy/

import (
	"context"
	"testing"

	"github.com/glassops-platform/glassops/internal/core/facts"
)

// Finding: Outcome.Denies is true for "fail" only, so any outcome outside the five known ones
// allows. A Go rule with a typo, or a Rego evaluator that maps a policy's verdict to an
// outcome string (phase6/integration), fails open on "FAIL", "deny" or "".
func TestRTOpenUnknownOutcomeAllows(t *testing.T) {
	for _, o := range []Outcome{"FAIL", "Fail", "deny", "error", ""} {
		evs := []Evaluator{rtEval{name: "go", results: []Result{{Rule: "rt", Outcome: o}}}}
		if allow, _, _ := EvaluateAll(context.Background(), evs, Input{Facts: facts.New(), Now: rtNow}); allow {
			t.Errorf("outcome %q allowed", o)
		}
	}
}

// Finding: an evaluator that returns no results and no error allows. gate refuses an empty
// rule list, but not a rule set that says nothing, such as a Rego file with no rules in it.
func TestRTOpenEvaluatorWithNoResultsAllows(t *testing.T) {
	evs := []Evaluator{rtEval{name: "go"}}
	if allow, _, _ := EvaluateAll(context.Background(), evs, Input{Facts: facts.New(), Now: rtNow}); allow {
		t.Fatal("an evaluator that returned nothing allowed")
	}
}
