// Package policy evaluates the v1 rules over a facts document.
package policy

import (
	"time"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/core/facts"
)

// Outcome is one rule result's verdict.
type Outcome string

const (
	Pass                Outcome = "pass"
	Fail                Outcome = "fail"
	Skipped             Outcome = "skipped"
	WaivedByConfig      Outcome = "waived_by_config"
	SatisfiedByApproval Outcome = "satisfied_by_approval"
)

// Result is one rule's verdict.
type Result struct {
	// Rule names the rule, such as "plan_not_empty", "window.prod", "blackout" or
	// "coverage_min".
	Rule    string  `json:"rule"`
	Outcome Outcome `json:"outcome"`
	Detail  string  `json:"detail"`
	// Mechanism and Approver are set for SatisfiedByApproval.
	Mechanism string `json:"mechanism,omitempty"`
	Approver  string `json:"approver,omitempty"`
}

// Denies reports whether the outcome blocks a deploy. Only Fail does.
func (o Outcome) Denies() bool { return o == Fail }

// Input is everything a rule may read.
type Input struct {
	Facts  *facts.Document
	Config *config.Config
	// Now is the injectable clock, already read.
	Now time.Time
}

// Rule is one v1 rule.
type Rule interface {
	Name() string
	// Evaluate may return several results, such as one per tier window.
	Evaluate(in Input) []Result
}

// Evaluate runs rules in order and reports allow when no result Denies. Every rule runs, so
// the results list every reason a deploy would be refused.
func Evaluate(rules []Rule, in Input) (allow bool, results []Result) {
	allow = true
	results = []Result{}
	for _, r := range rules {
		for _, res := range r.Evaluate(in) {
			if res.Outcome.Denies() {
				allow = false
			}
			results = append(results, res)
		}
	}
	return allow, results
}

// Default returns the v1 rule set.
func Default() []Rule {
	return nil
}
