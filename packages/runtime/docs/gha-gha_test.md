---
type: Documentation
domain: runtime
origin: packages/runtime/internal/gha/gha_test.go
last_modified: 2026-01-31
generated: true
source: packages/runtime/internal/gha/gha_test.go
generated_at: 2026-01-31T09:05:16.998327
hash: 78792b077f78dcf3f17aa320e2dd3e7fb2cc6d08b94f25a00d6add71f4b7b6b3
---

## Package gha Documentation

This package provides functions for interacting with the GitHub Actions environment. It allows reading input parameters, setting output variables, and reporting status information (errors, warnings, and success/failure) in a format understood by GitHub Actions.  We designed this package to simplify the creation of custom GitHub Actions.

**Key Concepts:**

The package centers around the idea of abstracting the specific mechanisms GitHub Actions uses for input, output, and status reporting. This allows actions written with this package to be more portable and easier to test.

**Key Functions:**

*   **`GetInput(name string) string`**:  Retrieves the value of an input parameter defined in the GitHub Actions workflow. It first checks for environment variables prefixed with `INPUT_`. If not found, it falls back to checking for variables prefixed with `GLASSOPS_`.  If neither prefix is found, an empty string is returned.  `INPUT_` prefixed variables take precedence. You should call this function to access values passed into your action.

*   **`GetInputWithDefault(name string, defaultValue string) string`**:  Similar to `GetInput`, but provides a default value if the input parameter is not set in the environment.  It follows the same precedence rules for `INPUT_` and `GLASSOPS_` prefixes.

*   **`SetOutput(name string, value string)`**: Sets an output variable that can be used by subsequent steps in the GitHub Actions workflow.  It formats the output string according to the GitHub Actions standard (`::set-output name=<name>::<value>`) and writes it to standard output. If the `GITHUB_OUTPUT` environment variable is not set, the output is written directly to standard output.

*   **`SetSecret(value string)`**:  Sets a secret value. This function formats the output string according to the GitHub Actions standard (`::add-mask::<value>`) and writes it to standard output, masking the value in the logs.

*   **`SetFailed(message string)`**:  Marks the current step as failed. It formats the error message according to the GitHub Actions standard (`::error::<message>`) and writes it to standard output. This will cause the workflow to stop.

*   **`Warning(message string)`**:  Outputs a warning message to the workflow log. It formats the message according to the GitHub Actions standard (`::warning::<message>`) and writes it to standard output.

*   **`StartGroup(name string)`**: Starts a named group in the workflow log.  This helps organize the output. It formats the output string according to the GitHub Actions standard (`::group::<name>`) and writes it to standard output.

*   **`EndGroup()`**: Ends the current group in the workflow log. It formats the output string according to the GitHub Actions standard (`::endgroup::`) and writes it to standard output.

**Error Handling:**

The functions in this package generally do not return explicit error values. Instead, they rely on the standard output stream for reporting status and errors to the GitHub Actions environment.  `SetFailed` is the primary mechanism for signaling a critical error that should halt the workflow.  Other functions provide default behavior (e.g., returning an empty string for missing inputs) rather than panicking.

**Concurrency:**

This package does not explicitly use goroutines or channels.  The functions are designed to be called sequentially within a single workflow step.

**Design Decisions:**

*   **Environment Variable Focus:** The package relies heavily on environment variables to receive input and communicate status. This is consistent with the way GitHub Actions provides information to steps.
*   **Standard Output Formatting:**  The package formats output according to the GitHub Actions standard, ensuring compatibility with the platform.
*   **Prefix Prioritization:** The `INPUT_` prefix is given priority over the `GLASSOPS_` prefix for input variables. This allows users to override default values provided by the package.