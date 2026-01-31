---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/services/health.test.ts
last_modified: 2026-01-29
generated: true
source: packages/runtime-ts/src/services/health.test.ts
generated_at: 2026-01-29T20:59:15.981662
hash: e417c599a11eef37d982678f0b02565e75438c27c15f6d8eff509eddf7989e6e
---

## Health Check Service Documentation

**1. Introduction**

This document details the functionality of the health check service. This service is designed to verify the operational status and version of a command-line interface (CLI). It provides a simple mechanism to determine if the CLI is available and functioning correctly.

**2. Functionality**

The core function, `healthCheck()`, executes a CLI command and parses its output to determine health status and version information. The service relies on the `@actions/exec` package for executing external commands.

**3. Operation**

Upon execution, `healthCheck()` performs the following steps:

*   Executes the CLI command.
*   Captures the standard output from the command.
*   Attempts to parse the output as JSON.
*   Extracts the CLI version from the JSON response. The service supports two expected JSON structures:
    *   `{ cliVersion: "version_string" }`
    *   `{ result: { cliVersion: "version_string" } }`
*   If the CLI execution is successful and a valid version is found, the service returns a `healthy` status of `true` along with the extracted version.
*   If the CLI execution fails or the JSON response does not contain a recognizable version field, the service returns a `healthy` status of `false` and an error message.
*   If the JSON format is unexpected, the version is reported as “unknown” while maintaining a `healthy` status of `true`.

**4. Return Value**

The `healthCheck()` function returns an object with the following properties:

*   `healthy`: A boolean indicating the health status of the CLI. `true` if healthy, `false` otherwise.
*   `version`: A string representing the CLI version. This will be “unknown” if the version cannot be determined from the output.  Will be undefined if the check is unhealthy.
*   `error`: A string containing an error message if the health check failed.  Undefined if the check is healthy.

**5. Error Handling**

The service handles the following error conditions:

*   **CLI Execution Failure:** If the CLI command fails to execute (e.g., command not found), the `healthy` status is set to `false`, and the `error` property is populated with the error message.
*   **Invalid JSON Response:** If the CLI output is not valid JSON, the service attempts to proceed, but may report an “unknown” version.
*   **Unexpected JSON Format:** If the JSON response does not contain the expected `cliVersion` field, the `version` is set to “unknown”, and the `healthy` status remains `true`.
*   **Non-Error Exceptions:** The service handles exceptions that are not instances of the `Error` class by setting the `error` property to the exception value.

**6. Dependencies**

*   `@actions/exec`: Used for executing the CLI command.

**7. Usage**

You can call the `healthCheck()` function directly to assess the health of the CLI.  The returned object provides information about the CLI’s status and version.