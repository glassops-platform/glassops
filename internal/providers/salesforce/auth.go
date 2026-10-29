package salesforce

import (
	"context"
	"errors"

	"github.com/glassops-platform/glassops/internal/providers"
)

// Authenticate logs in to t in a new session. Phase 2 workstream 1 implements it.
func (p *Provider) Authenticate(ctx context.Context, t providers.Target) (providers.Session, error) {
	return nil, errors.New("salesforce: authenticate is not implemented yet")
}
