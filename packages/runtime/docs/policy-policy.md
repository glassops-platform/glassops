---
type: Documentation
domain: runtime
last_modified: 2026-02-03
generated: true
source: packages/runtime/internal/policy/policy.go
generated_at: 2026-02-03T18:08:18.210143
hash: 69a9c9d14eede75a910ad11f6529908cd08c13edbcad87dc2c2c6e9e471235f9
---

## Policy Engine Documentation

This document describes the Policy Engine package, responsible for governing runtime behavior based on a configurable policy. It provides mechanisms for controlling deployments, validating plugins, and enforcing static analysis standards.

### Package Responsibilities

The `policy` package handles:

- Loading governance policies from a JSON configuration file.
- Validating the configuration for correctness.
- Enforcing freeze windows to block deployments during specified times.
- Managing a whitelist of allowed plugins and their version constraints.
- Providing configuration data for runtime environments, such as CLI and Node versions.

### Key Types

- **`Config`**: The top-level structure representing the complete governance configuration. It contains `GovernanceConfig` and `RuntimeConfig`.
- **`GovernanceConfig`**: Holds settings related to governance, including enabling/disabling governance, defining freeze windows, specifying a plugin whitelist, and configuring static analysis.
- **`FreezeWindow`**: Defines a period when deployments are prohibited. It includes the `Day` of the week, `Start` time, and `End` time.
- **`AnalyzerConfig`**: Configures static analysis behavior, including enabling/disabling the analyzer, setting a severity threshold, and specifying rulesets.
- **`RuntimeConfig`**: Stores runtime environment settings like `CLIVersion` and `NodeVersion`.
- **`Engine`**:  The core type that manages policy loading and enforcement. It holds the path to the configuration file.

### Important Functions

- **`New()`**: Creates a new `Engine` instance. It determines the configuration file path from the `GLASSOPS_CONFIG_PATH` environment variable, defaulting to `config/devops-config.json`. If the `GITHUB_WORKSPACE` environment variable is set, the path is relative to that workspace; otherwise, it's relative to the current directory.
- **`Load()`**: Reads the governance configuration from the configured file path. If the file does not exist, it loads a default, unsafe policy with governance disabled. It parses the JSON data into a `Config` struct, applies default values for missing runtime settings, validates freeze window formats, and returns the configuration. Errors during file reading or JSON parsing are returned.
- **`CheckFreeze(config *Config)`**: Checks if the current time falls within any defined freeze window in the provided configuration. If a freeze window is active, it returns an error indicating that deployments are blocked.
- **`ValidatePluginWhitelist(config *Config, pluginName string)`**: Determines if a given plugin is allowed based on the configured whitelist. If the whitelist is empty, all plugins are allowed. It compares the provided `pluginName` against the whitelisted entries.
- **`GetPluginVersionConstraint(config *Config, pluginName string)`**: Retrieves the version constraint for a given plugin from the whitelist, if it exists.
- **`extractPluginName(entry string)`**: Extracts the plugin name from a string that may include version information (e.g., "@scope/package@1.0.0" returns "@scope/package").
- **`extractVersionConstraint(entry string)`**: Extracts the version constraint from a string that may include version information (e.g., "@scope/package@1.0.0" returns "1.0.0").

### Error Handling

The package employs standard Go error handling practices. Functions return an error value when operations fail, such as:

- File not found during configuration loading.
- Errors reading the configuration file.
- Invalid JSON format in the configuration file.
- Invalid data within the configuration (e.g., incorrect freeze window format).
- Deployment blocked due to an active freeze window.

Errors are often wrapped using `fmt.Errorf` with `%w` to preserve the original error context.

### Concurrency

This package does not explicitly use goroutines or channels. Configuration loading is a synchronous operation.

### Design Decisions

- **Configuration File Path:** The configuration file path is configurable via the `GLASSOPS_CONFIG_PATH` environment variable, providing flexibility in deployment environments.
- **Default Configuration:** A default, unsafe configuration is loaded if the configuration file is missing, allowing the system to function (albeit without governance) in the absence of a policy.
- **Plugin Whitelisting:** The plugin whitelist provides a mechanism for controlling which plugins are permitted, enhancing security and stability. Version constraints can be specified alongside plugin names.
- **Freeze Windows:** Freeze windows offer a way to prevent deployments during critical periods, such as peak usage times or scheduled maintenance.
- **Validation:** Input validation is performed on freeze window data to ensure correctness and prevent unexpected behavior.