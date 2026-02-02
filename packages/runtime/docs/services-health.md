---
type: Documentation
domain: runtime
origin: packages/runtime/internal/services/health.go
last_modified: 2026-02-01
generated: true
source: packages/runtime/internal/services/health.go
generated_at: 2026-02-01T19:44:09.438919
hash: dc0d532648eaba96338eac5b57315c16984bf5a279ba8dc55a8413c0a9c7fb2d
---

## Health Service Documentation

This document describes the `services` package, specifically the `health` service, responsible for verifying the availability and version of the Salesforce CLI.

**Package Purpose:**

The `services` package provides internal services for the larger application. The `health` service focuses on ensuring the Salesforce CLI is correctly installed and functioning, a prerequisite for many operations. It provides a simple health check mechanism.

**Key Types:**

*   `HealthCheckResult`: This structure encapsulates the outcome of the health check.
    *   `Healthy`: A boolean indicating whether the Salesforce CLI is considered healthy (available and responding).
    *   `Version`: A string representing the version of the Salesforce CLI, if available.  Defaults to "unknown" if the version cannot be determined.
    *   `Error`: A string containing an error message if the health check failed.  Empty if the check passed.

**Important Functions:**

*   `HealthCheck()`: This function performs the health check. It executes the `sf version --json` command and parses the output.
    1.  **Command Execution:** It uses the `exec.Command` function to run the `sf version --json` command. This command requests the Salesforce CLI to output its version information in JSON format.
    2.  **Error Handling:** If the command execution fails (e.g., `sf` is not found in the system's PATH), the function captures the error. It specifically handles `exec.ExitError` to extract the standard error output from the CLI, providing a more informative error message.
    3.  **JSON Parsing:** The output of the command is parsed as JSON into a temporary structure.
    4.  **Version Extraction:** The function attempts to extract the CLI version from the parsed JSON. It handles variations in the JSON structure, checking both `result.CLIVersion` and `result.Result.CLIVersion`.
    5.  **Result Construction:** Finally, it constructs a `HealthCheckResult` structure based on the outcome. If the command executed successfully and a version was found, `Healthy` is set to `true`. Otherwise, `Healthy` is set to `false`, and the `Error` field is populated with the relevant error message.

**Error Handling:**

The `HealthCheck` function employs robust error handling. It checks for errors during command execution and JSON parsing. When an error occurs, it returns a `HealthCheckResult` with `Healthy` set to `false` and a descriptive error message in the `Error` field. The function attempts to provide specific error messages from the Salesforce CLI when available.

**Concurrency:**

This service does not currently employ goroutines or channels. The `HealthCheck` function is a synchronous operation.

**Design Decisions:**

*   **External Dependency:** The service relies on the external Salesforce CLI being installed and accessible in the system's PATH.
*   **JSON Parsing:** Using the `--json` flag with the `sf version` command allows for reliable parsing of the CLI version, avoiding potential issues with text-based parsing.
*   **Error Message Clarity:** The function prioritizes providing informative error messages to aid in troubleshooting.
*   **Version Fallback:** The code includes a fallback mechanism to retrieve the CLI version from different locations within the JSON response, accommodating potential changes in the CLI's output format.