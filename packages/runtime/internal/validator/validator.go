package validator

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/glassops-platform/glassops/packages/runtime/internal/gha"
)

// ValidateEnvironment checks for required environment variables.
func ValidateEnvironment() error {
	requiredEnvVars := []string{
		"GITHUB_WORKSPACE",
		"GITHUB_ACTOR",
		"GITHUB_REPOSITORY",
	}

	var missingEnvVars []string
	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			missingEnvVars = append(missingEnvVars, envVar)
		}
	}
	if len(missingEnvVars) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missingEnvVars, ", "))
	}
	return nil
}

// ValidateInputs checks for required inputs and sanitizes them.
func ValidateInputs() error {
	requiredInputs := []string{"client_id", "jwt_key", "username"}
	var missingInputs []string
	for _, input := range requiredInputs {
		if gha.GetInput(input) == "" {
			missingInputs = append(missingInputs, input)
		}
	}
	if len(missingInputs) > 0 {
		return fmt.Errorf("missing required inputs: %s", strings.Join(missingInputs, ", "))
	}

	// Validate JWT key format (skip if skip_auth is true)
	jwtKey := gha.GetInput("jwt_key")
	if gha.GetInput("skip_auth") != "true" {
		if !strings.Contains(jwtKey, "BEGIN") || !strings.Contains(jwtKey, "END") {
			return fmt.Errorf("invalid JWT key format - must contain BEGIN and END markers")
		}
	}
	if jwtKey != "" {
		gha.SetSecret(jwtKey)
	}

	return nil
}

// ValidateContext checks for valid repository and PR context.
func ValidateContext() error {
	if os.Getenv("GITHUB_EVENT_NAME") == "pull_request" {
		if os.Getenv("GITHUB_HEAD_REF") == "" {
			return fmt.Errorf("invalid pull request context - missing GITHUB_HEAD_REF")
		}
		if strings.Contains(os.Getenv("GITHUB_HEAD_REF"), ":") {
			gha.Warning("Running on forked repository - additional security validations recommended")
		}
	}

	// Validate repository format
	repoPattern := regexp.MustCompile(`^[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+$`)
	if !repoPattern.MatchString(os.Getenv("GITHUB_REPOSITORY")) {
		return fmt.Errorf("invalid repository format: %s", os.Getenv("GITHUB_REPOSITORY"))
	}

	return nil
}

// EnsureValidInstanceURL validates the Salesforce instance URL.
func EnsureValidInstanceURL(url string) error {
	if len(url) <= 8 || (!strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://")) {
		return fmt.Errorf("invalid instance URL: %s", url)
	}
	return nil
}
