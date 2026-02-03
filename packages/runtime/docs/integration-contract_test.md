---
type: Documentation
domain: runtime
last_modified: 2026-02-02
generated: true
source: packages/runtime/internal/integration/contract_test.go
generated_at: 2026-02-02T22:37:09.655732
hash: 7ab1e02da31199caf92781e9d2d113b04fe958b0db1d429763f36d052718bce2
---

## Integration Test Documentation: Contract Validation

This document details the integration tests for the contract package, focusing on validation and serialization functionality. The purpose of these tests is to ensure the contract object behaves as expected and enforces defined constraints.

**Package Responsibility:**

The `integration` package contains tests that verify the interaction between the components and the `contract` package. These tests are designed to confirm that the contract object can be created, validated, and serialized correctly.

**Key Types and Structures:**

The primary type involved is the `contract.Contract` structure (defined in the `github.com/glassops-platform/glassops/packages/runtime/internal/contract` package). This structure represents the core agreement or specification for a runtime operation.  Key fields within the `Contract` structure include:

*   `SchemaVersion`: A string indicating the version of the contract schema.
*   `Status`: A string representing the current status of the contract (e.g., "Succeeded", "Failed").
*   `Meta`: A nested structure containing metadata about the contract, including the execution `Engine` and a `Timestamp`.
*   `Quality`: A nested structure containing quality metrics, such as code `Coverage` and test `Results`.
*   `Audit`: A nested structure containing audit information, such as the user who `TriggeredBy` the contract, the `OrgID`, `Repository`, and `Commit`.

**Important Functions and Behavior:**

The tests focus on the following behaviors:

*   **Contract Creation with Defaults:**  The `contract.New()` function is tested to ensure it creates a contract object with sensible default values for key fields like `SchemaVersion`, `Status`, and `Meta.Engine`.
*   **Status Validation:** The `Validate()` method of the `Contract` structure is tested to verify that only valid status values are accepted.  Valid statuses include "Succeeded", "Failed", and "Blocked".  Empty strings and invalid statuses trigger an error.
*   **Engine Type Validation:** The `Validate()` method is also tested to ensure that only supported engine types are accepted. Valid engines include "native", "hardis", and "custom". Empty strings and invalid engine types result in an error.
*   **Coverage Bounds Validation:** The `Validate()` method checks that `Coverage.Actual` and `Coverage.Required` values are within acceptable bounds (non-negative and not exceeding 100). Values outside these bounds cause validation to fail.
*   **JSON Serialization:** The `ToJSON()` method is tested to confirm that a contract object can be serialized into a valid JSON representation. The test verifies that the resulting JSON data is not empty.
*   **Coverage Met Calculation:** The tests verify that the `Coverage.Met` field is correctly calculated based on whether `Coverage.Actual` is greater than or equal to `Coverage.Required`.
*   **Test Data Integration:** Tests leverage predefined test data structures (`TestData`) to populate contract fields and validate the overall contract state.

**Error Handling:**

The `Validate()` method is central to error handling. It returns an error when:

*   The `Status` field contains an invalid value.
*   The `Meta.Engine` field contains an invalid value.
*   `Coverage.Actual` or `Coverage.Required` are outside the valid range (0-100).

The tests assert that errors are returned when expected and that no errors are returned when validation should succeed.

**Concurrency:**

These integration tests do not explicitly test concurrent behavior. They focus on the functional correctness of the contract object and its validation logic in a single-threaded environment.

**Design Decisions:**

*   **Explicit Validation:** The `Validate()` method provides a clear and centralized point for enforcing contract constraints.
*   **Default Values:** Providing default values through `contract.New()` simplifies contract creation and ensures a consistent starting point.
*   **Test-Driven Approach:** The tests are structured to cover various scenarios, including valid and invalid inputs, to ensure robust validation logic.