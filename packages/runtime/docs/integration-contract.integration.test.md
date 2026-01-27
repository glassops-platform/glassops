---
type: Documentation
domain: runtime
origin: packages/runtime/src/integration/contract.integration.test.ts
last_modified: 2026-01-26
generated: true
source: packages/runtime/src/integration/contract.integration.test.ts
generated_at: 2026-01-26T14:15:06.372Z
hash: 35002dc1a490132e8a7a8fa414b99ad797f08d140e1c5bdf51a682ad5b426ad0
---

## Deployment Contract Integration Test Documentation

This document details the integration tests for the Deployment Contract schema, verifying its functionality with diverse inputs and boundary conditions. These tests ensure the contract accurately represents deployment metadata and enforces data integrity.

**Overview**

The core component under test is the `DeploymentContractSchema`, which defines the structure and validation rules for deployment contracts. A deployment contract encapsulates information about a deployment, including its status, quality metrics, and audit details.  I validate that the schema correctly parses valid contracts, applies default values where appropriate, and rejects invalid data.

**Key Areas of Testing**

The integration tests are organized into two primary sections:

1.  **Schema Validation:** This suite confirms the schema’s ability to validate contract data against defined rules. Tests cover:
    *   **Complete Contract Validation:** Verifies successful parsing of a fully populated contract.
    *   **Default Schema Version:** Confirms the schema automatically assigns a default version when one is not provided.
    *   **Engine Type Support:** Validates compatibility with supported engine types ("native", "hardis", "custom").
    *   **Status Value Support:**  Ensures the schema accepts valid status values ("Succeeded", "Failed", "Blocked").
    *   **Invalid Input Rejection:**  Tests the schema’s ability to reject contracts with invalid engine types, out-of-range coverage values (outside 0-100), negative test counts, and improperly formatted timestamps.

2.  **Coverage Calculation Integration:** This section focuses on the accurate determination of coverage status (met/not met) based on actual and required coverage percentages.  It tests various scenarios to ensure the `met` field is correctly calculated.

**Data Structures**

The tests primarily interact with the `DeploymentContract` data structure, which is defined and validated by the `DeploymentContractSchema`. This structure includes the following key fields:

*   `schemaVersion`:  The version of the contract schema.
*   `meta`: Metadata about the deployment, including the adapter, engine, timestamp, and trigger.
*   `status`: The deployment status (e.g., "Succeeded", "Failed").
*   `quality`: Quality metrics, including code coverage and test results.
*   `audit`: Audit information, such as the triggering user, organization ID, repository, and commit hash.

**Usage & Configuration**

You do not need to configure anything to run these tests. They are designed to be executed as part of the standard testing process.  The tests are self-contained and do not rely on external dependencies beyond the core libraries.

**Expected Behavior**

Successful execution of these tests confirms the reliability and correctness of the `DeploymentContractSchema`.  Any test failures indicate a potential issue with the schema definition or validation logic, requiring investigation and correction.