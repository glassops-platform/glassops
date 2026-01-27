---
type: Documentation
domain: runtime
origin: packages/runtime/src/integration/policy.integration.test.ts
last_modified: 2026-01-26
generated: true
source: packages/runtime/src/integration/policy.integration.test.ts
generated_at: 2026-01-26T14:15:45.941Z
hash: 5c1285434f3a01499ed5367df43d69588dc3342c10cf418543eff87ef608a892
---

## Policy Integration Test Documentation

This document details the integration tests for the ProtocolPolicy class, verifying its functionality with file system operations and configuration loading. These tests ensure the policy enforcement mechanisms operate as expected.

**Scope:**

The tests cover the following core areas:

*   **Configuration Loading:**  Loading and merging policy configurations from JSON files, handling missing files, and validating JSON and schema integrity.
*   **Plugin Whitelist Validation:** Validating installed plugins against a configured whitelist, extracting version constraints, and handling cases with and without whitelists.
*   **Freeze Window Validation:**  Determining whether a deployment should be blocked based on configured freeze windows (time-based restrictions).

**Test Environment:**

The tests create a temporary workspace directory (`test-workspace`) to isolate the tests from the existing environment.  The `GITHUB_WORKSPACE` environment variable is temporarily modified to point to this directory during each test.  This directory is cleaned up after each test suite and after all tests complete.

**Configuration:**

The tests read a configuration file named `devops-config.json` from the test workspace. This file defines governance and runtime settings, including:

*   `governance.enabled`:  A boolean indicating whether governance policies are active.
*   `governance.freeze_windows`: An array of objects defining time windows during which deployments are blocked. Each object includes a `day` (e.g., "Saturday"), `start` time (e.g., "00:00"), and `end` time (e.g., "23:59").
*   `governance.plugin_whitelist`: An array of strings representing allowed plugins and their version constraints (e.g., "sfdx-hardis@^4.0.0").
*   `runtime.cli_version`: The expected CLI version.
*   `runtime.node_version`: The expected Node.js version.

**Key Functionality Verified:**

*   **Policy Loading:** The `ProtocolPolicy` class correctly loads and merges configuration data from a JSON file.  It provides default values when the configuration file is missing. Errors are thrown for invalid JSON or schema.
*   **Plugin Validation:** The `validatePluginWhitelist` method checks if a given plugin is present in the configured whitelist. The `getPluginVersionConstraint` method retrieves the version constraint for a plugin, if specified.  If no whitelist is configured, all plugins are considered valid.
*   **Freeze Window Enforcement:** The `checkFreeze` method determines if the current time falls within a configured freeze window. If so, it throws an error, blocking the deployment.

**Error Handling:**

The tests verify that appropriate errors are thrown in the following scenarios:

*   Invalid JSON format in the configuration file.
*   Configuration schema validation failures.
*   Attempting a deployment during a configured freeze window.

**Dependencies:**

*   Node.js
*   `fs` module (file system operations)
*   `path` module (path manipulation)
*   Jest testing framework
*   `@actions/core` (mocked for testing purposes)