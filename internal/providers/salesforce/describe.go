package salesforce

import (
	"context"
	"errors"

	"github.com/glassops-platform/glassops/internal/providers"
)

// Describe reads the logged-in org. Phase 2 workstream 2 implements it.
func (p *Provider) Describe(ctx context.Context, s providers.Session) (providers.TargetInfo, error) {
	return providers.TargetInfo{}, errors.New("salesforce: describe is not implemented yet")
}
