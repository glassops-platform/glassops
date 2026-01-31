---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/protocol/contract.test.ts
last_modified: 2026-01-29
generated: true
source: packages/runtime-ts/src/protocol/contract.test.ts
generated_at: 2026-01-29T20:56:39.278177
hash: 606b673c382755f685c8f3302c0198fe680a1046fdc81b1970846e49007a58c0
---

## Deployment Contract Schema Documentation

This document details the structure and validation rules for the Deployment Contract Schema. This schema defines a standardized format for representing deployment information, ensuring data consistency and reliability.

**Purpose**

The Deployment Contract Schema provides a way to formally define the outcome of a deployment process. It captures metadata about the deployment, its status, quality metrics, and audit information.  This allows for consistent reporting and automated decision-making based on deployment results.

**Schema Structure**

The schema consists of the following top-level properties:

*   **schemaVersion:** (String, optional, default: "1.0") – Indicates the version of the schema used.
*   **meta:** (Object) – Contains metadata about the deployment environment and trigger.
    *   **adapter:** (String) – Specifies the adapter used for the deployment.
    *   **engine:** (String) – Specifies the engine used for the deployment.  Valid values are limited to "native".
    *   **timestamp:** (String) –  A timestamp representing when the deployment occurred, formatted as an ISO 8601 string (e.g., "2024-01-21T18:00:00.000Z").
    *   **trigger:** (String) – Indicates the event that triggered the deployment (e.g., "push").
*   **status:** (String) – Represents the overall status of the deployment. Valid values are limited to "Succeeded".
*   **quality:** (Object) – Contains quality metrics related to the deployment.
    *   **coverage:** (Object) – Details code coverage information.
        *   **actual:** (Number) – The actual code coverage percentage. Must be between 0 and 100, inclusive.
        *   **required:** (Number) – The required code coverage percentage.
        *   **met:** (Boolean) – Indicates whether the actual coverage meets the required coverage.
    *   **tests:** (Object) – Details test execution results.
        *   **total:** (Number) – The total number of tests executed. Must be a non-negative integer.
        *   **passed:** (Number) – The number of tests that passed.
        *   **failed:** (Number) – The number of tests that failed.
*   **audit:** (Object) – Contains audit information about the deployment.
    *   **triggeredBy:** (String) – The user or system that triggered the deployment.
    *   **orgId:** (String) – The organization ID associated with the deployment.
    *   **repository:** (String) – The repository where the code was deployed from.
    *   **commit:** (String) – The commit hash associated with the deployment.

**Validation Rules**

We enforce strict validation to ensure data integrity. The schema will reject contracts that:

*   Use an invalid value for the `engine` field (only "native" is permitted).
*   Use an invalid value for the `status` field (only "Succeeded" is permitted).
*   Have an incorrectly formatted `timestamp` (must be a valid ISO 8601 string).
*   Contain negative values for `tests.total`.
*   Have `coverage.actual` values outside the range of 0-100.

**Usage**

You can use the `DeploymentContractSchema.parse()` method to validate a deployment contract object against this schema. If the object is valid, the method will return the parsed object. If the object is invalid, the method will throw an error.