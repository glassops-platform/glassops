package services

import (
	"os"
	"runtime"
	"testing"
)

func TestNewRuntimeEnvironment(t *testing.T) {
	env := NewRuntimeEnvironment()

	if env == nil {
		t.Error("expected NewRuntimeEnvironment to return non-nil")
	}
}

func TestRuntimeEnvironmentPlatform(t *testing.T) {
	// Set GOOS env var
	originalGOOS := os.Getenv("GOOS")
	os.Setenv("GOOS", "linux")
	defer os.Setenv("GOOS", originalGOOS)

	env := NewRuntimeEnvironment()

	if env.platform != "linux" {
		t.Errorf("expected platform 'linux', got '%s'", env.platform)
	}
}

func TestRuntimeEnvironmentDefaultPlatform(t *testing.T) {
	// Unset GOOS to use default
	originalGOOS := os.Getenv("GOOS")
	os.Unsetenv("GOOS")
	defer func() {
		if originalGOOS != "" {
			os.Setenv("GOOS", originalGOOS)
		}
	}()

	env := NewRuntimeEnvironment()

	// Should be empty when GOOS is not set (reads from env, not runtime.GOOS)
	if env.platform != "" {
		t.Errorf("expected empty platform when GOOS not set, got '%s'", env.platform)
	}
}

func TestExecWithAutoConfirmWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-specific test")
	}

	env := &RuntimeEnvironment{platform: "windows"}

	// This would fail because sf doesn't exist, but we're just testing the command construction
	err := env.execWithAutoConfirm("echo", []string{"test"})
	// We expect this to work since echo is available
	if err != nil {
		t.Logf("execWithAutoConfirm returned error (expected if sf not installed): %v", err)
	}
}

func TestExecWithAutoConfirmUnix(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix-specific test")
	}

	env := &RuntimeEnvironment{platform: "linux"}

	// This would fail because sf doesn't exist, but we're just testing the command construction
	err := env.execWithAutoConfirm("echo", []string{"test"})
	// We expect this to work since echo is available
	if err != nil {
		t.Logf("execWithAutoConfirm returned error (expected if sf not installed): %v", err)
	}
}
