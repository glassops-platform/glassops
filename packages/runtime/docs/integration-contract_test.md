---
type: Documentation
domain: runtime
origin: packages/runtime/internal/integration/contract_test.go
last_modified: 2026-01-31
generated: true
source: packages/runtime/internal/integration/contract_test.go
generated_at: 2026-01-31T10:01:39.001623
hash: 7ab1e02da31199caf92781e9d2d113b04fe958b0db1d429763f36d052718bce2
---

## Integration Test Documentation: Contract Validation

This document details the integration tests for the contract package, focusing on validation and serialization functionality. The purpose of these tests is to ensure the contract object behaves as expected and enforces defined constraints.

**Package Responsibility:**

The `integration` package contains tests that verify the interaction between the components and the `contract` package. These tests are designed to confirm that the contract object can be created, validated, and serialized correctly.

**Key Types and Structures:**

The primary type involved in these tests is the `contract.Contract` struct (defined in the `github.com/glassops-platform/glassops/packages/runtime/internal/contract` package). This struct represents the core data structure for defining and validating operational agreements.  Key fields within the `Contract` struct, as tested here, include:

*   `SchemaVersion`:  A string representing the version of the contract schema.
*   `Status`: A string indicating the current status of the contract (e.g., "Succeeded", "Failed").
*   `Meta`: A nested struct containing metadata about the contract, including the `Engine` used.
*   `Quality`: A nested struct containing quality metrics, such as `Coverage` and `Tests`.
*   `Audit`: A nested struct containing audit information, such as `TriggeredBy`, `OrgID`, and `Repository`.

**Important Functions and Behavior:**

*   `contract.New()`: This function creates a new `Contract` object with default values. The tests verify that these defaults are set correctly (schema version, status, engine).
*   `contract.Validate()`: This function validates the `Contract` object, checking for invalid values in various fields.  The tests extensively use this function with different input values to confirm expected error conditions and successful validation.
*   `contract.ToJSON()`: This function serializes the `Contract` object into a JSON string. The tests verify that the serialization process does not result in errors and produces a non-empty JSON output.

**Error Handling:**

The `Validate()` function is central to error handling.  The tests systematically check that `Validate()` returns an error when invalid data is provided (e.g., an invalid status string, an unsupported engine type, negative coverage values).  Conversely, the tests confirm that `Validate()` does *not* return an error when valid data is provided.  The tests use `t.Errorf` to report validation failures, providing details about the expected and actual results.

**Test Cases and Scenarios:**

The integration tests cover the following scenarios:

*   **Default Values:** Verifies that a newly created contract has the expected default values for schema version, status, and engine.
*   **Status Validation:** Tests the validation of the `Status` field, ensuring that only allowed values are accepted.
*   **Engine Validation:** Tests the validation of the `Meta.Engine` field, ensuring that only supported engine types are accepted.
*   **Coverage Validation:** Tests the validation of `Quality.Coverage.Actual` and `Quality.Coverage.Required`, ensuring that values are non-negative and within acceptable bounds. It also verifies the `Met` field is correctly calculated.
*   **JSON Serialization:** Tests the serialization of a contract object to JSON, verifying that the process is successful and produces valid output.
*   **Test Data Integration:** Demonstrates how to populate the contract with data from external test data structures (`TestData`) and validate the resulting contract.
*   **Failing Coverage Scenario:** Specifically tests a scenario where coverage requirements are not met, verifying the correct behavior of the `Met` field.

**Design Decisions:**

*   **Table-Driven Tests:** The tests for status, engine, and coverage validation use a table-driven approach, making it easy to add new test cases and maintain the tests.
*   **Clear Error Messages:** The tests provide informative error messages that clearly indicate the expected and actual values when a test fails.
*   **Skip in Short Mode:** The entire test suite is skipped when running in short mode (`testing.Short()`), allowing for faster testing during development.