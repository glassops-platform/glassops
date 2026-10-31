// Package exec is the exec provider. It runs the commands in the exec section of
// glassops.yaml for plan, describe, validate, deploy and verify, so any deploy tool gets the
// gate, the PERMIT and the evidence. It is never detected from the project files. A project
// chooses it with provider: exec.
package exec

import (
	"context"
	"errors"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/providers"
)

// Name is the provider name in glassops.yaml.
const Name = config.ExecProvider

// errNotImplemented is what every operation returns until the provider lands.
var errNotImplemented = errors.New("exec provider is not implemented yet")

// Provider is the exec provider.
type Provider struct {
	// cfg is the exec section, set by Configure.
	cfg config.Exec
}

// New returns the exec provider.
func New() *Provider { return &Provider{} }

// The exec provider implements these optional capabilities.
var (
	_ providers.Provider     = (*Provider)(nil)
	_ providers.Verifier     = (*Provider)(nil)
	_ providers.Configurable = (*Provider)(nil)
)

func (p *Provider) Name() string { return Name }

// Detect always reports false, so the exec provider is used only when the config names it.
func (p *Provider) Detect(string) (bool, error) { return false, nil }

// Tools lists nothing yet.
func (p *Provider) Tools() []providers.Tool { return nil }

// CanonicalID compares Ids exactly.
func (p *Provider) CanonicalID(id string) string { return id }

// Configure keeps the exec section of cfg.
func (p *Provider) Configure(cfg *config.Config) error {
	if cfg == nil {
		p.cfg = config.Exec{}
		return nil
	}
	p.cfg = cfg.Exec
	return nil
}

// Authenticate is not implemented yet.
func (p *Provider) Authenticate(context.Context, providers.Target) (providers.Session, error) {
	return nil, errNotImplemented
}

// Describe is not implemented yet.
func (p *Provider) Describe(context.Context, providers.Session) (providers.TargetInfo, error) {
	return providers.TargetInfo{}, errNotImplemented
}

// Plan is not implemented yet.
func (p *Provider) Plan(context.Context, providers.PlanRequest) (providers.ChangeSet, error) {
	return providers.ChangeSet{}, errNotImplemented
}

// Validate is not implemented yet.
func (p *Provider) Validate(context.Context, providers.Session, providers.ChangeSet, providers.ValidateOptions) (providers.ValidateResult, error) {
	return providers.ValidateResult{}, errNotImplemented
}

// Deploy is not implemented yet.
func (p *Provider) Deploy(context.Context, providers.Session, providers.ChangeSet, providers.DeployOptions) (providers.DeployResult, error) {
	return providers.DeployResult{}, errNotImplemented
}

// Verify is not implemented yet.
func (p *Provider) Verify(context.Context, providers.Session, providers.VerifyRequest) (providers.VerifyResult, error) {
	return providers.VerifyResult{}, errNotImplemented
}
