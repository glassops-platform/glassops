---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/integration/test-helpers.ts
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/src/integration/test-helpers.ts
generated_at: 2026-01-31T10:09:30.630539
hash: af660f4ebe5e84e895d5f3226a2423c92c2523cdf16760c61bfe97bb4ee28393
---

## Integration Test Helpers

This document details a collection of utilities designed to facilitate robust and repeatable integration testing. These helpers provide functionality for environment setup, data creation, and mock object management, streamlining the testing process.

### Overview

The purpose of this module is to offer reusable components for integration tests, reducing redundancy and improving test maintainability. It addresses common needs such as workspace creation, configuration management, and the simulation of external dependencies.

### Core Components

**1. Test Workspace Management:**

*   `TEST_WORKSPACE`: A constant defining the path to a dedicated test workspace directory. This directory is used for creating test-specific files and configurations.
*   `setupTestWorkspace(config)`: Creates the test workspace directory if it doesn’t exist and writes a default or provided configuration file (`devops-config.json`) to it.  The `config` parameter accepts a configuration object; if omitted, a default configuration is used.
*   `cleanupTestWorkspace()`: Removes the test workspace directory and its contents.

**2. Mock Data Creation:**

*   `createMockEnvironment(overrides)`: Generates a set of environment variables commonly used in GitHub Actions environments, with optional overrides provided via the `overrides` parameter.
*   `createMockInputs(overrides)`: Creates mock GitHub Action inputs, allowing for controlled testing of input-driven behavior.  The `overrides` parameter allows customization of specific input values.
*   `createTempJwtKey(content)`: Generates a temporary file containing a JWT key. The `content` parameter allows specifying the key's content; a default key is provided if none is specified.

**3. Mocking External Dependencies:**

*   `mockSuccessfulCliExecution(mockExec, response)`: Mocks the successful execution of the Salesforce CLI (`sf`).  It simulates the CLI’s output, allowing tests to verify behavior based on expected results. The `response` parameter allows specifying custom JSON output.
*   `mockFailedCliExecution(mockExec, errorMessage)`: Mocks a failed Salesforce CLI execution, raising an error with a specified message.
*   `mockCacheOperations(mockRestoreCache, shouldRestore)`: Mocks cache restoration operations, controlling whether a cache key is returned.
*   `mockWhichOperations(mockWhich, availableCommands)`: Mocks the `which` command, allowing control over the availability of commands in the simulated environment.

**4. Test Data Factories:**

*   `TestData`: An object containing pre-defined test data for common scenarios, including:
    *   `validJwtKey`, `invalidJwtKey`: Example JWT keys.
    *   `testResults`:  Objects representing various test result scenarios (valid, empty, all passed, all failed).
    *   `coverageData`: Objects representing different code coverage scenarios.
    *   `freezeWindows`:  Arrays defining time windows for freezing operations.
    *   `pluginConfigs`:  Configurations for Salesforce CLI plugins.
    *   `salesforceEnvironments`:  URLs for different Salesforce environments.
    *   `repositoryFormats`: Valid and invalid repository formats.

**5. Assertion Helpers:**

*   `Assertions`: An object providing helper functions for common test assertions:
    *   `expectSuccessfulExecution()`: Verifies that output functions were called with expected values after a successful execution.
    *   `expectErrorLogged()`: Asserts that an error message was logged.
    *   `expectWarningLogged()`: Asserts that a warning message was logged.
    *   `expectInfoLogged()`: Asserts that an info message was logged.
    *   `expectCliCommandExecuted()`: Verifies that the Salesforce CLI was executed with the expected arguments.
    *   `expectFileExists()`: Asserts that a file exists and returns its content.
    *   `expectFileNotExists()`: Asserts that a file does not exist.

**6. Test Scenario Builder:**

*   `TestScenarioBuilder`: A class that allows building complex test scenarios by chaining configuration options.
    *   `withEnvironment()`: Sets environment variables for the scenario.
    *   `withInputs()`: Sets input values for the scenario.
    *   `withConfig()`: Sets the configuration for the scenario.
    *   `withCacheRestored()`: Configures whether the cache is restored.
    *   `withAvailableCommands()`: Specifies available commands for the scenario.
    *   `build()`: Returns the complete scenario configuration.

### Usage

You can import and use these helpers within your integration tests to create controlled and repeatable test environments. For example:

```typescript
import { setupTestWorkspace, createMockEnvironment, Assertions } from './test-helpers';

setupTestWorkspace();
const env = createMockEnvironment();

// ... your test logic ...

Assertions.expectSuccessfulExecution(mockSetOutput, { key: 'value' });