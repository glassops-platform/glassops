---
type: Documentation
domain: runtime
origin: packages/runtime/internal/gha/gha.go
last_modified: 2026-01-31
generated: true
source: packages/runtime/internal/gha/gha.go
generated_at: 2026-01-31T10:00:32.445109
hash: eda50c569f48d9c3082b8b62388ee96fc7c75c3c418c698edbb6248843fc078d
---

## GitHub Actions Integration Package Documentation

This package provides a set of utilities for interacting with the GitHub Actions environment. It aims to replicate functionality found in the `@actions/core` package commonly used in TypeScript-based GitHub Actions, offering equivalent capabilities within a Go environment.

**Package Responsibilities:**

The primary responsibility of this package is to simplify the process of reading inputs, setting outputs, and managing logging within a GitHub Actions workflow. It handles the specific environment variable conventions and output formatting required by GitHub Actions.

**Key Types and Interfaces:**

This package does not define any explicit types or interfaces. It operates directly on strings and utilizes environment variables.

**Important Functions:**

*   **`GetInput(name string) string`**: This function retrieves the value of an action input. It searches for the input value in the following order:
    1.  Environment variable named `INPUT_<NAME>` (where `<NAME>` is the input name in uppercase).
    2.  Environment variable named `GLASSOPS_<NAME>` (where `<NAME>` is the input name in uppercase).
    3.  If neither environment variable is found, it returns an empty string.

    Example:
    ```go
    inputValue := gha.GetInput("myInput")
    ```

*   **`GetInputWithDefault(name, defaultValue string) string`**: This function retrieves an action input, similar to `GetInput`. If the input is not found (i.e., `GetInput` returns an empty string), it returns the provided `defaultValue`.

    Example:
    ```go
    inputValue := gha.GetInputWithDefault("myInput", "default value")
    ```

*   **`SetOutput(name, value string)`**: This function sets an action output. It attempts to write the output to the file specified by the `GITHUB_OUTPUT` environment variable. If `GITHUB_OUTPUT` is not set, it falls back to the older `::set-output` command format.

    Example:
    ```go
    gha.SetOutput("myOutput", "output value")
    ```

*   **`SetSecret(secret string)`**: This function masks a sensitive value in the GitHub Actions logs. It uses the `::add-mask::` command to hide the value.

    Example:
    ```go
    gha.SetSecret("mySecretValue")
    ```

*   **`SetFailed(message string)`**: This function marks the GitHub Action as failed and sets an error message. It uses the `::error::` command to indicate failure.

    Example:
    ```go
    gha.SetFailed("An error occurred during processing.")
    ```

*   **`Info(message string)`**: This function logs an informational message to the GitHub Actions log.

    Example:
    ```go
    gha.Info("Starting the process...")
    ```

*   **`Warning(message string)`**: This function logs a warning message to the GitHub Actions log. It uses the `::warning::` command.

    Example:
    ```go
    gha.Warning("This is a warning message.")
    ```

*   **`Error(message string)`**: This function logs an error message to the GitHub Actions log. It uses the `::error::` command.

    Example:
    ```go
    gha.Error("An unexpected error occurred.")
    ```

*   **`StartGroup(name string)`**: This function starts a new log group with the given name. Log messages emitted after calling this function will be nested within the group. It uses the `::group::` command.

    Example:
    ```go
    gha.StartGroup("My Group")
    // Log messages here will be part of "My Group"
    ```

*   **`EndGroup()`**: This function ends the current log group. It uses the `::endgroup::` command.

    Example:
    ```go
    gha.EndGroup()
    ```

**Error Handling:**

The package primarily relies on the standard Go error handling pattern. The `SetOutput` function checks for errors when opening the output file, but generally, errors are not explicitly returned by the functions. Instead, failures are often signaled through the `SetFailed` function, which marks the entire action as failed.

**Concurrency:**

This package does not explicitly use goroutines or channels. The functions are designed to be called sequentially within the context of a single GitHub Actions workflow step.

**Design Decisions:**

*   **Environment Variable Priority:** The package prioritizes `INPUT_<NAME>` environment variables, aligning with the standard GitHub Actions input convention. It also supports a `GLASSOPS_<NAME>` prefix for potential internal use or customization.
*   **Output Handling:** The package supports both the newer file-based output mechanism (using `GITHUB_OUTPUT`) and the older `::set-output` command, providing backward compatibility.
*   **Logging:** The package leverages the `::` command syntax for structured logging in GitHub Actions, ensuring that messages are correctly formatted and displayed in the workflow logs.