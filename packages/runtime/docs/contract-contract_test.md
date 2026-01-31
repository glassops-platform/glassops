---
type: Documentation
domain: runtime
origin: packages/runtime/internal/contract/contract_test.go
last_modified: 2026-01-31
generated: true
source: packages/runtime/internal/contract/contract_test.go
generated_at: 2026-01-31T09:03:33.246353
hash: 3a35b367629b06cac4dba35c15628d04648cf4fb51ce53f91f0bdbdf0ae3e0d1
---

## Contract Package Documentation

This package defines the `Contract` type and associated methods for representing and validating operational agreements. It provides a structured way to define expectations around deployments, quality gates, and auditability. We aim to provide a consistent and verifiable representation of these agreements.

**Key Types and Interfaces**

*   **Contract:** The central type in this package. It encapsulates information about the agreement, including schema version, metadata (adapter, engine), status, quality criteria (coverage), and audit details.  It is the primary object you will interact with.

    *   `SchemaVersion` (string): Indicates the version of the contract schema. Currently set to "1.0".
    *   `Meta` (struct): Contains metadata about the contract, specifically the adapter and engine used.
        *   `Adapter` (string): Specifies the adapter used for the contract, defaulting to "native".
        *   `Engine` (string): Specifies the engine used for the contract, defaulting to "native".
    *   `Status` (string): Represents the current status of the contract ("Succeeded", "Failed", "Blocked").
    *   `Quality` (struct): Holds quality-related information.
        *   `Coverage` (struct): Details about code coverage requirements.
            *   `Actual` (float64): The actual code coverage achieved.
            *   `Required` (float64): The required code coverage percentage.
            *   `Met` (bool): Indicates whether the actual coverage meets the required coverage.
    *   `Audit` (struct): Contains audit information related to the contract.
        *   `TriggeredBy` (string): The user or system that triggered the contract.
        *   `OrgID` (string): The organization ID associated with the contract.
        *   `Repository` (string): The repository associated with the contract.
        *   `Commit` (string): The commit associated with the contract.

**Important Functions**

*   **New(): Contract:**  This function creates and returns a new `Contract` instance with default values.  It initializes the contract with a schema version of "1.0", adapter and engine set to "native", status set to "Succeeded", and coverage required set to 80.
*   **ToJSON(): ([]byte, error):** This method converts the `Contract` object into a JSON byte array. It returns an error if the conversion fails. You can use this to serialize the contract for storage or transmission.
*   **Validate(): error:** This method validates the `Contract` instance, checking for valid status, engine, and coverage values. It returns an error if any validation fails; otherwise, it returns nil.
*   **CoverageMet():** This function is not a standalone function but a property calculation within the `Contract` type. It determines if the actual code coverage meets the required coverage and sets the `Met` field accordingly.

**Error Handling**

The package employs standard Go error handling practices. Functions return an `error` value to indicate failure.  The `Validate` function specifically returns errors for invalid status values, unsupported engine types, and invalid coverage values (negative values or values outside the 0-100 range).  The `ToJSON` function returns an error if the contract cannot be serialized into JSON.

**Concurrency**

This package does not explicitly use goroutines or channels. The `Contract` type itself is not designed for concurrent access; it is expected to be used within a single goroutine.

**Design Decisions**

*   **Default Values:** The `New()` function provides sensible default values for the `Contract` fields, simplifying initialization.
*   **Validation:** The `Validate()` method ensures data integrity and enforces constraints on the contract's state.
*   **JSON Serialization:** The `ToJSON()` method allows for easy serialization of the contract for persistence or communication.
*   **Explicit Validation Rules:**  Specific validation tests are included for status, engine, and coverage to ensure adherence to defined constraints.