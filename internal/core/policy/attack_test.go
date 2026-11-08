package policy

// Red team tests for the gate decision. A test here passes when the attack is refused.
// Attacks that currently succeed are in attack_open_test.go, behind the redteam_open build
// tag.

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/core/facts"
)

// rtEval is a synthetic evaluator with fixed results and error.
type rtEval struct {
	name    string
	results []Result
	err     error
}

func (e rtEval) Name() string { return e.name }
func (e rtEval) Evaluate(context.Context, Input) ([]Result, error) {
	return e.results, e.err
}

var rtNow = time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)

func TestRTOneFailAnywhereDenies(t *testing.T) {
	pass := Result{Rule: "a", Outcome: Pass}
	fail := Result{Rule: "b", Outcome: Fail}
	cases := map[string][]Evaluator{
		"fail first, approvals after": {rtEval{name: "go", results: []Result{fail, {Rule: "c", Outcome: SatisfiedByApproval}, {Rule: "d", Outcome: WaivedByConfig}}}},
		"fail in a later evaluator":   {rtEval{name: "go", results: []Result{pass}}, rtEval{name: "x", results: []Result{fail}}},
		"evaluator error, no results": {rtEval{name: "go", results: []Result{pass}}, rtEval{name: "x", err: errors.New("rt: cannot decide")}},
		"evaluator error with passes": {rtEval{name: "x", results: []Result{pass, pass}, err: errors.New("rt: partial")}},
	}
	for name, evs := range cases {
		t.Run(name, func(t *testing.T) {
			allow, results, _ := EvaluateAll(context.Background(), evs, Input{Facts: facts.New(), Now: rtNow})
			if allow {
				t.Fatalf("allowed: %+v", results)
			}
		})
	}
}

func TestRTRegoOnThisBaseFailsClosed(t *testing.T) {
	cfg := &config.Config{Policy: config.Policy{Rego: config.Rego{Files: []string{"rt.rego"}}}}
	allow, _, err := EvaluateAll(context.Background(), Evaluators([]Rule{PlanNotEmpty{}}, cfg),
		Input{Facts: doc(map[string]any{facts.ChangeEmpty: false}), Config: cfg, Now: rtNow})
	if allow || err == nil {
		t.Fatalf("allow %v, err %v: a Rego config that cannot be evaluated did not deny", allow, err)
	}
}

// Missing evidence never passes: the real rules over an empty facts document deny.
func TestRTDefaultRulesDenyWithNoEvidence(t *testing.T) {
	cfg := &config.Config{Policy: config.Policy{Rules: config.Rules{
		SeparationOfDuties:    &config.SoDRule{Tiers: []string{"prod"}},
		DestructiveNeedsLabel: &config.DestructiveLabelRule{Label: "rt", Tiers: []string{"prod"}},
	}}}
	allow, results := Evaluate(Default(), Input{Facts: facts.New(), Config: cfg, Now: rtNow})
	if allow {
		t.Fatalf("allowed with no facts: %+v", results)
	}
}

// Evidence of the wrong type is not evidence. The fact names come from facts, and gate is
// what sets them, but a result file merged into the document can hold any JSON.
func TestRTWrongTypedEvidenceDoesNotPass(t *testing.T) {
	cfg := &config.Config{Policy: config.Policy{Rules: config.Rules{
		SeparationOfDuties:    &config.SoDRule{Tiers: []string{"prod"}},
		DestructiveNeedsLabel: &config.DestructiveLabelRule{Label: "rt", Tiers: []string{"prod"}},
	}}}
	cases := []struct {
		name, rule string
		kv         map[string]any
	}{
		{"change.empty as a string", "plan_not_empty", map[string]any{facts.ChangeEmpty: "false", facts.TargetTier: "prod"}},
		{"destructive count as text", "destructive_needs_label", map[string]any{facts.ChangeEmpty: false, facts.TargetTier: "prod",
			facts.ChangeDestructiveCount: "0", facts.CIPRAuthor: "a", facts.CIActor: "b"}},
		{"destructive count as a fraction", "destructive_needs_label", map[string]any{facts.ChangeEmpty: false, facts.TargetTier: "prod",
			facts.ChangeDestructiveCount: 0.5, facts.CIPRAuthor: "a", facts.CIActor: "b"}},
		{"author as a list", "separation_of_duties", map[string]any{facts.ChangeEmpty: false, facts.TargetTier: "prod", facts.ChangeDestructiveCount: 0,
			facts.CIPRAuthor: []any{"a"}, facts.CIActor: "b"}},
		{"same person in another case", "separation_of_duties", map[string]any{facts.ChangeEmpty: false, facts.TargetTier: "prod", facts.ChangeDestructiveCount: 0,
			facts.CIPRAuthor: "Author", facts.CIActor: "author"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			allow, results := Evaluate(Default(), Input{Facts: doc(tc.kv), Config: cfg, Now: rtNow})
			if allow {
				t.Fatalf("allowed: %+v", results)
			}
			for _, r := range results {
				if r.Rule == tc.rule && r.Outcome == Fail {
					return
				}
			}
			t.Fatalf("denied, but not by %s: %+v", tc.rule, results)
		})
	}
}
