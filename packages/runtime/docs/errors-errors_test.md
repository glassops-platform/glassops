---
type: Documentation
domain: runtime
origin: packages/runtime/internal/errors/errors_test.go
last_modified: 2026-02-01
generated: true
source: packages/runtime/internal/errors/errors_test.go
generated_at: 2026-02-01T19:40:19.414684
hash: 42983b4d752082e092e6310fbded062202a1677bca7087b44c5f8dc4d436d12c
---

## GlassOps Error Handling Package Documentation

This package defines custom error types and functions for managing errors within the GlassOps runtime environment. It provides a structured approach to error representation, allowing for consistent error handling and reporting across different components.

**Key Concepts**

The package centers around the `GlassOpsError` type, which extends the standard Go `error` interface. This allows for wrapping existing errors and adding contextual information specific to the GlassOps system.  The design emphasizes providing detailed error messages, associating errors with specific phases of operation, and assigning unique error codes for easier identification and programmatic handling.

**Types and Interfaces**

*   **`GlassOpsError`**: This is the core error type. It includes the following fields:
    *   `Message`: A human-readable error message.
    *   `Phase`: A string representing the phase of the operation where the error occurred (e.g., "Policy", "Bootstrap").
    *   `Code`: A unique string identifying the specific error.
    *   `Cause`: The underlying error that caused this error, allowing for error chaining.  This field implements the `Unwrap` method from the `errors` package.

*   **Specific Error Types**: The package defines several specialized error types that inherit from `GlassOpsError`, each representing a specific category of error:
    *   `PolicyError`:  Errors related to policy violations.
    *   `BootstrapError`: Errors occurring during the bootstrap process.
    *   `IdentityError`: Errors related to identity and authentication.
    *   `ContractError`: Errors during contract generation.
    *   `AnalyzerError`: Errors encountered during analysis.
    *   `FreezeError`: Errors related to scheduled freezes.  This type includes additional fields: `Day`, `StartTime`, and `EndTime`.

**Functions**

*   **`NewPolicyError(message string, cause error) *PolicyError`**: Creates a new `PolicyError` with the given message and optional cause.
*   **`NewBootstrapError(message string, cause error) *BootstrapError`**: Creates a new `BootstrapError` with the given message and optional cause.
*   **`NewIdentityError(message string, cause error) *IdentityError`**: Creates a new `IdentityError` with the given message and optional cause.
*   **`NewContractError(message string, cause error) *ContractError`**: Creates a new `ContractError` with the given message and optional cause.
*   **`NewAnalyzerError(message string, cause error) *AnalyzerError`**: Creates a new `AnalyzerError` with the given message and optional cause.
*   **`NewFreezeError(day string, startTime string, endTime string) *FreezeError`**: Creates a new `FreezeError` with the given day, start time, and end time.
*   **`IsGlassOpsError(err error) bool`**: Checks if an error is a `GlassOpsError` or one of its subtypes. You can use this function to determine if an error originates from this package.
*   **`GetPhase(err error) string`**: Extracts the `Phase` from a `GlassOpsError`. If the error is not a `GlassOpsError`, it returns "Unknown".
*   **`GetCode(err error) string`**: Extracts the `Code` from a `GlassOpsError`. If the error is not a `GlassOpsError`, it returns "UNKNOWN_ERROR".

**Error Handling Patterns**

The package promotes a layered error handling approach.  Specific error types allow components to handle errors based on their category. The `Cause` field enables tracing errors back to their root cause.  The `IsGlassOpsError` function provides a way to identify errors originating from this package, allowing for centralized handling of GlassOps-specific issues.

**Design Decisions**

*   **Explicit Error Types**:  Defining specific error types for different scenarios improves code clarity and allows for more targeted error handling.
*   **Error Codes**:  Using error codes facilitates programmatic error identification and automated responses.
*   **Error Wrapping**:  The `Cause` field allows for preserving the original error context when creating new errors, aiding in debugging.
*   **Phase Identification**: Including a `Phase` field provides valuable context about where the error occurred within the system.