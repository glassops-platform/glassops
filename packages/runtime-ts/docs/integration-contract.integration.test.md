---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/integration/contract.integration.test.ts
last_modified: 2026-01-30
generated: true
source: packages/runtime-ts/src/integration/contract.integration.test.ts
generated_at: 2026-01-30T22:52:00.822929
hash: 6b2a352f66d3d0aefc00ef36dddba62abd550228de051275c4a090ad1519375b
---

## Deployment Contract Integration Test Documentation

This document details the integration tests for the Deployment Contract system. These tests ensure the contract schema functions as expected with diverse inputs and potential edge cases.

**Overview**

The Deployment Contract represents a standardized format for describing deployment results.  It includes metadata about the deployment, its status, quality metrics, and audit information.  The schema validation ensures data conforms to the defined structure and constraints.

**Functionality Tested**

The integration tests cover the following key areas:

*   **Schema Validation:** Verifies that valid contract data can be parsed and that invalid data is rejected with appropriate errors. This includes checks for:
    *   Correct schema version.
    *   Supported engine types (native, hardis, custom).
    *   Supported status values (Succeeded, Failed, Blocked).
    *   Valid data types and ranges for numerical values (coverage percentages, test counts).
    *   Correct timestamp formatting.
*   **Coverage Calculation:** Confirms the accurate determination of whether coverage requirements are met based on actual and required coverage percentages.

**Data Structure**

The core data structure being tested is the `DeploymentContract`, defined by the `DeploymentContractSchema`.  Key properties include:

*   `schemaVersion`:  Indicates the version of the contract schema used. Defaults to "1.0" if not provided.
*   `meta`: Contains metadata about the deployment, including:
    *   `adapter`: The adapter used for deployment.
    *   `engine`: The deployment engine.
    *   `timestamp`: The time of deployment (ISO 8601 format).
    *   `trigger`: The event that initiated the deployment.
*   `status`: The overall status of the deployment (Succeeded, Failed, Blocked).
*   `quality`:  Contains quality metrics, including:
    *   `coverage`:  Details about code coverage, including `actual`, `required`, and `met` (boolean indicating if the requirement is met).
    *   `tests`:  Details about test execution, including `total`, `passed`, and `failed` counts.
*   `audit`:  Contains audit information, including:
    *   `triggeredBy`: The user or system that triggered the deployment.
    *   `orgId`: The organization ID.
    *   `repository`: The repository where the code is located.
    *   `commit`: The commit hash.

**Usage Notes**

These tests are designed to automatically validate the contract schema.  If you modify the schema, you should run these tests to ensure compatibility and prevent regressions.  

You can extend these tests by adding new scenarios to cover additional edge cases or validation rules.