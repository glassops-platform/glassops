---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/protocol/contract.ts
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/src/protocol/contract.ts
generated_at: 2026-01-31T10:10:16.914480
hash: 1b4bf84390a17dfe0003a547e8c270362a2a45bc6e276dae6611f4a4d7806f41
---

## Deployment Contract Specification

This document details the structure and content of the Deployment Contract, a standardized format for communicating deployment information. It ensures consistent data exchange between different components within the system.

### Overview

The Deployment Contract defines a common interface for representing the outcome of a deployment process. It encompasses metadata about the deployment, its status, quality metrics, and audit information. This contract is validated using a schema to guarantee data integrity.

### Schema Definition

The core of the Deployment Contract is defined by a Zod schema, ensuring type safety and validation. The schema specifies the following properties:

*   **schemaVersion**: `string` (default: "1.0"). Indicates the version of the contract schema used. This allows for future evolution of the contract without breaking compatibility.
*   **meta**: `object`. Contains metadata about the deployment environment and trigger.
    *   **adapter**: `string`. Identifies the adapter used for the deployment.
    *   **engine**: `enum` (`"native"`, `"hardis"`, `"custom"`). Specifies the execution engine used during deployment.
    *   **timestamp**: `string` (datetime). Records the time the deployment was initiated.
    *   **trigger**: `string`. Describes the event that initiated the deployment.
*   **status**: `enum` (`"Succeeded"`, `"Failed"`, `"Blocked"`). Represents the overall outcome of the deployment.
*   **quality**: `object`. Provides metrics related to the quality of the deployment.
    *   **coverage**: `object`. Details code coverage information.
        *   **actual**: `number` (min: 0, max: 100). The actual code coverage achieved.
        *   **required**: `number` (min: 0, max: 100). The required code coverage threshold.
        *   **met**: `boolean`. Indicates whether the actual coverage meets the required threshold.
    *   **tests**: `object`. Summarizes test execution results.
        *   **total**: `number` (min: 0). The total number of tests executed.
        *   **passed**: `number` (min: 0). The number of tests that passed.
        *   **failed**: `number` (min: 0). The number of tests that failed.
*   **audit**: `object`. Contains information for auditing and traceability.
    *   **triggeredBy**: `string`. Identifies the user or system that triggered the deployment.
    *   **orgId**: `string`. The organization ID associated with the deployment.
    *   **repository**: `string`. The repository where the code resides.
    *   **commit**: `string`. The commit hash associated with the deployed code.

### Type Definition

The `DeploymentContract` type is derived directly from the `DeploymentContractSchema` using `z.infer`. This ensures type safety and allows you to work with the contract data in a strongly-typed manner within your TypeScript code.

```typescript
export type DeploymentContract = z.infer<typeof DeploymentContractSchema>;
```

### Usage

You can use this contract to:

1.  **Validate Deployment Data**: Ensure that incoming deployment data conforms to the defined schema.
2.  **Standardize Data Exchange**: Facilitate consistent communication of deployment information between different system components.
3.  **Improve Auditability**: Provide a structured format for tracking deployment history and identifying potential issues.

To validate data against the schema, you can use the `DeploymentContractSchema.parse()` method. For example:

```typescript
const deploymentData = {
  schemaVersion: "1.0",
  meta: { adapter: "example", engine: "native", timestamp: "2024-01-01T00:00:00Z", trigger: "manual" },
  status: "Succeeded",
  quality: { coverage: { actual: 90, required: 80, met: true }, tests: { total: 100, passed: 95, failed: 5 } },
  audit: { triggeredBy: "user123", orgId: "org456", repository: "repo789", commit: "abcdef123456" },
};

const parsedData = DeploymentContractSchema.parse(deploymentData);
// parsedData will be of type DeploymentContract
```

If the data does not conform to the schema, `DeploymentContractSchema.parse()` will throw an error.