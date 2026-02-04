---
type: Documentation
domain: runtime
last_modified: 2026-02-03
generated: true
source: packages/runtime/internal/contract/contract.go
generated_at: 2026-02-03T18:07:17.613568
hash: 78d28f45ec51c0421025bba3fa29a952f8c6ded5d1bb950ca12f5b04e6d69bdb
---

## Deployment Contract Package Documentation

This package defines the schema and functionality for a deployment contract, representing the governance output of a deployment process. The contract captures metadata, status, quality metrics, and audit information related to a deployment. It is designed to be generated during runtime and persisted for record-keeping and potential downstream actions.

**Key Types and Interfaces**

*   **`DeploymentContract`**: The central structure representing the complete deployment contract. It contains fields for schema version, metadata, status, quality, and audit information.
*   **`Meta`**:  Holds execution metadata such as the adapter used, the engine type (native, hardis, or custom), timestamp, and the trigger event.
*   **`Quality`**: Encapsulates code quality metrics, including code coverage and test results.
*   **`Coverage`**:  Represents code coverage data, tracking actual coverage, required coverage, and whether the requirement is met.
*   **`TestResults`**: Stores the results of test execution, including the total number of tests, the number passed, and the number failed.
*   **`Audit`**: Contains audit trail information, including who triggered the deployment, the organization ID, the repository, and the commit hash.
*   **`ValidationError`**: A custom error type used to signal validation failures within the contract. It includes the field that failed validation and a descriptive message.

**Important Functions**

*   **`New()`**: Creates a new `DeploymentContract` instance with default values.  Defaults include schema version "1.0", a native adapter and engine, the current UTC timestamp, a "Succeeded" status, and a coverage requirement of 80%.
*   **`ToJSON()`**: Serializes a `DeploymentContract` instance into a JSON byte slice with indentation for readability.
*   **`Validate()`**: Validates the `DeploymentContract` to ensure all required fields are present and contain valid values. It checks the status, engine type, and coverage percentages. Returns a `ValidationError` if validation fails, otherwise returns nil.
*   **`Generate(orgID string)`**:  Generates a `DeploymentContract` based on input data (primarily from environment variables and GitHub Actions inputs), writes it to a file named `glassops-contract.json` within a `.glassops` directory in the workspace, and returns the file path.  It retrieves test results and coverage data from inputs, defaulting to zero values if inputs are missing or invalid. It also populates audit information using environment variables.

**Error Handling**

The package employs standard Go error handling practices. Functions return an error value when operations fail.  A custom error type, `ValidationError`, is used for contract validation failures, providing specific information about the invalid field and the reason for the failure.  Errors encountered during JSON marshaling, file system operations, or input parsing are also returned.

**Concurrency**

This package does not explicitly use goroutines or channels. It operates in a single-threaded manner.

**Design Decisions**

*   **JSON Serialization**: The contract is serialized to JSON for portability and ease of integration with other systems.
*   **Input-Driven Generation**: The `Generate` function is designed to be driven by external inputs (environment variables and GitHub Actions inputs), making it flexible and adaptable to different deployment environments.
*   **Default Values**: Sensible default values are provided in the `New` function to simplify contract creation and reduce the amount of required configuration.
*   **Validation**:  The `Validate` function ensures data integrity and prevents invalid contracts from being persisted.
*   **Directory Structure**: The contract is written to a `.glassops` directory within the workspace to keep it separate from other project files. The directory is created if it does not exist.
*   **Environment Variable Fallback**: The `hasEnvOr` function provides a mechanism to retrieve environment variables with a fallback value if the variable is not set.