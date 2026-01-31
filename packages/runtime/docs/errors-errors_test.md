---
type: Documentation
domain: runtime
origin: packages/runtime/internal/errors/errors_test.go
last_modified: 2026-01-29
generated: true
source: packages/runtime/internal/errors/errors_test.go
generated_at: 2026-01-29T21:22:50.606971
hash: 42983b4d752082e092e6310fbded062202a1677bca7087b44c5f8dc4d436d12c
---

## GlassOps Error Handling Package Documentation

This package defines custom error types and utility functions for managing errors within the GlassOps system. It provides a structured approach to error reporting, including phases, codes, and underlying causes, to aid in debugging and incident response.

**Key Types and Interfaces**

*   **GlassOpsError:** This is the base type for all custom errors within the system. It includes the following fields:
    *   `Message`: A human-readable error message.
    *   `Phase`:  A string representing the stage of the process where the error occurred (e.g., "Policy", "Bootstrap").
    *   `Code`: A unique error code for programmatic identification.
    *   `Cause`: An optional error that represents the underlying cause of this error, allowing for error chaining.  It implements the `error` interface and supports `Unwrap()` for accessing the root cause.

*   **Specific Error Types:** The package defines several concrete error types that extend `GlassOpsError`, each representing a specific category of error:
    *   `PolicyError`:  Errors related to policy violations.
    *   `BootstrapError`: Errors occurring during the bootstrapping process.
    *   `IdentityError`: Errors related to identity and authentication.
    *   `ContractError`: Errors during contract generation.
    *   `AnalyzerError`: Errors encountered during analysis.
    *   `FreezeError`: Errors related to scheduled freezes.  This type includes an additional `Day` field.

**Important Functions**

*   **NewPolicyError(message string, cause error) error:** Constructs a new `PolicyError` with the given message and optional cause.
*   **NewBootstrapError(message string, cause error) error:** Constructs a new `BootstrapError` with the given message and optional cause.
*   **NewIdentityError(message string, cause error) error:** Constructs a new `IdentityError` with the given message and optional cause.
*   **NewContractError(message string, cause error) error:** Constructs a new `ContractError` with the given message and optional cause.
*   **NewAnalyzerError(message string, cause error) error:** Constructs a new `AnalyzerError` with the given message and optional cause.
*   **NewFreezeError(day string, startTime string, endTime string) error:** Constructs a new `FreezeError` with the given day and time range.
*   **IsGlassOpsError(err error) bool:**  Checks if a given error is an instance of a `GlassOpsError` or any of its subtypes.  This allows you to determine if an error originates from within the GlassOps system.
*   **GetPhase(err error) string:**  Extracts the `Phase` from a `GlassOpsError`. If the error is not a `GlassOpsError`, it returns "Unknown".
*   **GetCode(err error) string:** Extracts the `Code` from a `GlassOpsError`. If the error is not a `GlassOpsError`, it returns "UNKNOWN_ERROR".

**Error Handling Patterns**

The package promotes a layered error handling approach.  Errors are often created with a `Cause` field, allowing you to trace the origin of an error through multiple layers of abstraction. The `Unwrap()` method on `GlassOpsError` allows you to access the underlying cause.

**Concurrency**

This package does not directly employ goroutines or channels.  It focuses solely on error definition and handling.

**Design Decisions**

*   **Structured Errors:** The use of custom error types with specific fields (Phase, Code) provides a standardized way to categorize and identify errors.
*   **Error Wrapping:**  The `Cause` field enables error wrapping, preserving the original error context while adding additional information.
*   **Type Safety:**  The use of specific error types allows for more precise error handling and avoids the need for type assertions.
*   **Utility Functions:** The `IsGlassOpsError`, `GetPhase`, and `GetCode` functions provide convenient ways to inspect and categorize errors.