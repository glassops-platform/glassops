package salesforce

import (
	"context"
	"errors"

	"github.com/glassops-platform/glassops/internal/providers"
)

// Plan runs sfdx-git-delta. Phase 2 workstream 3 implements it.
func (p *Provider) Plan(ctx context.Context, req providers.PlanRequest) (providers.ChangeSet, error) {
	return providers.ChangeSet{}, errors.New("salesforce: plan is not implemented yet")
}
