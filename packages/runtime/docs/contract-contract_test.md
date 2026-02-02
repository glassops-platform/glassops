---
type: Documentation
domain: runtime
origin: packages/runtime/internal/contract/contract_test.go
last_modified: 2026-02-01
generated: true
source: packages/runtime/internal/contract/contract_test.go
generated_at: 2026-02-01T19:39:37.212546
hash: 3a35b367629b06cac4dba35c15628d04648cf4fb51ce53f91f0bdbdf0ae3e0d1
---

## Contract Package Documentation

This package defines the `Contract` type and associated methods for representing and validating operational agreements. It provides a structured way to define expectations around deployments, quality gates, and auditability. We aim to provide a consistent and verifiable representation of these agreements.

**Key Types and Interfaces**

*   **`Contract`**: This is the central type in the package. It encapsulates all information related to a contract, including schema version, metadata (adapter, engine), status, quality metrics (coverage), and audit details.

    *   `SchemaVersion` (string): Indicates the version of the contract schema. Currently set to "1.0".
    *   `Meta` (struct): Contains metadata about the contract, specifically the adapter and engine used.
        *   `Adapter` (string): The adapter used for the contract, defaulting to "native".
        *   `Engine` (string): The engine used for the contract, defaulting to "native".
    *   `Status` (string): Represents the current status of the contract (e.g., "Succeeded", "Failed", "Blocked").
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

*   **`New()`**: This function creates and returns a new `Contract` instance with default values.  The default values ensure a baseline configuration for all contracts.
*   **`ToJSON()`**: This function converts the `Contract` object into a JSON byte slice. It returns an error if the conversion fails. You can use this to serialize the contract for storage or transmission.
*   **`Validate()`**: This function validates the `Contract` instance, checking for valid status and engine values, and valid coverage metrics. It returns an error if any validation fails. This is important to ensure data integrity.

**Error Handling**

The package employs standard Go error handling practices. Functions return an `error` value to indicate failure.  The `Validate()` function is central to error detection, ensuring the contract adheres to defined constraints.  Errors are checked after calling functions like `ToJSON()` and `Validate()` to handle potential issues.

**Concurrency**

This package does not currently employ goroutines or channels, and is therefore not explicitly concurrent. The `Contract` type itself is not designed to be shared concurrently without external synchronization mechanisms.

**Design Decisions**

*   **Default Values:** The `New()` function provides sensible default values for key fields, simplifying contract creation.
*   **Validation:** The `Validate()` function enforces constraints on the contract data, ensuring consistency and reliability.
*   **JSON Serialization:** The `ToJSON()` function allows for easy serialization of the contract data for storage or transmission.
*   **Explicit Validation Rules:**  Specific tests validate the allowed values for `Status` and `Engine`, and the acceptable range for coverage metrics. This approach promotes clarity and maintainability.