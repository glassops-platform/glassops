---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/integration/policy.integration.test.ts
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/src/integration/policy.integration.test.ts
generated_at: 2026-01-31T09:12:49.614202
hash: cfc958a5318263c5233c93e4cdba61064edb5b9626a53970db4367f2295e7737
---

## Policy Integration Test Documentation

This document details the integration tests for the ProtocolPolicy class, verifying its functionality with file system operations and configuration loading. These tests ensure the policy enforcement mechanisms operate as expected.

**Scope:**

The tests cover the following core areas:

*   Configuration Loading: Validating the loading and merging of policies from configuration files, handling missing files, and detecting invalid JSON or schema.
*   Plugin Whitelist Validation: Confirming that plugin validation against a configured whitelist functions correctly, including version constraint extraction.
*   Freeze Window Validation: Ensuring deployment blocking during defined freeze windows and allowing deployments outside of these periods.

**Test Environment:**

The tests create a temporary workspace directory (`test-workspace`) to isolate the tests from the existing environment. The `GITHUB_WORKSPACE` environment variable is temporarily modified to point to this directory during each test.  This directory is cleaned up after each test suite and after all tests complete.

**Configuration:**

The tests read a configuration file named `devops-config.json` from the test workspace. This file defines governance and runtime settings, including:

*   `governance.enabled`:  A boolean indicating whether governance policies are active.
*   `governance.freeze_windows`: An array of objects defining time periods where deployments are blocked. Each object includes a `day` (e.g., "Saturday"), `start` time (e.g., "00:00"), and `end` time (e.g., "23:59").
*   `governance.plugin_whitelist`: An array of strings representing allowed plugins and their version constraints.
*   `runtime.cli_version`: The expected CLI version.
*   `runtime.node_version`: The expected Node.js version.

**Key Functionality Verified:**

*   **Policy Loading:** The `ProtocolPolicy` class correctly loads and merges configuration data from a JSON file. Default values are used when the configuration file is missing. Errors are thrown for invalid JSON or schema.
*   **Plugin Validation:** The `validatePluginWhitelist` method accurately checks if a given plugin is present in the configured whitelist. The `getPluginVersionConstraint` method correctly extracts version constraints from the whitelist entries.  If no whitelist is configured, all plugins are permitted.
*   **Freeze Window Enforcement:** The `checkFreeze` method correctly identifies whether the current time falls within a defined freeze window, throwing an error ("FROZEN") if it does. Deployments are allowed outside of freeze windows.

**RTM Coverage:**

These tests provide coverage for the following requirements:

*   BR-001: Additive Policy Merge
*   BR-007: Pass/Fail Adjudication - Freeze Windows
*   BR-020: Plugin Whitelist Management

**Usage:**

You can run these tests using a standard testing framework like Jest. Ensure the necessary dependencies are installed and the test environment is properly configured.