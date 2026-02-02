---
type: Documentation
domain: runtime
origin: packages/runtime/internal/gha/gha_test.go
last_modified: 2026-02-01
generated: true
source: packages/runtime/internal/gha/gha_test.go
generated_at: 2026-02-01T19:41:00.139841
hash: 78792b077f78dcf3f17aa320e2dd3e7fb2cc6d08b94f25a00d6add71f4b7b6b3
---

## Package gha Documentation

This package provides functions for interacting with the GitHub Actions environment. It allows reading input parameters, setting output variables, and emitting diagnostic messages. The package is designed to be a lightweight abstraction over the GitHub Actions workflow environment variables and standard output.

**Key Concepts:**

The package centers around the idea of interacting with a specific environment – GitHub Actions. It reads configuration from environment variables prefixed with `INPUT_` or `GLASSOPS_` and writes status and output information to standard output in a format understood by GitHub Actions.

**Key Functions:**

*   `GetInput(name string) string`: This function retrieves the value of an input variable. It first checks for an environment variable prefixed with `INPUT_`. If not found, it falls back to checking for a variable prefixed with `GLASSOPS_`. If neither is found, it returns an empty string.  The `INPUT_` prefix takes precedence over `GLASSOPS_`.

*   `GetInputWithDefault(name string, defaultValue string) string`: This function retrieves the value of an input variable, similar to `GetInput`. However, if the variable is not found in the environment, it returns the provided `defaultValue`.

*   `SetOutput(name string, value string)`: This function sets a GitHub Actions output variable. It writes a formatted string to standard output that GitHub Actions parses to set the output. If the `GITHUB_OUTPUT` environment variable is not set, it defaults to writing to standard output.

*   `SetSecret(value string)`: This function sets a secret in GitHub Actions. It writes a formatted string to standard output that masks the provided `value` in the GitHub Actions logs.

*   `SetFailed(message string)`: This function marks the GitHub Actions workflow as failed. It writes a formatted error message to standard output.

*   `Warning(message string)`: This function emits a warning message to the GitHub Actions log. It writes a formatted warning message to standard output.

*   `StartGroup(name string)`: This function starts a named group in the GitHub Actions log.  Subsequent log messages will be indented until `EndGroup` is called.

*   `EndGroup()`: This function ends the currently active group in the GitHub Actions log, removing the indentation.

**Error Handling:**

The functions in this package do not typically return explicit error values. Instead, failures are often indicated by the absence of an expected environment variable or by the inability to format the output correctly for GitHub Actions. The `SetFailed` function is used to explicitly signal a workflow failure.

**Design Decisions:**

*   **Environment Variable Prefixes:** The use of `INPUT_` and `GLASSOPS_` prefixes allows for a degree of flexibility in configuring the tool, while also providing a fallback mechanism.
*   **Standard Output for Output/Status:**  Writing output and status information to standard output is a common pattern in GitHub Actions and allows for easy integration with the platform's logging and output mechanisms.
*   **No Explicit Error Returns:** The decision to avoid explicit error returns simplifies the function signatures and aligns with the typical usage pattern in GitHub Actions, where the absence of a value or a formatted error message on standard output indicates a problem.