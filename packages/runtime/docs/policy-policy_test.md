---
type: Documentation
domain: runtime
last_modified: 2026-02-04
generated: true
source: packages/runtime/internal/policy/policy_test.go
generated_at: 2026-02-04T01:35:08.021799
hash: a9cb94ab2fcf9819bfc11e411ffa8cc68281199ffdbea91af1173cd121a7885c
---

## Policy Package Documentation

This package defines the policy engine responsible for governing runtime behavior. It handles configuration loading, freeze window enforcement, and plugin whitelisting. The primary goal is to provide a mechanism for controlling when and how operations can be performed, and which tools are permitted.

**Key Types and Interfaces**

*   **Config:** This structure holds the entire configuration loaded from a JSON file or environment variables. It contains two main fields:
    *   `Governance`:  A `GovernanceConfig` struct that manages governance-related settings.
    *   `Runtime`: A struct managing runtime-specific settings like CLI and Node versions.
*   **GovernanceConfig:**  This structure encapsulates governance settings, including:
    *   `Enabled`: A boolean indicating whether governance is active.
    *   `FreezeWindows`: A slice of `FreezeWindow` structs defining periods when operations are prohibited.
    *   `PluginWhitelist`: A slice of strings representing allowed plugin names.
*   **FreezeWindow:** This structure defines a specific freeze period with:
    *   `Day`: The day of the week the freeze window applies to (e.g., "Friday").
    *   `Start`: The start time of the freeze window in HH:MM format (e.g., "17:00").
    *   `End`: The end time of the freeze window in HH:MM format (e.g., "23:59").
*   **Engine:** This type represents the policy engine itself. It provides methods for loading the configuration, checking for freeze windows, and validating plugins.

**Important Functions**

*   **New(): Engine:** This function creates and returns a new instance of the `Engine`.
*   **Load(): (*Config, error):** This function loads the configuration. It first checks for the `GLASSOPS_CONFIG_PATH` environment variable. If set, it attempts to read the configuration from the specified file. If the environment variable is not set, it attempts to load a default configuration.  It returns a pointer to the `Config` struct and an error if loading fails.
*   **CheckFreeze(*Config): error:** This function checks if the current time falls within a defined freeze window. It iterates through the `FreezeWindows` in the configuration and returns an error if a match is found. If no freeze windows are defined or the current time is outside of any window, it returns nil.
*   **ValidatePluginWhitelist(*Config, string): bool:** This function checks if a given plugin name is present in the `PluginWhitelist`. It returns `true` if the plugin is whitelisted or if the whitelist is empty (allowing all plugins), and `false` otherwise.
*   **GetPluginVersionConstraint(*Config, string): string:** This function retrieves the version constraint for a given plugin from the `PluginWhitelist`. It parses the plugin name and version (if present) and returns the version constraint string. If the plugin is not found in the whitelist or no version is specified, it returns an empty string.
*   **extractPluginName(string): string:** This helper function extracts the plugin name from a string that may include a version constraint (e.g., "sfdx-hardis@^4.0.0").
*   **extractVersionConstraint(string): string:** This helper function extracts the version constraint from a string that may include a plugin name (e.g., "sfdx-hardis@^4.0.0").

**Error Handling**

The package uses standard Go error handling patterns. Functions return an error value as the second return parameter.  The caller is responsible for checking the error value and handling it appropriately.  Errors typically indicate file I/O problems during configuration loading or issues with the configuration data itself.

**Concurrency**

This package does not explicitly use goroutines or channels. It is designed to be used in a single-threaded manner.

**Design Decisions**

*   **Configuration Loading:** The package supports loading configuration from a JSON file specified by the `GLASSOPS_CONFIG_PATH` environment variable. This allows for flexible configuration management.  Default values are applied if the configuration file is empty or missing.
*   **Plugin Whitelisting:** The plugin whitelisting mechanism allows for controlling which plugins are permitted to run. The version constraint parsing provides a way to enforce specific plugin versions.
*   **Freeze Windows:** The freeze window functionality provides a way to prevent operations during specific times, such as during critical maintenance periods.