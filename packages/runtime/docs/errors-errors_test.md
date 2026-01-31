---
type: Documentation
domain: runtime
origin: packages/runtime/internal/errors/errors_test.go
last_modified: 2026-01-31
generated: true
source: packages/runtime/internal/errors/errors_test.go
generated_at: 2026-01-31T09:04:33.761503
hash: 42983b4d752082e092e6310fbded062202a1677bca7087b44c5f8dc4d436d12c
---

## GlassOps Error Handling Package Documentation

This package defines custom error types and functions for managing errors within the GlassOps runtime environment. It provides a structured approach to error reporting, including phases, codes, and underlying causes, to aid in debugging and incident response.

**Key Concepts**

The core idea is to create specific error types for different parts of the system, allowing for more targeted error handling and reporting.  Errors are designed to be easily identifiable and contain contextual information.

**Types**

*   **GlassOpsError:** This is the base type for all GlassOps-specific errors. It includes a `Message` (human-readable description), `Phase` (the component where the error occurred), `Code` (a machine-readable error code), and an optional `Cause` (the underlying error that triggered this error).  The `Error()` method formats the error message as "[CODE] MESSAGE".  It also supports error unwrapping using the `Unwrap()` method to access the underlying cause.
*   **PolicyError:** Represents errors related to policy enforcement. The `Phase` is always "Policy" and the `Code` is "POLICY_VIOLATION".
*   **BootstrapError:** Represents errors occurring during the bootstrapping process. The `Phase` is always "Bootstrap" and the `Code` is "BOOTSTRAP_FAILED".
*   **IdentityError:** Represents errors related to identity and authentication. The `Phase` is always "Identity" and the `Code` is "AUTHENTICATION_FAILED".
*   **ContractError:** Represents errors during contract generation. The `Phase` is always "Contract" and the `Code` is "CONTRACT_GENERATION_FAILED".
*   **AnalyzerError:** Represents errors occurring during analysis. The `Phase` is always "Analyzer" and the `Code` is "ANALYSIS_FAILED".
*   **FreezeError:** Represents errors related to scheduled freezes. The `Phase` is always "Policy" and the `Code` is "FROZEN". It also includes a `Day` field to indicate the day of the freeze.

**Functions**

*   **NewPolicyError(message string, cause error) error:** Creates a new `PolicyError` with the given message and optional cause.
*   **NewBootstrapError(message string, cause error) error:** Creates a new `BootstrapError` with the given message and optional cause.
*   **NewIdentityError(message string, cause error) error:** Creates a new `IdentityError` with the given message and optional cause.
*   **NewContractError(message string, cause error) error:** Creates a new `ContractError` with the given message and optional cause.
*   **NewAnalyzerError(message string, cause error) error:** Creates a new `AnalyzerError` with the given message and optional cause.
*   **NewFreezeError(day string, startTime string, endTime string) error:** Creates a new `FreezeError` with the given day, start time, and end time.
*   **IsGlassOpsError(err error) bool:** Checks if the given error is a `GlassOpsError` or one of its subtypes. You can use this to determine if an error should be handled by the GlassOps error handling system.
*   **GetPhase(err error) string:** Returns the phase associated with the error. If the error is not a `GlassOpsError`, it returns "Unknown".
*   **GetCode(err error) string:** Returns the error code associated with the error. If the error is not a `GlassOpsError`, it returns "UNKNOWN_ERROR".

**Error Handling Patterns**

The package encourages wrapping underlying errors with `GlassOpsError` to provide context.  The `Unwrap()` method allows you to access the original error for more detailed investigation.  The `IsGlassOpsError()` function provides a way to identify errors originating from this package.

**Design Decisions**

*   **Specific Error Types:**  Using specific error types for different components improves code clarity and allows for more precise error handling.
*   **Error Codes:** Machine-readable error codes facilitate automated error processing and reporting.
*   **Error Wrapping:** Wrapping underlying errors preserves the original error information while adding contextual details.
*   **Phase Identification:** The `Phase` field helps pinpoint the source of the error within the system.