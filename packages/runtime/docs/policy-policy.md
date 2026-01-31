---
type: Documentation
domain: runtime
origin: packages/runtime/internal/policy/policy.go
last_modified: 2026-01-29
generated: true
source: packages/runtime/internal/policy/policy.go
generated_at: 2026-01-29T21:25:45.744377
hash: 8a1a02f4d3862565e4162e9128051a7cf16adfedea8d446cdb6684bf679de150
---

## Policy Engine Documentation

This document describes the Policy Engine, a component responsible for enforcing governance rules within the system. It manages configuration loading, freeze window validation, and plugin whitelisting.

**Package Purpose and Responsibilities**

The `policy` package provides the functionality to load, parse, and apply governance policies. These policies control aspects such as deployment timing (freeze windows) and permitted plugins. The engine aims to provide a centralized point for managing and enforcing these rules, enhancing security and stability.

**Key Types and Interfaces**

*   **`Config`**: This structure represents the complete governance configuration, encompassing both governance-specific settings and runtime environment settings.
*   **`GovernanceConfig`**: Contains settings related to governance enforcement, including whether governance is enabled, defined freeze windows, and a whitelist of allowed plugins.
*   **`RuntimeConfig`**: Holds runtime environment details like CLI and Node versions.
*   **`FreezeWindow`**: Defines a specific time window (day and time range) during which deployments are prohibited.
*   **`AnalyzerConfig`**: Configures static analysis settings, including enabling/disabling analysis, setting a severity threshold, and specifying rulesets.
*   **`Engine`**: The core type that manages policy loading and enforcement. It holds the path to the configuration file.

**Important Functions and Their Behavior**

*   **`New()`**:  This function creates a new `Engine` instance. It determines the configuration file path by first checking the `GLASSOPS_CONFIG_PATH` environment variable. If not set, it defaults to `"config/devops-config.json"`. It also uses the `GITHUB_WORKSPACE` environment variable to resolve the path, defaulting to the current directory if not set.
*   **`Load()`**: This function reads the governance configuration from the configured file path. If the file does not exist, it logs a warning and returns a default configuration with governance disabled. It parses the file as JSON and applies default values if certain configuration options are missing (CLI version, Node version, Analyzer severity threshold). It also validates the format of freeze window times and days, returning an error if invalid.
*   **`CheckFreeze(config *Config)`**: This function checks if the current time falls within any defined freeze windows. It iterates through the `FreezeWindows` in the provided `Config` and compares the current day and time against each window's settings. If a match is found, it returns an error indicating that deployments are blocked.
*   **`ValidatePluginWhitelist(config *Config, pluginName string)`**: This function determines whether a given plugin is allowed based on the configured whitelist. If the whitelist is empty, all plugins are allowed. Otherwise, it checks if the plugin name exists within the whitelist.
*   **`GetPluginVersionConstraint(config *Config, pluginName string)`**: This function retrieves the version constraint for a given plugin from the whitelist, if it exists.
*   **`extractPluginName(entry string)`**: This helper function extracts the plugin name from a string that may include version information (e.g., "@scope/package@1.0.0" returns "@scope/package").
*   **`extractVersionConstraint(entry string)`**: This helper function extracts the version constraint from a string that may include version information (e.g., "@scope/package@1.0.0" returns "1.0.0").

**Error Handling Patterns**

The functions in this package employ standard Go error handling. Errors are returned as the second return value, allowing the caller to check for and handle potential issues. Errors are often wrapped using `fmt.Errorf` with `%w` to preserve the original error context.  Specific errors are returned for invalid configuration file paths, invalid JSON format, invalid freeze window settings, and deployment blocks due to freeze windows.

**Concurrency Patterns**

This package does not currently employ explicit concurrency patterns like goroutines or channels. The functions are designed to be called sequentially.

**Notable Design Decisions**

*   **Configuration File Path**: The configuration file path is determined by environment variables (`GLASSOPS_CONFIG_PATH` and `GITHUB_WORKSPACE`) to provide flexibility and integration with different environments.
*   **Default Configuration**: A default configuration is provided when the configuration file is missing, ensuring that the system can still operate (albeit with governance disabled) in the absence of a valid configuration.
*   **Freeze Window Validation**:  Input validation is performed on freeze window settings to prevent invalid configurations from causing unexpected behavior.
*   **Plugin Whitelisting**: The plugin whitelisting mechanism allows for fine-grained control over which plugins are permitted, enhancing security. The functions to extract plugin names and version constraints provide flexibility in how whitelist entries are formatted.