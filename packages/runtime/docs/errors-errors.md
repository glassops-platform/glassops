---
type: Documentation
domain: runtime
origin: packages/runtime/internal/errors/errors.go
last_modified: 2026-01-29
generated: true
source: packages/runtime/internal/errors/errors.go
generated_at: 2026-01-29T21:22:26.891471
hash: 8d74b0b3de8008cf135ed510e21fe41f8422da3ed98dbdf964fe62fb5608c129
---

## GlassOps Runtime Error Package Documentation

This package defines a structured error handling system for the GlassOps runtime. It provides a consistent way to categorize and report errors, which is important for telemetry, debugging, and integration with automation tools like GitHub Actions.

**Core Concepts**

The package centers around the `GlassOpsError` type, which serves as the base for all errors originating from the runtime.  Specific error types are then defined as structs embedding `GlassOpsError`, adding context relevant to the particular failure scenario.

**Key Types and Interfaces**

*   **`GlassOpsError`**: This struct represents a generic GlassOps runtime error.
    *   `Message`: A human-readable description of the error.
    *   `Phase`:  Indicates the stage of the runtime where the error occurred (e.g., "Policy", "Bootstrap").
    *   `Code`: A machine-readable error code for categorization and automated handling.
    *   `Cause`: An optional underlying error that led to this error, enabling error chaining.
    *   `Error()`:  Implements the `error` interface, providing a formatted string representation of the error.
    *   `Unwrap()`: Allows access to the underlying `Cause` error, supporting error wrapping and inspection.

*   **Specific Error Types**: The package defines several error types, each representing a distinct category of failure:
    *   `PolicyError`:  Indicates a violation of a governance policy.
    *   `BootstrapError`: Indicates a failure during the CLI bootstrap process.
    *   `IdentityError`: Indicates an authentication or identity-related failure.
    *   `ContractError`: Indicates a failure during contract generation or validation.
    *   `AnalyzerError`: Indicates a failure during code analysis.
    *   `FreezeError`: Indicates a deployment is blocked due to a pre-defined freeze window.  Includes `Day`, `Start`, and `End` fields to specify the freeze window details.

**Important Functions**

*   **`NewPolicyError\[message string, cause error]`**: Creates a new `PolicyError` instance. You should provide a descriptive message and any underlying error that caused the policy violation.
*   **`NewBootstrapError\[message string, cause error]`**: Creates a new `BootstrapError` instance.
*   **`NewIdentityError\[message string, cause error]`**: Creates a new `IdentityError` instance.
*   **`NewContractError\[message string, cause error]`**: Creates a new `ContractError` instance.
*   **`NewAnalyzerError\[message string, cause error]`**: Creates a new `AnalyzerError` instance.
*   **`NewFreezeError\[day string, start string, end string]`**: Creates a new `FreezeError` instance.  You must provide the day, start time, and end time of the freeze window.
*   **`IsGlassOpsError(err error)`**:  Checks if a given error is a `GlassOpsError` or one of its specific subtypes. This is useful for determining if an error originates from the GlassOps runtime.
*   **`GetPhase(err error)`**: Extracts the `Phase` value from a `GlassOpsError`. Returns "Unknown" if the error is not a `GlassOpsError`.
*   **`GetCode(err error)`**: Extracts the `Code` value from a `GlassOpsError`. Returns "UNKNOWN\_ERROR" if the error is not a `GlassOpsError`.

**Error Handling Patterns**

The package promotes error wrapping.  When creating a new error, You should include the original error as the `Cause` to preserve the error history.  The `Unwrap()` method on `GlassOpsError` allows You to traverse this chain.

**Design Decisions**

*   **Structured Errors**: The use of structured errors with defined fields (Message, Phase, Code, Cause) allows for more detailed error reporting and analysis.
*   **Error Categorization**:  Specific error types provide a clear categorization of different failure scenarios.
*   **Error Codes**: Machine-readable error codes facilitate automated error handling and integration with monitoring systems.
*   **Error Wrapping**:  The inclusion of a `Cause` field enables error wrapping, preserving the original error context.