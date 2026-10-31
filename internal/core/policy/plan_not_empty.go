package policy

import "github.com/glassops-platform/glassops/internal/core/facts"

// PlanNotEmpty fails when the plan has no changes. A plan that says nothing is also a failure,
// because missing evidence never passes.
type PlanNotEmpty struct{}

// Name returns the rule name.
func (PlanNotEmpty) Name() string { return "plan_not_empty" }

// Evaluate checks change.empty.
func (PlanNotEmpty) Evaluate(in Input) []Result {
	empty, ok := in.Facts.Bool(facts.ChangeEmpty)
	switch {
	case !ok:
		return one("plan_not_empty", Fail, "the facts do not say whether the plan is empty (%s is missing)", facts.ChangeEmpty)
	case empty:
		return one("plan_not_empty", Fail, "the plan has no changes")
	}
	return one("plan_not_empty", Pass, "the plan has changes")
}
