---
type: Documentation
domain: runtime
origin: packages/runtime/internal/policy/policy_test.go
last_modified: 2026-01-31
generated: true
source: packages/runtime/internal/policy/policy_test.go
generated_at: 2026-01-31T09:07:33.497457
hash: 441ac8ad7b79a0cf34d3855a487021291d095753f60c207942d2646ead09cd39
---

## Policy Package Documentation

This package defines the policy engine responsible for governing runtime behavior. It manages configuration loading, freeze window enforcement, and plugin whitelisting.  We aim to provide a flexible and configurable system for controlling operations within a defined environment.

**Key Types and Interfaces**

*   **Config:** This structure holds the entire configuration loaded from a file or defaults. It contains `Governance` and `Runtime` sections.
*   **GovernanceConfig:**  Nested within `Config`, this structure defines governance-related settings, including whether governance is enabled, a list of freeze windows, and a whitelist of allowed plugins.
*   **FreezeWindow:** Represents a time period during which certain operations are restricted. It includes the `Day` of the week and `Start` and `End` times.
*   **PolicyEngine:** The primary type for interacting with the policy system. It provides methods for loading the configuration, checking for freeze windows, and validating plugins.

**Important Functions**

*   **New(): PolicyEngine:**  Creates and returns a new instance of the `PolicyEngine`.
*   **Load(): (Config, error):** Loads the configuration. It first attempts to load from the file specified by the `GLASSOPS_CONFIG_PATH` environment variable. If the file is not found, it returns a default configuration.  Errors are returned if there are issues reading the configuration file.
*   **CheckFreeze(config *Config): error:**  Determines if the current time falls within a defined freeze window. It returns an error if a freeze window is active; otherwise, it returns nil.
*   **ValidatePluginWhitelist(config *Config, plugin string): bool:** Checks if a given plugin is allowed based on the configured whitelist.  If the whitelist is empty, all plugins are permitted.
*   **GetPluginVersionConstraint(config *Config, plugin string): string:** Retrieves the version constraint for a given plugin from the whitelist. If the plugin is not found in the whitelist, an empty string is returned.
*   **extractPluginName(pluginWithVersion string) string:** A helper function that extracts the plugin name from a string that may include a version constraint (e.g., "sfdx-hardis@^4.0.0").
*   **extractVersionConstraint(pluginWithVersion string) string:** A helper function that extracts the version constraint from a string that may include a plugin name (e.g., "sfdx-hardis@^4.0.0").

**Error Handling**

The package employs standard Go error handling practices. Functions return an `error` value to indicate failure.  The `Load` function specifically returns an error if it cannot read the configuration file.  Other functions return errors when a policy check fails (e.g., `CheckFreeze` when a freeze window is active).

**Concurrency**

This package does not explicitly use goroutines or channels. It is designed to be used in a single-threaded manner within the application.

**Design Decisions**

*   **Configuration Source:** The package prioritizes loading configuration from an environment variable (`GLASSOPS_CONFIG_PATH`) to allow for flexible deployment and configuration management.
*   **Default Configuration:** A default configuration is provided to ensure the system functions even without an explicit configuration file.
*   **Plugin Whitelisting:** The plugin whitelisting mechanism allows for fine-grained control over which plugins are permitted to run, enhancing security and stability.
*   **Freeze Windows:** The freeze window feature enables scheduled restrictions on operations, preventing changes during critical periods.
*   **Helper Functions:** The `extractPluginName` and `extractVersionConstraint` functions are designed to parse plugin strings and provide flexibility in how plugin information is handled. You can use these functions to process plugin names and versions independently.