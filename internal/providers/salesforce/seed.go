package salesforce

import (
	"context"
	"errors"

	"github.com/glassops-platform/glassops/internal/providers"
)

// Seed deploys the base state a destructive change needs to an ephemeral org. It is not
// implemented yet.
func (p *Provider) Seed(_ context.Context, _ providers.Session, _ providers.SeedRequest) (providers.SeedResult, error) {
	return providers.SeedResult{}, errors.New("salesforce seed is not implemented yet")
}
