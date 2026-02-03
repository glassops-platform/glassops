---
type: Documentation
domain: runtime
last_modified: 2026-02-02
generated: true
source: packages/runtime/internal/validator/validator.go
generated_at: 2026-02-02T22:42:04.572189
hash: acfd2f4925dc27469ace03cf98b76bba53c90a78319da04a39b51c0c967a7414
---

## Package validator - Documentation

This package is responsible for validating the environment, inputs, and context in which the application operates. It ensures that all necessary components are present and correctly configured before proceeding with core functionality. We aim to provide clear and actionable error messages to assist users in resolving any issues.

### Key Types and Interfaces

This package does not define any custom types or interfaces. It relies on standard Go types like `string` and utilizes functions from the `github.com/glassops-platform/glassops/packages/runtime/internal/gha` package.

### Important Functions

**`ValidateEnvironment()`**

This function checks for the presence of required environment variables: `GITHUB_WORKSPACE`, `GITHUB_ACTOR`, and `GITHUB_REPOSITORY`. It iterates through a predefined list of required variables and returns an error if any are missing. The error message lists the missing variables.

```go
func ValidateEnvironment() error {
	// ... implementation details ...
}
```

**`ValidateInputs()`**

This function validates required inputs obtained through the GitHub Actions interface. It checks for the existence of `client_id`, `jwt_key`, and `username`.  If any are missing, it returns an error listing them. It also validates the format of the `jwt_key` if authentication is not skipped (`skip_auth` is not "true"). The `jwt_key` must contain both "BEGIN" and "END" markers to be considered valid. If a valid `jwt_key` is provided, it is set as a secret using `gha.SetSecret()`.

```go
func ValidateInputs() error {
	// ... implementation details ...
}
```

**`ValidateContext()`**

This function validates the context of the execution, specifically for pull requests. If the `GITHUB_EVENT_NAME` is `pull_request`, it verifies the presence of the `GITHUB_HEAD_REF` environment variable. If `GITHUB_HEAD_REF` is missing, an error is returned. If the `GITHUB_HEAD_REF` contains a colon (":"), a warning is issued indicating that the action is running on a forked repository and recommends additional security validations. It also validates the format of the `GITHUB_REPOSITORY` environment variable using a regular expression to ensure it conforms to the pattern `^[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+$`.

```go
func ValidateContext() error {
	// ... implementation details ...
}
```

**`EnsureValidInstanceURL(url string) error`**

This function validates a Salesforce instance URL. It checks if the URL's length is greater than 8 characters and if it begins with either "http://" or "https://". If either of these conditions is not met, it returns an error.

```go
func EnsureValidInstanceURL(url string) error {
	// ... implementation details ...
}
```

### Error Handling

All validation functions return an `error` type.  Errors are created using `fmt.Errorf()` and include descriptive messages indicating the nature of the validation failure.  The error messages are designed to be informative and help users quickly identify and resolve issues.  Missing variables or invalid formats are clearly indicated in the error messages.

### Concurrency

This package does not employ goroutines or channels. All operations are performed synchronously.

### Design Decisions

- **Separation of Concerns:** The validation logic is divided into separate functions, each responsible for validating a specific aspect of the environment, inputs, or context. This promotes modularity and maintainability.
- **Clear Error Messages:**  We prioritize providing clear and actionable error messages to simplify troubleshooting.
- **External Dependency:** The package relies on the `github.com/glassops-platform/glassops/packages/runtime/internal/gha` package for accessing inputs and setting secrets, promoting code reuse and consistency within the larger system.