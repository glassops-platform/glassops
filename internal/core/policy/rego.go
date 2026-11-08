package policy

import (
	"context"
	"errors"

	"github.com/glassops-platform/glassops/internal/core/config"
)

// errRegoNotImplemented is what the Rego evaluator returns until it lands. gate fails closed
// on it.
var errRegoNotImplemented = errors.New("rego policy evaluation is not implemented yet")

// regoEvaluator evaluates the Rego files in policy.rego. It is not implemented yet.
type regoEvaluator struct {
	cfg config.Rego
}

// NewRego returns the evaluator for the Rego files cfg names.
func NewRego(cfg config.Rego) Evaluator { return regoEvaluator{cfg: cfg} }

// Name returns "rego".
func (regoEvaluator) Name() string { return "rego" }

// Evaluate is not implemented yet, so it always returns an error.
func (regoEvaluator) Evaluate(context.Context, Input) ([]Result, error) {
	return nil, errRegoNotImplemented
}
