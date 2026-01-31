---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/integration/test-helpers.ts
last_modified: 2026-01-29
generated: true
source: packages/runtime-ts/src/integration/test-helpers.ts
generated_at: 2026-01-29T20:56:21.166554
hash: af660f4ebe5e84e895d5f3226a2423c92c2523cdf16760c61bfe97bb4ee28393
---

## Integration Test Helpers

This document details a collection of utilities designed to facilitate robust and repeatable integration testing. These helpers provide functionality for environment setup, data creation, and mock object management, streamlining the testing process.

**Purpose**

The primary goal of these tools is to provide a consistent and isolated environment for integration tests, reducing dependencies on external systems and ensuring predictable results.

**Key Components**

*   **Test Workspace:** A dedicated directory (`test-workspace`) used for storing test-related files, including a configuration file (`devops-config.json`). Functions are provided to create and clean up this workspace.
*   **Configuration:** A default configuration object (`DEFAULT_CONFIG`) is available, and can be overridden to customize test scenarios.
*   **Mocking:**  Functions are included to mock environment variables, GitHub Action inputs, Salesforce CLI execution (success and failure), cache operations, and file system commands.
*   **Test Data:** A `TestData` object contains pre-defined data sets for common scenarios, such as valid and invalid JWT keys, test results, coverage data, and plugin configurations.
*   **Assertions:**  An `Assertions` object offers helper functions for verifying expected outcomes, including checking output values, error messages, and CLI command execution.
*   **Scenario Builder:** A `TestScenarioBuilder` class allows for the creation of complex test scenarios by combining different configurations, inputs, and mock behaviors.

**Functionality Overview**

*   **Workspace Management:**
    *   `setupTestWorkspace()`: Creates the test workspace directory and initializes the configuration file.
    *   `cleanupTestWorkspace()`: Removes the test workspace directory and its contents.
*   **Mock Data Creation:**
    *   `createMockEnvironment()`: Generates a set of mock environment variables.
    *   `createMockInputs()`: Generates mock GitHub Action inputs.
    *   `createTempJwtKey()`: Creates a temporary file containing a JWT key.
*   **CLI Mocking:**
    *   `mockSuccessfulCliExecution()`: Simulates successful Salesforce CLI command execution.
    *   `mockFailedCliExecution()`: Simulates failed Salesforce CLI command execution.
*   **System Mocking:**
    *   `mockCacheOperations()`: Mocks cache restore operations.
    *   `mockWhichOperations()`: Mocks the `which` command, controlling the availability of commands.
*   **Scenario Composition:**
    *   `TestScenarioBuilder`:  A fluent interface for building comprehensive test scenarios with configurable environments, inputs, configurations, and mock behaviors.

**Usage**

You can leverage these helpers within your integration tests to:

1.  Set up a controlled test environment.
2.  Inject mock data to isolate dependencies.
3.  Simulate various outcomes of external commands.
4.  Verify expected behavior using assertion helpers.
5.  Construct complex test scenarios using the `TestScenarioBuilder`.

These tools are designed to improve the reliability, maintainability, and efficiency of your integration testing efforts.