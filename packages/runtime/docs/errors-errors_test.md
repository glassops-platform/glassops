---
type: Documentation
domain: runtime
origin: packages/runtime/internal/errors/errors_test.go
last_modified: 2026-01-31
generated: true
source: packages/runtime/internal/errors/errors_test.go
generated_at: 2026-01-31T10:00:09.300059
hash: 42983b4d752082e092e6310fbded062202a1677bca7087b44c5f8dc4d436d12c
---

## GlassOps Error Handling Package Documentation

This package defines custom error types and functions for managing errors within the GlassOps system. It provides a structured approach to error representation, allowing for consistent error handling and reporting across different components.

**Key Concepts**

The package centers around the `GlassOpsError` type, which serves as a base for more specific error types related to different phases of operation. This allows for a standardized way to represent errors, including a human-readable message, a phase identifier, and an error code.  The package also provides functions to determine if an error is a `GlassOpsError`, extract the error phase, and retrieve the error code.

**Types**

*   **`GlassOpsError`**: This is the core error type. It includes:
    *   `Message`: A string providing a human-readable description of the error.
    *   `Phase`: A string indicating the stage or component where the error occurred (e.g., "Policy", "Bootstrap").
    *   `Code`: A string representing a unique error code for programmatic identification.
    *   `Cause`: An optional error that represents the underlying cause of this error, allowing for error chaining.
    *   The `Error()` method returns a formatted string combining the code and message.
    *   The `Unwrap()` method returns the underlying `Cause` error, if present.

*   **Specific Error Types**: The package defines several concrete error types derived from `GlassOpsError`, each representing errors specific to a particular operation:
    *   `PolicyError`: Errors related to policy enforcement.
    *   `BootstrapError`: Errors occurring during the bootstrapping process.
    *   `IdentityError`: Errors related to identity and authentication.
    *   `ContractError`: Errors during contract generation.
    *   `AnalyzerError`: Errors encountered during analysis.
    *   `FreezeError`: Errors related to freezing operations, including a `Day` field to indicate the day of the freeze.

**Functions**

*   **`NewPolicyError(message string, cause error) error`**: Creates a new `PolicyError` with the given message and optional cause.
*   **`NewBootstrapError(message string, cause error) error`**: Creates a new `BootstrapError` with the given message and optional cause.
*   **`NewIdentityError(message string, cause error) error`**: Creates a new `IdentityError` with the given message and optional cause.
*   **`NewContractError(message string, cause error) error`**: Creates a new `ContractError` with the given message and optional cause.
*   **`NewAnalyzerError(message string, cause error) error`**: Creates a new `AnalyzerError` with the given message and optional cause.
*   **`NewFreezeError(day string, startTime string, endTime string) error`**: Creates a new `FreezeError` with the specified day and time range.
*   **`IsGlassOpsError(err error) bool`**: Checks if the given error is a `GlassOpsError` or one of its subtypes.  You can use this function to determine if an error originated from within the GlassOps system.
*   **`GetPhase(err error) string`**: Extracts the `Phase` from a `GlassOpsError`. If the error is not a `GlassOpsError`, it returns "Unknown".
*   **`GetCode(err error) string`**: Extracts the `Code` from a `GlassOpsError`. If the error is not a `GlassOpsError`, it returns "UNKNOWN_ERROR".

**Error Handling Patterns**

The package promotes a layered error handling approach. Specific error types inherit from the base `GlassOpsError`, allowing for targeted error handling based on the error's phase or code. The `Cause` field enables error chaining, providing context about the root cause of an error.

**Concurrency**

This package does not directly involve concurrency patterns like goroutines or channels. It focuses solely on error definition and handling.

**Design Decisions**

*   **Structured Errors**: The use of custom error types with specific fields (Phase, Code) provides a structured way to represent errors, making them easier to handle and analyze.
*   **Error Chaining**: The `Cause` field allows for propagating underlying errors, providing valuable debugging information.
*   **Phase-Specific Errors**: Defining separate error types for each phase of operation improves code clarity and allows for more targeted error handling.