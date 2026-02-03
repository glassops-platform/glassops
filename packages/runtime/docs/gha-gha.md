---
type: Documentation
domain: runtime
last_modified: 2026-02-02
generated: true
source: packages/runtime/internal/gha/gha.go
generated_at: 2026-02-02T22:36:37.420154
hash: eda50c569f48d9c3082b8b62388ee96fc7c75c3c418c698edbb6248843fc078d
---

## GitHub Actions Integration Package Documentation

This package provides a set of utilities for interacting with the GitHub Actions environment. It aims to replicate functionality found in the `@actions/core` package commonly used in TypeScript-based GitHub Actions, offering equivalent capabilities within a Go environment.

**Package Responsibilities:**

The primary responsibility of this package is to simplify the process of reading inputs, setting outputs, and managing logging within a GitHub Actions workflow. It handles the specific environment variable conventions and output formatting required by GitHub Actions.

**Key Types and Interfaces:**

This package does not define any explicit types or interfaces. It operates directly on strings and utilizes environment variables.

**Important Functions and Their Behavior:**

*   **`GetInput(name string) string`**: This function retrieves the value of an action input. It searches for the input value in the following order:
    1.  Environment variable named `INPUT_<NAME>` (where `<NAME>` is the input name in uppercase).
    2.  Environment variable named `GLASSOPS_<NAME>` (where `<NAME>` is the input name in uppercase).
    3.  If neither environment variable is found, it returns an empty string.

*   **`GetInputWithDefault(name, defaultValue string) string`**: This function attempts to retrieve an action input using `GetInput()`. If `GetInput()` returns an empty string, it returns the provided `defaultValue`.

*   **`SetOutput(name, value string)`**: This function sets an action output parameter. It first checks for the presence of the `GITHUB_OUTPUT` environment variable.
    1.  If `GITHUB_OUTPUT` is set, it appends a line to the file specified by the variable in the format `name=value`. This is the preferred method for setting outputs in modern GitHub Actions.
    2.  If `GITHUB_OUTPUT` is not set, it falls back to the older `::set-output` command format, printing to standard output.

*   **`SetSecret(secret string)`**: This function masks a sensitive value in the GitHub Actions logs. It uses the `::add-mask::` command to hide the value.

*   **`SetFailed(message string)`**: This function marks the GitHub Action as failed and sets an error message. It uses the `::error::` command to report the error.

*   **`Info(message string)`**: This function logs an informational message to the GitHub Actions log. It prints the message to standard output.

*   **`Warning(message string)`**: This function logs a warning message to the GitHub Actions log. It uses the `::warning::` command to format the warning.

*   **`Error(message string)`**: This function logs an error message to the GitHub Actions log. It uses the `::error::` command to format the error.

*   **`StartGroup(name string)`**: This function starts a new log group in the GitHub Actions log. It uses the `::group::` command to begin the group.

*   **`EndGroup()`**: This function ends the current log group in the GitHub Actions log. It prints `::endgroup::` to standard output.

**Error Handling:**

The package handles errors primarily through the `SetFailed` function.  Functions like `SetOutput` include basic error checking (specifically when opening the output file) but do not return errors directly. Instead, they rely on the `SetFailed` function to signal a failure to the GitHub Actions workflow.

**Concurrency:**

This package is not inherently concurrent. The functions are designed to be called sequentially within a single workflow step. There are no goroutines or channels used within the provided code.

**Design Decisions:**

*   **Environment Variable Preference:** The `GetInput` function prioritizes the `INPUT_<NAME>` environment variable convention, aligning with standard GitHub Actions practices. The `GLASSOPS_<NAME>` prefix provides a fallback for custom integrations.
*   **Output Method Flexibility:** The `SetOutput` function supports both the modern file-based output mechanism (using `GITHUB_OUTPUT`) and the older `::set-output` command, ensuring compatibility with a wider range of GitHub Actions environments.
*   **Logging Consistency:** The package provides a consistent set of logging functions (`Info`, `Warning`, `Error`) that leverage the GitHub Actions command format for structured logging.