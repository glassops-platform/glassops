---
type: Documentation
domain: runtime
last_modified: 2026-02-02
generated: true
source: packages/runtime/internal/services/health.go
generated_at: 2026-02-02T22:39:50.638629
hash: dc0d532648eaba96338eac5b57315c16984bf5a279ba8dc55a8413c0a9c7fb2d
---

## Health Service Documentation

This document describes the `health` service package, responsible for verifying the operational status of dependent command-line tools, specifically the Salesforce CLI (sf). It provides a mechanism to determine if the CLI is installed, accessible, and functioning correctly.

**Package Responsibilities:**

The primary responsibility of this package is to perform a health check on the Salesforce CLI. This involves executing a CLI command and parsing its output to determine the CLI’s health and version. This information is then returned in a structured format.

**Key Types:**

*   `HealthCheckResult`: This struct encapsulates the outcome of the health check.
    *   `Healthy`: A boolean indicating whether the Salesforce CLI is considered healthy (i.e., accessible and responding).
    *   `Version`: A string representing the version of the Salesforce CLI.  If the version cannot be determined, it defaults to "unknown".
    *   `Error`: A string containing an error message if the health check failed.  This field is empty if the check was successful.

**Important Functions:**

*   `HealthCheck()`: This function performs the health check.
    1.  It executes the command `sf version --json` using `exec.Command`.
    2.  It captures the standard output and any errors from the command.
    3.  If an error occurs during command execution (e.g., the `sf` command is not found), it constructs a `HealthCheckResult` with `Healthy` set to `false` and the error message populated. It attempts to extract the error message from the standard error stream of the command if available.
    4.  If the command executes successfully, it attempts to parse the JSON output using `json.Unmarshal`.
    5.  If JSON parsing fails, it returns a `HealthCheckResult` with `Healthy` set to `false` and an appropriate error message.
    6.  If parsing is successful, it extracts the CLI version from the JSON response. The function handles variations in the JSON structure, checking both `result.CLIVersion` and `result.Result.CLIVersion`.
    7.  Finally, it returns a `HealthCheckResult` with `Healthy` set to `true`, the extracted version, and an empty error string.

**Error Handling:**

The `HealthCheck` function employs robust error handling. It checks for errors during command execution and JSON parsing. When an error occurs, it constructs a `HealthCheckResult` with the `Healthy` flag set to `false` and a descriptive error message. The function specifically handles `exec.ExitError` to capture standard error output from the CLI command, providing more informative error messages to the user.

**Concurrency:**

This package does not currently employ goroutines or channels. The `HealthCheck` function is synchronous and executes sequentially.

**Design Decisions:**

*   **JSON Parsing:** The function relies on the Salesforce CLI providing version information in JSON format. This allows for structured parsing and avoids fragile string manipulation.
*   **Error Message Extraction:** The function attempts to extract error messages from the standard error stream of the CLI command, providing more context to the user when the CLI fails.
*   **Version Fallback:** The function includes fallback logic to handle potential variations in the JSON response structure, ensuring that the version is extracted correctly whenever possible.
*   **Simple Result Structure:** The `HealthCheckResult` struct provides a clear and concise representation of the health check outcome.

**Usage:**

You can call the `HealthCheck` function to determine the health of the Salesforce CLI. You should inspect the `Healthy` field of the returned `HealthCheckResult` to determine if the CLI is functioning correctly. If `Healthy` is `false`, you should examine the `Error` field for details about the failure.