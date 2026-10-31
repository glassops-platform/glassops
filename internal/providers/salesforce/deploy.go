package salesforce

import (
	"context"
	"errors"

	"github.com/glassops-platform/glassops/internal/providers"
)

// Deploy changes the target org. It is not implemented yet and always returns an error.
func (p *Provider) Deploy(_ context.Context, _ providers.Session, _ providers.ChangeSet, _ providers.DeployOptions) (providers.DeployResult, error) {
	return providers.DeployResult{}, errors.New("salesforce deploy is not implemented yet")
}
