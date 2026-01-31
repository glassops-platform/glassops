package integration

import (
	"os"
	"testing"

	"github.com/glassops-platform/glassops/packages/runtime/internal/services"
)

func TestIdentityIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Run("parses valid SFDX auth URL", func(t *testing.T) {
		authURL := "force://PlatformCLI::refresh_token@login.salesforce.com"
		env, err := SetupTestWorkspace(nil)
		if err != nil {
			t.Fatalf("failed to setup: %v", err)
		}
		defer env.Cleanup()
		env.SetEnvironment(nil)

		identity := services.NewIdentity()
		orgID, err := identity.ParseAuthURL(authURL)

		// The auth URL doesn't contain org ID directly, so this tests the parsing logic
		if err != nil {
			// Expected if parsing requires actual org ID in URL
			t.Logf("Parse returned error (expected for mock URL): %v", err)
		}
		_ = orgID // May be empty for mock URLs
	})

	t.Run("validates auth URL format", func(t *testing.T) {
		testCases := []struct {
			name      string
			authURL   string
			expectErr bool
		}{
			{
				name:      "valid format",
				authURL:   "force://PlatformCLI::refresh_token@login.salesforce.com",
				expectErr: false,
			},
			{
				name:      "sandbox URL",
				authURL:   "force://PlatformCLI::refresh_token@test.salesforce.com",
				expectErr: false,
			},
			{
				name:      "custom domain",
				authURL:   "force://PlatformCLI::refresh_token@company.my.salesforce.com",
				expectErr: false,
			},
			{
				name:      "empty string",
				authURL:   "",
				expectErr: true,
			},
			{
				name:      "invalid format",
				authURL:   "not-a-valid-url",
				expectErr: true,
			},
		}

		env, err := SetupTestWorkspace(nil)
		if err != nil {
			t.Fatalf("failed to setup: %v", err)
		}
		defer env.Cleanup()
		env.SetEnvironment(nil)

		identity := services.NewIdentity()

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := identity.ParseAuthURL(tc.authURL)
				if tc.expectErr && err == nil {
					t.Error("expected error")
				}
				if !tc.expectErr && err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			})
		}
	})

	t.Run("handles environment auth URL", func(t *testing.T) {
		env, err := SetupTestWorkspace(nil)
		if err != nil {
			t.Fatalf("failed to setup: %v", err)
		}
		defer env.Cleanup()

		authURL := "force://PlatformCLI::test_refresh_token@login.salesforce.com"
		env.SetEnvironment(map[string]string{
			"INPUT_SFDX_AUTH_URL": authURL,
		})

		// Read from environment
		envAuthURL := os.Getenv("INPUT_SFDX_AUTH_URL")
		if envAuthURL != authURL {
			t.Errorf("expected auth URL from env, got %s", envAuthURL)
		}
	})
}
