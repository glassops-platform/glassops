---
type: Documentation
domain: runtime
origin: packages/runtime/config/jest.integration.config.js
last_modified: 2026-01-26
generated: true
source: packages/runtime/config/jest.integration.config.js
generated_at: 2026-01-26T14:07:04.680Z
hash: 89181d826588c6b16971358a5521db0acc288f097f3e7c9b42ca6f89b8d10363
---

## Jest Integration Test Configuration

**Document Version:** 1.0
**Date:** October 26, 2023
**Author:** Principal Architect

**1. Introduction**

This document details the configuration for Jest integration tests within the runtime package. These tests verify the interaction between different components of the system, ensuring they function correctly when combined. This configuration is designed to provide comprehensive test coverage and reliable results.

**2. Core Configuration**

*   **Test Runner:** Jest is utilized as the test runner.
*   **TypeScript Support:**  `ts-jest` is employed to enable direct testing of TypeScript code without pre-compilation.
*   **Test Environment:** Tests are executed within a Node.js environment.
*   **Source Code Root:** The root directory for source code is set to `../src`.
*   **Test File Pattern:** Tests are identified by the filename pattern `**/*.integration.test.ts`.

**3. Test Execution & Performance**

*   **Timeout:**  A test timeout of 30 seconds (30000 milliseconds) is enforced for each integration test, accommodating potentially longer execution times inherent in integration testing.
*   **Module Transformation:**  Node modules are generally excluded from the transformation process to improve performance. Specific packages requiring transformation can be explicitly included if necessary.

**4. Code Coverage**

*   **Coverage Enabled:** Code coverage is enabled to measure the extent to which the codebase is exercised by the integration tests.
*   **Coverage Directory:** Coverage reports are generated and stored in the `../coverage/integration` directory.
*   **Coverage Reporters:**  Multiple coverage report formats are generated:
    *   **Text:**  A human-readable text summary in the console.
    *   **Lcov:**  A format suitable for integration with external coverage analysis tools.
    *   **HTML:**  An interactive HTML report for detailed code coverage visualization.
    *   **JSON Summary:** A concise JSON summary of coverage metrics.
*   **Coverage Scope:** Coverage is collected from all TypeScript files (`**/*.ts`) *except*:
    *   Test files (`**/*.test.ts`, `**/*.integration.test.ts`).
    *   Files within `node_modules`.
    *   The main entry point file (`index.ts`), which is covered by End-to-End (E2E) tests.
    *   Integration test helper files (`integration/test-helpers.ts`), which contain test utilities and are not part of the production code.

**5. Test Setup**

*   **Setup File:** The `jest.integration.setup.js` file (located in `../config`) is executed before each integration test to configure the test environment, mock dependencies, and perform any necessary initialization. This ensures a consistent and isolated testing environment.