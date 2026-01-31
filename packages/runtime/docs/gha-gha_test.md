---
type: Documentation
domain: runtime
origin: packages/runtime/internal/gha/gha_test.go
last_modified: 2026-01-31
generated: true
source: packages/runtime/internal/gha/gha_test.go
generated_at: 2026-01-31T10:00:49.871291
hash: 78792b077f78dcf3f17aa320e2dd3e7fb2cc6d08b94f25a00d6add71f4b7b6b3
---

## Package gha Documentation

This package provides functions for interacting with the GitHub Actions environment. It allows reading input parameters, setting output variables, and emitting diagnostic messages. The package is designed to be a lightweight abstraction over the GitHub Actions workflow environment variables and standard output.

**Key Concepts:**

The package centers around the idea of interacting with a specific environment – GitHub Actions. It reads environment variables prefixed with `INPUT_` or `GLASSOPS_` to retrieve input values and writes to standard output using a specific format to set outputs, report errors, warnings, and secrets.

**Key Functions:**

*   `GetInput(name string) string`: This function retrieves the value of an input variable. It first checks for an environment variable prefixed with `INPUT_`. If not found, it falls back to checking for a variable prefixed with `GLASSOPS_`. If neither is found, it returns an empty string.  The `INPUT_` prefix takes precedence over `GLASSOPS_`.

*   `GetInputWithDefault(name string, defaultValue string) string`: This function retrieves the value of an input variable, similar to `GetInput`. However, if the variable is not found in the environment, it returns the provided `defaultValue`.

*   `SetOutput(name string, value string)`: This function sets a GitHub Actions output variable. It writes a formatted string to standard output that GitHub Actions recognizes as an output definition. If the `GITHUB_OUTPUT` environment variable is not set, it defaults to writing to standard output.

*   `SetSecret(value string)`: This function sets a secret in GitHub Actions. It writes a formatted string to standard output that instructs GitHub Actions to mask the provided `value` in logs.

*   `SetFailed(message string)`: This function marks the GitHub Actions workflow as failed. It writes a formatted error message to standard output.

*   `Warning(message string)`: This function emits a warning message to the GitHub Actions log. It writes a formatted warning message to standard output.

*   `StartGroup(name string)`: This function starts a named group in the GitHub Actions log.  Subsequent log messages will be indented until `EndGroup` is called.

*   `EndGroup()`: This function ends the currently active group in the GitHub Actions log, removing the indentation.

**Error Handling:**

The functions in this package do not typically return explicit error values. Instead, failures are often indicated by the absence of an expected environment variable or by the inability to format the output correctly. The `SetFailed` function is used to explicitly signal a workflow failure.

**Concurrency:**

This package does not appear to use goroutines or channels. The functions are designed to be called sequentially within a single workflow step.

**Design Decisions:**

*   **Prefix Prioritization:** The package prioritizes `INPUT_` prefixed environment variables over `GLASSOPS_` prefixed variables. This allows users to override default values provided by the package with their own custom inputs.
*   **Standard Output for Outputs:** The package uses standard output to communicate outputs, secrets, errors, and warnings to GitHub Actions. This is the standard mechanism for interacting with the GitHub Actions environment.
*   **No Explicit Error Returns:** The decision to not return errors simplifies the API, but requires users to rely on the absence of expected values or the `SetFailed` function to detect failures.