---
type: Documentation
domain: runtime
origin: packages/runtime/src/protocol/policy.test.ts
last_modified: 2026-01-26
generated: true
source: packages/runtime/src/protocol/policy.test.ts
generated_at: 2026-01-26T14:18:53.493Z
hash: 5c9303f9bb265d309d786dcb4512df7e8ad9b4c77e49a3511aa509b5890d4598
---

## Protocol Policy Document

This document details the Protocol Policy component, responsible for enforcing governance and security constraints during runtime operations. It outlines the policy’s functionality, including freeze window checks and plugin whitelisting.

**Overview**

The Protocol Policy component provides a mechanism to control when and how operations can be performed. It reads configuration from a `devops-config.json` file (if present) to determine governance settings. If the configuration file is missing, a default, permissive policy is applied.  The policy enforces rules related to deployment timing (freeze windows) and permitted plugins.

**Key Features**

*   **Freeze Window Enforcement:** Prevents deployments during specified time windows, such as weekends or specific maintenance periods. The policy checks the current time against configured freeze windows and throws an error if a deployment is attempted during a frozen period.
*   **Plugin Whitelisting:**  Ensures that only approved plugins are used.  The policy validates plugins against a whitelist defined in the configuration.
*   **Configuration Loading:** Loads governance settings from a `devops-config.json` file. Handles missing or invalid configuration files gracefully, falling back to a default policy or throwing an error.
*   **Version Constraints:** Supports version constraints for whitelisted plugins, allowing for specific versions or ranges of versions to be approved.

**Functionality Details**

1.  **`checkFreeze(config)`:**
    *   Takes a configuration object as input.
    *   Determines if the current time falls within a defined freeze window.
    *   Throws an error if the current time is within a freeze window, halting execution.
    *   Does nothing if no freeze windows are defined in the configuration.

2.  **`validatePluginWhitelist(config, pluginName)`:**
    *   Takes a configuration object and a plugin name as input.
    *   Checks if the specified plugin is present in the configured whitelist.
    *   Returns `true` if the plugin is whitelisted or if no whitelist is configured.
    *   Returns `false` if the plugin is not in the whitelist.

3.  **`getPluginVersionConstraint(config, pluginName)`:**
    *   Takes a configuration object and a plugin name as input.
    *   Retrieves the version constraint for the specified plugin from the whitelist, if one exists.
    *   Returns the version constraint string (e.g., "^6.0.0") if found.
    *   Returns `null` if the plugin is not in the whitelist or if no version constraint is specified.

4.  **`load()`:**
    *   Attempts to load governance settings from a `devops-config.json` file.
    *   If the file is missing, loads a default, permissive policy and issues a warning.
    *   If the file exists, parses the JSON content and validates it against a predefined schema.
    *   Throws an error if the JSON is invalid or if the configuration schema is invalid.
    *   Handles potential errors during file reading and parsing.

**Configuration**

The policy is configured through a `devops-config.json` file. The following settings are supported:

*   `governance.enabled`: A boolean indicating whether governance policies are enabled.
*   `governance.freeze_windows`: An array of objects defining freeze windows. Each object includes:
    *   `day`: The day of the week (e.g., "Friday").
    *   `start`: The start time of the freeze window (e.g., "09:00").
    *   `end`: The end time of the freeze window (e.g., "17:00").
*   `governance.plugin_whitelist`: An array of strings representing whitelisted plugins, optionally including version constraints (e.g., "sfdx-hardis@^6.0.0").
*   `runtime.cli_version`: The CLI version.
*   `runtime.node_version`: The Node.js version.

**Error Handling**

The policy handles errors related to configuration loading and validation.  Invalid JSON or schema errors result in an exception being thrown, preventing execution.  Freeze window violations also result in an exception.  Missing configuration files result in a warning and the application of a default policy.