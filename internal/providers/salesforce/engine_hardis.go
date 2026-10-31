package salesforce

import (
	"context"
	"errors"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/providers"
)

// errHardisNotImplemented is what the hardis engine returns until it lands. Commands report
// it as a tool error, exit 2.
var errHardisNotImplemented = errors.New("salesforce engine hardis is not implemented yet")

// hardisEngine runs plan, validate and deploy through sfdx-hardis, which the user installs.
// It is not implemented yet.
type hardisEngine struct {
	p   *Provider
	cfg config.Hardis
}

func newHardisEngine(p *Provider, cfg config.Hardis) Engine {
	return hardisEngine{p: p, cfg: cfg}
}

func (e hardisEngine) Name() string { return config.EngineHardis }

func (e hardisEngine) Plan(context.Context, providers.PlanRequest) (providers.ChangeSet, error) {
	return providers.ChangeSet{}, errHardisNotImplemented
}

func (e hardisEngine) Validate(context.Context, providers.Session, providers.ChangeSet, providers.ValidateOptions) (providers.ValidateResult, error) {
	return providers.ValidateResult{}, errHardisNotImplemented
}

func (e hardisEngine) Deploy(context.Context, providers.Session, providers.ChangeSet, providers.DeployOptions) (providers.DeployResult, error) {
	return providers.DeployResult{}, errHardisNotImplemented
}
