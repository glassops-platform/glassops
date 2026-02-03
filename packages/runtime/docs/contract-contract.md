---
type: Documentation
domain: runtime
last_modified: 2026-02-02
generated: true
source: packages/runtime/internal/contract/contract.go
generated_at: 2026-02-02T22:35:24.362952
hash: 617e74d7d32ea1207809060db3935c860af6d30ba7cba135faede149ae063c59
---

## Deployment Contract Package Documentation

This package defines the schema and functionality for a deployment contract, representing the outcome and quality metrics of a software deployment process. It provides a structured way to record governance outputs and facilitate automated decision-making.

**Key Types and Interfaces**

*   **`DeploymentContract`**: The central data structure. It encapsulates the overall deployment status, metadata, quality metrics, and audit information.  It consists of the following fields:
    *   `SchemaVersion`:  The version of the contract schema being used.
    *   `Meta`: Metadata about the execution environment.
    *   `Status`:  The overall deployment status ("Succeeded", "Failed", or "Blocked").
    *   `Quality`:  Code quality metrics.
    *   `Audit`:  Audit trail information.

*   **`Meta`**: Contains execution metadata such as the adapter used, the engine type, timestamp, and trigger.
    *   `Adapter`: The adapter used for deployment.
    *   `Engine`: The execution engine ("native", "hardis", "custom").
    *   `Timestamp`: The time of execution in RFC3339 format.
    *   `Trigger`: The event that initiated the deployment.

*   **`Quality`**:  Holds code quality metrics.
    *   `Coverage`: Code coverage information.
    *   `Tests`: Test execution results.

*   **`Coverage`**: Tracks code coverage details.
    *   `Actual`: The actual code coverage percentage.
    *   `Required`: The required code coverage percentage.
    *   `Met`: A boolean indicating whether the coverage requirement is met.

*   **`TestResults`**:  Represents the results of test execution.
    *   `Total`: The total number of tests executed.
    *   `Passed`: The number of tests that passed.
    *   `Failed`: The number of tests that failed.

*   **`Audit`**: Contains audit trail information.
    *   `TriggeredBy`: The user or system that triggered the deployment.
    *   `OrgID`: The organization ID.
    *   `Repository`: The repository name.
    *   `Commit`: The commit SHA.

*   **`ValidationError`**: A custom error type used to indicate validation failures within the contract. It includes the field that failed validation and a descriptive message.

**Important Functions**

*   **`New()`**: Creates a new `DeploymentContract` instance with default values.  The default status is "Succeeded", the adapter and engine are "native", and a timestamp is automatically set. Coverage is initialized with a required value of 80.
*   **`ToJSON()`**: Serializes a `DeploymentContract` instance to a JSON byte slice with indentation for readability.
*   **`Validate()`**: Validates the `DeploymentContract` to ensure all required fields are present and have valid values. It checks the `Status` and `Engine` against allowed values and validates that coverage percentages are within the range of 0 to 100. Returns a `ValidationError` if validation fails, otherwise returns nil.
*   **`Generate(orgID string)`**:  Generates a `DeploymentContract` based on input parameters and environment variables, then writes it to a file named `glassops-contract.json` in the workspace. It retrieves test results and coverage data from environment variables (using `gha.GetInput`). It populates the `Audit` section with information from environment variables.  The function returns the path to the generated contract file or an error if one occurred.
*   **`hasEnvOr(key, fallback string)`**:  Retrieves the value of an environment variable. If the variable is not set, it returns the provided fallback value.
*   **`parseFloat(s string, defaultVal float64)`**: Attempts to parse a string as a float64. If the string is empty or parsing fails, it returns the provided default value.

**Error Handling**

The package employs standard Go error handling practices. Functions return an error value when operations fail.  A custom error type, `ValidationError`, is used for contract validation failures, providing specific information about the invalid field and the reason for the error.  Errors are also handled during JSON unmarshaling and file writing. Warnings are logged when test results input is invalid, allowing the process to continue with default values.

**Concurrency**

This package does not explicitly use goroutines or channels. It operates in a synchronous manner.

**Design Decisions**

*   **JSON Serialization**: The use of JSON for serialization allows for easy integration with other systems and tools.
*   **Validation**:  The `Validate()` method ensures data integrity and prevents invalid contracts from being used.
*   **Environment Variable Configuration**:  The `Generate()` function relies heavily on environment variables for configuration, making it suitable for use in CI/CD pipelines.
*   **Default Values**: Providing default values in `New()` and `parseFloat()` simplifies contract creation and reduces the need for explicit configuration in all cases.