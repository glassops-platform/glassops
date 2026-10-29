package salesforce

import (
	"context"
	"errors"

	"github.com/glassops-platform/glassops/internal/providers"
)

// Validate runs a check-only deploy validate. Phase 2 workstream 4 implements it.
func (p *Provider) Validate(ctx context.Context, s providers.Session, cs providers.ChangeSet, opts providers.ValidateOptions) (providers.ValidateResult, error) {
	return providers.ValidateResult{}, errors.New("salesforce: validate is not implemented yet")
}
