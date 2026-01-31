---
type: Documentation
domain: runtime
origin: packages/runtime/internal/integration/helpers.go
last_modified: 2026-01-31
generated: true
source: packages/runtime/internal/integration/helpers.go
generated_at: 2026-01-31T10:02:06.718135
hash: e52dc4d6d6a8ca54108f60c771cbec995f0fbcd31ed3a63fa63e608e54b9e2fc
---

## Integration Test Helpers Documentation

This package provides a collection of helper functions and data structures to simplify the creation and management of test environments for integration tests. It focuses on providing a consistent and isolated environment for reliable test execution.

### Package Responsibilities

The primary responsibilities of this package are:

*   Creating temporary workspaces for tests.
*   Managing test configuration files.
*   Setting and restoring environment variables required by tests.
*   Providing pre-defined test data for common scenarios.
*   Cleaning up test artifacts after execution.

### Key Types and Interfaces

*   **TestEnvironment:** This struct encapsulates the state of a test environment. It holds the path to the workspace directory (`WorkspacePath`), the path to the configuration file (`ConfigPath`), and a map of the original environment variables (`OriginalEnv`) before the test environment was set up.

*   **FreezeWindowData:**  A struct containing pre-defined freeze window configurations for testing purposes.  It includes configurations for weekend, weekday, and multiple freeze windows.

*   **PluginConfigData:** A struct containing pre-defined plugin configurations, including whitelists, configurations without whitelists, versioned plugins, and scoped plugins.

*   **TestResultsData:** A struct containing pre-defined test result data, including valid, empty, all-passed, and all-failed test result sets.

*   **TestResults:** A struct representing the results of a test run, containing the total number of tests, the number of tests that passed, and the number of tests that failed.

*   **CoverageTestData:** A struct containing pre-defined code coverage data, including good, borderline, failing, and perfect coverage scenarios.

*   **Coverage:** A struct representing code coverage information, containing the actual coverage percentage and the required coverage percentage.

### Important Functions

*   **SetupTestWorkspace(config map[string]interface{}) (*TestEnvironment, error):** This function creates a temporary workspace directory, creates a `config` subdirectory within it, and writes a JSON configuration file (`devops-config.json`) to the `config` directory.  If the provided `config` is nil, it uses a `DefaultConfig` map. It returns a pointer to a `TestEnvironment` struct representing the created environment, or an error if creation fails.

*   **SetEnvironment(vars map[string]string):** This function sets environment variables required for the tests. It starts with a set of default environment variables (e.g., `GITHUB_WORKSPACE`, `GITHUB_ACTOR`) and merges them with any overrides provided in the `vars` map. Before setting the variables, it saves the original values in the `TestEnvironment`’s `OriginalEnv` map.

*   **Cleanup():** This function restores the original environment variables saved during `SetEnvironment` and removes the temporary workspace directory. If an environment variable was not previously set, it is unset.

*   **WriteConfig(config map[string]interface{}) error:** This function updates the configuration file within the test workspace with the provided `config` data. It marshals the `config` map into JSON format and writes it to the `ConfigPath` specified in the `TestEnvironment`.

### Error Handling

The functions in this package generally return an `error` value to indicate failure.  Common error scenarios include:

*   Failure to create directories.
*   Failure to write to files.
*   Failure to marshal data to JSON.

In case of an error, the `SetupTestWorkspace` function attempts to remove the created workspace to ensure a clean state.

### Concurrency

This package does not explicitly use goroutines or channels.  Its operations are primarily synchronous and file-system based.

### Design Decisions

*   **Configuration Management:** The package provides a mechanism for managing test configuration through a JSON file. This allows for flexible and customizable test setups.
*   **Environment Isolation:** The use of temporary workspaces and environment variable management ensures that tests are isolated from each other and from the host environment.
*   **Test Data Factories:** The `TestData` variable provides pre-defined test data for common scenarios, reducing the need for repetitive data creation in tests.
*   **Cleanup:** The `Cleanup` function ensures that test artifacts are removed after execution, preventing resource leaks and ensuring a clean testing environment.