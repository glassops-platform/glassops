---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/protocol/policy.test.ts
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/src/protocol/policy.test.ts
generated_at: 2026-01-31T10:10:43.910721
hash: 3f4533598237bae7d4dbe66cd3c23671496fb9bde5d1b088263469c91d0aeb2e
---

## Protocol Policy Documentation

This document details the functionality of the Protocol Policy component. This component enforces governance and security constraints during runtime operations.

**Overview**

The Protocol Policy provides mechanisms to control deployment timing and permitted plugins. It reads configuration from a `devops-config.json` file, falling back to a default, permissive policy if the file is not found.

**Key Features**

*   **Freeze Window Enforcement:** Prevents deployments during specified time windows.
*   **Plugin Whitelisting:** Restricts execution to a pre-approved list of plugins.
*   **Configuration Loading:** Dynamically loads policy settings from a JSON file.
*   **Version Constraints:** Supports specifying version ranges for whitelisted plugins.

**Functionality**

**1. `checkFreeze(config)`**

This function verifies if the current time falls within a defined freeze window specified in the configuration.

*   **Parameters:**
    *   `config`: The Protocol Configuration object.
*   **Behavior:**
    *   If a freeze window is defined and the current time is within that window, an error is thrown with the message "FROZEN".
    *   If no freeze windows are defined, the function completes without error.

**2. `validatePluginWhitelist(config, pluginName)`**

This function checks if a given plugin is permitted based on the configured whitelist.

*   **Parameters:**
    *   `config`: The Protocol Configuration object.
    *   `pluginName`: The name of the plugin to validate.
*   **Behavior:**
    *   If no whitelist is configured, the function returns `true` (plugin is allowed).
    *   If the plugin name is present in the whitelist, the function returns `true`.
    *   If the plugin name, including any version constraint, is present in the whitelist, the function returns `true`.
    *   Otherwise, the function returns `false` (plugin is not allowed).

**3. `getPluginVersionConstraint(config, pluginName)`**

This function retrieves the version constraint for a given plugin from the whitelist, if specified.

*   **Parameters:**
    *   `config`: The Protocol Configuration object.
    *   `pluginName`: The name of the plugin to query.
*   **Behavior:**
    *   If no whitelist is configured, the function returns `null`.
    *   If the plugin is not in the whitelist, the function returns `null`.
    *   If the plugin is in the whitelist with a version constraint (e.g., `@^6.0.0`), the function returns the constraint string.
    *   If the plugin is in the whitelist without a version constraint, the function returns `null`.

**4. `load()`**

This asynchronous function loads the policy configuration from the `devops-config.json` file.

*   **Behavior:**
    *   It first checks for the existence of the `devops-config.json` file in the `config/` directory.
    *   If the file is not found, it logs a warning and returns a default, unsafe policy (governance disabled).
    *   If the file exists, it reads and parses the JSON content.
    *   If the JSON is invalid, or the configuration schema is invalid, it throws an error with the message "Invalid Governance Policy".
    *   If any other error occurs during file reading, it also throws an error with the message "Invalid Governance Policy".

**Configuration Schema**

The `devops-config.json` file should adhere to the following structure:

```json
{
  "governance": {
    "enabled": true | false,
    "freeze_windows": [
      {
        "day": "Monday" | "Tuesday" | "Wednesday" | "Thursday" | "Friday" | "Saturday" | "Sunday",
        "start": "HH:MM",
        "end": "HH:MM"
      }
    ],
    "plugin_whitelist": [
      "plugin-name",
      "plugin-name@version-constraint"
    ]
  },
  "runtime": {
    "cli_version": "latest",
    "node_version": "20"
  }
}
```

**Error Handling**

The component handles potential errors during configuration loading, including:

*   Missing configuration file.
*   Invalid JSON format.
*   Invalid configuration schema.
*   File reading errors.

In all error cases, an error with the message "Invalid Governance Policy" is thrown.