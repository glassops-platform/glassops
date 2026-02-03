---
type: Documentation
domain: runtime
last_modified: 2026-02-02
generated: true
source: packages/runtime/internal/errors/errors.go
generated_at: 2026-02-02T22:35:59.940933
hash: 8d74b0b3de8008cf135ed510e21fe41f8422da3ed98dbdf964fe62fb5608c129
---

## GlassOps Runtime Error Package Documentation

This package defines a structured error handling system for the GlassOps runtime. The primary goal is to provide consistent error categorization, which supports improved telemetry and integration with automation tools like GitHub Actions.

**Key Concepts**

The package introduces a base error type, `GlassOpsError`, and several specialized error types derived from it. This hierarchy allows for specific error identification while maintaining a common structure for handling and reporting.

**GlassOpsError Type**

The `GlassOpsError` struct is the foundation for all errors within the GlassOps runtime.

```go
type GlassOpsError struct {
	Message string
	Phase   string
	Code    string
	Cause   error
}
```

*   `Message`: A human-readable description of the error.
*   `Phase`:  Indicates the stage of the runtime where the error occurred (e.g., "Policy", "Bootstrap").
*   `Code`: A unique identifier for the error type, intended for programmatic handling and analysis.
*   `Cause`: An optional nested error that provides additional context or the original error that triggered this error. This supports error chaining.

The `Error()` method satisfies the `error` interface, providing a formatted string representation of the error, including the code, message, and underlying cause if present. The `Unwrap()` method allows access to the underlying `Cause` error, enabling error inspection and handling.

**Specialized Error Types**

The package defines several specific error types, each representing a distinct category of failure:

*   `PolicyError`:  Indicates a violation of a governance policy.
*   `BootstrapError`: Indicates a failure during the CLI bootstrap process.
*   `IdentityError`: Indicates an authentication or identity-related failure.
*   `ContractError`: Indicates a failure during contract generation or validation.
*   `AnalyzerError`: Indicates a failure during code analysis.
*   `FreezeError`: Indicates that a deployment is blocked due to a pre-defined governance freeze window. This type includes additional fields: `Day`, `Start`, and `End` to specify the freeze window details.

Each specialized error type is a struct embedding the `GlassOpsError` type, adding specific context relevant to that error category.  Each has a corresponding `New...Error` function to create instances. For example:

```go
func NewPolicyError(message string, cause error) *PolicyError {
	return &PolicyError{
		GlassOpsError: GlassOpsError{
			Message: message,
			Phase:   "Policy",
			Code:    "POLICY_VIOLATION",
			Cause:   cause,
		},
	}
}
```

These `New...Error` functions simplify error creation and ensure consistent error structure.

**Utility Functions**

The package provides utility functions for working with `GlassOpsError` types:

*   `IsGlassOpsError(err error) bool`:  Determines if a given error is a `GlassOpsError` or one of its specialized types. This allows you to check if an error belongs to this system.
*   `GetPhase(err error) string`: Extracts the `Phase` value from a `GlassOpsError`. Returns "Unknown" if the error is not a `GlassOpsError`.
*   `GetCode(err error) string`: Extracts the `Code` value from a `GlassOpsError`. Returns "UNKNOWN_ERROR" if the error is not a `GlassOpsError`.

**Error Handling Patterns**

The package encourages the use of error chaining through the `Cause` field in the `GlassOpsError` struct.  When creating a new error, you should include the original error as the `Cause` whenever possible. This provides a complete history of the error and aids in debugging.

**Concurrency**

This package does not explicitly use goroutines or channels. It focuses on defining error types and providing utility functions for handling them.  Error handling itself is not inherently concurrent.

**Design Decisions**

*   **Structured Errors:** The use of a structured error type allows for programmatic analysis and reporting of errors.
*   **Error Codes:**  Unique error codes facilitate automated error handling and correlation with external systems.
*   **Error Chaining:** The `Cause` field enables tracing the origin of errors and understanding the sequence of events that led to the failure.
*   **Phase Categorization:** The `Phase` field provides a high-level categorization of errors, aiding in identifying problematic areas of the runtime.