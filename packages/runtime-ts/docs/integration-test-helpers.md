---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/integration/test-helpers.ts
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/src/integration/test-helpers.ts
generated_at: 2026-01-31T09:13:04.924696
hash: af660f4ebe5e84e895d5f3226a2423c92c2523cdf16760c61bfe97bb4ee28393
---

## Integration Test Helpers

This document details a collection of utilities designed to facilitate robust and repeatable integration testing. These helpers provide functionality for environment setup, data creation, and mock object management, streamlining the testing process.

**Purpose**

The primary goal of these helpers is to provide a consistent and isolated environment for integration tests, reducing dependencies on external systems and ensuring predictable results.

**Key Components**

*   **Test Workspace:** A dedicated directory (`test-workspace`) used for storing test-related files, including a configuration file (`devops-config.json`). Functions are provided to create and clean up this workspace.
*   **Configuration:** A default configuration object (`DEFAULT_CONFIG`) is available, which can be overridden to customize test scenarios.
*   **Mocking:**  Functions are included to mock environment variables, GitHub Action inputs, Salesforce CLI execution, cache operations, and file system interactions. This allows for controlled testing of specific code paths without relying on actual external services.
*   **Test Data:** A `TestData` object provides pre-defined data sets for common test scenarios, such as valid and invalid JWT keys, test results, coverage data, and plugin configurations.
*   **Assertions:**  An `Assertions` object offers helper functions for verifying expected outcomes, including checking output values, error messages, and CLI command execution.
*   **Scenario Builder:** A `TestScenarioBuilder` class allows for the creation of complex test scenarios by combining different environment settings, inputs, and configurations.

**Functions & Objects**

*   `setupTestWorkspace(config)`: Creates the test workspace directory and initializes the configuration file. Accepts an optional configuration object.
*   `cleanupTestWorkspace()`: Removes the test workspace directory and its contents.
*   `createMockEnvironment(overrides)`: Generates a set of mock environment variables for testing, allowing overrides for specific variables.
*   `createMockInputs(overrides)`: Creates mock GitHub Action inputs, with the ability to override default values.
*   `createTempJwtKey(content)`: Generates a temporary file containing a JWT key.
*   `mockSuccessfulCliExecution(mockExec, response)`: Mocks successful execution of the Salesforce CLI.
*   `mockFailedCliExecution(mockExec, errorMessage)`: Mocks failed execution of the Salesforce CLI.
*   `mockCacheOperations(mockRestoreCache, shouldRestore)`: Mocks cache restore operations.
*   `mockWhichOperations(mockWhich, availableCommands)`: Mocks the `which` command, controlling which commands are reported as available.
*   `TestData`: An object containing pre-defined test data.
*   `Assertions`: An object containing assertion helper functions.
*   `TestScenarioBuilder`: A class for building complex test scenarios.

**Usage**

You can use these helpers within your integration tests to:

1.  Set up a controlled test environment.
2.  Provide mock data and dependencies.
3.  Execute the code under test.
4.  Assert the expected outcomes.

The `TestScenarioBuilder` is particularly useful for creating complex test setups with multiple configurations and dependencies.