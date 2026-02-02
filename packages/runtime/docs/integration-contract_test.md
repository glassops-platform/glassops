---
type: Documentation
domain: runtime
origin: packages/runtime/internal/integration/contract_test.go
last_modified: 2026-02-01
generated: true
source: packages/runtime/internal/integration/contract_test.go
generated_at: 2026-02-01T19:41:15.126602
hash: 7ab1e02da31199caf92781e9d2d113b04fe958b0db1d429763f36d052718bce2
---

## Integration Test Documentation: Contract Validation

This document details the integration tests for the contract package, focusing on validation and serialization functionality. The purpose of these tests is to ensure the contract object behaves as expected and enforces defined constraints.

**Package Responsibility:**

The `integration` package contains tests that verify the interaction between the components and the `contract` package. These tests are designed to confirm that the contract object can be created, validated, and serialized correctly.

**Key Types and Structures:**

The primary type involved is the `contract.Contract` structure (defined in the `github.com/glassops-platform/glassops/packages/runtime/internal/contract` package). This structure represents the core contract object, containing fields related to status, metadata, quality metrics, and audit information.  The tests also make use of custom test data structures, such as `TestData.CoverageData` and `TestData.TestResults`, to provide various input scenarios.

**Important Functions and Behavior:**

*   **`contract.New()`**: This function creates a new `contract.Contract` instance with default values. The tests verify that these defaults are set correctly (schema version, status, engine, timestamp).
*   **`contract.Validate()`**: This function validates the contract object, checking for invalid values in key fields. The tests extensively use this function with various input values to confirm that validation works as expected.  Specifically, it checks the validity of the `Status` and `Meta.Engine` fields, as well as the `Quality.Coverage.Actual` and `Quality.Coverage.Required` values.
*   **`contract.ToJSON()`**: This function serializes the contract object into a JSON string. The tests verify that the serialization process does not result in errors and produces a non-empty JSON output.

**Error Handling:**

The `contract.Validate()` function returns an error when validation fails. The tests check for the presence of errors when invalid input is provided and verify that no errors are returned when valid input is used.  The tests use `t.Errorf` and `t.Fatalf` to report validation failures and serialization errors, respectively.

**Test Cases and Scenarios:**

The integration tests cover the following scenarios:

*   **Default Values:** Verifies that a newly created contract has the expected default values.
*   **Status Validation:** Tests the validation of the `Status` field, ensuring that only allowed values are accepted. Invalid or empty statuses trigger an error.
*   **Engine Validation:** Tests the validation of the `Meta.Engine` field, ensuring that only allowed engine types are accepted. Invalid or empty engine types trigger an error.
*   **Coverage Bounds Validation:** Validates the `Quality.Coverage.Actual` and `Quality.Coverage.Required` values, ensuring they are within acceptable ranges (non-negative and not exceeding 100).
*   **JSON Serialization:** Confirms that the contract object can be successfully serialized into a JSON string.
*   **Valid Contract with Test Data:** Creates a contract using predefined test data and verifies that validation passes and coverage is met.
*   **Failing Coverage Scenario:** Creates a contract with test data designed to result in failing coverage and verifies that the `Quality.Coverage.Met` field is correctly set to false.

**Design Decisions:**

The tests employ a table-driven approach for validating status and engine types, making the tests more concise and easier to maintain.  The use of dedicated test data structures (`TestData`) promotes code reusability and improves test readability. The tests are skipped when running in short mode (`testing.Short()`) to reduce execution time during development.