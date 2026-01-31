---
type: Documentation
domain: runtime
origin: packages/runtime/internal/services/health.go
last_modified: 2026-01-31
generated: true
source: packages/runtime/internal/services/health.go
generated_at: 2026-01-31T09:08:38.133790
hash: dc0d532648eaba96338eac5b57315c16984bf5a279ba8dc55a8413c0a9c7fb2d
---

## Health Service Documentation

This document describes the Health Service package, responsible for verifying the availability and version of the Salesforce CLI. It provides a health check mechanism to ensure the required dependencies are present and functioning correctly.

**Package Responsibilities:**

The primary responsibility of this package is to execute a health check for the Salesforce CLI. This check confirms the CLI is installed, accessible in the system’s PATH, and capable of returning a version number. This information is presented in a structured format for consumption by other components.

**Key Types:**

* **HealthCheckResult:** This structure encapsulates the outcome of the health check.
    * `Healthy` (bool): Indicates whether the Salesforce CLI is functioning as expected. `true` signifies a successful check, `false` indicates a problem.
    * `Version` (string):  Stores the version string of the Salesforce CLI, if available.  If the version cannot be determined, it defaults to "unknown".
    * `Error` (string): Contains an error message if the health check failed. This provides details about the reason for the failure.

**Important Functions:**

* **HealthCheck(): HealthCheckResult**
    This function performs the core health check. It executes the `sf version --json` command and analyzes the output.
    1. **Command Execution:** It uses the `exec.Command` function to run the `sf version --json` command. This command requests the Salesforce CLI to output its version information in JSON format.
    2. **Error Handling:**  If the command execution fails (e.g., the `sf` command is not found), the function captures the error. It specifically handles `exec.ExitError` to extract the standard error output from the CLI, providing a more informative error message.
    3. **JSON Parsing:** If the command executes successfully, the function attempts to parse the JSON output using `json.Unmarshal`.  If parsing fails, it returns a `HealthCheckResult` indicating an error.
    4. **Version Extraction:** The function extracts the CLI version from the parsed JSON. It handles potential variations in the JSON structure by checking both `result.CLIVersion` and `result.Result.CLIVersion`.
    5. **Result Construction:** Finally, the function constructs and returns a `HealthCheckResult` containing the health status, version information, and any error messages.

**Error Handling:**

The `HealthCheck` function employs robust error handling. It captures errors during command execution and JSON parsing. When an error occurs, it constructs a `HealthCheckResult` with `Healthy` set to `false` and the `Error` field populated with a descriptive message. The function prioritizes providing specific error details, including the standard error output from the Salesforce CLI when available.

**Concurrency:**

This package does not currently employ goroutines or channels. The health check is performed synchronously.

**Design Decisions:**

* **External Dependency Check:** The service focuses on verifying an external dependency (Salesforce CLI) rather than internal service state.
* **JSON Output:** The use of `sf version --json` ensures a machine-readable output format, simplifying parsing and error handling.
* **Error Message Clarity:**  We prioritize providing informative error messages to aid in troubleshooting.  The extraction of standard error from the CLI is a key aspect of this.
* **Version Fallback:** The function includes logic to handle variations in the JSON output format of the Salesforce CLI, ensuring version information is retrieved whenever possible.