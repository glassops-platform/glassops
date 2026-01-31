---
type: Documentation
domain: runtime
origin: packages/runtime/internal/policy/policy.go
last_modified: 2026-01-31
generated: true
source: packages/runtime/internal/policy/policy.go
generated_at: 2026-01-31T10:03:13.172560
hash: 8a1a02f4d3862565e4162e9128051a7cf16adfedea8d446cdb6684bf679de150
---

## Policy Engine Documentation

This document describes the policy engine package, responsible for governing deployments and runtime environments. It provides a mechanism to enforce rules related to deployment timing, allowed plugins, and runtime configurations.

**Package Responsibilities:**

The `policy` package handles loading, validating, and enforcing governance policies. It reads a configuration file, checks for deployment freeze windows, and validates plugin usage against a defined whitelist.

**Key Types:**

*   **`Config`**: The top-level structure representing the complete governance configuration. It contains `GovernanceConfig` and `RuntimeConfig`.
*   **`GovernanceConfig`**: Holds settings related to governance, including whether governance is enabled, defined freeze windows, a plugin whitelist, and analyzer configuration.
*   **`RuntimeConfig`**: Stores runtime environment settings such as CLI and Node.js versions.
*   **`FreezeWindow`**: Defines a specific time window (day and time range) during which deployments are prohibited.
*   **`AnalyzerConfig`**: Configures static analysis settings, including enabling/disabling the analyzer, setting a severity threshold, and specifying rulesets.
*   **`Engine`**:  Manages the policy loading and enforcement process. It encapsulates the configuration path.

**Important Functions:**

*   **`New()`**: Creates a new `Engine` instance. It determines the configuration file path from the `GLASSOPS_CONFIG_PATH` environment variable, defaulting to `config/devops-config.json` if the variable is not set. It also uses the `GITHUB_WORKSPACE` environment variable to determine the base path, defaulting to the current directory if not set.
*   **`Load()`**: Reads the governance configuration from the configured file path. If the file does not exist, it logs a warning and returns a default configuration with governance disabled. It unmarshals the JSON data into a `Config` struct, applies default values for missing runtime settings (CLI version defaults to "latest", Node version defaults to "20", analyzer severity defaults to 1), and validates the freeze window definitions. Returns a pointer to the `Config` struct and an error if any issues occur during file reading or parsing.
*   **`CheckFreeze(config *Config)`**: Checks if the current time falls within any defined freeze windows. It iterates through the `FreezeWindows` in the provided `Config` and returns an error if a match is found, indicating a deployment block. If no freeze windows are defined or the current time is outside of all windows, it returns nil.
*   **`ValidatePluginWhitelist(config *Config, pluginName string)`**: Checks if a given plugin is allowed based on the configured whitelist. If the whitelist is empty, all plugins are allowed. It iterates through the whitelist, extracting the plugin name and comparing it to the provided `pluginName`. Returns `true` if the plugin is whitelisted, `false` otherwise.
*   **`GetPluginVersionConstraint(config *Config, pluginName string)`**: Retrieves the version constraint for a given plugin from the whitelist. If the plugin is found in the whitelist, the version constraint associated with it is returned. If the plugin is not found or the whitelist is empty, an empty string is returned.
*   **`extractPluginName(entry string)`**: Extracts the plugin name from a string that may include a version. Handles both scoped packages (e.g., `@scope/package@1.0.0`) and regular packages (e.g., `package@1.0.0`).
*   **`extractVersionConstraint(entry string)`**: Extracts the version constraint from a string that may include a plugin name.

**Error Handling:**

The package employs standard Go error handling practices. Functions return an error value alongside their primary return value. Errors are often wrapped using `fmt.Errorf` to provide context and preserve the original error. Specific error conditions, such as invalid freeze window formats or missing configuration files, are handled with informative error messages.

**Concurrency:**

This package does not explicitly use goroutines or channels. It is designed to be used in a synchronous manner.

**Design Decisions:**

*   **Configuration File:** The policy engine relies on a JSON configuration file for defining governance rules. This allows for easy modification and version control of policies.
*   **Default Values:** Sensible default values are applied for runtime settings and analyzer severity to ensure reasonable behavior even with incomplete configurations.
*   **Whitelist Approach:** The plugin whitelist uses an allowlist approach, meaning only explicitly listed plugins are permitted. This provides a strong security posture.
*   **Time Zone:** The `CheckFreeze` function uses UTC time to avoid ambiguity related to time zones.
*   **Environment Variables:** The configuration file path is configurable via the `GLASSOPS_CONFIG_PATH` environment variable, providing flexibility in different environments.