---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/protocol/contract.ts
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/src/protocol/contract.ts
generated_at: 2026-01-31T09:13:49.112262
hash: 1b4bf84390a17dfe0003a547e8c270362a2a45bc6e276dae6611f4a4d7806f41
---

## Deployment Contract Specification

This document details the structure and content of the Deployment Contract, a standardized format for communicating deployment information. It ensures consistent data exchange between systems involved in software delivery.

**Purpose**

The Deployment Contract provides a comprehensive record of a deployment attempt, including its configuration, execution status, quality metrics, and audit trail. You can use this contract to track, analyze, and manage deployments across different environments and platforms.

**Schema Definition**

The contract is defined using a schema that validates the data structure and types. This schema is based on the Zod library and guarantees data integrity.

**Contract Fields**

The Deployment Contract consists of the following key fields:

*   **schemaVersion:** (String, default: "1.0") – Indicates the version of the contract schema used. This allows for future compatibility and evolution of the contract format.
*   **meta:** (Object) – Contains metadata about the deployment.
    *   **adapter:** (String) – Identifies the adapter used for the deployment.
    *   **engine:** (Enum: "native", "hardis", "custom") – Specifies the execution engine used during deployment.
    *   **timestamp:** (DateTime String) – Records the date and time when the deployment was triggered.
    *   **trigger:** (String) – Describes the event or process that initiated the deployment.
*   **status:** (Enum: "Succeeded", "Failed", "Blocked") – Represents the overall outcome of the deployment attempt.
*   **quality:** (Object) – Provides metrics related to the quality of the deployed code.
    *   **coverage:** (Object) – Details code coverage information.
        *   **actual:** (Number, 0-100) – The actual code coverage achieved.
        *   **required:** (Number, 0-100) – The minimum required code coverage.
        *   **met:** (Boolean) – Indicates whether the required coverage was met.
    *   **tests:** (Object) – Summarizes test execution results.
        *   **total:** (Number, >=0) – The total number of tests executed.
        *   **passed:** (Number, >=0) – The number of tests that passed.
        *   **failed:** (Number, >=0) – The number of tests that failed.
*   **audit:** (Object) – Contains information for auditing and traceability.
    *   **triggeredBy:** (String) – Identifies the user or system that triggered the deployment.
    *   **orgId:** (String) – The organization identifier associated with the deployment.
    *   **repository:** (String) – The repository where the deployed code resides.
    *   **commit:** (String) – The commit hash of the deployed code.

**Data Type**

The `DeploymentContract` type is derived directly from the `DeploymentContractSchema`, ensuring type safety and consistency. 

**Implementation Notes**

I have designed this contract to be extensible. Future versions may include additional fields to accommodate evolving requirements. We will maintain backward compatibility whenever possible.