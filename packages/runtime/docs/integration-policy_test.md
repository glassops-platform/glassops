---
type: Documentation
domain: runtime
origin: packages/runtime/internal/integration/policy_test.go
last_modified: 2026-01-29
generated: true
source: packages/runtime/internal/integration/policy_test.go
generated_at: 2026-01-29T21:25:20.092062
hash: 043050be51f99529cb2e5a145137b1818e025062b4529569289a697c01e2321b
---

## Policy Integration Test Documentation

This document details the integration tests for the policy engine. These tests verify the engine’s ability to load configurations, validate plugin whitelists, and extract version constraints. The tests are designed to ensure the policy engine functions correctly in various scenarios, including missing configurations, valid configurations, and invalid configurations.

**Package Responsibilities:**

The `integration` package contains tests that exercise the `policy` package by simulating real-world scenarios. These tests focus on the interaction between the policy engine and its configuration, verifying expected behavior across different input conditions.

**Key Types and Interfaces:**

*   **`policy.Engine`**: This type, defined in the `policy` package, is the core component responsible for loading, parsing, and validating policy configurations. It provides methods for accessing and interpreting the policy rules.
*   **`policy.Config`**: This type represents the loaded policy configuration. It contains settings related to governance, runtime, and plugin whitelists.
*   **`SetupTestWorkspace`**: A function used to create a temporary workspace for testing. It allows writing configuration files and setting environment variables. It returns an environment object with a `Cleanup` method for resource management.

**Important Functions and Behavior:**

*   **`TestPolicyIntegration`**: This is the main test function that orchestrates all individual test cases. It uses subtests (`t.Run`) to organize and execute specific scenarios.
*   **`SetupTestWorkspace(testConfig map\[string]interface{})`**: This function sets up a test environment. It accepts an optional configuration map (`testConfig`) which is written to a temporary configuration file. It returns an environment object that can be used to manage the test workspace and clean up resources after the test.
*   **`engine.Load()`**: This method of the `policy.Engine` loads the policy configuration from the configured file path (determined by the `GLASSOPS_CONFIG_PATH` environment variable). If the configuration file is missing, it loads a default configuration. It returns a `policy.Config` object and an error if loading fails.
*   **`engine.ValidatePluginWhitelist(config *policy.Config, pluginName string) bool`**: This method validates whether a given plugin name is present in the configured whitelist. It returns `true` if the plugin is whitelisted, and `false` otherwise.
*   **`engine.GetPluginVersionConstraint(config *policy.Config, pluginName string) string`**: This method retrieves the version constraint for a given plugin name from the configuration. It returns the version constraint string (e.g., "^4.0.0") if found, or an empty string if not found.

**Error Handling:**

The tests extensively check for errors returned by the `policy.Engine` methods, particularly `Load()`.  If an error is expected, the test verifies that an error is indeed returned. If an error is *not* expected, the test fails if an error is encountered. Invalid configuration files are specifically tested to ensure appropriate error handling.

**Concurrency:**

This code does not exhibit any explicit concurrency patterns (goroutines, channels). The tests are designed to be executed sequentially.

**Notable Design Decisions:**

*   **Configuration Loading**: The policy engine supports loading configurations from a file specified by the `GLASSOPS_CONFIG_PATH` environment variable. If the file is missing, it falls back to a default configuration.
*   **Plugin Whitelisting**: The engine provides a mechanism to whitelist specific plugins, enhancing security by preventing the execution of unauthorized plugins.
*   **Version Constraints**: The engine supports specifying version constraints for whitelisted plugins, allowing for fine-grained control over plugin versions.
*   **Test Workspace**: The use of `SetupTestWorkspace` provides a clean and isolated environment for each test, preventing interference between tests. The `Cleanup` method ensures that temporary resources are released after each test.