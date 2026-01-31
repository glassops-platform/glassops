---
type: Documentation
domain: runtime
origin: packages/runtime/internal/gha/gha.go
last_modified: 2026-01-31
generated: true
source: packages/runtime/internal/gha/gha.go
generated_at: 2026-01-31T09:04:54.254234
hash: eda50c569f48d9c3082b8b62388ee96fc7c75c3c418c698edbb6248843fc078d
---

## GitHub Actions Integration Package Documentation

This package provides a set of utilities designed to replicate the functionality of the `@actions/core` package commonly used in TypeScript-based GitHub Actions, but implemented in Go. It allows Go programs running within a GitHub Actions workflow to interact with the Actions environment – retrieving inputs, setting outputs, and managing logging.

**Key Responsibilities:**

The primary purpose of this package is to abstract away the specifics of the GitHub Actions environment, providing a consistent interface for accessing and manipulating action data. This simplifies the development of Go-based Actions and promotes portability.

**Key Types and Interfaces:**

This package does not define any explicit types or interfaces. It operates directly on environment variables and standard output.

**Important Functions:**

*   **`GetInput(name string) string`**: This function retrieves the value of an action input. It searches for the input value in the following order:
    1.  Environment variable named `INPUT_<NAME>` (where `<NAME>` is the input name in uppercase).
    2.  Environment variable named `GLASSOPS_<NAME>` (where `<NAME>` is the input name in uppercase).
    3.  If neither environment variable is found, it returns an empty string.

*   **`GetInputWithDefault(name, defaultValue string) string`**: This function attempts to retrieve an action input using `GetInput()`. If `GetInput()` returns an empty string, it returns the provided `defaultValue`.

*   **`SetOutput(name, value string)`**: This function sets an action output. It first attempts to write the output to the file specified by the `GITHUB_OUTPUT` environment variable (the preferred method). If `GITHUB_OUTPUT` is not set, it falls back to using the older `::set-output` command format, writing directly to standard output.

*   **`SetSecret(secret string)`**: This function masks a sensitive value in the GitHub Actions logs. It uses the `::add-mask::` command to prevent the secret from being displayed.

*   **`SetFailed(message string)`**: This function marks the current action as failed and sets an error message. It uses the `::error::` command to report the error to the GitHub Actions workflow.

*   **`Info(message string)`**: This function logs an informational message to the GitHub Actions log.

*   **`Warning(message string)`**: This function logs a warning message to the GitHub Actions log, using the `::warning::` command.

*   **`Error(message string)`**: This function logs an error message to the GitHub Actions log, using the `::error::` command.

*   **`StartGroup(name string)`**: This function starts a new log group with the given name, using the `::group::` command.

*   **`EndGroup()`**: This function ends the current log group, using the `::endgroup::` command.

**Error Handling:**

The package handles errors primarily through the `SetFailed` function.  Functions like `SetOutput` check for errors when opening files, but do not return errors directly. Instead, they rely on the `GITHUB_OUTPUT` environment variable being correctly set and handle the fallback to standard output if file operations fail.  The `SetFailed` function is intended to be called when a critical error occurs that should halt the action's execution.

**Concurrency:**

This package does not explicitly use goroutines or channels. All functions are designed to be called synchronously within the main execution flow of the action.

**Design Decisions:**

*   **Environment Variable Priority:** The `GetInput` function prioritizes `INPUT_<NAME>` environment variables, aligning with the standard convention for GitHub Actions inputs. The `GLASSOPS_<NAME>` prefix provides a mechanism for internal use or overrides.
*   **Output Handling:** The package supports both the modern `GITHUB_OUTPUT` file-based output mechanism and the older `::set-output` command, ensuring compatibility with different GitHub Actions environments.
*   **Logging:** The package leverages the GitHub Actions command-line logging format (`::command::`) for structured logging, enabling better integration with the Actions platform.