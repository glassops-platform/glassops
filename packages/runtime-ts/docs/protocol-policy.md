---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/protocol/policy.ts
last_modified: 2026-01-29
generated: true
source: packages/runtime-ts/src/protocol/policy.ts
generated_at: 2026-01-29T20:57:47.249992
hash: 010bcd70ed0165b80fec5d3c7beef6127deacdd2ca8158e77fc56047f2ebd024
---

## Protocol Policy Document

This document details the Protocol Policy component, responsible for governing runtime behavior and enforcing operational constraints. It provides a centralized mechanism for managing Salesforce development and deployment processes.

**Overview**

The Protocol Policy component manages configuration settings related to governance and runtime environments. It allows administrators to define rules for plugin usage, schedule deployment freezes, and control the versions of required tools.  This component is designed to enhance security, maintain stability, and enforce best practices within the development lifecycle.

**Configuration**

The system’s behavior is driven by a configuration file, `devops-config.json`, located by default in the `config` directory of the workspace (or specified by the `GLASSOPS_CONFIG_PATH` environment variable). The configuration adheres to a defined schema, ensuring data integrity and predictability.

The configuration is structured into two primary sections:

*   **Governance:** Controls operational aspects such as deployment freezes and plugin whitelisting.
    *   `enabled`: A boolean flag to globally enable or disable governance features. Defaults to `true`.
    *   `freeze_windows`: Defines specific time windows during which deployments are blocked. Each window includes a `day` of the week and a `start` and `end` time (HH:MM format).
    *   `plugin_whitelist`:  An optional list of allowed Salesforce CLI plugins. Version constraints can be included (e.g., `sfdx-hardis@^4.0.0`). If not specified, all plugins are permitted.
    *   `analyzer`: Configures the static code analyzer.
        *   `enabled`: Enables or disables the analyzer. Defaults to `false`.
        *   `severity_threshold`: Sets the minimum severity level for analyzer findings (1-3). Defaults to `1`.
        *   `rulesets`: An optional array of rulesets to apply.
        *   `opinionated`: Determines whether to prioritize the `sf code-analyzer` over the `sf scanner`. Defaults to `true`.
*   **Runtime:** Specifies runtime environment settings.
    *   `cli_version`: The desired Salesforce CLI version. Defaults to `latest`.
    *   `node_version`: The required Node.js version. Defaults to `20`.

**Functionality**

The Protocol Policy component provides the following key functions:

*   **Configuration Loading:**  Loads the configuration from the `devops-config.json` file. If the file is missing, a default, less restrictive policy is applied, and a warning is issued.  Errors during file parsing result in an exception.
*   **Deployment Freeze Check:**  Evaluates whether the current time falls within a defined freeze window. If a match is found, a deployment error is thrown, blocking the operation.  Time comparisons are performed using UTC to ensure consistency.
*   **Plugin Whitelist Validation:**  Determines whether a given plugin is permitted based on the configured whitelist. If no whitelist is defined, all plugins are allowed.
*   **Plugin Version Constraint Retrieval:** Extracts the version constraint for a specified plugin from the whitelist, if available.

**Usage**

You interact with this component through its methods after instantiating the `ProtocolPolicy` class.  

1.  Instantiate the `ProtocolPolicy` class.
2.  Call the `load()` method to retrieve the configuration.
3.  Use the loaded configuration with other methods, such as `checkFreeze()` or `validatePluginWhitelist()`, to enforce policies.

**Error Handling**

The component includes robust error handling:

*   Invalid configuration files will result in an error being thrown, providing details about the parsing failure.
*   Deployment freeze violations will trigger an error, clearly indicating the blocked time window.
*   Missing configuration files are handled gracefully with a warning and the application of a default policy.