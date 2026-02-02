---
type: Documentation
domain: runtime
origin: packages/runtime/internal/validator/validator.go
last_modified: 2026-02-01
generated: true
source: packages/runtime/internal/validator/validator.go
generated_at: 2026-02-01T19:46:29.138413
hash: acfd2f4925dc27469ace03cf98b76bba53c90a78319da04a39b51c0c967a7414
---

## Validator Package Documentation

This package provides functions for validating the environment, inputs, and context in which the application is running. It ensures that necessary conditions are met before proceeding with operations, enhancing reliability and security.

**Key Responsibilities:**

*   Verifying the presence of required environment variables.
*   Checking for required inputs and performing basic sanitization.
*   Validating the context of execution (e.g., pull request vs. direct commit).
*   Validating the format of the Salesforce instance URL.

**Key Types and Interfaces:**

This package does not define any custom types or interfaces. It relies on built-in Go types like `string` and standard library functions.

**Important Functions:**

*   **`ValidateEnvironment()`:**
    This function checks for the existence of essential environment variables: `GITHUB_WORKSPACE`, `GITHUB_ACTOR`, and `GITHUB_REPOSITORY`. It iterates through a predefined list of required variables and returns an error if any are missing.  If all variables are present, it returns `nil`.

    ```go
    func ValidateEnvironment() error {
    	// ... implementation details ...
    }
    ```

*   **`ValidateInputs()`:**
    This function validates required inputs obtained through the GitHub Actions interface. It checks for the presence of `client_id`, `jwt_key`, and `username`.  It also validates the format of the `jwt_key` to ensure it contains "BEGIN" and "END" markers, unless `skip_auth` is set to "true".  If the `jwt_key` is provided and valid, it is marked as a secret.

    ```go
    func ValidateInputs() error {
    	// ... implementation details ...
    }
    ```

*   **`ValidateContext()`:**
    This function validates the execution context. If the `GITHUB_EVENT_NAME` is `pull_request`, it verifies the presence of the `GITHUB_HEAD_REF` environment variable. It also checks if the repository is a fork (by looking for a colon in `GITHUB_HEAD_REF`) and issues a warning if it is, recommending additional security validations.  Additionally, it validates that the `GITHUB_REPOSITORY` conforms to a standard format using a regular expression.

    ```go
    func ValidateContext() error {
    	// ... implementation details ...
    }
    ```

*   **`EnsureValidInstanceURL(url string) error`:**
    This function validates a provided Salesforce instance URL. It checks if the URL's length is greater than 8 characters and if it begins with either "http://" or "https://".  It returns an error if either of these conditions is not met.

    ```go
    func EnsureValidInstanceURL(url string) error {
    	// ... implementation details ...
    }
    ```

**Error Handling:**

All validation functions return an `error` value.  If a validation check fails, a descriptive error message is returned using `fmt.Errorf`.  The error messages typically indicate which specific requirement was not met.  If a validation check passes, the functions return `nil`.

**Concurrency:**

This package does not employ goroutines or channels and is therefore not inherently concurrent. The functions are designed to be executed sequentially.

**Design Decisions:**

*   **External Dependency:** The package depends on the `github.com/glassops-platform/glassops/packages/runtime/internal/gha` package for accessing GitHub Actions inputs and setting secrets.
*   **Regular Expression for Repository Validation:** A regular expression is used to validate the format of the `GITHUB_REPOSITORY` environment variable, providing a flexible and robust validation mechanism.
*   **JWT Key Validation:** The JWT key validation checks for the presence of "BEGIN" and "END" markers, providing a basic level of format verification.
*   **Warning for Forked Repositories:** A warning is issued when running on a forked repository to alert users to potential security implications.