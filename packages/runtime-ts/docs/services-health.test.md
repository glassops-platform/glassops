---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/services/health.test.ts
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/src/services/health.test.ts
generated_at: 2026-01-31T10:12:38.941298
hash: e417c599a11eef37d982678f0b02565e75438c27c15f6d8eff509eddf7989e6e
---

## Health Check Service Documentation

This document details the functionality and behavior of the health check service. This service is designed to verify the operational status and version of a command-line interface (CLI).

### Overview

The health check service provides a mechanism to determine if the required CLI is available and functioning correctly. It executes a command and parses the output to determine health status and CLI version. The service is intended for use in automated environments to ensure dependencies are met before proceeding with operations.

### Functionality

The primary function provided is `healthCheck()`. This function performs the following actions:

1.  Executes a predefined CLI command.
2.  Captures the standard output from the command execution.
3.  Parses the output, expecting a JSON response.
4.  Determines the health status based on command execution success and the presence of version information.
5.  Returns a result object containing health status, version information, and any error messages.

### Return Value

The `healthCheck()` function returns an object with the following properties:

*   `healthy`: A boolean value indicating the health status. `true` signifies a healthy state; `false` indicates an issue.
*   `version`: A string representing the CLI version. This will be “unknown” if the version cannot be determined from the CLI output.  This property is `undefined` if the CLI execution fails.
*   `error`: A string containing an error message if the CLI execution failed or if an error occurred during parsing. This property is `undefined` when the service is healthy.

### Error Handling

The service handles the following error conditions:

*   **CLI Execution Failure:** If the CLI command fails to execute (e.g., command not found), the `healthy` property is set to `false`, the `error` property is populated with the error message, and the `version` property is set to `undefined`.
*   **Non-Error Exceptions:**  If the CLI execution throws an exception that is not an `Error` object (e.g., a string), the `healthy` property is set to `false`, the `error` property is populated with the exception value, and the `version` property is set to `undefined`.
*   **Unexpected Output Format:** If the CLI output is valid JSON but does not contain the expected `cliVersion` or `result.cliVersion` field, the `healthy` property is set to `true`, the `version` property is set to “unknown”, and the `error` property remains `undefined`.

### Usage

You can call the `healthCheck()` function directly to assess the health of the CLI.  

```typescript
const result = await healthCheck();

if (result.healthy) {
  console.log(`CLI is healthy. Version: ${result.version}`);
} else {
  console.error(`CLI is unhealthy. Error: ${result.error}`);
}
```

### Dependencies

This service depends on the `@actions/exec` package for executing CLI commands.  The tests mock this dependency to ensure isolated testing.