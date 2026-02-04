---
type: Documentation
domain: runtime
last_modified: 2026-02-04
generated: true
source: packages/runtime/internal/integration/policy_test.go
generated_at: 2026-02-04T01:34:10.397400
hash: 69ef7cdaad7b2601051764a37468519a12593e49e8dfc03c57560a7ae26e3061
---

## Policy Integration Test Documentation

This document describes the integration tests for the policy engine. These tests verify the correct behavior of the policy loading, validation, and application processes.

**Package Purpose:**

The `integration` package contains tests that exercise the policy engine with realistic configurations and scenarios. These tests ensure that the policy engine interacts correctly with its configuration and provides expected results.

**Key Types and Interfaces:**

- `policy.Engine`: This type (defined in the `policy` package) represents the core policy engine. It is responsible for loading, validating, and applying policies.  We interact with it through its methods to test the configuration process.
- `map[string]interface{}`: This is used extensively to represent the configuration data. It allows for flexible configuration structures.

**Important Functions and Behavior:**

- `TestPolicyIntegration`: This is the main test function, containing several sub-tests (using `t.Run`). It sets up a test workspace, loads configurations, and asserts expected behavior.
- `SetupTestWorkspace(testConfig map[string]interface{})`: This function creates a temporary workspace for testing. It takes a configuration map as input and sets up the environment accordingly. It also handles cleanup after the test.
- `engine.Load()`: This method of the `policy.Engine` loads the configuration from the test workspace and populates the internal policy state. It returns a configuration object and an error if loading fails.
- `engine.ValidatePluginWhitelist(config, pluginName string) bool`: This method checks if a given plugin name is present in the configured whitelist. It returns `true` if the plugin is allowed, and `false` otherwise.
- `engine.GetPluginVersionConstraint(config, pluginName string) string`: This method retrieves the version constraint for a given plugin from the configuration.

**Test Cases:**

1. **Loads default config when file has empty values:** This test verifies that when the configuration file contains empty values, the policy engine applies default values. Specifically, it checks that `Governance.Enabled` is `false` when the `governance` section of the configuration is empty.

2. **Loads valid governance config:** This test validates that the policy engine correctly loads a valid configuration with governance enabled, freeze windows defined, and a plugin whitelist. It asserts that the loaded configuration matches the expected values.

3. **Validates plugin whitelist:** This test checks the functionality of the plugin whitelist validation. It verifies that whitelisted plugins are allowed, non-whitelisted plugins are blocked, and scoped plugins are handled correctly.

4. **Extracts version constraints:** This test confirms that the policy engine can correctly extract version constraints for plugins from the configuration. It checks that the correct constraints are returned for both regular and scoped plugins.

5. **Rejects invalid config:** This test ensures that the policy engine handles invalid configuration data gracefully. It writes an invalid configuration (specifically, an invalid day in a freeze window) and verifies that the `Load()` method returns an error.

**Error Handling:**

The tests extensively check for errors returned by the `SetupTestWorkspace` and `engine.Load()` functions.  If an error is encountered, the test immediately fails using `t.Fatalf` or `t.Error`. This ensures that any issues during configuration loading or validation are detected and reported.

**Concurrency:**

This test suite does not currently employ goroutines or channels, as the tests are focused on synchronous configuration loading and validation.