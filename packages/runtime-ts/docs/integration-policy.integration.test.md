---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/integration/policy.integration.test.ts
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/src/integration/policy.integration.test.ts
generated_at: 2026-01-31T10:09:07.881210
hash: cfc958a5318263c5233c93e4cdba61064edb5b9626a53970db4367f2295e7737
---

## Policy Integration Test Documentation

This document details the integration tests for the ProtocolPolicy class, verifying its functionality with file system operations and configuration loading. These tests ensure the policy enforcement mechanisms operate as expected.

### Overview

The integration tests focus on three key areas: configuration loading, plugin whitelist validation, and freeze window validation.  The tests utilize a temporary workspace to simulate a real-world environment, loading configuration from a JSON file and exercising the policy’s validation routines. Mock implementations of GitHub Actions modules are used to isolate the policy logic.

### Configuration Loading

These tests verify the correct loading and parsing of the governance policy configuration.

*   **Successful Load:**  The policy successfully loads a valid configuration file, correctly interpreting settings for governance and runtime parameters.
*   **Missing File:** When the configuration file is absent, the policy defaults to pre-defined values.
*   **Invalid JSON:**  An error is thrown if the configuration file contains invalid JSON syntax.
*   **Invalid Schema:** An error is thrown if the configuration file contains valid JSON but does not conform to the expected schema (e.g., incorrect data types).

### Plugin Whitelist Validation

These tests confirm the policy’s ability to validate plugins against a configured whitelist.

*   **Whitelist Validation:** The policy correctly identifies whitelisted and non-whitelisted plugins based on the loaded configuration.
*   **Version Constraints:** The policy extracts and applies version constraints specified in the whitelist. If no constraint is provided, it is handled appropriately.
*   **No Whitelist:** If no whitelist is configured, all plugins are considered valid.
*   **Scoped Packages:** The policy correctly handles scoped packages, including those without explicit version constraints.

### Freeze Window Validation

These tests ensure the policy correctly enforces deployment restrictions during defined freeze windows.

*   **Freeze Enforcement:** The policy throws an error when a deployment is attempted during a configured freeze window, based on the current day and time.
*   **Outside Freeze Window:** Deployments are permitted outside of configured freeze windows.
*   **No Freeze Windows:** If no freeze windows are defined, deployments are always allowed.

### Test Environment

The tests create a temporary workspace directory to isolate the configuration and prevent interference with the existing environment. This directory is cleaned up after each test suite and at the end of all tests. The `GITHUB_WORKSPACE` environment variable is temporarily modified to point to this test directory during test execution and restored afterward.