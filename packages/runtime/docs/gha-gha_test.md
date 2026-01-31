---
type: Documentation
domain: runtime
origin: packages/runtime/internal/gha/gha_test.go
last_modified: 2026-01-29
generated: true
source: packages/runtime/internal/gha/gha_test.go
generated_at: 2026-01-29T21:23:40.422398
hash: fc769e8020ad716f96454336672d4d2ed63ce6d3741e973fd3bf6b3a01ef394e
---

## Package gha Documentation

This package provides functions for interacting with the GitHub Actions environment. It allows reading input values, setting output variables, and emitting diagnostic messages (warnings, errors) formatted for GitHub Actions consumption. The package is designed to simplify the creation of GitHub Actions by abstracting away the specific syntax required for these operations.

**Key Concepts:**

The package centers around the idea of interacting with the environment variables and standard output stream in a way that GitHub Actions understands.  GitHub Actions exposes input parameters as environment variables prefixed with `INPUT_`, and outputs are set via a specific standard output format. This package handles these details.

**Key Functions:**

*   **`GetInput(name string) string`**:  Retrieves the value of an input variable. It first checks for environment variables prefixed with `INPUT_`. If not found, it falls back to checking for variables prefixed with `GLASSOPS_`.  `INPUT_` prefixed variables take precedence. If the variable is not found under either prefix, an empty string is returned.

*   **`GetInputWithDefault(name string, defaultValue string) string`**: Retrieves the value of an input variable, similar to `GetInput`. However, if the variable is not found in the environment (under either `INPUT_` or `GLASSOPS_` prefixes), it returns the provided `defaultValue`.

*   **`SetOutput(name string, value string)`**: Sets a GitHub Actions output variable. It formats the output as `::set-output name=<name>::<value>` and writes it to standard output.  This is how actions communicate results to subsequent steps.

*   **`SetSecret(value string)`**: Masks a value in the GitHub Actions log. It formats the output as `::add-mask::<value>` and writes it to standard output. This is used for sensitive information like passwords or API keys.

*   **`SetFailed(message string)`**: Marks the current action as failed. It formats the output as `::error::<message>` and writes it to standard output. This causes the workflow to stop and report an error.

*   **`Warning(message string)`**: Emits a warning message to the GitHub Actions log. It formats the output as `::warning::<message>` and writes it to standard output.

*   **`StartGroup(name string)`**: Starts a named group in the GitHub Actions log. It formats the output as `::group::<name>` and writes it to standard output. This helps organize the log output.

*   **`EndGroup()`**: Ends the current group in the GitHub Actions log. It formats the output as `::endgroup::` and writes it to standard output.

**Error Handling:**

The functions in this package do not typically return explicit error values. Instead, they rely on the behavior of GitHub Actions. For example, `GetInput` returns an empty string if the input is not found, and `SetFailed` signals failure to the workflow.  The tests verify the correct output is sent to standard output.

**Concurrency:**

This package does not explicitly use goroutines or channels. The functions are designed to be called sequentially within the context of a single GitHub Actions step.

**Design Decisions:**

*   **Prefix Fallback:** The decision to fall back to `GLASSOPS_` prefixed environment variables provides flexibility and allows for compatibility with existing systems.
*   **Standard Output for Actions Communication:**  Using standard output for setting outputs, secrets, and errors is consistent with the GitHub Actions specification.
*   **No Explicit Error Returns:** The package prioritizes simplicity and relies on GitHub Actions' built-in error handling mechanisms.