package services

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/glassops-platform/glassops/packages/runtime/internal/gha"
)

// AuthRequest contains Salesforce authentication parameters.
type AuthRequest struct {
	ClientID    string
	JWTKey      string
	Username    string
	InstanceURL string
}

// IdentityResolver handles Salesforce authentication.
type IdentityResolver struct{}

// NewIdentityResolver creates a new IdentityResolver.
func NewIdentityResolver() *IdentityResolver {
	return &IdentityResolver{}
}

// Authenticate performs JWT-based authentication with Salesforce.
func (i *IdentityResolver) Authenticate(req AuthRequest) (string, error) {
	gha.StartGroup("Authenticating Identity")
	defer gha.EndGroup()

	// 1. Sanitize JWT Key (handle environment variable escaping)
	// Replace literal "\n" with real newlines if they exist
	jwtKey := strings.ReplaceAll(req.JWTKey, `\n`, "\n")

	// Write JWT key to temp file with secure permissions
	keyPath := filepath.Join(os.TempDir(), fmt.Sprintf("glassops-jwt-%d.key", time.Now().UnixNano()))
	if err := os.WriteFile(keyPath, []byte(jwtKey), 0600); err != nil {
		return "", fmt.Errorf("failed to write JWT key: %w", err)
	}

	// Ensure cleanup
	defer func() {
		// Secure cleanup: overwrite with zeros before unlinking
		if stat, err := os.Stat(keyPath); err == nil {
			zeros := make([]byte, stat.Size())
			_ = os.WriteFile(keyPath, zeros, 0600)
		}
		os.Remove(keyPath)
	}()

	args := []string{
		"org", "login", "jwt",
		"--client-id", req.ClientID,
		"--jwt-key-file", keyPath,
		"--username", req.Username,
		"--set-default",
		"--json",
	}

	if req.InstanceURL != "" {
		args = append(args, "--instance-url", req.InstanceURL)
	}

	// Retry for transient Salesforce API failures
	var output []byte
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		cmd := exec.Command("sf", args...)

		// Capture stderr for debugging
		var stderr strings.Builder
		cmd.Stderr = &stderr

		var err error
		output, err = cmd.Output()
		if err == nil {
			break
		}

		lastErr = fmt.Errorf("%w (stderr: %s)", err, stderr.String())

		// Don't retry if it's a configuration error (e.g. invalid grant)
		if strings.Contains(stderr.String(), "error") || strings.Contains(stderr.String(), "invalid") {
			break
		}

		time.Sleep(time.Duration(2000*(1<<attempt)) * time.Millisecond)
	}

	if lastErr != nil {
		// Log the detailed error from sf
		gha.Error(fmt.Sprintf("Salesforce CLI Error: %v", lastErr))
		return "", fmt.Errorf("authentication failed")
	}

	var result struct {
		Result struct {
			OrgID       string `json:"orgId"`
			AccessToken string `json:"accessToken"`
		} `json:"result"`
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return "", fmt.Errorf("failed to parse authentication response: %w", err)
	}

	gha.Info(fmt.Sprintf("Authenticated as %s (%s)", req.Username, result.Result.OrgID))
	return result.Result.OrgID, nil
}
