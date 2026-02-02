---
type: Documentation
domain: runtime
origin: packages/runtime/internal/contract/contract.go
last_modified: 2026-02-01
generated: true
source: packages/runtime/internal/contract/contract.go
generated_at: 2026-02-01T19:39:18.195298
hash: 617e74d7d32ea1207809060db3935c860af6d30ba7cba135faede149ae063c59
---

## Deployment Contract Package Documentation

This package defines the schema and functionality for a Deployment Contract, representing the outcome of a governance process during software deployment. The contract captures metadata about the deployment, its status, quality metrics, and audit information. It is designed to be a standardized output for various deployment adapters and engines.

**Key Types:**

*   **DeploymentContract:** The central structure representing the complete deployment contract. It contains fields for schema version, metadata, status, quality, and audit details.
*   **Meta:**  Holds execution metadata such as the adapter used, the engine type ("native", "hardis", "custom"), a timestamp, and the trigger event.
*   **Quality:** Encapsulates code quality metrics, including code coverage and test results.
*   **Coverage:**  Details code coverage information, tracking actual coverage, required coverage, and whether the requirement is met.
*   **TestResults:** Stores the results of test execution, including the total number of tests, the number passed, and the number failed.
*   **Audit:** Contains audit trail information, including who triggered the deployment, the organization ID, the repository, and the commit hash.
*   **ValidationError:** A custom error type used to report validation failures during contract checks.

**Important Functions:**

*   **New():** Creates a new `DeploymentContract` instance with default values.  The default status is "Succeeded", the adapter and engine are "native", and a current timestamp is recorded. Coverage is initialized with a requirement of 80%.
*   **ToJSON():** Serializes a `DeploymentContract` instance into a JSON byte slice, formatted with indentation for readability.
*   **Validate():**  Performs validation checks on the `DeploymentContract` to ensure data integrity. It verifies that the `Status` field is one of the allowed values ("Succeeded", "Failed", "Blocked"), the `Engine` field is valid ("native", "hardis", "custom"), and that `Coverage` values are within the range of 0 to 100.  Returns a `ValidationError` if any check fails, otherwise returns nil.
*   **Generate(orgID string):**  Creates a `DeploymentContract` based on input parameters and environment variables, then writes it to a file named `glassops-contract.json` in the workspace. It retrieves test results and coverage data from GitHub Actions inputs.  If the inputs are invalid, it uses default values and logs a warning. It populates audit information using environment variables like `GITHUB_ACTOR`, `GITHUB_REPOSITORY`, and `GITHUB_SHA`. The function returns the path to the generated contract file or an error if one occurred.

**Error Handling:**

The package employs standard Go error handling practices. Functions return an error value when operations fail, allowing calling code to handle errors appropriately. A custom `ValidationError` type is defined for validation failures, providing specific information about the invalid field and the reason for the error.

**Concurrency:**

This package does not explicitly use goroutines or channels. It operates in a synchronous manner.

**Design Decisions:**

*   **JSON Serialization:** The contract is serialized to JSON for portability and ease of integration with other systems.
*   **Validation:**  The `Validate` function ensures the contract adheres to a defined schema, preventing invalid data from being processed.
*   **Environment Variable Handling:** The `Generate` function relies on environment variables for audit information and input parameters, making it suitable for use in CI/CD pipelines like GitHub Actions.  The `hasEnvOr` helper function provides a convenient way to retrieve environment variables with a fallback value.
*   **Input Parsing:** The `parseFloat` function handles the conversion of string inputs to float64, providing a default value if the input is invalid or empty.