---
type: Documentation
domain: runtime
last_modified: 2026-02-03
generated: true
source: packages/runtime/internal/validator/validator.go
generated_at: 2026-02-03T18:09:02.888664
hash: e3a5585f431e1c633476fa878dcdfd1d3e92ef8e4a937f56b71cbae11dae2c8e
---

## Package validator Documentation

This package is responsible for validating the environment, inputs, and context in which the application operates. It ensures that necessary conditions are met before proceeding with core functionality, enhancing reliability and security. We aim to provide clear and actionable error messages when validation fails.

### Key Types and Interfaces

This package does not define any custom types or interfaces. It relies on standard Go types like strings and errors.

### Functions

**`ValidateEnvironment()`**

This function verifies the presence of essential environment variables. It checks for `GITHUB_WORKSPACE`, `GITHUB_ACTOR`, and `GITHUB_REPOSITORY`. If any of these variables are not set, it returns an error listing the missing variables. Otherwise, it returns nil, indicating successful validation.

*Example:*

If `GITHUB_WORKSPACE` is not set, the function returns an error message similar to: `missing required environment variables: GITHUB_WORKSPACE`.

**`ValidateInputs()`**

This function validates required inputs obtained from the GitHub Actions environment. It checks for the existence of `client_id`, `jwt_key`, and `username` inputs using the `gha.GetInput()` function. If any input is missing, it returns an error listing the missing inputs. Additionally, it performs a basic format check on the `jwt_key`. If `skip_auth` is not set to "true", it verifies that the `jwt_key` contains both "BEGIN" and "END" markers. If the key is valid and not empty, it is set as a secret using `gha.SetSecret()`.

*Example:*

If `client_id` is missing, the function returns an error message similar to: `missing required inputs: client_id`.

**`ValidateContext()`**

This function validates the context of the execution, specifically checking for pull request and repository validity. If the `GITHUB_EVENT_NAME` is `pull_request`, it verifies the presence of the `GITHUB_HEAD_REF` environment variable. If `GITHUB_HEAD_REF` is missing, it returns an error. It also issues a warning via `gha.Warning()` if the pull request originates from a forked repository, recommending additional security validations.  It validates the format of the `GITHUB_REPOSITORY` environment variable using a regular expression to ensure it conforms to the expected `owner/repository` structure.

*Example:*

If `GITHUB_HEAD_REF` is missing during a pull request event, the function returns an error message: `invalid pull request context - missing GITHUB_HEAD_REF`.

**`EnsureValidInstanceURL(url string) error`**

This function validates a provided Salesforce instance URL. It checks if the URL's length is greater than 8 characters and if it begins with either "http://" or "https://". If either of these conditions is not met, it returns an error indicating an invalid URL. Otherwise, it returns nil.

*Example:*

If the provided URL is "example", the function returns an error message similar to: `invalid instance URL: example`.

### Error Handling

The package consistently uses the `error` type for signaling validation failures. Error messages are formatted to be informative, clearly indicating which variables or inputs are missing or invalid.  The `fmt.Errorf()` function is used to create error messages, allowing for dynamic inclusion of relevant details like missing variable names or invalid URL values.

### Concurrency

This package does not employ goroutines or channels. All operations are performed synchronously.

### Design Decisions

*   **Separation of Concerns:** The validation logic is divided into separate functions, each responsible for a specific aspect of validation (environment, inputs, context, URL). This promotes modularity and maintainability.
*   **Clear Error Messages:**  We prioritize providing clear and specific error messages to assist users in quickly identifying and resolving validation issues.
*   **GitHub Actions Integration:** The package is designed to work seamlessly within a GitHub Actions environment, leveraging environment variables and input parameters provided by the platform.
*   **Basic JWT Key Validation:** The JWT key validation is intentionally basic, checking only for the presence of "BEGIN" and "END" markers. More robust validation might be considered in the future, but this provides a reasonable initial check.