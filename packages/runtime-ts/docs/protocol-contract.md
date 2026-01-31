---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/protocol/contract.ts
last_modified: 2026-01-29
generated: true
source: packages/runtime-ts/src/protocol/contract.ts
generated_at: 2026-01-29T20:56:59.126741
hash: 1b4bf84390a17dfe0003a547e8c270362a2a45bc6e276dae6611f4a4d7806f41
---

## Deployment Contract Specification

This document details the structure and content of the Deployment Contract, a standardized format for communicating deployment information. It ensures consistent data exchange between systems involved in software delivery.

**Purpose**

The Deployment Contract provides a clear, machine-readable record of a deployment attempt, including its configuration, status, quality metrics, and audit trail. You can use this contract to track deployments, enforce policies, and analyze deployment outcomes.

**Schema Definition**

The contract is defined using a schema that validates the data structure. The schema guarantees data integrity and facilitates automated processing. 

The contract consists of the following key sections:

* **schemaVersion:** (String, default: "1.0") – Indicates the version of the contract schema used. This allows for future evolution of the contract format.
* **meta:** Contains metadata about the deployment.
    * **adapter:** (String) – Identifies the adapter used for the deployment.
    * **engine:** (Enum: "native", "hardis", "custom") – Specifies the execution engine used during deployment.
    * **timestamp:** (DateTime String) – Records the date and time of the deployment attempt.
    * **trigger:** (String) – Describes the event that initiated the deployment.
* **status:** (Enum: "Succeeded", "Failed", "Blocked") – Represents the final outcome of the deployment.
* **quality:** Provides metrics related to the quality of the deployed code.
    * **coverage:** Details code coverage information.
        * **actual:** (Number, 0-100) – The actual code coverage achieved.
        * **required:** (Number, 0-100) – The minimum required code coverage.
        * **met:** (Boolean) – Indicates whether the required coverage was met.
    * **tests:** Summarizes test execution results.
        * **total:** (Number, >=0) – The total number of tests executed.
        * **passed:** (Number, >=0) – The number of tests that passed.
        * **failed:** (Number, >=0) – The number of tests that failed.
* **audit:** Contains information for auditing and traceability.
    * **triggeredBy:** (String) – Identifies the user or system that triggered the deployment.
    * **orgId:** (String) – The organization identifier associated with the deployment.
    * **repository:** (String) – The repository containing the deployed code.
    * **commit:** (String) – The commit hash of the deployed code.

**Data Type**

The `DeploymentContract` type is derived directly from the schema, ensuring type safety and consistency. I define it as follows:

```typescript
export type DeploymentContract = z.infer<typeof DeploymentContractSchema>;
```

**Usage**

We designed this contract to be easily integrated into various deployment pipelines and monitoring systems. You should validate any data against the schema before processing it to ensure data quality. 

**Future Considerations**

I plan to extend this contract with additional fields to support more complex deployment scenarios and reporting requirements. Schema versioning will be used to maintain backward compatibility.