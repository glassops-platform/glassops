---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/services/health.ts
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/src/services/health.ts
generated_at: 2026-01-31T10:12:55.976710
hash: d643450fca14f67821903d531e6c29e9a99b9b9bca34c02860de6cdc68024520
---

## Health Check Service Documentation

This document details the functionality of the Health Check service, designed to verify the operational status of required dependencies. Specifically, it confirms the availability of the Salesforce CLI.

### Overview

The Health Check service provides a simple mechanism to assess the environment’s readiness for operations. It executes a command and interprets the output to determine if dependencies are functioning correctly. This service is essential for automated systems and provides informative feedback to users.

### HealthCheckResult Interface

The service returns a `HealthCheckResult` object with the following properties:

*   `healthy`: A boolean value indicating the health status. `true` signifies a successful check; `false` indicates a failure.
*   `version`: (Optional) A string representing the version of the checked component, if available.
*   `error`: (Optional) A string containing an error message if the health check failed.

### healthCheck Function

The `healthCheck()` function performs the core health verification.

**Function Signature:**

`async function healthCheck(): Promise<HealthCheckResult>`

**Behavior:**

This function checks for the presence and functionality of the Salesforce CLI (`sf`). It executes the command `sf version --json` and parses the output. 

*   **Success:** If the command executes successfully and returns valid JSON, the function parses the JSON to extract the CLI version. The function then returns a `HealthCheckResult` object with `healthy` set to `true` and the extracted `version`. If the version information is found in different possible locations within the JSON response, the function attempts to retrieve it from each location.
*   **Failure:** If the command fails to execute (e.g., `sf` is not found in the system’s PATH) or the output is not valid JSON, the function catches the error. It then returns a `HealthCheckResult` object with `healthy` set to `false` and the error message included in the `error` property.

**Example Usage:**

```typescript
import { healthCheck } from './health';

async function checkHealth() {
  const result = await healthCheck();

  if (result.healthy) {
    console.log('System is healthy. Salesforce CLI version:', result.version);
  } else {
    console.error('System is unhealthy:', result.error);
  }
}

checkHealth();
```

**Dependencies:**

*   `@actions/exec`: This package is used to execute shell commands.

**Error Handling:**

The function includes robust error handling. It catches exceptions during command execution and JSON parsing, providing informative error messages in the `HealthCheckResult`. You should review the `error` property of the returned object to diagnose issues.