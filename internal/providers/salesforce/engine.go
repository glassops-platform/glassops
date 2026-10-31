package salesforce

import (
	"context"
	"fmt"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/providers"
)

// Engine is how a Salesforce change runs. Only plan, validate and deploy go through it.
// GlassOps keeps auth, the org Id guard, gate, the PERMIT checks, verify and report whatever
// the engine.
type Engine interface {
	// Name is the value of salesforce.engine that selects the engine.
	Name() string
	Plan(ctx context.Context, req providers.PlanRequest) (providers.ChangeSet, error)
	Validate(ctx context.Context, s providers.Session, cs providers.ChangeSet, opts providers.ValidateOptions) (providers.ValidateResult, error)
	Deploy(ctx context.Context, s providers.Session, cs providers.ChangeSet, opts providers.DeployOptions) (providers.DeployResult, error)
}

// engineFor returns the engine salesforce.engine names.
func (p *Provider) engineFor(name string, cfg config.Salesforce) (Engine, error) {
	switch name {
	case "", config.EngineNative:
		return nativeEngine{p: p}, nil
	case config.EngineHardis:
		return newHardisEngine(p, cfg.Hardis), nil
	default:
		return nil, fmt.Errorf("salesforce: unknown engine %q (want %s or %s)", name, config.EngineNative, config.EngineHardis)
	}
}

// Configure selects the engine from salesforce.engine. Without a call the provider uses the
// native engine.
func (p *Provider) Configure(cfg *config.Config) error {
	if cfg == nil {
		p.engine = nil
		return nil
	}
	e, err := p.engineFor(cfg.Salesforce.EngineName(), cfg.Salesforce)
	if err != nil {
		return err
	}
	p.engine = e
	return nil
}

// Engine returns the engine plan, validate and deploy go through.
func (p *Provider) Engine() Engine {
	if p.engine == nil {
		return nativeEngine{p: p}
	}
	return p.engine
}

// Plan works out the change set through the selected engine. It needs no login.
func (p *Provider) Plan(ctx context.Context, req providers.PlanRequest) (providers.ChangeSet, error) {
	return p.Engine().Plan(ctx, req)
}

// Validate runs a check-only validate through the selected engine. It never changes the org.
func (p *Provider) Validate(ctx context.Context, s providers.Session, cs providers.ChangeSet, opts providers.ValidateOptions) (providers.ValidateResult, error) {
	return p.Engine().Validate(ctx, s, cs, opts)
}

// Deploy deploys through the selected engine. Callers must have verified a PERMIT first.
func (p *Provider) Deploy(ctx context.Context, s providers.Session, cs providers.ChangeSet, opts providers.DeployOptions) (providers.DeployResult, error) {
	return p.Engine().Deploy(ctx, s, cs, opts)
}

// nativeEngine is the sf CLI with sfdx-git-delta, as plan.go, validate.go and deploy.go
// implement it.
type nativeEngine struct{ p *Provider }

func (e nativeEngine) Name() string { return config.EngineNative }

func (e nativeEngine) Plan(ctx context.Context, req providers.PlanRequest) (providers.ChangeSet, error) {
	return e.p.nativePlan(ctx, req)
}

func (e nativeEngine) Validate(ctx context.Context, s providers.Session, cs providers.ChangeSet, opts providers.ValidateOptions) (providers.ValidateResult, error) {
	return e.p.nativeValidate(ctx, s, cs, opts)
}

func (e nativeEngine) Deploy(ctx context.Context, s providers.Session, cs providers.ChangeSet, opts providers.DeployOptions) (providers.DeployResult, error) {
	return e.p.nativeDeploy(ctx, s, cs, opts)
}
