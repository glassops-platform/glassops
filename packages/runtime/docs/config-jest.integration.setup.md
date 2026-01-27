---
type: Documentation
domain: runtime
origin: packages/runtime/config/jest.integration.setup.js
last_modified: 2026-01-26
generated: true
source: packages/runtime/config/jest.integration.setup.js
generated_at: 2026-01-26T14:07:22.445Z
hash: 28dc82c0c48fdebd11b234e6b847821b7d8d04b2aa13580313447ad7370cec5b
---

## Integration Test Environment Configuration

**Document Version:** 1.0
**Date:** October 26, 2023
**Author:** Principal Architect

**1. Introduction**

This document details the configuration applied to the integration test environment for the runtime package.  It outlines the environment variables established and modifications made to the Jest testing framework to facilitate reliable and reproducible integration tests.

**2. Purpose**

The primary goal of this configuration is to simulate a typical GitHub Actions environment for integration tests, even when running locally. This ensures tests accurately reflect real-world execution conditions.  Additionally, it adjusts testing parameters to accommodate the longer execution times often associated with integration tests.

**3. Environment Variable Setup**

The following environment variables are set to default values if they are not already defined in the execution environment. These variables mimic those provided by the GitHub Actions platform:

*   **`GITHUB_WORKSPACE`**:  Set to the current working directory (`process.cwd()`) if not already defined. Represents the project's workspace.
*   **`GITHUB_ACTOR`**: Set to "integration-test" if not already defined. Represents the user or application triggering the workflow.
*   **`GITHUB_REPOSITORY`**: Set to "test/integration" if not already defined. Represents the repository where the workflow is running.
*   **`GITHUB_RUN_ID`**: Set to "12345" if not already defined.  A unique identifier for the current workflow run.
*   **`GITHUB_SHA`**: Set to "abc123def456" if not already defined.  The commit SHA triggering the workflow.
*   **`GITHUB_EVENT_NAME`**: Set to "push" if not already defined.  The event that triggered the workflow (e.g., push, pull_request).

**4. Jest Configuration**

*   **Timeout Increase:** The default Jest timeout is increased to 30,000 milliseconds (30 seconds) to accommodate the potentially longer execution times of integration tests.  This prevents tests from failing prematurely due to timeout issues.
*   **Console Output Control:**  A helper is provided to optionally suppress console output during test execution.  Currently, the default console logging functions are preserved.  Commented-out code demonstrates how to replace them with Jest mock functions to silence console output if desired. This is useful for focusing on test results without extraneous logging.



**5. Dependencies**

*   Jest testing framework.
*   Node.js runtime environment.



**6. Future Considerations**

*   Explore more granular control over environment variable overrides via a configuration file.
*   Implement a mechanism to dynamically adjust the timeout based on test suite size or complexity.