package policy

import (
	"context"
	"errors"
	"fmt"

	"github.com/glassops-platform/glassops/internal/core/config"
)

// Evaluator is one source of rule results for gate, such as the built-in Go rules or Rego.
// An error means the evaluator could not decide, and gate fails closed on it.
type Evaluator interface {
	Name() string
	Evaluate(ctx context.Context, in Input) ([]Result, error)
}

// GoRules evaluates the built-in Go rules.
type GoRules struct {
	Rules []Rule
}

// Name returns "go".
func (GoRules) Name() string { return "go" }

// Evaluate runs every rule in order. It never returns an error.
func (g GoRules) Evaluate(_ context.Context, in Input) ([]Result, error) {
	_, results := Evaluate(g.Rules, in)
	return results, nil
}

// Evaluators returns what gate evaluates for cfg: the Go rules first, then Rego when
// policy.rego names files. The Go rules always run, and later evaluators can only add results.
func Evaluators(rules []Rule, cfg *config.Config) []Evaluator {
	evs := []Evaluator{GoRules{Rules: rules}}
	if cfg != nil && cfg.Policy.Rego.Enabled() {
		evs = append(evs, NewRego(cfg.Policy.Rego))
	}
	return evs
}

// EvaluateAll runs every evaluator in order and merges their results. allow is false when any
// result Denies or any evaluator fails. A failed evaluator adds a Fail result named after it
// and its error is returned, so the caller can fail closed. Every evaluator runs even after
// one fails.
func EvaluateAll(ctx context.Context, evs []Evaluator, in Input) (allow bool, results []Result, err error) {
	allow = true
	results = []Result{}
	var errs []error
	for _, ev := range evs {
		out, evErr := ev.Evaluate(ctx, in)
		for _, r := range out {
			if r.Outcome.Denies() {
				allow = false
			}
			results = append(results, r)
		}
		if evErr != nil {
			allow = false
			results = append(results, Result{Rule: ev.Name(), Outcome: Fail, Detail: evErr.Error()})
			errs = append(errs, fmt.Errorf("%s: %w", ev.Name(), evErr))
		}
	}
	return allow, results, errors.Join(errs...)
}
