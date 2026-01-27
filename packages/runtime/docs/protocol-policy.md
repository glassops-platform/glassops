---
type: Documentation
domain: runtime
origin: packages/runtime/src/protocol/policy.ts
last_modified: 2026-01-26
generated: true
source: packages/runtime/src/protocol/policy.ts
generated_at: 2026-01-26T14:19:12.635Z
hash: 5170423c5806e2247635e0617b74790d79c84fe7b1a86cd8d328fe8f660ff03e
---

## Protocol Policy Document

**1. Introduction**

This document details the Protocol Policy, a component responsible for governing runtime behavior and enforcing operational constraints. It manages configurations related to governance, including deployment freezes and plugin whitelisting, as well as runtime settings like CLI and Node.js versions.

**2. Configuration**

The policy is driven by a configuration file (`devops-config.json`) located in the root of the GitHub workspace (or the current directory if not running within a workspace).  If this file is absent, a default, less restrictive policy is applied. The configuration is validated against a defined schema.

**2.1. Governance Settings**

*   **Enabled:** A boolean flag to globally enable or disable governance features. Defaults to `true`.
*   **Freeze Windows:** Defines specific time windows during which deployments are blocked. Each window includes a `day` (Monday-Sunday), `start` time (HH:MM), and `end` time (HH:MM).  Times are evaluated using UTC.
*   **Plugin Whitelist:** A list of allowed Salesforce CLI plugins. Entries can optionally include version constraints (e.g., `sfdx-hardis@^4.0.0`). If no whitelist is provided, all plugins are permitted.
*   **Analyzer:** Configures the Salesforce code analyzer.
    *   **Enabled:** Enables or disables the analyzer. Defaults to `false`.
    *   **Severity Threshold:** Sets the minimum severity level for analyzer findings (1-3). Defaults to `1`.
    *   **Rulesets:** An optional array of rulesets to apply.
    *   **Opinionated:**  Determines whether the `sf code-analyzer` is preferred over `sf scanner`. Defaults to `true`.

**2.2. Runtime Settings**

*   **CLI Version:** Specifies the desired Salesforce CLI version. Defaults to `latest`.
*   **Node Version:** Specifies the desired Node.js version. Defaults to `20`.

**3. Core Functionality**

*   **Configuration Loading:** The `ProtocolPolicy` class loads the configuration from `devops-config.json`, validating it against the defined schema.  Errors during loading result in an exception.
*   **Deployment Freeze Check:** The `checkFreeze` method evaluates the current time against configured freeze windows. If the current time falls within a freeze window, a descriptive error is thrown, blocking the deployment.
*   **Plugin Validation:**
    *   `validatePluginWhitelist`: Checks if a given plugin is allowed based on the configured whitelist. If no whitelist is defined, all plugins are allowed.
    *   `getPluginVersionConstraint`: Retrieves the version constraint for a given plugin from the whitelist, if specified.
*   **Plugin Name and Version Extraction:**  Internal methods (`extractPluginName`, `extractVersionConstraint`) parse whitelist entries to determine the plugin name and associated version constraint. These methods handle both scoped packages (e.g., `@scope/package@version`) and regular packages (e.g., `package@version`).

**4. Usage**

You can instantiate the `ProtocolPolicy` class to access its functionality.  

```typescript
const policy = new ProtocolPolicy();
const config = await policy.load();
policy.checkFreeze(config); // Check for deployment freezes
const isPluginAllowed = policy.validatePluginWhitelist(config, "sfdx-cli"); // Validate a plugin
const versionConstraint = policy.getPluginVersionConstraint(config, "sfdx-cli"); // Get version constraint
```

**5. Error Handling**

The `ProtocolPolicy` includes error handling for invalid configuration files and deployment freeze violations.  Informative error messages are provided to aid in troubleshooting.  A warning is logged if the configuration file is not found, and a default policy is applied.