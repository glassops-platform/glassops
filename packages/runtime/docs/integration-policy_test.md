---
type: Documentation
domain: runtime
origin: packages/runtime/internal/integration/policy_test.go
last_modified: 2026-01-31
generated: true
source: packages/runtime/internal/integration/policy_test.go
generated_at: 2026-01-31T09:06:46.871197
hash: 043050be51f99529cb2e5a145137b1818e025062b4529569289a697c01e2321b
---

## Policy Integration Test Documentation

This document describes the integration tests for the policy engine. These tests verify the correct behavior of the policy loading, validation, and application processes.

**Package Purpose:**

The `integration` package contains tests that exercise the `policy` package by simulating real-world scenarios. These tests ensure that the policy engine interacts correctly with configuration data and enforces defined rules.

**Key Types and Interfaces:**

*   **`policy.Engine`**: This type, defined in the `policy` package, is the core component responsible for loading, parsing, and validating policy configurations. It provides methods for accessing and interpreting policy rules.
*   **`config` (returned by `engine.Load()`):** Represents the loaded policy configuration. It contains settings for governance (enabled status, freeze windows, plugin whitelist) and runtime environment details (CLI and Node versions).

**Important Functions and Behavior:**

*   **`SetupTestWorkspace(testConfig map[string]interface{})`**: This function creates a temporary workspace for running tests. It allows writing a configuration file to disk, simulating a real-world deployment scenario. It returns an environment object with cleanup capabilities and an error if setup fails.
*   **`engine.Load()`**: This function loads the policy configuration from the configured file path (defined by the `GLASSOPS_CONFIG_PATH` environment variable). If the configuration file is missing, it loads a default configuration. It returns the loaded configuration and an error if loading fails (e.g., invalid JSON).
*   **`engine.ValidatePluginWhitelist(config, pluginName string)`**: This function checks if a given plugin is allowed based on the configured plugin whitelist. It returns `true` if the plugin is whitelisted, and `false` otherwise.
*   **`engine.GetPluginVersionConstraint(config, pluginName string)`**: This function retrieves the version constraint for a given plugin from the configuration.

**Error Handling:**

The tests extensively check for errors returned by the `SetupTestWorkspace` and `engine.Load` functions.  Failures during workspace setup or configuration loading result in test failures.  The tests also verify that invalid configuration data (e.g., incorrect freeze window day) results in an error during the `Load` process.

**Concurrency:**

These tests do not explicitly use goroutines or channels. They are designed as sequential integration tests.

**Notable Design Decisions:**

*   **Configuration File Handling:** The tests simulate missing configuration files to verify the default configuration behavior.
*   **Test Workspace:** The `SetupTestWorkspace` function provides a clean and isolated environment for each test, preventing interference between tests. The `defer env.Cleanup()` ensures that the workspace is removed after each test, regardless of success or failure.
*   **Comprehensive Validation:** The tests cover various scenarios, including valid and invalid configurations, plugin whitelisting, and version constraint extraction, to ensure the policy engine functions correctly under different conditions.
*   **Environment Variables:** The tests use the `GLASSOPS_CONFIG_PATH` environment variable to control the location of the configuration file, mimicking how the policy engine would be used in a production environment.