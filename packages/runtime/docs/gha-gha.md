---
type: Documentation
domain: runtime
origin: packages/runtime/internal/gha/gha.go
last_modified: 2026-02-01
generated: true
source: packages/runtime/internal/gha/gha.go
generated_at: 2026-02-01T19:40:41.266387
hash: eda50c569f48d9c3082b8b62388ee96fc7c75c3c418c698edbb6248843fc078d
---

## GitHub Actions Integration Package Documentation

This package provides a set of utilities for interacting with the GitHub Actions environment. It aims to replicate functionality found in the `@actions/core` package commonly used in TypeScript-based GitHub Actions, offering equivalent capabilities within a Go environment.

**Package Responsibilities:**

The primary responsibility of this package is to simplify the process of reading inputs, setting outputs, and managing logging within a GitHub Actions workflow. It handles the specific environment variable conventions and output formatting required by the GitHub Actions platform.

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

*   **`SetSecret(secret string)`**: This function masks a given string in the GitHub Actions logs, preventing sensitive information from being displayed.

    Example:
    ```go
    gha.SetSecret("mySecretValue")
    ```

*   **`SetFailed(message string)`**: This function marks the GitHub Action as failed and sets an error message. This will cause the workflow to terminate with a failure status.

    Example:
    ```go
    gha.SetFailed("An error occurred during processing.")
    ```

*   **`Info(message string)`**: Logs an informational message to the GitHub Actions log.

    Example:
    ```go
    gha.Info("Starting the process...")
    ```

*   **`Warning(message string)`**: Logs a warning message to the GitHub Actions log.

    Example:
    ```go
    gha.Warning("Low disk space detected.")
    ```

*   **`Error(message string)`**: Logs an error message to the GitHub Actions log. This does not automatically fail the action; `SetFailed` should be used for that purpose.

    Example:
    ```go
    gha.Error("Failed to connect to the database.")
    ```

*   **`StartGroup(name string)`**: Starts a new log group with the given name. This helps organize the logs in the GitHub Actions UI.

    Example:
    ```go
    gha.StartGroup("Database Connection")
    ```

*   **`EndGroup()`**: Ends the current log group.

    Example:
    ```go
    gha.EndGroup()
    ```

**Error Handling:**

The package primarily handles errors internally, particularly within `SetOutput` when attempting to open or write to the `GITHUB_OUTPUT` file.  If an error occurs during output file operations, it is logged, and the function falls back to the older output format.  Other functions do not explicitly return errors; instead, they rely on the standard GitHub Actions logging mechanisms to indicate success or failure.  `SetFailed` is the primary method for signaling a workflow failure.

**Concurrency:**

This package does not explicitly use goroutines or channels. Its functions are designed to be called sequentially within the context of a GitHub Actions workflow step.

**Design Decisions:**

*   **Input Prioritization:** The package prioritizes `INPUT_<NAME>` environment variables, aligning with the standard GitHub Actions input convention. The `GLASSOPS_<NAME>` prefix provides a mechanism for custom input sources if needed.
*   **Output Flexibility:** The package supports both the modern `GITHUB_OUTPUT` file-based output mechanism and the older `::set-output` command, ensuring compatibility with different GitHub Actions runtime environments.
*   **Logging Consistency:** The package uses the `::` notation for logging messages (warning, error, group) to ensure proper formatting and display within the GitHub Actions UI.