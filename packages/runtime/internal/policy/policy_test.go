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

func TestValidatePluginWhitelistEmpty(t *testing.T) {
	engine := New()
	config := &Config{
		Governance: GovernanceConfig{
			PluginWhitelist: nil,
		},
	}

	// Empty whitelist allows all plugins
	if !engine.ValidatePluginWhitelist(config, "any-plugin") {
		t.Error("expected empty whitelist to allow all plugins")
	}
}

func TestValidatePluginWhitelistMatch(t *testing.T) {
	engine := New()
	config := &Config{
		Governance: GovernanceConfig{
			PluginWhitelist: []string{
				"sfdx-hardis@^4.0.0",
				"@salesforce/plugin-deploy-retrieve",
			},
		},
	}

	if !engine.ValidatePluginWhitelist(config, "sfdx-hardis") {
		t.Error("expected sfdx-hardis to be whitelisted")
	}

	if !engine.ValidatePluginWhitelist(config, "@salesforce/plugin-deploy-retrieve") {
		t.Error("expected @salesforce/plugin-deploy-retrieve to be whitelisted")
	}

	if engine.ValidatePluginWhitelist(config, "not-whitelisted") {
		t.Error("expected not-whitelisted to be rejected")
	}
}

func TestGetPluginVersionConstraint(t *testing.T) {
	engine := New()
	config := &Config{
		Governance: GovernanceConfig{
			PluginWhitelist: []string{
				"sfdx-hardis@^4.0.0",
				"@salesforce/plugin-deploy-retrieve@2.0.0",
				"no-version-plugin",
			},
		},
	}

	constraint := engine.GetPluginVersionConstraint(config, "sfdx-hardis")
	if constraint != "^4.0.0" {
		t.Errorf("expected '^4.0.0', got '%s'", constraint)
	}

	constraint = engine.GetPluginVersionConstraint(config, "@salesforce/plugin-deploy-retrieve")
	if constraint != "2.0.0" {
		t.Errorf("expected '2.0.0', got '%s'", constraint)
	}

	constraint = engine.GetPluginVersionConstraint(config, "no-version-plugin")
	if constraint != "" {
		t.Errorf("expected empty string for no version, got '%s'", constraint)
	}

	constraint = engine.GetPluginVersionConstraint(config, "not-in-list")
	if constraint != "" {
		t.Errorf("expected empty string for missing plugin, got '%s'", constraint)
	}
}

func TestExtractPluginName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"sfdx-hardis@^4.0.0", "sfdx-hardis"},
		{"sfdx-hardis", "sfdx-hardis"},
		{"@salesforce/plugin-deploy-retrieve@2.0.0", "@salesforce/plugin-deploy-retrieve"},
		{"@salesforce/plugin-deploy-retrieve", "@salesforce/plugin-deploy-retrieve"},
		{"@scope/package@1.0.0-beta.1", "@scope/package"},
	}

	for _, test := range tests {
		result := extractPluginName(test.input)
		if result != test.expected {
			t.Errorf("extractPluginName(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestExtractVersionConstraint(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"sfdx-hardis@^4.0.0", "^4.0.0"},
		{"sfdx-hardis", ""},
		{"@salesforce/plugin-deploy-retrieve@2.0.0", "2.0.0"},
		{"@salesforce/plugin-deploy-retrieve", ""},
		{"@scope/package@1.0.0-beta.1", "1.0.0-beta.1"},
	}

	for _, test := range tests {
		result := extractVersionConstraint(test.input)
		if result != test.expected {
			t.Errorf("extractVersionConstraint(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}
