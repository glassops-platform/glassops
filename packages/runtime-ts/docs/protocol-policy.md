---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/protocol/policy.ts
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/src/protocol/policy.ts
generated_at: 2026-01-31T10:11:07.724962
hash: 010bcd70ed0165b80fec5d3c7beef6127deacdd2ca8158e77fc56047f2ebd024
---

## Protocol Policy Document

This document details the Protocol Policy component, responsible for governing runtime behavior and enforcing operational constraints. It provides a mechanism for configuring and validating Salesforce development and deployment processes.

### Overview

The Protocol Policy component manages configuration settings related to governance and runtime environments. It allows administrators to define rules for plugin usage, schedule deployment freezes, and control analyzer behavior.  I provide a centralized location for defining and enforcing these policies.

### Configuration

The policy is driven by a JSON configuration file, `devops-config.json`, located by default in the `config` directory within the workspace (or as specified by the `GLASSOPS_CONFIG_PATH` environment variable). The configuration adheres to the following schema:

```json
{
  "governance": {
    "enabled": true,
    "freeze_windows": [
      {
        "day": "Monday",
        "start": "09:00",
        "end": "17:00"
      }
    ],
    "plugin_whitelist": [
      "sfdx-hardis@^4.0.0",
      "@salesforce/plugin-deploy-retrieve"
    ],
    "analyzer": {
      "enabled": false,
      "severity_threshold": 1,
      "rulesets": [
        "namespace1/rule1",
        "namespace2/rule2"
      ],
      "opinionated": true
    }
  },
  "runtime": {
    "cli_version": "latest",
    "node_version": "20"
  }
}
```

**Configuration Details:**

*   **governance.enabled**:  A boolean indicating whether governance features are active. Defaults to `true`.
*   **governance.freeze_windows**: An optional array of objects defining time windows during which deployments are blocked. Each object requires a `day` (Monday-Sunday), `start` (HH:MM), and `end` (HH:MM) time.
*   **governance.plugin_whitelist**: An optional array of strings representing allowed Salesforce CLI plugins. Version constraints can be included (e.g., `sfdx-hardis@^4.0.0`).
*   **governance.analyzer.enabled**: A boolean indicating whether the analyzer is enabled. Defaults to `false`.
*   **governance.analyzer.severity_threshold**: An integer (1-3) defining the minimum severity level for analyzer findings to be reported. Defaults to `1`.
*   **governance.analyzer.rulesets**: An optional array of strings specifying the rulesets to be used by the analyzer.
*   **governance.analyzer.opinionated**: A boolean indicating whether to prioritize the `sf code-analyzer` over the `sf scanner`. Defaults to `true`.
*   **runtime.cli_version**: The desired Salesforce CLI version. Defaults to `latest`.
*   **runtime.node_version**: The desired Node.js version. Defaults to `20`.

### Core Functionality

*   **Loading Configuration:** The `load()` method reads the configuration file and validates it against the defined schema. If the file is missing, a default, less restrictive policy is applied. Errors during file reading or schema validation result in exceptions.
*   **Freeze Window Check:** The `checkFreeze()` method evaluates whether the current time falls within a defined freeze window. If a match is found, an error is thrown, blocking further execution.  This method uses UTC for consistent behavior.
*   **Plugin Whitelist Validation:** The `validatePluginWhitelist()` method determines if a given plugin is allowed based on the configured whitelist. If no whitelist is defined, all plugins are permitted.
*   **Version Constraint Retrieval:** The `getPluginVersionConstraint()` method retrieves the version constraint associated with a specific plugin from the whitelist, if available.

### Usage

You can instantiate the `ProtocolPolicy` class and use its methods to enforce policies within your workflows. For example:

```typescript
import { ProtocolPolicy } from './policy';

const policy = new ProtocolPolicy();

try {
  const config = await policy.load();
  policy.checkFreeze(config);

  if (!policy.validatePluginWhitelist(config, 'sfdx-cli')) {
    throw new Error('Plugin not allowed');
  }

  const versionConstraint = policy.getPluginVersionConstraint(config, 'sfdx-hardis');
  console.log(`Version constraint for sfdx-hardis: ${versionConstraint}`);

} catch (error) {
  core.setFailed(error.message);
}
```

### Error Handling

The component includes robust error handling. Invalid configuration files or violations of governance rules will result in descriptive error messages, facilitating troubleshooting and policy correction.