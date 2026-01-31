---
type: Documentation
domain: runtime
origin: packages/runtime/internal/integration/helpers.go
last_modified: 2026-01-29
generated: true
source: packages/runtime/internal/integration/helpers.go
generated_at: 2026-01-29T21:24:28.686024
hash: e52dc4d6d6a8ca54108f60c771cbec995f0fbcd31ed3a63fa63e608e54b9e2fc
---

## Integration Test Helpers Documentation

This package provides a collection of helper functions and data structures designed to simplify the creation and management of test environments for integration tests. It focuses on providing a consistent and isolated environment for reliable test execution.

### Key Types and Interfaces

*   **TestEnvironment:** This struct encapsulates the state of a test environment. It holds the path to the workspace directory, the path to the configuration file, and a map of the original environment variables before test setup.
*   **FreezeWindowData:**  A struct containing pre-defined freeze window configurations for testing purposes. These configurations are represented as slices of maps, defining start and end times for specific days.
*   **PluginConfigData:** A struct holding pre-defined plugin configurations, including whitelists, and configurations with and without versioning.
*   **TestResultsData:** A struct containing pre-defined test result sets for various scenarios (valid, empty, all passed, all failed).
*   **TestResults:** A struct representing the outcome of a test run, including the total number of tests, the number of passed tests, and the number of failed tests.
*   **CoverageTestData:** A struct containing pre-defined code coverage data for testing purposes.
*   **Coverage:** A struct representing code coverage information, including the actual coverage percentage and the required coverage percentage.

### Important Functions

*   **SetupTestWorkspace\[config map\[string]interface{}](config map\[string]interface{}) (*TestEnvironment, error):** This function creates a temporary directory to serve as the test workspace. It creates a `config` subdirectory within the workspace and writes a JSON configuration file (`devops-config.json`) to it. If no configuration is provided, it uses a `DefaultConfig`. It returns a pointer to a `TestEnvironment` struct and an error, if any.
*   **SetEnvironment(vars map\[string]string):** This method, associated with the `TestEnvironment` struct, sets up the environment variables required for the tests. It merges provided overrides (`vars`) with a set of default environment variables, including `GITHUB_WORKSPACE`, `GITHUB_ACTOR`, and `GLASSOPS_CONFIG_PATH`. It saves the original values of the environment variables before setting the new ones.
*   **Cleanup():** This method, associated with the `TestEnvironment` struct, restores the original environment variables and removes the temporary test workspace. If an environment variable was not previously set, it is unset. Otherwise, its original value is restored.
*   **WriteConfig(config map\[string]interface{}) error:** This method, associated with the `TestEnvironment` struct, updates the configuration file within the test workspace with the provided configuration data. It marshals the configuration to JSON and writes it to the `devops-config.json` file.

### Error Handling

The functions in this package generally return an `error` value to indicate failure. Common error scenarios include:

*   Failure to create the temporary workspace directory.
*   Failure to create the configuration directory.
*   Failure to marshal the configuration data to JSON.
*   Failure to write the configuration data to the file.

In all error cases, the workspace is removed to ensure a clean state.

### Concurrency

This package does not explicitly use goroutines or channels.

### Design Decisions

*   **Temporary Workspace:** The use of a temporary workspace ensures that each test run operates in an isolated environment, preventing interference between tests.
*   **Configuration File:** The use of a JSON configuration file allows for easy customization of the test environment.
*   **Environment Variable Management:**  Saving and restoring environment variables ensures that tests do not inadvertently affect the host environment.
*   **Predefined Test Data:** The `TestData` variable provides a convenient way to access common test data, reducing code duplication and improving test maintainability.
*   **Default Configuration:** The `DefaultConfig` provides a sensible default configuration when no specific configuration is provided.