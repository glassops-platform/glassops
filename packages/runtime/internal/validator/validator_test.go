package validator

import (
	"os"
	"testing"
)

func TestValidateEnvironment(t *testing.T) {
	// Setup - Save original env
	originalVars := map[string]string{
		"GITHUB_WORKSPACE":  os.Getenv("GITHUB_WORKSPACE"),
		"GITHUB_ACTOR":      os.Getenv("GITHUB_ACTOR"),
		"GITHUB_REPOSITORY": os.Getenv("GITHUB_REPOSITORY"),
	}
	defer func() {
		for k, v := range originalVars {
			os.Setenv(k, v)
		}
	}()

	t.Run("Success", func(t *testing.T) {
		os.Setenv("GITHUB_WORKSPACE", "/tmp")
		os.Setenv("GITHUB_ACTOR", "me")
		os.Setenv("GITHUB_REPOSITORY", "owner/repo")
		if err := ValidateEnvironment(); err != nil {
			t.Errorf("ValidateEnvironment() error = %v, want nil", err)
		}
	})

	t.Run("Missing Variable", func(t *testing.T) {
		os.Unsetenv("GITHUB_WORKSPACE")
		if err := ValidateEnvironment(); err == nil {
			t.Error("ValidateEnvironment() expected error, got nil")
		}
	})
}

func TestEnsureValidInstanceURL(t *testing.T) {
	tests := []struct {
		url     string
		wantErr bool
	}{
		{"https://login.salesforce.com", false},
		{"https://test.salesforce.com", false},
		{"https://my-domain.my.salesforce.com", false},
		{"http://insecure.com", false}, // Validator currently allows http prefix check
		{"ftp://invalid.com", true},
		{"invalid-url", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			if err := EnsureValidInstanceURL(tt.url); (err != nil) != tt.wantErr {
				t.Errorf("EnsureValidInstanceURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
