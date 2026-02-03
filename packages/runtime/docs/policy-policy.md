---
type: Documentation
domain: runtime
last_modified: 2026-02-02
generated: true
source: packages/runtime/internal/policy/policy.go
generated_at: 2026-02-02T22:38:52.063444
hash: 8a1a02f4d3862565e4162e9128051a7cf16adfedea8d446cdb6684bf679de150
---

## Policy Package Documentation

This package implements the governance policy engine, responsible for enforcing organizational rules and constraints during deployments and runtime operations. It provides mechanisms for controlling when deployments can occur, which plugins are permitted, and how runtime environments are configured.

### Key Types

*   **`Config`**: The top-level structure representing the complete governance configuration. It contains `GovernanceConfig` and `RuntimeConfig`.
*   **`GovernanceConfig`**: Holds settings related to governance, including enabling/disabling governance, defining freeze windows, and managing a plugin whitelist.
*   **`FreezeWindow`**: Defines a specific time period (day and time range) during which deployments are prohibited.
*   **`AnalyzerConfig`**: Configures static analysis settings, such as enabling the analyzer, setting a severity threshold, and specifying rulesets.
*   **`RuntimeConfig`**: Stores runtime environment settings, like the required CLI and Node.js versions.
*   **`Engine`**:  Manages the loading and enforcement of the governance policy.

### Engine Functions

*   **`New()`**: Creates a new `Engine` instance. It determines the configuration file path by first checking the `GLASSOPS_CONFIG_PATH` environment variable, and falling back to a default path (`config/devops-config.json`) if the variable is not set. It also determines the workspace path using the `GITHUB_WORKSPACE` environment variable, defaulting to the current directory if not set.
*   **`Load()`**: Reads, parses, and validates the governance configuration from the file path specified during `Engine` creation. If the configuration file does not exist, it logs a warning and returns a default configuration with governance disabled. It handles potential errors during file reading and JSON unmarshaling. Default values are applied if certain configuration fields are missing (CLI version, Node version, Analyzer severity threshold). It validates the format of freeze window days and times, returning an error if invalid.
*   **`CheckFreeze(config *Config)`**:  Determines if the current time falls within any defined freeze windows. It returns an error if a freeze window is active, blocking deployment. If no freeze windows are defined, or the current time is outside of any windows, it returns nil.
*   **`ValidatePluginWhitelist(config *Config, pluginName string)`**: Checks if a given plugin is permitted based on the configured whitelist. If the whitelist is empty, all plugins are allowed. It extracts the plugin name from the whitelist entries and compares it to the provided `pluginName`.
*   **`GetPluginVersionConstraint(config *Config, pluginName string)`**: Retrieves the version constraint for a given plugin from the whitelist, if available.
*   **`extractPluginName(entry string)`**: Extracts the plugin name from a string that may include version information (e.g., "@scope/package@1.0.0" returns "@scope/package").
*   **`extractVersionConstraint(entry string)`**: Extracts the version constraint from a string that may include version information (e.g., "@scope/package@1.0.0" returns "1.0.0").

### Error Handling

The package employs standard Go error handling practices. Functions return an error value alongside their primary return value. Errors are often wrapped using `fmt.Errorf` to provide context and preserve the original error. Specific error messages are provided for invalid configuration file paths, invalid JSON format, invalid freeze window definitions, and active freeze windows.

### Concurrency

This package does not explicitly use goroutines or channels. It operates primarily on configuration data loaded during initialization and performs synchronous checks during runtime.

### Design Decisions

*   **Configuration File Location:** The configuration file path is configurable via the `GLASSOPS_CONFIG_PATH` environment variable, providing flexibility in different environments.
*   **Default Configuration:** A default configuration is provided when the configuration file is missing, ensuring the system can operate (albeit with reduced governance) even without a custom configuration.
*   **Plugin Whitelisting:** The plugin whitelist allows for fine-grained control over which plugins are permitted, enhancing security and compliance.
*   **Freeze Windows:** Freeze windows provide a mechanism to prevent deployments during critical periods, such as peak hours or scheduled maintenance.
*   **Validation:** Input validation is performed on freeze window data to ensure data integrity and prevent unexpected behavior.