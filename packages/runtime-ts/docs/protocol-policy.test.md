---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/protocol/policy.test.ts
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/src/protocol/policy.test.ts
generated_at: 2026-01-31T09:14:17.040920
hash: 3f4533598237bae7d4dbe66cd3c23671496fb9bde5d1b088263469c91d0aeb2e
---

## Protocol Policy Documentation

This document details the functionality of the Protocol Policy component, responsible for enforcing governance and security constraints during runtime operations. It outlines the policy’s features, configuration, and behavior.

**Overview**

The Protocol Policy component provides mechanisms to control deployment timing and permitted plugins. It reads configuration from a `devops-config.json` file, falling back to a default, permissive policy if the file is absent.  The policy is designed to enhance operational stability and security.

**Key Features**

*   **Deployment Freeze Windows:**  The policy can prevent deployments during specified days and times, ensuring critical systems are not modified during peak usage or maintenance periods.
*   **Plugin Whitelisting:** The policy restricts the use of plugins to a predefined list, mitigating the risk of malicious or unauthorized extensions.
*   **Version Constraints:** Plugin entries within the whitelist can specify version constraints, allowing for controlled updates and compatibility.
*   **Configuration Loading:** The policy loads its configuration from a `devops-config.json` file located in the `config/` directory.

**Functionality**

**1. `checkFreeze(config)`**

This function determines if the current time falls within a defined freeze window specified in the configuration.

*   **Input:** `config` – The protocol configuration object.
*   **Behavior:**
    *   If `governance.enabled` is false, or `governance.freeze_windows` is not defined, the function does not throw an error.
    *   If the current time is within a freeze window, the function throws an error with the message “FROZEN”.
    *   Otherwise, the function completes without error.

**2. `validatePluginWhitelist(config, pluginName)`**

This function verifies if a given plugin is permitted based on the configured whitelist.

*   **Input:**
    *   `config` – The protocol configuration object.
    *   `pluginName` – The name of the plugin to validate.
*   **Behavior:**
    *   If no `plugin_whitelist` is defined in the configuration, the function returns `true` (plugin is allowed).
    *   If the `pluginName` is present in the `plugin_whitelist`, the function returns `true`.
    *   If the `pluginName` is present in the `plugin_whitelist` with a version constraint, the function returns `true`.
    *   Otherwise, the function returns `false` (plugin is not allowed).

**3. `getPluginVersionConstraint(config, pluginName)`**

This function retrieves the version constraint for a given plugin from the configured whitelist.

*   **Input:**
    *   `config` – The protocol configuration object.
    *   `pluginName` – The name of the plugin to query.
*   **Behavior:**
    *   If no `plugin_whitelist` is defined, the function returns `null`.
    *   If the `pluginName` is not found in the `plugin_whitelist`, the function returns `null`.
    *   If the `pluginName` is found with a version constraint (e.g., “@^6.0.0”), the function returns the constraint string.
    *   If the `pluginName` is found without a version constraint, the function returns `null`.

**4. `load()`**

This asynchronous function loads the protocol policy configuration.

*   **Behavior:**
    *   Attempts to read the `devops-config.json` file from the `config/` directory.
    *   If the file is not found, logs a warning and returns a default, unsafe policy (governance disabled).
    *   If the file is found, parses the JSON content.
    *   If the JSON is invalid, throws an error: “Invalid Governance Policy”.
    *   If the configuration schema is invalid, throws an error: “Invalid Governance Policy”.
    *   Handles potential errors during file reading and parsing, throwing “Invalid Governance Policy” in case of failure.

**Configuration**

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

The policy component handles errors related to configuration loading and validation.  Invalid JSON or schema errors result in an “Invalid Governance Policy” error being thrown. Missing configuration files result in a warning and the use of a default, permissive policy.