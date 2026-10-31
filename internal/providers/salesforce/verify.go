package salesforce

import (
	"context"
	"errors"

	"github.com/glassops-platform/glassops/internal/providers"
)

// Verify reads the org after a deploy and checks it against the plan. It is not implemented
// yet.
func (p *Provider) Verify(_ context.Context, _ providers.Session, _ providers.VerifyRequest) (providers.VerifyResult, error) {
	return providers.VerifyResult{}, errors.New("salesforce verify is not implemented yet")
}
