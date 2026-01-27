---
type: Documentation
domain: runtime
origin: packages/runtime/src/protocol/contract.test.ts
last_modified: 2026-01-26
generated: true
source: packages/runtime/src/protocol/contract.test.ts
generated_at: 2026-01-26T14:18:08.241Z
hash: 3e77ac798207584d0d1a1af15ee460c7adf6afd5ace7293b16347eb50bbff883
---

## Deployment Contract Schema Documentation

This document describes the `DeploymentContractSchema`, which defines the structure and validation rules for deployment contracts. These contracts represent the outcome of a deployment process and contain metadata, status information, quality metrics, and audit details.

**Purpose**

The schema ensures that deployment contracts adhere to a consistent format, enabling reliable processing and analysis of deployment results. It validates the data to maintain data integrity and prevent errors.

**Schema Structure**

A valid deployment contract must conform to the following structure:

*   **`schemaVersion`** (string, optional):  The version of the schema used. If not provided, it defaults to "1.0".
*   **`meta`** (object): Contains metadata about the deployment.
    *   **`adapter`** (string): The adapter used for the deployment.
    *   **`engine`** (string): The engine used for the deployment.  Acceptable values are limited to "native".
    *   **`timestamp`** (string): The timestamp of the deployment event in ISO 8601 format (e.g., "2024-01-21T18:00:00.000Z").
    *   **`trigger`** (string): The event that triggered the deployment (e.g., "push").
*   **`status`** (string): The overall status of the deployment. Acceptable values are limited to "Succeeded".
*   **`quality`** (object): Contains quality metrics for the deployment.
    *   **`coverage`** (object): Code coverage information.
        *   **`actual`** (number): The actual code coverage percentage. Must be between 0 and 100.
        *   **`required`** (number): The required code coverage percentage.
        *   **`met`** (boolean): Indicates whether the required coverage was met.
    *   **`tests`** (object): Test execution results.
        *   **`total`** (number): The total number of tests executed. Must be a non-negative integer.
        *   **`passed`** (number): The number of tests that passed.
        *   **`failed`** (number): The number of tests that failed.
*   **`audit`** (object): Contains audit information about the deployment.
    *   **`triggeredBy`** (string): The user or system that triggered the deployment.
    *   **`orgId`** (string): The organization ID associated with the deployment.
    *   **`repository`** (string): The repository where the code was deployed from (e.g., "org/repo").
    *   **`commit`** (string): The commit hash associated with the deployment.

**Validation Rules**

I enforce the following validation rules:

*   The `engine` field must be set to "native".
*   The `status` field must be set to "Succeeded".
*   The `timestamp` field must be a valid ISO 8601 formatted string.
*   Test counts (`total`, `passed`, `failed`) must be non-negative integers.
*   Code coverage percentages (`actual`, `required`) must be between 0 and 100.

**Usage**

You can parse a deployment contract using `DeploymentContractSchema.parse(contractData)`. This method will validate the `contractData` against the schema and return the parsed contract if it is valid. If the contract is invalid, an error will be thrown.