package salesforce

import (
	"context"
	"errors"
	"time"

	"github.com/glassops-platform/glassops/internal/providers"
)

// errEnvNotImplemented is what the environment methods return until they land.
var errEnvNotImplemented = errors.New("salesforce environments are not implemented yet")

// CreateEnv creates a scratch org from the Dev Hub session hub. It is not implemented yet.
func (p *Provider) CreateEnv(_ context.Context, _ providers.Session, _ providers.EnvRequest) (providers.Env, error) {
	return providers.Env{}, errEnvNotImplemented
}

// DeleteEnv deletes a scratch org from the Dev Hub session hub. It is not implemented yet.
func (p *Provider) DeleteEnv(_ context.Context, _ providers.Session, _ providers.Env) error {
	return errEnvNotImplemented
}

// SweepEnvs deletes old scratch orgs named with prefix. It is not implemented yet.
func (p *Provider) SweepEnvs(_ context.Context, _ providers.Session, _ string, _ time.Duration) ([]providers.Env, error) {
	return nil, errEnvNotImplemented
}
