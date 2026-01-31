---
type: Documentation
domain: runtime
origin: packages/runtime/internal/contract/contract.go
last_modified: 2026-01-31
generated: true
source: packages/runtime/internal/contract/contract.go
generated_at: 2026-01-31T09:03:12.621387
hash: a3f26726bb5f1246580a0587bbde1927f7284f9f574ef6261e317958a6f2bb26
---

## Deployment Contract Package Documentation

This package defines the schema for a Deployment Contract, representing the outcome of a deployment governance process. It provides a structured way to record details about the deployment, its quality, and an audit trail.

**Key Types and Interfaces**

*   **DeploymentContract:** The central structure. It encapsulates all information related to a deployment’s governance outcome, including metadata, status, quality metrics, and audit details.
*   **Meta:** Contains metadata about the deployment execution, such as the adapter used, the execution engine, the timestamp of execution, and the triggering event.
*   **Quality:** Holds code quality metrics, specifically code coverage and test results.
*   **Coverage:** Represents code coverage information, including the actual coverage achieved, the required coverage, and whether the requirement was met.
*   **TestResults:** Stores the results of test execution, including the total number of tests, the number of tests that passed, and the number that failed.
*   **Audit:** Contains audit trail information, identifying who triggered the deployment, the organization ID, the repository, and the commit hash.
*   **ValidationError:** A custom error type used to report validation failures within the contract. It includes the specific field that failed validation and a descriptive message.

**Important Functions**

*   **New():**  Creates and returns a pointer to a new `DeploymentContract` instance initialized with default values. The default engine and adapter are set to "native", the schema version to "1.0", the status to "Succeeded", and a required code coverage of 80% is established. The timestamp is automatically set to the current UTC time.
*   **ToJSON():** Serializes a `DeploymentContract` instance into a JSON byte slice with indentation for readability. It returns the JSON data and any error encountered during serialization.
*   **Validate():**  Performs validation checks on the `DeploymentContract` instance. It verifies that the `Status` and `Engine` fields contain allowed values and that `Coverage.Actual` and `Coverage.Required` fall within the range of 0 to 100.  If any validation fails, it returns a `ValidationError` containing details about the failure. Otherwise, it returns nil.

**Error Handling**

The package employs a custom error type, `ValidationError`, to provide specific information about validation failures. This allows You to easily identify which field in the contract is invalid and understand the reason for the failure.  Functions return errors to signal issues during operations like validation or JSON serialization.

**Design Decisions**

*   **JSON Serialization:** The use of JSON for serialization allows for easy interoperability with other systems and services.
*   **Explicit Validation:** The `Validate` function provides a clear and centralized point for ensuring the integrity of the contract data.
*   **Default Values:** The `New` function provides sensible default values, simplifying contract creation and reducing boilerplate code.
*   **String-Based Status and Engine:** Using strings for `Status` and `Engine` allows for flexibility and easy extension, while validation ensures only permitted values are used.