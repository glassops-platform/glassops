---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/protocol/policy.ts
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/src/protocol/policy.ts
generated_at: 2026-01-31T09:14:34.716270
hash: 010bcd70ed0165b80fec5d3c7beef6127deacdd2ca8158e77fc56047f2ebd024
---

## Protocol Policy Document

This document details the Protocol Policy component, responsible for governing runtime behavior and enforcing operational constraints. It provides a centralized mechanism for managing Salesforce development and deployment processes.

**Overview**

The Protocol Policy component manages configuration settings related to governance and runtime environments. It allows administrators to define rules for plugin usage, schedule deployment freezes, and control analyzer behavior.  The policy is loaded from a configuration file, defaulting to an unsafe policy if the file is not found.

**Configuration**

The system’s behavior is determined by a JSON configuration file, validated against a defined schema. Key configuration sections include:

*   **Governance:** Controls operational restrictions.
    *   `enabled`: A boolean flag to enable or disable governance features. Defaults to `true`.
    *   `freeze_windows`: Defines time periods during which deployments are blocked. Each window specifies a `day` of the week and a `start` and `end` time (HH:MM format).
    *   `plugin_whitelist`:  An optional list of allowed Salesforce CLI plugins, potentially including version constraints (e.g., `sfdx-hardis@^4.0.0`). If empty, all plugins are permitted.
    *   `analyzer`: Configures the code analyzer.
        *   `enabled`: Enables or disables the analyzer. Defaults to `false`.
        *   `severity_threshold`: Sets the minimum severity level for analyzer findings (1-3). Defaults to `1`.
        *   `rulesets`: An optional array of rulesets to apply.
        *   `opinionated`:  Determines whether the `sf code-analyzer` is preferred over `sf scanner`. Defaults to `true`.
*   **Runtime:** Specifies runtime environment settings.
    *   `cli_version`: The desired Salesforce CLI version. Defaults to `latest`.
    *   `node_version`: The Node.js version to use. Defaults to `20`.

**Functionality**

The Protocol Policy component provides the following core functions:

*   **Configuration Loading:**  Loads the configuration from a JSON file specified by the `GLASSOPS_CONFIG_PATH` environment variable, or defaults to `config/devops-config.json`.  If the file is missing, a default, less restrictive policy is applied.  Invalid configuration files will result in an error.
*   **Deployment Freeze Check:**  Evaluates whether the current time falls within a defined freeze window. If a match is found, deployment is blocked with an error message.  This function uses UTC time for consistency.
*   **Plugin Whitelist Validation:**  Determines whether a given plugin is permitted based on the configured whitelist. If no whitelist is defined, all plugins are allowed.
*   **Plugin Version Constraint Retrieval:**  If a plugin is whitelisted with a version constraint, this function retrieves that constraint.

**Usage**

You can interact with the Protocol Policy component through its class methods.  After instantiating the `ProtocolPolicy` class, you should call the `load()` method to retrieve the configuration.  Then, you can use the `checkFreeze()`, `validatePluginWhitelist()`, and `getPluginVersionConstraint()` methods to enforce governance rules.

**Error Handling**

The component includes error handling for invalid configuration files and deployment freeze violations.  Warnings are issued if the configuration file is not found.  Detailed error messages are provided to aid in troubleshooting.