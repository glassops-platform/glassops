---
type: Documentation
domain: runtime
origin: packages/runtime/internal/integration/contract_test.go
last_modified: 2026-01-29
generated: true
source: packages/runtime/internal/integration/contract_test.go
generated_at: 2026-01-29T21:24:02.940157
hash: 7ab1e02da31199caf92781e9d2d113b04fe958b0db1d429763f36d052718bce2
---

## Integration Test Documentation: Contract Validation and Serialization

This document details the integration tests for the contract package, focusing on validation and serialization functionality. These tests verify the package’s behavior with various inputs and ensure adherence to expected contract standards.

**Package Responsibility:**

The `integration` package contains tests that interact with the `contract` package to confirm its functionality in a realistic context. These tests are designed to catch issues that might not be apparent in unit tests, particularly those related to data validation and serialization.

**Key Types and Interfaces:**

The primary type involved is the `contract.Contract` struct (defined in the `github.com/glassops-platform/glassops/packages/runtime/internal/contract` package). This struct represents the core contract object, containing fields related to status, metadata, quality metrics, and audit information.  The tests do not directly interact with interfaces, but rely on the methods exposed by the `contract.Contract` type.

**Important Functions and Behavior:**

The tests focus on the following key behaviors:

*   **`contract.New()`**: This function creates a new `contract.Contract` instance with default values. The tests verify that these defaults are correctly set (SchemaVersion = "1.0", Status = "Succeeded", Engine = "native", Timestamp is populated).
*   **`contract.Validate()`**: This method validates the contract’s data, ensuring that fields conform to defined rules. The tests cover validation of:
    *   `Status`: Valid statuses are "Succeeded", "Failed", and "Blocked".  Invalid or empty statuses result in an error.
    *   `Meta.Engine`: Valid engines are "native", "hardis", and "custom". Invalid or empty engine values result in an error.
    *   `Quality.Coverage.Actual` and `Quality.Coverage.Required`: These values must be non-negative and not exceed 100.
*   **`contract.ToJSON()`**: This method serializes the `contract.Contract` struct into a JSON string. The tests verify that the serialization process does not return an error and produces a non-empty JSON output.

**Error Handling:**

The `contract.Validate()` and `contract.ToJSON()` methods return an `error` value. The tests extensively check for expected errors based on invalid input data.  The tests assert that errors are returned when validation fails and that no errors are returned when validation succeeds.  Serialization errors are also checked.

**Concurrency:**

This test suite does not involve any concurrency patterns (goroutines or channels). The tests are designed to be executed sequentially.

**Design Decisions:**

*   **Table-Driven Tests:** The tests for status and engine validation employ a table-driven approach, making it easy to add new test cases and maintain the test suite.
*   **Test Data Helpers:** The "contract with test data" test case uses a `TestData` structure to provide pre-defined test data, improving readability and maintainability.
*   **Coverage Calculation:** The tests explicitly calculate the `Quality.Coverage.Met` field based on `Actual` and `Required` coverage values, ensuring the logic is correct.

**Test Cases Summary:**

*   **Creates valid contract with defaults:** Verifies the correct initialization of a new contract.
*   **Validates contract status:** Tests valid and invalid status values.
*   **Validates engine types:** Tests valid and invalid engine types.
*   **Validates coverage bounds:** Tests valid and invalid coverage values (actual and required).
*   **Serializes to JSON:** Verifies successful serialization to JSON.
*   **Contract with test data:** Validates a contract using predefined test data and confirms coverage is met.
*   **Failing coverage scenario:** Validates a contract where coverage is not met.

**Usage Notes:**

You may need to run these tests with the `-short` flag to skip the integration tests during rapid development cycles. However, for thorough validation, it is recommended to run the tests without the `-short` flag.