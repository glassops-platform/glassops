---
type: Documentation
domain: runtime
origin: packages/runtime/internal/errors/errors.go
last_modified: 2026-02-01
generated: true
source: packages/runtime/internal/errors/errors.go
generated_at: 2026-02-01T19:39:59.293373
hash: 8d74b0b3de8008cf135ed510e21fe41f8422da3ed98dbdf964fe62fb5608c129
---

## GlassOps Runtime Error Package Documentation

This package defines a structured error handling system for the GlassOps runtime. It provides a consistent way to categorize and report errors, which is important for telemetry, debugging, and integration with automation tools like GitHub Actions.

**Core Concepts**

The package centers around the `GlassOpsError` type, which serves as the base for all errors originating from the runtime.  This allows for consistent error handling and information gathering.  Specific error types are defined to represent different failure scenarios within the GlassOps system.

**Key Types**

*   **`GlassOpsError`**: This struct represents a generic GlassOps runtime error. It contains the following fields:
    *   `Message`: A human-readable description of the error.
    *   `Phase`:  Indicates the stage of the GlassOps process where the error occurred (e.g., "Policy", "Bootstrap").
    *   `Code`: A machine-readable error code for categorization and automated handling.
    *   `Cause`: An optional underlying error that led to this error, enabling error chaining.
    The `Error()` method formats the error for display, and the `Unwrap()` method allows access to the underlying cause.

*   **`PolicyError`**: Represents a violation of a governance policy. It embeds a `GlassOpsError` and adds no additional fields.
*   **`BootstrapError`**: Indicates a failure during the CLI bootstrap process. It embeds a `GlassOpsError` and adds no additional fields.
*   **`IdentityError`**: Represents an authentication or identity-related failure. It embeds a `GlassOpsError` and adds no additional fields.
*   **`ContractError`**: Indicates a failure during contract generation or validation. It embeds a `GlassOpsError` and adds no additional fields.
*   **`AnalyzerError`**: Represents a failure during code analysis. It embeds a `GlassOpsError` and adds no additional fields.
*   **`FreezeError`**: Indicates that a deployment is blocked due to a pre-defined governance freeze window. It embeds a `GlassOpsError` and includes:
    *   `Day`: The day the freeze window applies to.
    *   `Start`: The start time of the freeze window.
    *   `End`: The end time of the freeze window.

**Functions**

*   **`NewPolicyError(message string, cause error) *PolicyError`**: Creates a new `PolicyError` with the given message and underlying cause.
*   **`NewBootstrapError(message string, cause error) *BootstrapError`**: Creates a new `BootstrapError` with the given message and underlying cause.
*   **`NewIdentityError(message string, cause error) *IdentityError`**: Creates a new `IdentityError` with the given message and underlying cause.
*   **`NewContractError(message string, cause error) *ContractError`**: Creates a new `ContractError` with the given message and underlying cause.
*   **`NewAnalyzerError(message string, cause error) *AnalyzerError`**: Creates a new `AnalyzerError` with the given message and underlying cause.
*   **`NewFreezeError(day, start, end string) *FreezeError`**: Creates a new `FreezeError` representing a deployment blocked by a freeze window.
*   **`IsGlassOpsError(err error) bool`**: Checks if a given error is a `GlassOpsError` or one of its specific subtypes. This allows You to easily identify errors originating from this package.
*   **`GetPhase(err error) string`**: Extracts the `Phase` value from a `GlassOpsError`. Returns "Unknown" if the error is not a `GlassOpsError`.
*   **`GetCode(err error) string`**: Extracts the `Code` value from a `GlassOpsError`. Returns "UNKNOWN\_ERROR" if the error is not a `GlassOpsError`.

**Error Handling**

The package promotes a layered error handling approach.  Errors are often wrapped with additional context using the `Cause` field of the `GlassOpsError` struct. This allows You to trace the origin of an error and understand the sequence of events that led to it. The `Unwrap()` method facilitates this error chaining.

**Design Considerations**

*   **Structured Errors:** The use of specific error types with defined fields provides a standardized way to represent and handle errors.
*   **Error Codes:**  Machine-readable error codes enable automated error handling and reporting.
*   **Error Chaining:** The `Cause` field allows for the propagation of underlying errors, providing more detailed information about the root cause of a failure.
*   **Phase Identification:** The `Phase` field helps pinpoint where in the GlassOps process an error occurred.