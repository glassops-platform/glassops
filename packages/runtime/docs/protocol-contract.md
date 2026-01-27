---
type: Documentation
domain: runtime
origin: packages/runtime/src/protocol/contract.ts
last_modified: 2026-01-26
generated: true
source: packages/runtime/src/protocol/contract.ts
generated_at: 2026-01-26T14:18:29.753Z
hash: 8930f157afb3a0ec0364c5c0c3cbcf6057b699d0bf10b9125ba1ff9e1de6d48c
---

## Deployment Contract Specification

This document details the structure and content of the Deployment Contract, a standardized format for representing deployment metadata. It provides a consistent way to communicate the outcome and quality of deployments across different systems and environments.

**Purpose**

The Deployment Contract serves as a single source of truth for deployment information. It enables automated validation, reporting, and auditing of deployments.

**Schema Overview**

The contract is defined using a schema that ensures data integrity and consistency. The schema consists of the following key sections:

**1. Schema Version**

*   `schemaVersion`: (String, default: "1.0") – Indicates the version of the contract schema used. This allows for future evolution of the contract format while maintaining backward compatibility.

**2. Metadata**

*   `adapter`: (String) – Identifies the adapter used for the deployment.
*   `engine`: (Enum: "native", "hardis", "custom") – Specifies the execution engine used during deployment.
*   `timestamp`: (DateTime String) – Records the date and time of the deployment.
*   `trigger`: (String) – Describes the event or process that initiated the deployment.

**3. Status**

*   `status`: (Enum: "Succeeded", "Failed", "Blocked") – Represents the overall outcome of the deployment.

**4. Quality**

This section provides metrics related to the quality of the deployed code.

*   `coverage`:
    *   `actual`: (Number, 0-100) – The actual code coverage achieved during testing.
    *   `required`: (Number, 0-100) – The minimum code coverage required for the deployment.
    *   `met`: (Boolean) – Indicates whether the actual coverage meets the required coverage.
*   `tests`:
    *   `total`: (Number, >=0) – The total number of tests executed.
    *   `passed`: (Number, >=0) – The number of tests that passed.
    *   `failed`: (Number, >=0) – The number of tests that failed.

**5. Audit Information**

*   `triggeredBy`: (String) – Identifies the user or system that initiated the deployment.
*   `orgId`: (String) – The organization associated with the deployment.
*   `repository`: (String) – The repository containing the deployed code.
*   `commit`: (String) – The specific commit hash deployed.

**Data Type**

The `DeploymentContract` type is derived directly from the `DeploymentContractSchema`, ensuring type safety and consistency. 

**Usage**

I designed this contract to be used in automated deployment pipelines. You can validate deployment data against this schema to ensure it conforms to the expected format. I provide tools to generate and parse contracts programmatically. 

**Future Considerations**

We plan to extend this contract with additional metadata, such as performance metrics and security scan results, to provide a more comprehensive view of deployment quality.