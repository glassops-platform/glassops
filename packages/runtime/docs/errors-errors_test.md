---
type: Documentation
domain: runtime
last_modified: 2026-02-02
generated: true
source: packages/runtime/internal/errors/errors_test.go
generated_at: 2026-02-02T22:36:18.145265
hash: 42983b4d752082e092e6310fbded062202a1677bca7087b44c5f8dc4d436d12c
---

## GlassOps Error Handling Package Documentation

This package defines custom error types and functions for managing errors within the GlassOps runtime environment. It provides a structured approach to error representation, allowing for consistent error handling and reporting across different components.

**Key Concepts**

The core idea is to create specific error types for different phases of operation (Policy, Bootstrap, Identity, Contract, Analyzer, and a general GlassOps error) and to provide functions to extract relevant information from these errors. This facilitates better error analysis and recovery.

**Types**

*   **GlassOpsError:** This is a base type for all GlassOps-specific errors. It includes a `Message` (string), `Phase` (string), `Code` (string), and an optional `Cause` (error). The `Error()` method formats the error message as "[CODE] MESSAGE". The `Unwrap()` method allows access to the underlying cause of the error, supporting error chaining.
*   **PolicyError:** Represents an error occurring during policy evaluation. It inherits from `GlassOpsError` and sets the `Phase` to "Policy" and the `Code` to "POLICY_VIOLATION".
*   **BootstrapError:** Represents an error during the bootstrap process. It inherits from `GlassOpsError`, sets the `Phase` to "Bootstrap", and the `Code` to "BOOTSTRAP_FAILED".
*   **IdentityError:** Represents an error related to identity and authentication. It inherits from `GlassOpsError`, sets the `Phase` to "Identity", and the `Code` to "AUTHENTICATION_FAILED".
*   **ContractError:** Represents an error during contract generation. It inherits from `GlassOpsError`, sets the `Phase` to "Contract", and the `Code` to "CONTRACT_GENERATION_FAILED".
*   **AnalyzerError:** Represents an error during analysis. It inherits from `GlassOpsError`, sets the `Phase` to "Analyzer", and the `Code` to "ANALYSIS_FAILED".
*   **FreezeError:** Represents an error related to a freeze policy. It inherits from `GlassOpsError`, sets the `Phase` to "Policy", the `Code` to "FROZEN", and includes a `Day` field (string) representing the day of the freeze.

**Functions**

*   **NewPolicyError(message string, cause error) error:** Constructs a new `PolicyError` with the given message and optional cause.
*   **NewBootstrapError(message string, cause error) error:** Constructs a new `BootstrapError` with the given message and optional cause.
*   **NewIdentityError(message string, cause error) error:** Constructs a new `IdentityError` with the given message and optional cause.
*   **NewContractError(message string, cause error) error:** Constructs a new `ContractError` with the given message and optional cause.
*   **NewAnalyzerError(message string, cause error) error:** Constructs a new `AnalyzerError` with the given message and optional cause.
*   **NewFreezeError(day string, startTime string, endTime string) error:** Constructs a new `FreezeError` with the given day, start time, and end time.
*   **IsGlassOpsError(err error) bool:** Checks if the given error is a `GlassOpsError` or one of its subtypes. It returns `true` if it is, and `false` otherwise.
*   **GetPhase(err error) string:** Extracts the `Phase` from a `GlassOpsError`. If the error is not a `GlassOpsError`, it returns "Unknown".
*   **GetCode(err error) string:** Extracts the `Code` from a `GlassOpsError`. If the error is not a `GlassOpsError`, it returns "UNKNOWN_ERROR".

**Error Handling Patterns**

The package promotes the use of error wrapping using the `errors.New()` function and the `Cause` field in `GlassOpsError`. This allows for preserving the original error context while adding more specific information at each layer of the application.  You can use `errors.Unwrap()` to access the root cause of an error.

**Design Decisions**

*   **Specific Error Types:** Defining specific error types for each phase allows for more targeted error handling and reporting.
*   **Error Codes:** Using error codes provides a standardized way to identify and categorize errors.
*   **Error Wrapping:**  Error wrapping preserves the original error context, aiding in debugging and troubleshooting.
*   **Phase Identification:** The `GetPhase` function provides a convenient way to determine the origin of an error within the GlassOps runtime.