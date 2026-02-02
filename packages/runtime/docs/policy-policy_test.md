---
type: Documentation
domain: runtime
origin: packages/runtime/internal/policy/policy_test.go
last_modified: 2026-02-01
generated: true
source: packages/runtime/internal/policy/policy_test.go
generated_at: 2026-02-01T19:43:22.842119
hash: 441ac8ad7b79a0cf34d3855a487021291d095753f60c207942d2646ead09cd39
---

## Policy Package Documentation

This package defines the policy engine responsible for governing operations within a runtime environment. It manages configuration loading, freeze window enforcement, and plugin whitelisting to ensure controlled and secure execution.

**Key Types and Interfaces**

*   **Config:** This structure holds the entire configuration loaded from a file or defaults. It contains `Governance` and `Runtime` settings.
*   **GovernanceConfig:**  Nested within `Config`, this structure defines governance-related settings, including whether governance is enabled, a list of freeze windows, and a whitelist of allowed plugins.
*   **FreezeWindow:** Represents a time period during which certain operations are restricted. It includes the `Day` of the week and the `Start` and `End` times.
*   **Engine:** The primary type representing the policy engine. It provides methods for loading configuration, checking for freeze windows, and validating plugins.

**Important Functions**

*   **New(): Engine:**  This function creates and returns a new instance of the `Engine`. It initializes the engine with default settings.
*   **Load(): (Config, error):** This method loads the configuration. It first checks for the `GLASSOPS_CONFIG_PATH` environment variable. If the variable is set, it attempts to load the configuration from the specified file. If the file does not exist or is invalid, it returns a default configuration.  If the environment variable is not set, it returns a default configuration.
*   **CheckFreeze(config *Config): error:** This function checks if the current time falls within a defined freeze window specified in the provided `Config`. If governance is not enabled or no freeze windows are defined, it returns no error. Otherwise, it evaluates the current day and time against the configured freeze windows and returns an error if a freeze window is active.
*   **ValidatePluginWhitelist(config *Config, plugin string): bool:** This function checks if a given plugin is allowed based on the configured whitelist. If the whitelist is empty, all plugins are allowed. Otherwise, it checks if the plugin name exists in the whitelist.
*   **GetPluginVersionConstraint(config *Config, plugin string): string:** This function retrieves the version constraint for a given plugin from the whitelist. If the plugin is not found in the whitelist, or if no version constraint is specified for the plugin, it returns an empty string.
*   **extractPluginName(pluginWithVersion string) string:** This helper function extracts the plugin name from a string that may include a version constraint (e.g., "sfdx-hardis@^4.0.0").
*   **extractVersionConstraint(pluginWithVersion string) string:** This helper function extracts the version constraint from a string that may include a plugin name (e.g., "sfdx-hardis@^4.0.0").

**Error Handling**

The package employs standard Go error handling practices. Functions return an `error` value to indicate failure.  The `Load` function specifically handles file not found and invalid JSON errors.  Other functions return errors when validation checks fail, such as when a freeze window is active.

**Concurrency**

This package does not currently employ goroutines or channels. It operates in a single-threaded manner.

**Design Decisions**

*   **Configuration Loading:** The package prioritizes loading from a configuration file specified by the `GLASSOPS_CONFIG_PATH` environment variable. If the file is unavailable, it falls back to a default configuration, ensuring the system remains operational.
*   **Plugin Whitelisting:** The whitelist approach provides a simple and effective mechanism for controlling which plugins are permitted to run.
*   **Freeze Windows:** The use of named days of the week and time ranges for freeze windows offers a flexible way to define restricted periods.
*   **Helper Functions:** The `extractPluginName` and `extractVersionConstraint` functions promote code reusability and clarity by encapsulating the logic for parsing plugin strings.