---
type: Documentation
domain: runtime
origin: packages/runtime/internal/policy/policy_test.go
last_modified: 2026-01-31
generated: true
source: packages/runtime/internal/policy/policy_test.go
generated_at: 2026-01-31T10:03:33.769790
hash: 441ac8ad7b79a0cf34d3855a487021291d095753f60c207942d2646ead09cd39
---

## Policy Package Documentation

This package defines the policy engine responsible for governing operations within a runtime environment. It manages configuration loading, freeze window enforcement, and plugin whitelisting to ensure controlled and secure execution.

**Key Types and Interfaces**

*   **Config:** This structure holds the entire configuration loaded from a file or defaults. It contains `Governance` and `Runtime` fields.
*   **GovernanceConfig:**  Embedded within `Config`, this structure defines governance-related settings, including whether governance is enabled, a list of freeze windows, and a whitelist of allowed plugins.
*   **FreezeWindow:** Represents a time period during which certain operations are restricted. It includes the `Day` of the week and `Start` and `End` times.
*   **Engine:** The primary type representing the policy engine. It provides methods for loading configuration, checking for freeze windows, and validating plugins.

**Important Functions**

*   **New(): Engine:**  This function creates and returns a new instance of the `Engine`.
*   **Load(): (Config, error):** This method loads the configuration. It first attempts to load from the file specified by the `GLASSOPS_CONFIG_PATH` environment variable. If the file is not found, it returns a default configuration.  An error is returned if there are issues during file reading or configuration parsing.
*   **CheckFreeze(config *Config): error:** This function checks if the current time falls within a defined freeze window specified in the provided `Config`. If governance is not enabled or no freeze windows are defined, it returns no error. Otherwise, it returns an error if the current time is within a freeze window.
*   **ValidatePluginWhitelist(config *Config, plugin string): bool:** This function checks if a given plugin is allowed based on the `PluginWhitelist` defined in the `Config`. If the whitelist is empty, all plugins are considered allowed.
*   **GetPluginVersionConstraint(config *Config, plugin string): string:** This function retrieves the version constraint for a given plugin from the `PluginWhitelist`. If the plugin is not found in the whitelist or no version constraint is specified, an empty string is returned.
*   **extractPluginName(pluginWithVersion string) string:** This helper function extracts the plugin name from a string that may include a version constraint (e.g., "sfdx-hardis@^4.0.0").
*   **extractVersionConstraint(pluginWithVersion string) string:** This helper function extracts the version constraint from a string that may include a plugin name (e.g., "sfdx-hardis@^4.0.0").

**Error Handling**

The package employs standard Go error handling practices. Functions return an `error` value to indicate failure.  The `Load` function specifically handles errors related to file access and configuration parsing.  `CheckFreeze` returns an error when a freeze window is active.

**Concurrency**

This package does not explicitly use goroutines or channels. It is designed to be used in a single-threaded manner within the context of the larger application.

**Design Decisions**

*   **Configuration Source:** The package relies on an environment variable (`GLASSOPS_CONFIG_PATH`) to determine the location of the configuration file. This allows for flexibility in deployment environments.
*   **Default Configuration:** A default configuration is provided to ensure the system functions even without an explicit configuration file.
*   **Plugin Whitelisting:** The plugin whitelisting mechanism allows for fine-grained control over which plugins are permitted to run, enhancing security.
*   **Freeze Windows:** The freeze window feature enables scheduled restrictions on operations, preventing actions during critical periods.