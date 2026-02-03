---
type: Documentation
domain: runtime
last_modified: 2026-02-02
generated: true
source: packages/runtime/internal/policy/policy_test.go
generated_at: 2026-02-02T22:39:08.444515
hash: 441ac8ad7b79a0cf34d3855a487021291d095753f60c207942d2646ead09cd39
---

## Policy Package Documentation

This package defines the policy engine responsible for governing runtime behavior. It manages configuration loading, freeze window enforcement, and plugin whitelisting. The primary goal is to provide a mechanism for controlling when and how actions can be performed within the runtime environment.

**Key Types and Interfaces**

*   **Config:** This structure holds the entire configuration loaded from a file or defaults. It contains `Governance` and `Runtime` fields.
*   **GovernanceConfig:**  Embedded within `Config`, this structure defines governance-related settings, including whether governance is enabled, a list of freeze windows, and a whitelist of allowed plugins.
*   **FreezeWindow:** Represents a time window during which certain operations are prohibited. It includes the `Day` of the week and the `Start` and `End` times.
*   **Engine:** The central component of the policy package. It provides methods for loading the configuration, checking for freeze windows, and validating plugin whitelists.

**Important Functions**

*   **New(): Engine:**  This function creates and returns a new instance of the `Engine`. It initializes the engine with default settings.
*   **Load(): (Config, error):** This method loads the configuration. It first checks for the `GLASSOPS_CONFIG_PATH` environment variable. If the variable is set, it attempts to load the configuration from the specified file. If the file does not exist or is invalid, it returns a default configuration.
*   **CheckFreeze(config *Config): error:** This function checks if the current time falls within a defined freeze window. It iterates through the `FreezeWindows` in the provided `Config` and returns an error if a match is found. If no freeze windows are defined or the current time is outside of any window, it returns nil.
*   **ValidatePluginWhitelist(config *Config, plugin string): bool:** This function checks if a given plugin is allowed based on the `PluginWhitelist` in the `Config`. If the whitelist is empty, all plugins are allowed. Otherwise, it checks if the plugin name exists in the whitelist.
*   **GetPluginVersionConstraint(config *Config, plugin string): string:** This function retrieves the version constraint for a given plugin from the `PluginWhitelist`. It parses the plugin string (e.g., "sfdx-hardis@^4.0.0") and returns the version constraint (e.g., "^4.0.0"). If the plugin is not found in the whitelist or no version is specified, it returns an empty string.
*   **extractPluginName(plugin string): string:** This helper function extracts the plugin name from a string that may include a version constraint.
*   **extractVersionConstraint(plugin string): string:** This helper function extracts the version constraint from a string that may include a plugin name.

**Error Handling**

The `Load()` function returns an error if it fails to load the configuration file.  `CheckFreeze()` returns an error if the current time falls within a freeze window. Other functions return boolean values to indicate success or failure, with errors being returned in specific cases.

**Concurrency**

This package does not explicitly use goroutines or channels. It is designed to be used in a single-threaded manner within the runtime environment.

**Design Decisions**

*   **Configuration Loading:** The package prioritizes loading from a configuration file specified by the `GLASSOPS_CONFIG_PATH` environment variable. If the file is not found, it falls back to a default configuration.
*   **Plugin Whitelisting:** The package uses a simple string-based whitelist for plugins. The `GetPluginVersionConstraint` function allows for specifying version constraints for whitelisted plugins.
*   **Freeze Windows:** Freeze windows are defined by day of the week and time ranges, providing a flexible way to restrict operations during specific periods.
*   **Testability:** The package is designed with testability in mind, with clear separation of concerns and well-defined interfaces. The tests cover various scenarios, including loading default configurations, loading valid configurations, checking freeze windows, and validating plugin whitelists.