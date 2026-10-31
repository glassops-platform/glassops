package salesforce

import "github.com/glassops-platform/glassops/internal/providers"

// The Salesforce provider implements every optional capability.
var (
	_ providers.Environments = (*Provider)(nil)
	_ providers.Seeder       = (*Provider)(nil)
	_ providers.Tester       = (*Provider)(nil)
	_ providers.Verifier     = (*Provider)(nil)
	_ providers.Configurable = (*Provider)(nil)
)
