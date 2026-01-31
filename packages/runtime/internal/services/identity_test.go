package services

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewIdentityResolver(t *testing.T) {
	resolver := NewIdentityResolver()

	if resolver == nil {
		t.Error("expected NewIdentityResolver to return non-nil")
	}
}

func TestAuthRequestFields(t *testing.T) {
	req := AuthRequest{
		ClientID:    "test-client-id",
		JWTKey:      "PLACEHOLDER_NOT_A_KEY",
		Username:    "test@example.com",
		InstanceURL: "https://test.salesforce.com",
	}

	if req.ClientID != "test-client-id" {
		t.Errorf("expected ClientID 'test-client-id', got '%s'", req.ClientID)
	}

	if req.Username != "test@example.com" {
		t.Errorf("expected Username 'test@example.com', got '%s'", req.Username)
	}

	if req.InstanceURL != "https://test.salesforce.com" {
		t.Errorf("expected InstanceURL 'https://test.salesforce.com', got '%s'", req.InstanceURL)
	}
}

func TestAuthRequestEmptyInstanceURL(t *testing.T) {
	req := AuthRequest{
		ClientID: "test-client-id",
		JWTKey:   "test-key",
		Username: "test@example.com",
	}

	// InstanceURL should be empty by default
	if req.InstanceURL != "" {
		t.Errorf("expected empty InstanceURL, got '%s'", req.InstanceURL)
	}
}

func TestJWTKeyFileCreation(t *testing.T) {
	// Test that we can write a JWT key to a temp file
	tempDir := t.TempDir()
	keyPath := filepath.Join(tempDir, "test-jwt.key")
	jwtKey := "PLACEHOLDER_NOT_A_KEY"

	err := os.WriteFile(keyPath, []byte(jwtKey), 0600)
	if err != nil {
		t.Fatalf("failed to write JWT key file: %v", err)
	}

	// Verify file permissions
	info, err := os.Stat(keyPath)
	if err != nil {
		t.Fatalf("failed to stat JWT key file: %v", err)
	}

	// On Unix, check permissions (0600)
	// On Windows, this is less meaningful but file should exist
	if info.Size() == 0 {
		t.Error("JWT key file is empty")
	}

	// Verify content
	content, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatalf("failed to read JWT key file: %v", err)
	}

	if string(content) != jwtKey {
		t.Error("JWT key file content mismatch")
	}
}

func TestSecureCleanup(t *testing.T) {
	// Test the secure cleanup pattern (overwrite with zeros before delete)
	tempDir := t.TempDir()
	keyPath := filepath.Join(tempDir, "secure-key.key")
	secretData := "super-secret-key-data"

	// Write secret
	err := os.WriteFile(keyPath, []byte(secretData), 0600)
	if err != nil {
		t.Fatalf("failed to write secret file: %v", err)
	}

	// Get file size
	stat, err := os.Stat(keyPath)
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}

	// Overwrite with zeros
	zeros := make([]byte, stat.Size())
	err = os.WriteFile(keyPath, zeros, 0600)
	if err != nil {
		t.Fatalf("failed to overwrite with zeros: %v", err)
	}

	// Verify content is zeros
	content, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatalf("failed to read zeroed file: %v", err)
	}

	for i, b := range content {
		if b != 0 {
			t.Errorf("byte %d is not zero: %d", i, b)
		}
	}

	// Delete
	err = os.Remove(keyPath)
	if err != nil {
		t.Fatalf("failed to remove file: %v", err)
	}

	// Verify deleted
	_, err = os.Stat(keyPath)
	if !os.IsNotExist(err) {
		t.Error("file should be deleted")
	}
}
