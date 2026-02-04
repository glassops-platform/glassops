package services

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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

// Identity provides utilities for handling Salesforce identity and auth URLs.
type Identity struct{}

// NewIdentity creates a new Identity instance.
func NewIdentity() *Identity {
	return &Identity{}
}

// ParseAuthURL parses and validates a Salesforce SFDX auth URL.
// Format: force://PlatformCLI::<refresh_token>@<instance_url>
func (i *Identity) ParseAuthURL(authURL string) (string, error) {
	if authURL == "" {
		return "", fmt.Errorf("auth URL is empty")
	}

	// Validate basic format
	if !strings.HasPrefix(authURL, "force://") {
		return "", fmt.Errorf("invalid auth URL format: must start with force://")
	}

	// Extract instance URL
	atIndex := strings.LastIndex(authURL, "@")
	if atIndex == -1 {
		return "", fmt.Errorf("invalid auth URL format: missing @ separator")
	}

	instanceURL := authURL[atIndex+1:]

	// Validate instance URL is a Salesforce domain
	validDomains := regexp.MustCompile(`(?i)\.salesforce\.com$`)
	if !validDomains.MatchString(instanceURL) {
		return "", fmt.Errorf("invalid auth URL: instance URL must be a salesforce.com domain")
	}

	// The org ID is not directly in the auth URL - it would need to be retrieved
	// by authenticating with the URL. Return the instance URL for now.
	return instanceURL, nil
}

// AuthenticateWithURL authenticates using an SFDX auth URL.
func (i *Identity) AuthenticateWithURL(authURL string) (string, error) {
	if _, err := i.ParseAuthURL(authURL); err != nil {
		return "", err
	}

	// Write auth URL to temp file
	urlPath := filepath.Join(os.TempDir(), fmt.Sprintf("glassops-auth-%d.txt", time.Now().UnixNano()))
	if err := os.WriteFile(urlPath, []byte(authURL), 0600); err != nil {
		return "", fmt.Errorf("failed to write auth URL: %w", err)
	}
	defer func() {
		// Secure cleanup
		if stat, err := os.Stat(urlPath); err == nil {
			zeros := make([]byte, stat.Size())
			_ = os.WriteFile(urlPath, zeros, 0600)
		}
		os.Remove(urlPath)
	}()

	// Use sf org login sfdx-url command
	cmd := exec.Command("sf", "org", "login", "sfdx-url",
		"--sfdx-url-file", urlPath,
		"--set-default",
		"--json",
	)

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("authentication with SFDX URL failed: %w", err)
	}

	var result struct {
		Result struct {
			OrgID string `json:"orgId"`
		} `json:"result"`
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return "", fmt.Errorf("failed to parse authentication response: %w", err)
	}

	return result.Result.OrgID, nil
}
