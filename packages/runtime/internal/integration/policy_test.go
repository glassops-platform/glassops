package integration

import (
	"testing"

	"github.com/glassops-platform/glassops/packages/runtime/internal/policy"
)

func TestPolicyIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Run("loads default config when file missing", func(t *testing.T) {
		env, err := SetupTestWorkspace(nil)
		if err != nil {
			t.Fatalf("failed to setup workspace: %v", err)
		}
		defer env.Cleanup()

		// Remove the config file to test default behavior
		env.SetEnvironment(map[string]string{
			"GLASSOPS_CONFIG_PATH": "nonexistent/config.json",
		})

		engine := policy.New()
		config, err := engine.Load()
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if config.Governance.Enabled {
			t.Error("expected governance to be disabled for missing config")
		}
	})

	t.Run("loads valid governance config", func(t *testing.T) {
		testConfig := map[string]interface{}{
			"governance": map[string]interface{}{
				"enabled": true,
				"freeze_windows": []map[string]string{
					{"day": "Saturday", "start": "00:00", "end": "23:59"},
				},
				"plugin_whitelist": []string{"sfdx-hardis@^4.0.0"},
			},
			"runtime": map[string]interface{}{
				"cli_version":  "2.0.0",
				"node_version": "20",
			},
		}

		env, err := SetupTestWorkspace(testConfig)
		if err != nil {
			t.Fatalf("failed to setup workspace: %v", err)
		}
		defer env.Cleanup()
		env.SetEnvironment(nil)

		engine := policy.New()
		config, err := engine.Load()
		if err != nil {
			t.Fatalf("failed to load config: %v", err)
		}

		if !config.Governance.Enabled {
			t.Error("expected governance to be enabled")
		}
		if len(config.Governance.FreezeWindows) != 1 {
			t.Errorf("expected 1 freeze window, got %d", len(config.Governance.FreezeWindows))
		}
		if len(config.Governance.PluginWhitelist) != 1 {
			t.Errorf("expected 1 plugin in whitelist, got %d", len(config.Governance.PluginWhitelist))
		}
		if config.Runtime.CLIVersion != "2.0.0" {
			t.Errorf("expected CLI version 2.0.0, got %s", config.Runtime.CLIVersion)
		}
	})

	t.Run("validates plugin whitelist", func(t *testing.T) {
		testConfig := map[string]interface{}{
			"governance": map[string]interface{}{
				"enabled":          true,
				"plugin_whitelist": TestData.PluginConfigs.Whitelist,
			},
			"runtime": map[string]interface{}{},
		}

		env, err := SetupTestWorkspace(testConfig)
		if err != nil {
			t.Fatalf("failed to setup workspace: %v", err)
		}
		defer env.Cleanup()
		env.SetEnvironment(nil)

		engine := policy.New()
		config, err := engine.Load()
		if err != nil {
			t.Fatalf("failed to load config: %v", err)
		}

		// Should allow whitelisted plugin
		if !engine.ValidatePluginWhitelist(config, "sfdx-hardis") {
			t.Error("expected sfdx-hardis to be whitelisted")
		}

		// Should block non-whitelisted plugin
		if engine.ValidatePluginWhitelist(config, "malicious-plugin") {
			t.Error("expected malicious-plugin to be blocked")
		}

		// Should allow scoped plugin
		if !engine.ValidatePluginWhitelist(config, "@salesforce/plugin-deploy-retrieve") {
			t.Error("expected @salesforce/plugin-deploy-retrieve to be whitelisted")
		}
	})

	t.Run("extracts version constraints", func(t *testing.T) {
		testConfig := map[string]interface{}{
			"governance": map[string]interface{}{
				"enabled":          true,
				"plugin_whitelist": []string{"sfdx-hardis@^4.0.0", "@salesforce/cli@2.x"},
			},
			"runtime": map[string]interface{}{},
		}

		env, err := SetupTestWorkspace(testConfig)
		if err != nil {
			t.Fatalf("failed to setup workspace: %v", err)
		}
		defer env.Cleanup()
		env.SetEnvironment(nil)

		engine := policy.New()
		config, err := engine.Load()
		if err != nil {
			t.Fatalf("failed to load config: %v", err)
		}

		constraint := engine.GetPluginVersionConstraint(config, "sfdx-hardis")
		if constraint != "^4.0.0" {
			t.Errorf("expected ^4.0.0, got %s", constraint)
		}

		scopedConstraint := engine.GetPluginVersionConstraint(config, "@salesforce/cli")
		if scopedConstraint != "2.x" {
			t.Errorf("expected 2.x, got %s", scopedConstraint)
		}
	})

	t.Run("rejects invalid config", func(t *testing.T) {
		env, err := SetupTestWorkspace(nil)
		if err != nil {
			t.Fatalf("failed to setup workspace: %v", err)
		}
		defer env.Cleanup()
		env.SetEnvironment(nil)

		// Write invalid JSON
		env.WriteConfig(map[string]interface{}{
			"governance": map[string]interface{}{
				"enabled": true,
				"freeze_windows": []map[string]string{
					{"day": "InvalidDay", "start": "00:00", "end": "23:59"},
				},
			},
			"runtime": map[string]interface{}{},
		})

		engine := policy.New()
		_, err = engine.Load()
		if err == nil {
			t.Error("expected error for invalid freeze window day")
		}
	})
}
