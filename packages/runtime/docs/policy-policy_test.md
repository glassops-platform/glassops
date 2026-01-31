---
type: Documentation
domain: runtime
origin: packages/runtime/internal/policy/policy_test.go
last_modified: 2026-01-29
generated: true
source: packages/runtime/internal/policy/policy_test.go
generated_at: 2026-01-29T21:26:14.128050
hash: 441ac8ad7b79a0cf34d3855a487021291d095753f60c207942d2646ead09cd39
---

## Policy Package Documentation

This package defines the policy engine responsible for governing runtime behavior. It manages configuration loading, freeze window enforcement, and plugin whitelisting. The primary goal is to provide a mechanism for controlling when and how operations can be performed, enhancing stability and security.

**Key Types and Interfaces**

*   **Config:** This struct represents the overall configuration loaded from a file or defaults. It contains `Governance` and `Runtime` fields.
*   **GovernanceConfig:**  Embedded within `Config`, this struct holds settings related to governance, including a boolean `Enabled` flag, a slice of `FreezeWindow` objects, and a slice of strings representing the `PluginWhitelist`.
*   **FreezeWindow:** This struct defines a time window during which operations are restricted. It includes the `Day` of the week, `Start` time, and `End` time.
*   **Engine:** This type encapsulates the policy logic. It provides methods for loading the configuration, checking for freeze windows, and validating plugins.

**Important Functions**

*   **New() -> \*Engine:** This function creates and returns a new instance of the `Engine`.
*   **Load() (Config, error):** This method loads the configuration. It first attempts to load from the file specified by the `GLASSOPS_CONFIG_PATH` environment variable. If the file does not exist, it loads a default configuration.  It returns the loaded `Config` and any error encountered during the process.
*   **CheckFreeze(\[config \*Config]) error:** This function checks if the current time falls within a defined freeze window specified in the provided `Config`. It returns an error if a freeze window is active; otherwise, it returns nil.
*   **ValidatePluginWhitelist(\[config \*Config], pluginName string) bool:** This function checks if a given `pluginName` is present in the `PluginWhitelist` of the provided `Config`. It returns `true` if the plugin is whitelisted (or the whitelist is empty, allowing all plugins), and `false` otherwise.
*   **GetPluginVersionConstraint(\[config \*Config], pluginName string) string:** This function retrieves the version constraint for a given plugin from the `PluginWhitelist`. It parses the plugin string (e.g., "sfdx-hardis@^4.0.0") and returns the version constraint (e.g., "^4.0.0"). If the plugin is not found or no version is specified, it returns an empty string.
*   **extractPluginName(plugin string) string:** This helper function extracts the plugin name from a string that may include a version constraint (e.g., "sfdx-hardis@^4.0.0" returns "sfdx-hardis").
*   **extractVersionConstraint(plugin string) string:** This helper function extracts the version constraint from a string that may include a plugin name (e.g., "sfdx-hardis@^4.0.0" returns "^4.0.0").

**Error Handling**

The package employs standard Go error handling practices. Functions return an `error` value to indicate failure.  The calling code is responsible for checking and handling these errors.  In the provided tests, errors are checked using `if err != nil { t.Fatalf(...) }` to immediately halt execution if an unexpected error occurs.

**Concurrency**

This package does not explicitly use goroutines or channels. It operates in a single-threaded manner.

**Design Decisions**

*   **Configuration Loading:** The package supports loading configuration from a file specified by an environment variable (`GLASSOPS_CONFIG_PATH`). This allows for flexibility in deployment and configuration management. A default configuration is used if no file is found.
*   **Plugin Whitelisting:** The package uses a simple string-based whitelist for plugins. This approach is easy to implement and maintain. The `GetPluginVersionConstraint` function provides a way to retrieve version constraints associated with whitelisted plugins.
*   **Freeze Windows:** Freeze windows are defined by day and time ranges. This allows for restricting operations during specific periods, such as off-peak hours or maintenance windows.
*   **Testability:** The package is designed with testability in mind. The use of dependency injection (through the `Engine` type) and clear function signatures makes it easy to write unit tests.