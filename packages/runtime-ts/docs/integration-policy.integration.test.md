---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/integration/policy.integration.test.ts
last_modified: 2026-01-29
generated: true
source: packages/runtime-ts/src/integration/policy.integration.test.ts
generated_at: 2026-01-29T20:56:06.473351
hash: cfc958a5318263c5233c93e4cdba61064edb5b9626a53970db4367f2295e7737
---

## Policy Integration Test Documentation

This document details the integration tests for the ProtocolPolicy class, verifying its functionality with file system operations and configuration loading. These tests ensure the policy enforcement mechanisms operate as expected.

**Scope:**

The tests cover the following core areas:

*   Configuration Loading: Validating the loading and merging of policies from configuration files, handling missing files, and detecting invalid JSON or schema.
*   Plugin Whitelist Validation: Confirming that plugin validation against a configured whitelist functions correctly, including version constraint extraction and handling of scoped packages.
*   Freeze Window Validation: Verifying the enforcement of deployment freezes during specified time windows.

**Test Environment:**

The tests create a temporary workspace directory (`test-workspace`) to isolate the tests from the existing environment. The `GITHUB_WORKSPACE` environment variable is temporarily modified to point to this directory during each test.  The workspace is cleaned up before and after each test suite.

**Configuration:**

The tests read configuration from a JSON file (`devops-config.json`) within the test workspace. This file defines governance settings, including:

*   `governance.enabled`:  A boolean indicating whether governance policies are active.
*   `governance.freeze_windows`: An array of objects defining time windows during which deployments are blocked. Each object includes a `day` (e.g., "Saturday"), `start` time (e.g., "00:00"), and `end` time (e.g., "23:59").
*   `governance.plugin_whitelist`: An array of strings representing allowed plugins and their version constraints.
*   `runtime.cli_version`: The expected CLI version.
*   `runtime.node_version`: The expected Node.js version.

**Key Functionality Tested:**

*   **Loading and Merging Policies:** The system correctly loads and merges configuration data from the `devops-config.json` file. Default values are applied when the file is missing.
*   **Plugin Validation:** The `validatePluginWhitelist` function accurately determines whether a given plugin is permitted based on the configured whitelist. Version constraints are correctly parsed and applied.
*   **Freeze Window Enforcement:** The `checkFreeze` function correctly identifies whether the current time falls within a configured freeze window, throwing an error if a freeze is active.
*   **Error Handling:** The system gracefully handles invalid JSON and schema errors in the configuration file, providing informative error messages.

**Usage:**

You can run these tests using a standard testing framework like Jest. No specific user interaction is required beyond executing the test suite.

**Relevant Requirements:**

These tests provide coverage for the following requirements:

*   BR-001: Additive Policy Merge
*   BR-007: Pass/Fail Adjudication - Freeze Windows
*   BR-020: Plugin Whitelist Management