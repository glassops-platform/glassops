---
type: Documentation
domain: runtime
origin: packages/runtime/internal/contract/contract.go
last_modified: 2026-01-29
generated: true
source: packages/runtime/internal/contract/contract.go
generated_at: 2026-01-29T21:21:18.789038
hash: a3f26726bb5f1246580a0587bbde1927f7284f9f574ef6261e317958a6f2bb26
---

## Deployment Contract Package Documentation

This package defines the schema for a Deployment Contract, representing the outcome of a deployment governance process. It provides structures for capturing metadata, status, quality metrics, and audit information related to a deployment.

**Key Types and Interfaces**

*   **DeploymentContract:** The central structure representing the complete deployment contract. It encapsulates metadata about the deployment, its status, quality assessment, and audit trail.
*   **Meta:**  Holds execution metadata such as the adapter used, the execution engine, the timestamp of execution, and the triggering event.
*   **Quality:** Contains code quality metrics, specifically code coverage and test results.
*   **Coverage:** Represents code coverage information, including the actual coverage achieved, the required coverage, and whether the requirement was met.
*   **TestResults:** Stores the results of test execution, including the total number of tests, the number of tests that passed, and the number that failed.
*   **Audit:**  Provides audit trail information, including the user or system that triggered the deployment, the organization ID, the repository, and the commit hash.
*   **ValidationError:** A custom error type used to signal validation failures within the contract. It includes the specific field that failed validation and a descriptive message.

**Important Functions**

*   **New() \[T any]**: This function creates and returns a pointer to a new `DeploymentContract` instance initialized with default values. The default values include a schema version of "1.0", a `Meta` structure with "native" as the adapter and engine, the current UTC timestamp, a `Status` of "Succeeded", and a `Quality` structure with a coverage requirement of 80%.
*   **ToJSON() \[T any]**: This method, associated with the `DeploymentContract` type, serializes the contract data into a JSON byte slice. It uses `json.MarshalIndent` to produce a human-readable JSON output with indentation. It returns the JSON data and any error encountered during serialization.
*   **Validate() \[T any]**: This method validates the `DeploymentContract` instance to ensure data integrity. It checks the validity of the `Status`, `Engine`, `Coverage.Actual`, and `Coverage.Required` fields against predefined allowed values and ranges. If any validation fails, it returns a `ValidationError` containing details about the failure. Otherwise, it returns `nil`.

**Error Handling**

The package employs a custom error type, `ValidationError`, to provide specific information about validation failures. This allows callers to easily identify which field(s) in the contract are invalid and why.  The `Error()` method on `ValidationError` provides a user-friendly error message.

**Design Decisions**

*   **JSON Serialization:** The use of JSON for serialization allows for easy interoperability with other systems and services.
*   **Explicit Validation:** The `Validate()` method provides a clear and centralized point for ensuring the contract's data integrity.
*   **Default Values:** The `New()` function provides sensible default values, simplifying contract creation and reducing boilerplate code.
*   **String-Based Status and Engine:** Using strings for `Status` and `Engine` allows for flexibility and easy extension, while validation ensures only permitted values are used.
*   **Custom Error Type:** The `ValidationError` type provides specific context for errors, making debugging and error handling more effective.

You can create a new contract using `New()`, populate its fields with relevant data, and then validate it using `Validate()` before processing it further. You can then serialize the contract to JSON using `ToJSON()` for storage or transmission.