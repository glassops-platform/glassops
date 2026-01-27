---
type: Documentation
domain: runtime
origin: packages/runtime/src/integration/test-helpers.ts
last_modified: 2026-01-26
generated: true
source: packages/runtime/src/integration/test-helpers.ts
generated_at: 2026-01-26T14:16:05.492Z
hash: 6132275461a546617ad1d176a073abb9da2b3e71a64dea6027475f4bdb734c54
---

## Integration Test Helpers

This document details a collection of utilities designed to support integration testing. These helpers streamline the process of setting up test environments, managing mock dependencies, and generating test data. They are intended for use by developers and testers working on the project.

**Core Functionality**

The provided tools address the following key areas:

*   **Test Workspace Management:**  Creation and cleanup of a dedicated test directory (`TEST_WORKSPACE`) and associated configuration file (`devops-config.json`).  The default configuration includes governance enabled and specific runtime settings for CLI and Node versions.
*   **Environment Mocking:** Generation of mock environment variables (`GITHUB_WORKSPACE`, `GITHUB_ACTOR`, etc.) to simulate a GitHub Actions environment.  These can be customized with overrides.
*   **Input Mocking:** Creation of mock GitHub Action inputs, including Salesforce-specific credentials and configuration options.  Customization via overrides is supported.
*   **Key Management:**  Creation of temporary JWT key files for testing authentication scenarios.
*   **CLI Mocking:**  Simulation of Salesforce CLI execution, allowing for control over success/failure scenarios and response data.  This includes handling JSON output and simulating standard output streams.
*   **Cache Mocking:**  Control over cache restoration behavior during tests.
*   **Command Availability Mocking:**  Mocking the `which` command to simulate the presence or absence of specific commands (e.g., `sf`).
*   **Test Data:**  Predefined test data sets for JWT keys (valid and invalid), test results (various scenarios), coverage data, freeze windows, plugin configurations, Salesforce environments, and repository formats.
*   **Assertions:**  Helper functions for verifying expected outcomes, such as successful execution, error logging, warning logging, and CLI command execution.  File existence checks are also included.
*   **Scenario Building:** A `TestScenarioBuilder` class provides a fluent interface for constructing complex test scenarios by combining environment variables, inputs, configuration, cache state, and command availability.

**Key Components**

*   **`TEST_WORKSPACE`:**  A constant defining the path to the test workspace directory.
*   **`DEFAULT_CONFIG`:**  An object containing default configuration settings for the test environment.
*   **`setupTestWorkspace()`:**  Creates the test workspace directory and configuration file.
*   **`cleanupTestWorkspace()`:**  Removes the test workspace directory and its contents.
*   **`createMockEnvironment()`:**  Generates a set of mock environment variables.
*   **`createMockInputs()`:**  Generates a set of mock GitHub Action inputs.
*   **`createTempJwtKey()`:**  Creates a temporary file containing a JWT key.
*   **`mockSuccessfulCliExecution()`:**  Mocks successful Salesforce CLI execution.
*   **`mockFailedCliExecution()`:**  Mocks failed Salesforce CLI execution.
*   **`mockCacheOperations()`:**  Mocks cache restoration behavior.
*   **`mockWhichOperations()`:**  Mocks the `which` command.
*   **`TestData`:**  An object containing pre-defined test data.
*   **`Assertions`:**  An object containing assertion helper functions.
*   **`TestScenarioBuilder`:** A class for building complex test scenarios.

**Usage**

You can use these helpers within your integration tests to create controlled and repeatable test environments. For example, you can use `setupTestWorkspace()` to prepare a test directory, `createMockEnvironment()` to simulate a GitHub Actions environment, and `mockSuccessfulCliExecution()` to control the behavior of the Salesforce CLI. The `TestScenarioBuilder` allows you to combine these elements into a cohesive test setup.