package policy

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadDefaultConfig(t *testing.T) {
	// Create temp config with minimal/empty values to test defaults are applied
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "devops-config.json")

	// Empty config - should result in default values being applied
	configContent := `{
		"governance": {},
		"runtime": {}
	}`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	os.Setenv("GLASSOPS_CONFIG_PATH", configPath)
	defer os.Unsetenv("GLASSOPS_CONFIG_PATH")

	engine := New()
	config, err := engine.Load()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if config.Governance.Enabled != false {
		t.Error("expected governance.enabled to be false for default config")
	}

	if config.Runtime.CLIVersion != "latest" {
		t.Errorf("expected cli_version 'latest', got '%s'", config.Runtime.CLIVersion)
	}
}

func TestLoadValidConfig(t *testing.T) {
	// Create temp config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "devops-config.json")

	configContent := `{
		"governance": {
			"enabled": true,
			"freeze_windows": [
				{"day": "Friday", "start": "17:00", "end": "23:59"}
			],
			"plugin_whitelist": ["sfdx-hardis@^4.0.0", "@salesforce/plugin-deploy-retrieve"]
		},
		"runtime": {
			"cli_version": "2.0.0",
			"node_version": "20"
		}
	}`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	os.Setenv("GITHUB_WORKSPACE", tempDir)
	os.Setenv("GLASSOPS_CONFIG_PATH", "devops-config.json")
	defer func() {
		os.Unsetenv("GITHUB_WORKSPACE")
		os.Unsetenv("GLASSOPS_CONFIG_PATH")
	}()

	engine := New()
	config, err := engine.Load()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !config.Governance.Enabled {
		t.Error("expected governance.enabled to be true")
	}

	if len(config.Governance.FreezeWindows) != 1 {
		t.Errorf("expected 1 freeze window, got %d", len(config.Governance.FreezeWindows))
	}

	if config.Governance.FreezeWindows[0].Day != "Friday" {
		t.Errorf("expected freeze window day 'Friday', got '%s'", config.Governance.FreezeWindows[0].Day)
	}

	if len(config.Governance.PluginWhitelist) != 2 {
		t.Errorf("expected 2 whitelisted plugins, got %d", len(config.Governance.PluginWhitelist))
	}

	if config.Runtime.CLIVersion != "2.0.0" {
		t.Errorf("expected cli_version '2.0.0', got '%s'", config.Runtime.CLIVersion)
	}
}

func TestCheckFreezeNoWindows(t *testing.T) {
	engine := New()
	config := &Config{
		Governance: GovernanceConfig{
			Enabled:       true,
			FreezeWindows: nil,
		},
	}

	err := engine.CheckFreeze(config)
	if err != nil {
		t.Errorf("expected no error for empty freeze windows, got %v", err)
	}
}

func TestCheckFreezeOutsideWindow(t *testing.T) {
	engine := New()

	// Use a day that is definitely not today (in UTC)
	now := time.Now().UTC()
	days := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	currentDay := days[now.Weekday()]

	// Pick a different day
	freezeDay := "Sunday"
	if currentDay == "Sunday" {
		freezeDay = "Monday"
	}

	config := &Config{
		Governance: GovernanceConfig{
			Enabled: true,
			FreezeWindows: []FreezeWindow{
				{Day: freezeDay, Start: "00:00", End: "23:59"},
			},
		},
	}

	err := engine.CheckFreeze(config)
	if err != nil {
		t.Errorf("expected no error outside freeze window, got %v", err)
	}
}
