---
type: Documentation
domain: runtime
origin: packages/runtime/internal/policy/policy.go
last_modified: 2026-02-01
generated: true
source: packages/runtime/internal/policy/policy.go
generated_at: 2026-02-01T19:43:05.088776
hash: 8a1a02f4d3862565e4162e9128051a7cf16adfedea8d446cdb6684bf679de150
---

## Policy Engine Documentation

This document describes the Policy Engine package, responsible for managing and enforcing governance policies within the system. It provides a mechanism to control deployments, plugin usage, and runtime environments based on configurable rules.

**Package Purpose and Responsibilities**

The `policy` package provides the core functionality for loading, validating, and applying governance policies. These policies define constraints on deployments, permitted plugins, and required runtime configurations. The engine aims to prevent unauthorized or risky actions by enforcing these policies.

**Key Types and Interfaces**

*   **`Config`**: This structure represents the complete governance configuration, encompassing both governance-specific settings and runtime environment settings.
*   **`GovernanceConfig`**: Contains settings related to governance enforcement, including enabling/disabling governance, defining freeze windows, and specifying a plugin whitelist.
*   **`RuntimeConfig`**: Holds configuration details about the runtime environment, such as the CLI version and Node.js version.
*   **`FreezeWindow`**: Defines a specific time window (day and time range) during which deployments are blocked.
*   **`AnalyzerConfig`**: Configures static analysis settings, including enabling/disabling analysis, setting a severity threshold, and specifying rulesets.
*   **`Engine`**: The central type that manages policy loading and enforcement. It encapsulates the configuration path and provides methods for interacting with the policy.

**Important Functions and Their Behavior**

*   **`New()`**: Creates a new `Engine` instance. It determines the configuration file path by first checking the `GLASSOPS_CONFIG_PATH` environment variable, and falling back to a default path (`config/devops-config.json`) if the variable is not set. It also determines the workspace path using the `GITHUB_WORKSPACE` environment variable, defaulting to the current directory if not set.
*   **`Load()`**: Reads the governance configuration from the specified file path. If the file does not exist, it logs a warning and returns a default configuration with governance disabled. It parses the JSON configuration file into a `Config` struct. It applies default values for missing runtime settings (CLI version defaults to "latest", Node version defaults to "20") and analyzer severity threshold (defaults to 1). It validates the format of freeze window times and days, returning an error if invalid.
*   **`CheckFreeze()`**: Validates whether the current time falls within any defined freeze windows. If a freeze window is active, it returns an error indicating that deployments are blocked.
*   **`ValidatePluginWhitelist()`**: Checks if a given plugin name is present in the configured plugin whitelist. If the whitelist is empty, all plugins are allowed.
*   **`GetPluginVersionConstraint()`**: Retrieves the version constraint associated with a whitelisted plugin. If the plugin is not whitelisted, an empty string is returned.
*   **`extractPluginName()`**: Extracts the plugin name from a string that may include a version constraint (e.g., "@scope/package@1.0.0" returns "@scope/package").
*   **`extractVersionConstraint()`**: Extracts the version constraint from a string that may include a plugin name (e.g., "@scope/package@1.0.0" returns "1.0.0").

**Error Handling Patterns**

The package employs standard Go error handling practices. Functions return an error value as the second return parameter. Errors are often wrapped using `fmt.Errorf` to provide context and preserve the original error using `%w`.  Missing configuration files or invalid JSON are handled gracefully, often with default values or informative error messages.

**Concurrency Patterns**

This package does not currently exhibit any explicit concurrency patterns (goroutines or channels). The operations are primarily focused on configuration loading and validation, which are generally performed sequentially.

**Notable Design Decisions**

*   **Configuration File Path:** The configuration file path is determined by environment variables (`GLASSOPS_CONFIG_PATH`, `GITHUB_WORKSPACE`) to provide flexibility and support different deployment environments.
*   **Default Values:** Sensible default values are applied for missing configuration parameters to ensure the system functions even with incomplete configurations.
*   **Plugin Whitelisting:** The plugin whitelisting mechanism allows for fine-grained control over permitted plugins and their versions. The `extractPluginName` and `extractVersionConstraint` functions support flexible whitelist entries.
*   **Freeze Windows:** The freeze window feature provides a mechanism to block deployments during critical periods, such as maintenance windows or holidays.