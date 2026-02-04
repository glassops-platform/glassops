---
type: Documentation
domain: runtime
last_modified: 2026-02-04
generated: true
source: packages/runtime/internal/policy/policy.go
generated_at: 2026-02-04T01:34:35.646538
hash: a2e69545b8af26d59f81e3a5adc1a62a6ff960bac5a6ff82e4ee07b1cb798125
---

## Governance Policy Engine Documentation

This document describes the governance policy engine, responsible for managing and enforcing organizational policies related to deployments and runtime environments.

**Package Purpose:**

The `policy` package provides the functionality to load, validate, and apply governance rules. It controls aspects like deployment freeze windows, plugin whitelisting, and runtime environment constraints. This package aims to provide a centralized and configurable system for maintaining operational safety and compliance.

**Key Types and Interfaces:**

*   **`Config`**: The top-level structure representing the entire governance configuration. It contains `GovernanceConfig` and `RuntimeConfig`.
*   **`GovernanceConfig`**: Holds settings related to governance rules, including enabling/disabling governance, defining freeze windows, and managing a plugin whitelist.
*   **`RuntimeConfig`**: Stores runtime environment settings such as CLI and Node.js versions.
*   **`FreezeWindow`**: Defines a specific time window (day and time range) during which deployments are prohibited.
*   **`AnalyzerConfig`**: Configures static analysis settings, including enabling/disabling the analyzer, setting a severity threshold, and specifying rulesets.
*   **`Engine`**: The core type that manages policy loading and enforcement. It encapsulates the configuration path.

**Important Functions:**

*   **`New()`**: Creates a new `Engine` instance. It determines the configuration file path, prioritizing the `GLASSOPS_CONFIG_PATH` environment variable, falling back to a default location (`config/devops-config.json`). If the `GITHUB_WORKSPACE` environment variable is set, the path is resolved relative to that workspace.
*   **`Load()`**: Reads the governance configuration from the configured file path. If the file does not exist, it returns a default configuration with governance disabled and sets default runtime versions. It handles JSON unmarshaling errors and validates the configuration data, including freeze window times and days. Default values are applied if certain runtime settings are missing.
*   **`CheckFreeze()`**: Validates whether the current time falls within any defined freeze windows. If a match is found, it returns an error indicating that deployments are blocked.
*   **`ValidatePluginWhitelist()`**: Checks if a given plugin name is present in the configured whitelist. If the whitelist is empty, all plugins are allowed.
*   **`GetPluginVersionConstraint()`**: Retrieves the version constraint associated with a whitelisted plugin. Returns an empty string if the plugin is not whitelisted or no version constraint is specified.
*   **`extractPluginName()`**: Helper function to extract the plugin name from a string that may include a version.
*   **`extractVersionConstraint()`**: Helper function to extract the version constraint from a string that may include a plugin name.

**Error Handling:**

The package employs standard Go error handling practices. Functions return an error value alongside their primary return value. Errors are often wrapped using `fmt.Errorf` to provide context and preserve the original error. Specific error conditions, such as file not found or invalid JSON, are handled gracefully, often with informative error messages.

**Concurrency:**

This package is not inherently concurrent. The `Load()` function performs file I/O, which could be made concurrent with goroutines if performance becomes a concern, but the current implementation is single-threaded.

**Design Decisions:**

*   **Configuration File Path:** The package supports both absolute and relative configuration file paths. The use of environment variables (`GLASSOPS_CONFIG_PATH`, `GITHUB_WORKSPACE`) allows for flexibility in different deployment environments.
*   **Default Configuration:** A default configuration is provided when the configuration file is missing, ensuring that the system can operate even without explicit configuration. Governance is disabled by default in this scenario.
*   **Whitelist Behavior:** An empty plugin whitelist allows all plugins, providing a simple way to disable whitelisting.
*   **Time Validation:** Freeze window times are validated to ensure they are in the correct format (HH:MM).
*   **Plugin Name Extraction:** The `extractPluginName` and `extractVersionConstraint` functions provide a robust way to parse plugin names and version constraints from strings, accommodating both scoped and unscoped packages.