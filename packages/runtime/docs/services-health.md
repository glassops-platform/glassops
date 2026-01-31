---
type: Documentation
domain: runtime
origin: packages/runtime/internal/services/health.go
last_modified: 2026-01-29
generated: true
source: packages/runtime/internal/services/health.go
generated_at: 2026-01-29T21:27:31.918385
hash: dc0d532648eaba96338eac5b57315c16984bf5a279ba8dc55a8413c0a9c7fb2d
---

## Health Service Documentation

This document describes the `services` package, specifically the health check functionality. This service is responsible for verifying the availability and version of the Salesforce CLI.

**Package Responsibilities:**

The primary responsibility of this package is to provide a health check mechanism for dependencies required by the larger system. Currently, it focuses solely on the Salesforce CLI. This allows the system to determine if the necessary tools are present and functioning correctly before proceeding with operations.

**Key Types:**

*   `HealthCheckResult`: This struct encapsulates the outcome of the health check.
    *   `Healthy`: A boolean indicating whether the Salesforce CLI is functioning as expected.
    *   `Version`: A string representing the version of the Salesforce CLI.  If the version cannot be determined, it defaults to "unknown".
    *   `Error`: A string containing an error message if the health check failed.  This field is empty when `Healthy` is true.

**Important Functions:**

*   `HealthCheck()`: This function performs the health check for the Salesforce CLI.
    *   Signature: `func HealthCheck() HealthCheckResult`
    *   Behavior:
        1.  Executes the `sf version --json` command using `exec.Command`.
        2.  Captures the command's output and any errors.
        3.  If an error occurs during command execution:
            *   It attempts to extract the standard error output from the error, providing a more informative error message.
            *   Returns a `HealthCheckResult` with `Healthy` set to `false` and the error message populated in the `Error` field.
        4.  If the command executes successfully, it attempts to parse the JSON output.
        5.  If JSON parsing fails, it returns a `HealthCheckResult` with `Healthy` set to `false` and an appropriate error message.
        6.  If parsing succeeds, it extracts the CLI version from the JSON response, attempting to find it in multiple possible locations within the structure.
        7.  Returns a `HealthCheckResult` with `Healthy` set to `true`, the extracted version in the `Version` field, and an empty `Error` field.

**Error Handling:**

The `HealthCheck` function employs robust error handling. It specifically checks for `exec.ExitError` to capture standard error output from the CLI command, providing more detailed error information to the user.  JSON parsing errors are also handled gracefully, preventing crashes and providing a meaningful error message.

**Concurrency:**

This service does not currently employ goroutines or channels. The `HealthCheck` function is synchronous and executes sequentially.

**Design Decisions:**

*   **External Dependency Check:** The design prioritizes verifying external dependencies (Salesforce CLI) before proceeding with core functionality. This proactive approach enhances system stability.
*   **JSON Parsing:** The use of JSON parsing allows for structured retrieval of the CLI version, making the process more reliable and less prone to errors compared to parsing plain text output.
*   **Error Message Clarity:** The function attempts to provide specific and informative error messages to aid in troubleshooting.
*   **Version Fallback:** The code includes fallback logic to retrieve the CLI version from different locations within the JSON response, accommodating potential variations in the CLI output format.