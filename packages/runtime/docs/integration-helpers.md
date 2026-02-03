---
type: Documentation
domain: runtime
last_modified: 2026-02-02
generated: true
source: packages/runtime/internal/integration/helpers.go
generated_at: 2026-02-02T22:37:30.683798
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

*   **TestEnvironment:** This struct encapsulates the state of a test environment. It holds the path to the workspace directory, the path to the configuration file, and a map of original environment variables that were overridden during test setup.
    ```go
    type TestEnvironment struct {
    	WorkspacePath string
    	ConfigPath    string
    	OriginalEnv   map[string]string
    }
    ```

*   **FreezeWindowData:**  A struct containing pre-defined freeze window configurations for testing purposes.
    ```go
    type FreezeWindowData struct {
    	Weekend  []map[string]string
    	Weekday  []map[string]string
    	Multiple []map[string]string
    }
    ```

*   **PluginConfigData:** A struct containing pre-defined plugin configuration lists for testing.
    ```go
    type PluginConfigData struct {
    	Whitelist   []string
    	NoWhitelist []string
    	Versioned   []string
    	Scoped      []string
    }
    ```

*   **TestResultsData:** A struct containing pre-defined test result sets for testing.
    ```go
    type TestResultsData struct {
    	Valid     TestResults
    	Empty     TestResults
    	AllPassed TestResults
    	AllFailed TestResults
    }
    ```

*   **TestResults:** A struct representing the outcome of a test run, including total, passed, and failed counts.
    ```go
    type TestResults struct {
    	Total  int
    	Passed int
    	Failed int
    }
    ```

*   **CoverageTestData:** A struct containing pre-defined code coverage data for testing.
    ```go
    type CoverageTestData struct {
    	Good       Coverage
    	Borderline Coverage
    	Failing    Coverage
    	Perfect    Coverage
    }
    ```

*   **Coverage:** A struct representing code coverage information, with actual and required coverage percentages.
    ```go
    type Coverage struct {
    	Actual   float64
    	Required float64
    }
    ```

### Important Functions

*   **SetupTestWorkspace(config map[string]interface{}) (*TestEnvironment, error):** This function creates a temporary workspace directory, creates a `config` subdirectory within it, and writes a JSON configuration file (`devops-config.json`) to the `config` directory.  If a `config` is not provided, it uses a `DefaultConfig`. It returns a pointer to a `TestEnvironment` struct containing information about the created workspace and configuration file, or an error if creation fails.
*   **SetEnvironment(vars map[string]string):** This method, associated with the `TestEnvironment` struct, sets environment variables required for tests. It merges provided overrides (`vars`) with a set of default environment variables, saving the original values in the `TestEnvironment`’s `OriginalEnv` map before setting them.
*   **Cleanup():** This method, associated with the `TestEnvironment` struct, restores the original environment variables saved during setup and removes the temporary workspace directory. If an environment variable was not previously set, it is unset.
*   **WriteConfig(config map[string]interface{}) error:** This method, associated with the `TestEnvironment` struct, updates the configuration file within the test workspace with the provided `config` data.
*   **TestData:** This is a struct containing various pre-defined data sets for testing different scenarios. It includes data for freeze windows, plugin configurations, test results, and code coverage.

### Error Handling

The functions in this package generally return an `error` value to indicate failure.  Common error scenarios include:

*   Failure to create temporary directories.
*   Failure to write configuration files.
*   Errors during JSON marshaling.

In case of an error, the `SetupTestWorkspace` function attempts to clean up any created workspace before returning the error.

### Concurrency

This package does not currently employ goroutines or channels. It operates synchronously.

### Design Decisions

*   **Temporary Workspace:** Using a temporary workspace ensures that tests do not interfere with each other or with the host system.
*   **Configuration Management:**  Providing a mechanism to manage test configuration allows for flexible and repeatable tests.
*   **Environment Variable Management:**  Setting and restoring environment variables ensures that tests have the correct context and do not pollute the host environment.
*   **Pre-defined Test Data:**  Including pre-defined test data simplifies test creation and reduces duplication.
*   **Default Configuration:** The `DefaultConfig` provides a sensible baseline for tests, reducing the need to specify configuration for every test case.