---
type: Documentation
domain: runtime
origin: packages/runtime/internal/contract/contract_test.go
last_modified: 2026-01-29
generated: true
source: packages/runtime/internal/contract/contract_test.go
generated_at: 2026-01-29T21:21:47.606378
hash: 3a35b367629b06cac4dba35c15628d04648cf4fb51ce53f91f0bdbdf0ae3e0d1
---

## Contract Package Documentation

This package defines the `Contract` type and associated methods for representing and validating operational agreements. It provides a structured way to define expectations around deployments, quality gates, and auditability. We aim to provide a consistent and verifiable representation of these agreements.

**Key Types and Interfaces**

*   **Contract:** The central type in this package. It encapsulates information about the contract’s schema version, metadata (adapter and engine), status, quality criteria (specifically code coverage), and audit details.  It is the primary object you will interact with.

    *   `SchemaVersion` (string):  Indicates the version of the contract schema. Currently fixed at "1.0".
    *   `Meta` (struct): Contains metadata about the contract.
        *   `Adapter` (string): Specifies the adapter used for the contract, defaulting to "native".
        *   `Engine` (string): Specifies the engine used for the contract, defaulting to "native".
    *   `Status` (string): Represents the current status of the contract (e.g., "Succeeded", "Failed", "Blocked").
    *   `Quality` (struct): Holds quality-related information.
        *   `Coverage` (struct): Details about code coverage requirements.
            *   `Actual` (float64): The actual code coverage achieved.
            *   `Required` (float64): The minimum required code coverage.
            *   `Met` (bool): Indicates whether the actual coverage meets the required coverage.
    *   `Audit` (struct): Contains audit information related to the contract.
        *   `TriggeredBy` (string): The user or system that triggered the contract.
        *   `OrgID` (string): The organization ID associated with the contract.
        *   `Repository` (string): The repository associated with the contract.
        *   `Commit` (string): The commit associated with the contract.

**Important Functions**

*   **New() -> \*Contract:**  This function creates and returns a new `Contract` instance with default values.  The default values ensure a baseline configuration for all contracts.
*   **ToJSON() (\[byte[], error]):**  This method serializes the `Contract` object into a JSON byte array. It returns an error if the serialization process fails. You can use this to persist or transmit the contract data.
*   **Validate() error:** This method validates the `Contract` instance, checking for valid status, engine, and coverage values. It returns an error if any validation fails; otherwise, it returns nil.
*   **Quality.Coverage.Met = Actual >= Required:** This is not a function, but a key calculation within the `Contract` type. It determines if the code coverage requirements are met by comparing the actual coverage to the required coverage.

**Error Handling**

The package employs standard Go error handling practices. Functions return an `error` value to indicate failure.  You should always check the error return value and handle it appropriately.  Validation errors are returned by the `Validate()` method, providing specific details about what failed.

**Concurrency**

This package does not explicitly use goroutines or channels. The `Contract` type itself is not designed for concurrent access; however, you can safely use it within concurrent applications if you provide appropriate synchronization mechanisms (e.g., mutexes) when accessing and modifying the `Contract` instance.

**Design Decisions**

*   **Immutability:** While the `Contract` struct is not strictly immutable, the design encourages treating instances as read-only after creation and validation. Modifications should generally result in the creation of a new `Contract` instance.
*   **Explicit Validation:** The `Validate()` method provides a centralized point for ensuring the integrity of the `Contract` data. This helps prevent unexpected behavior due to invalid configurations.
*   **JSON Serialization:**  The `ToJSON()` method facilitates easy integration with other systems and services that rely on JSON for data exchange.