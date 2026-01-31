---
type: Documentation
domain: runtime
origin: packages/runtime/internal/errors/errors.go
last_modified: 2026-01-31
generated: true
source: packages/runtime/internal/errors/errors.go
generated_at: 2026-01-31T09:59:44.974290
hash: 8d74b0b3de8008cf135ed510e21fe41f8422da3ed98dbdf964fe62fb5608c129
---

## GlassOps Runtime Error Package Documentation

This package defines a structured error handling system for the GlassOps runtime environment. It provides a consistent way to categorize and report errors, which is important for telemetry, debugging, and integration with automation tools like GitHub Actions.

**Core Concepts**

The package centers around the `GlassOpsError` type, which serves as the base for all errors originating from the runtime.  This allows for consistent error handling and information gathering across different components.  Errors are designed to be wrapped, providing context about the origin and cause of failures.

**Key Types**

*   **`GlassOpsError`**: This struct represents a generic GlassOps runtime error. It contains the following fields:
    *   `Message`: A human-readable description of the error.
    *   `Phase`:  Indicates the stage of the runtime where the error occurred (e.g., "Policy", "Bootstrap").
    *   `Code`: A unique string identifier for the error type. This is intended for programmatic identification and filtering.
    *   `Cause`: An optional nested error that provides more detail about the root cause of the problem. This supports error wrapping.
    *   The `Error()` method implements the `error` interface, providing a formatted string representation of the error.
    *   The `Unwrap()` method allows access to the underlying `Cause` error, enabling error chain traversal.

*   **Specific Error Types**:  The package defines several concrete error types that inherit from `GlassOpsError`, each representing a specific category of failure:
    *   `PolicyError`:  Indicates a violation of a governance policy.
    *   `BootstrapError`: Indicates a failure during the CLI bootstrap process.
    *   `IdentityError`: Indicates an authentication or identity-related failure.
    *   `ContractError`: Indicates a failure during contract generation or validation.
    *   `AnalyzerError`: Indicates a failure during code analysis.
    *   `FreezeError`: Indicates that a deployment is blocked due to a pre-defined governance freeze window. This type includes additional fields: `Day`, `Start`, and `End` to specify the freeze window details.

**Functions**

*   **`NewPolicyError(message string, cause error) *PolicyError`**: Creates a new `PolicyError` instance. You should provide a descriptive message and any underlying error that contributed to the policy violation.
*   **`NewBootstrapError(message string, cause error) *BootstrapError`**: Creates a new `BootstrapError` instance.
*   **`NewIdentityError(message string, cause error) *IdentityError`**: Creates a new `IdentityError` instance.
*   **`NewContractError(message string, cause error) *ContractError`**: Creates a new `ContractError` instance.
*   **`NewAnalyzerError(message string, cause error) *AnalyzerError`**: Creates a new `AnalyzerError` instance.
*   **`NewFreezeError(day, start, end string) *FreezeError`**: Creates a new `FreezeError` instance.  You must provide the day, start time, and end time of the freeze window.
*   **`IsGlassOpsError(err error) bool`**:  Checks if a given error is a `GlassOpsError` or one of its subtypes. This is useful for determining if an error originated from the GlassOps runtime.
*   **`GetPhase(err error) string`**: Extracts the `Phase` value from a `GlassOpsError`. If the error is not a `GlassOpsError`, it returns "Unknown".
*   **`GetCode(err error) string`**: Extracts the `Code` value from a `GlassOpsError`. If the error is not a `GlassOpsError`, it returns "UNKNOWN_ERROR".

**Error Handling and Wrapping**

We encourage wrapping errors to provide context.  The `Cause` field in `GlassOpsError` allows you to chain errors together, preserving the original error while adding information about where the error occurred.  The `Unwrap()` method allows you to traverse this chain.

**Design Considerations**

*   **Standardized Error Types**: The use of specific error types allows for more precise error handling and reporting.
*   **Error Codes**:  The `Code` field provides a machine-readable identifier for each error type, facilitating automated processing and analysis.
*   **Error Wrapping**:  The `Cause` field enables the creation of rich error contexts, making it easier to diagnose and resolve issues.
*   **Phase Identification**: The `Phase` field helps pinpoint the stage of the runtime where the error originated.