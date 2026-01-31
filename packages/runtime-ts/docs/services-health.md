---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/services/health.ts
last_modified: 2026-01-29
generated: true
source: packages/runtime-ts/src/services/health.ts
generated_at: 2026-01-29T20:59:35.200164
hash: d643450fca14f67821903d531e6c29e9a99b9b9bca34c02860de6cdc68024520
---

## Health Service Documentation

**1. Introduction**

This document details the Health Service, a component designed to verify the operational status of required dependencies. Specifically, it confirms the availability and version of the Salesforce CLI. This service provides a simple mechanism for assessing the environment before proceeding with core operations.

**2. Functionality**

The Health Service provides a single function, `healthCheck`, which performs the following actions:

*   **Dependency Verification:** Executes the `sf version --json` command to determine if the Salesforce CLI is installed and accessible.
*   **Version Extraction:** Parses the command output to extract the Salesforce CLI version. It attempts to locate the version information in multiple possible output structures.
*   **Status Reporting:** Returns a `HealthCheckResult` object indicating the health status, the detected version (if available), and any error messages encountered.

**3. HealthCheckResult Interface**

The `HealthCheckResult` interface defines the structure of the service’s output:

*   `healthy`: A boolean value indicating whether the Salesforce CLI is available. `true` signifies a healthy state; `false` indicates an issue.
*   `version`: (Optional) A string representing the version of the Salesforce CLI. This field is populated when the CLI is detected.
*   `error`: (Optional) A string containing an error message if the health check fails. This provides details about the reason for the failure.

**4. Usage**

To perform a health check, call the `healthCheck` function. 

```typescript
import { healthCheck } from './health';

async function exampleUsage() {
  const result = await healthCheck();

  if (result.healthy) {
    console.log('Salesforce CLI is healthy. Version:', result.version);
  } else {
    console.error('Salesforce CLI health check failed:', result.error);
  }
}

exampleUsage();
```

**5. Error Handling**

If the `sf version --json` command fails to execute, or if the output cannot be parsed, the `healthCheck` function will return a `HealthCheckResult` object with `healthy` set to `false` and the `error` field populated with a descriptive error message.

**6. Dependencies**

This service depends on the `@actions/exec` package for executing shell commands. Ensure this dependency is installed in your project.

**7. Future Considerations**

We plan to expand this service to include checks for other dependencies and system requirements as needed.