package salesforce

import (
	"context"
	"errors"

	"github.com/glassops-platform/glassops/internal/providers"
)

// RunTests runs Apex tests with code coverage. It is not implemented yet.
func (p *Provider) RunTests(_ context.Context, _ providers.Session, _ providers.TestRequest) (providers.TestResult, error) {
	return providers.TestResult{}, errors.New("salesforce apex tests are not implemented yet")
}
