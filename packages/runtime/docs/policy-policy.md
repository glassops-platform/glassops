---
type: Documentation
domain: runtime
origin: packages/runtime/internal/policy/policy.go
last_modified: 2026-01-31
generated: true
source: packages/runtime/internal/policy/policy.go
generated_at: 2026-01-31T09:07:12.631683
hash: 8a1a02f4d3862565e4162e9128051a7cf16adfedea8d446cdb6684bf679de150
---

## Policy Engine Documentation

This document describes the Policy Engine package, responsible for managing and enforcing governance policies within the system. It provides a mechanism to control deployments, manage plugin usage, and enforce runtime environment standards.

**Package Purpose and Responsibilities**

The `policy` package provides the core functionality for loading, validating, and applying governance policies. These policies define rules related to deployment timing (freeze windows), permitted plugins, and runtime environment requirements.  We aim to provide a flexible and configurable system to ensure deployments adhere to organizational standards and security best practices.

**Key Types and Interfaces**

*   **`Config`**:  The top-level structure representing the complete governance configuration. It contains `GovernanceConfig` and `RuntimeConfig`.
*   **`GovernanceConfig`**: Holds settings related to governance rules, including whether governance is enabled, defined freeze windows, a whitelist of allowed plugins, and static analyzer configuration.
*   **`RuntimeConfig`**: Stores runtime environment settings such as the required CLI version and Node.js version.
*   **`FreezeWindow`**: Defines a specific time window (day and time range) during which deployments are prohibited.
*   **`AnalyzerConfig`**: Configures the static analysis component, including enabling/disabling it, setting a severity threshold for reported issues, and specifying rulesets to apply.
*   **`Engine`**:  The central component responsible for loading and applying the governance policy. It encapsulates the configuration path and provides methods for policy enforcement.

**Important Functions and Their Behavior**

*   **`New()`**: Creates a new `Engine` instance. It determines the configuration file path, prioritizing the `GLASSOPS_CONFIG_PATH` environment variable, and falling back to a default location (`config/devops-config.json`) if the environment variable is not set. It also uses the `GITHUB_WORKSPACE` environment variable to resolve the path, defaulting to the current directory if not set.
*   **`Load()`**: Reads the governance configuration from the specified file. If the file does not exist, it logs a warning and returns a default configuration with governance disabled. It parses the JSON configuration, applies default values for missing fields (CLI version, Node version, analyzer severity), and validates the freeze window definitions.  Errors during file reading or JSON parsing are returned.
*   **`CheckFreeze()`**:  Determines if the current time falls within a defined freeze window. It iterates through the configured freeze windows and compares the current day and time against the window's settings. If a match is found, an error is returned indicating that deployments are blocked.
*   **`ValidatePluginWhitelist()`**: Checks if a given plugin is permitted based on the configured plugin whitelist. If the whitelist is empty, all plugins are allowed. Otherwise, it iterates through the whitelist and compares the plugin name against the whitelisted entries.
*   **`GetPluginVersionConstraint()`**: Retrieves the version constraint associated with a whitelisted plugin. If the plugin is whitelisted, the function extracts and returns the version constraint string.
*   **`extractPluginName()`**:  Helper function to extract the plugin name from a string that may include a version. It handles both scoped and regular packages.
*   **`extractVersionConstraint()`**: Helper function to extract the version constraint from a string that may include a version.

**Error Handling Patterns**

The package employs standard Go error handling practices. Functions return an error value alongside their primary return value. Errors are often wrapped using `fmt.Errorf("%w", err)` to provide context and preserve the original error for debugging.  Specific error messages are provided for invalid configuration data, such as incorrect freeze window formats or missing configuration files.

**Concurrency Patterns**

This package is primarily focused on configuration loading and validation, and does not currently employ goroutines or channels for concurrent operations.  It is designed to be thread-safe as its methods do not share mutable state.

**Notable Design Decisions**

*   **Configuration File Location:** The configuration file path is configurable via the `GLASSOPS_CONFIG_PATH` environment variable, providing flexibility in different deployment environments.
*   **Default Configuration:** A default configuration is provided when the configuration file is missing, allowing the system to function (albeit with governance disabled) even without explicit configuration.
*   **Plugin Whitelisting:** The plugin whitelist provides a mechanism to control which plugins are permitted, enhancing security and ensuring compliance with organizational standards.
*   **Freeze Windows:** The freeze window feature allows for scheduled maintenance or blackout periods during which deployments are blocked.
*   **Validation:** Input validation is performed on the configuration data to ensure data integrity and prevent unexpected behavior.