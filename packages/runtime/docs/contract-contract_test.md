---
type: Documentation
domain: runtime
last_modified: 2026-02-02
generated: true
source: packages/runtime/internal/contract/contract_test.go
generated_at: 2026-02-02T22:35:40.512407
hash: 3a35b367629b06cac4dba35c15628d04648cf4fb51ce53f91f0bdbdf0ae3e0d1
---

## Contract Package Documentation

This package defines the `Contract` type and associated methods for representing and validating operational agreements. It serves as a central structure for defining expectations around deployments, quality gates, and auditability. We aim to provide a standardized way to express these contracts, enabling automated verification and enforcement.

**Key Types and Interfaces**

*   **`Contract`**: This is the primary type in the package. It encapsulates all information related to a contract, including schema version, metadata, status, quality metrics, and audit details.

    *   `SchemaVersion` (string): Indicates the version of the contract schema. Currently set to "1.0".
    *   `Meta` (struct): Contains metadata about the contract, including the adapter and engine being used.
        *   `Adapter` (string): Specifies the adapter used for the contract, defaulting to "native".
        *   `Engine` (string): Specifies the engine used for the contract, defaulting to "native".
    *   `Status` (string): Represents the current status of the contract (e.g., "Succeeded", "Failed", "Blocked").
    *   `Quality` (struct): Holds quality-related information.
        *   `Coverage` (struct): Details code coverage metrics.
            *   `Actual` (float64): The actual code coverage achieved.
            *   `Required` (float64): The minimum required code coverage.
            *   `Met` (bool): Indicates whether the actual coverage meets the required coverage.
    *   `Audit` (struct): Stores audit information related to the contract.
        *   `TriggeredBy` (string): The user or system that triggered the contract.
        *   `OrgID` (string): The organization ID associated with the contract.
        *   `Repository` (string): The repository associated with the contract.
        *   `Commit` (string): The commit associated with the contract.

**Important Functions**

*   **`New()`**: This function creates and returns a new `Contract` instance with default values. This ensures a consistent starting point for contract creation.
*   **`ToJSON()`**: This function converts the `Contract` object into a JSON byte array. It is used for serialization and persistence of contract data. It returns an error if the conversion fails. You should check for this error to ensure data integrity.
*   **`Validate()`**: This function validates the `Contract` object, checking for valid status, engine, and coverage values. It returns an error if any validation fails, providing feedback on what needs to be corrected.

**Error Handling**

The package employs standard Go error handling practices. Functions return an `error` value to indicate failure. You should always check the returned error value and handle it appropriately.  Validation errors provide specific details about the validation failure, aiding in debugging and correction.

**Concurrency**

This package does not currently employ goroutines or channels. It is designed to operate synchronously.

**Design Decisions**

*   **Default Values:** The `New()` function provides sensible default values for contract fields, simplifying contract creation and reducing boilerplate code.
*   **Validation:** The `Validate()` function enforces constraints on contract fields, ensuring data integrity and preventing invalid states.
*   **JSON Serialization:** The `ToJSON()` function enables easy serialization of contracts for storage and transmission.
*   **Explicit Validation Rules:**  Separate validation tests exist for `Status`, `Engine`, and `Coverage` to provide clear and focused validation logic. This improves maintainability and readability.