---
type: Documentation
domain: runtime
origin: packages/runtime/internal/gha/gha.go
last_modified: 2026-01-29
generated: true
source: packages/runtime/internal/gha/gha.go
generated_at: 2026-01-29T21:23:12.261279
hash: eda50c569f48d9c3082b8b62388ee96fc7c75c3c418c698edbb6248843fc078d
---

## GitHub Actions Integration Package Documentation

This package provides a set of utilities designed to replicate the functionality of the `@actions/core` package commonly used in TypeScript-based GitHub Actions, but implemented in Go. It allows Go programs running within a GitHub Actions workflow to interact with the Actions environment – retrieving inputs, setting outputs, and managing logging.

**Key Responsibilities:**

*   Input retrieval from environment variables.
*   Output setting for use by subsequent workflow steps.
*   Logging and error reporting in a format understood by GitHub Actions.
*   Secret masking to prevent sensitive data from appearing in logs.

**Key Types and Interfaces:**

This package does not define any explicit types or interfaces. It operates directly on strings and environment variables.

**Important Functions:**

*   **`GetInput(name string) string`**: This function retrieves the value of an action input. It searches for the input value in the following order:
    1.  Environment variable named `INPUT_<NAME>` (where `<NAME>` is the input name in uppercase).
    2.  Environment variable named `GLASSOPS_<NAME>` (where `<NAME>` is the input name in uppercase).
    3.  If neither environment variable is found, it returns an empty string.

*   **`GetInputWithDefault(name, defaultValue string) string`**: This function attempts to retrieve an action input using `GetInput(name)`. If the input is not found (i.e., `GetInput` returns an empty string), it returns the provided `defaultValue`.

*   **`SetOutput(name, value string)`**: This function sets an action output parameter. It first attempts to write the output to the file specified by the `GITHUB_OUTPUT` environment variable. If `GITHUB_OUTPUT` is not set, it falls back to the older `::set-output` command format.  The output is written in the format `name=value` to the specified file or standard output.

*   **`SetSecret(secret string)`**: This function masks a given `secret` string in the GitHub Actions logs. This prevents sensitive information from being accidentally exposed.

*   **`SetFailed(message string)`**: This function marks the current action as failed and sets an error message. This message will be displayed in the GitHub Actions workflow run details.

*   **`Info(message string)`**: This function logs an informational message to the GitHub Actions log.

*   **`Warning(message string)`**: This function logs a warning message to the GitHub Actions log.

*   **`Error(message string)`**: This function logs an error message to the GitHub Actions log.

*   **`StartGroup(name string)`**: This function starts a new log group with the given `name`. Log messages emitted after calling this function will be visually grouped in the GitHub Actions log.

*   **`EndGroup()`**: This function ends the current log group.

**Error Handling:**

The package handles errors primarily through the `SetFailed` function.  Functions like `SetOutput` attempt to handle file operations gracefully, but errors during file writing are not explicitly returned; instead, the fallback mechanism to `::set-output` is used.  Other functions do not return errors; they rely on the caller to handle potential issues based on the function's behavior (e.g., an empty string returned by `GetInput` indicating a missing input).

**Concurrency:**

This package is not inherently concurrent. The functions are designed to be called sequentially within a single workflow step.  There are no goroutines or channels used within the package itself.

**Design Decisions:**

*   **Environment Variable Preference:** The `GetInput` function prioritizes the `INPUT_<NAME>` environment variable format, which is the standard for GitHub Actions. The `GLASSOPS_<NAME>` prefix provides a fallback for potential internal use or compatibility.
*   **Output Handling:** The `SetOutput` function supports both the modern `GITHUB_OUTPUT` file-based output mechanism and the older `::set-output` command, ensuring compatibility with different GitHub Actions runner versions.
*   **Logging Format:** The logging functions (`Info`, `Warning`, `Error`, `StartGroup`, `EndGroup`) use the `::` command format, which is recognized by GitHub Actions for structured logging and grouping.