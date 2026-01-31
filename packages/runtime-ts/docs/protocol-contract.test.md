---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/protocol/contract.test.ts
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/src/protocol/contract.test.ts
generated_at: 2026-01-31T10:09:51.727124
hash: 606b673c382755f685c8f3302c0198fe680a1046fdc81b1970846e49007a58c0
---

## Deployment Contract Schema Documentation

This document details the structure and validation rules for the Deployment Contract Schema. This schema defines a standardized format for representing deployment metadata, ensuring data integrity and consistency across systems.

### Overview

The `DeploymentContractSchema` provides a robust method for validating deployment-related information. It enforces specific data types, formats, and allowable values for key deployment attributes.  We designed this schema to ensure reliable data exchange and processing.

### Schema Structure

The schema accepts a JavaScript object with the following properties:

*   **`schemaVersion`**: (String, optional, default: "1.0") – Indicates the version of the schema used.
*   **`meta`**: (Object, required) – Contains metadata about the deployment.
    *   **`adapter`**: (String, required) – Specifies the adapter used for the deployment.
    *   **`engine`**: (String, required) – Defines the engine used for the deployment.  Acceptable values are limited to "native".
    *   **`timestamp`**: (String, required) – Represents the deployment timestamp in ISO 8601 format (e.g., "2024-01-21T18:00:00.000Z").
    *   **`trigger`**: (String, required) – Indicates the event that triggered the deployment (e.g., "push").
*   **`status`**: (String, required) – Represents the deployment status. Acceptable values are limited to "Succeeded".
*   **`quality`**: (Object, required) – Contains quality metrics for the deployment.
    *   **`coverage`**: (Object, required) – Details code coverage information.
        *   **`actual`**: (Number, required) – The actual code coverage percentage. Must be between 0 and 100, inclusive.
        *   **`required`**: (Number, required) – The required code coverage percentage.
        *   **`met`**: (Boolean, required) – Indicates whether the coverage requirement was met.
    *   **`tests`**: (Object, required) – Details test execution results.
        *   **`total`**: (Number, required) – The total number of tests executed. Must be a non-negative integer.
        *   **`passed`**: (Number, required) – The number of tests that passed.
        *   **`failed`**: (Number, required) – The number of tests that failed.
*   **`audit`**: (Object, required) – Contains audit information about the deployment.
    *   **`triggeredBy`**: (String, required) – The user or system that triggered the deployment.
    *   **`orgId`**: (String, required) – The organization ID associated with the deployment.
    *   **`repository`**: (String, required) – The repository where the code is located (e.g., "org/repo").
    *   **`commit`**: (String, required) – The commit hash associated with the deployment.

### Validation Rules

The `DeploymentContractSchema` enforces the following validation rules:

*   The `engine` value must be "native".
*   The `status` value must be "Succeeded".
*   The `timestamp` must adhere to the ISO 8601 format.
*   Test counts (`total`, `passed`, `failed`) must be non-negative integers.
*   Coverage percentages (`actual`, `required`) must be between 0 and 100, inclusive.

### Usage

You can validate a deployment contract using the `parse` method:

```typescript
const result = DeploymentContractSchema.parse(yourContractObject);
```

If the provided object conforms to the schema, the `parse` method will return the original object. If validation fails, an error will be thrown.

### Error Handling

If the input data does not conform to the schema, the `parse` method will throw an error. You should implement appropriate error handling mechanisms in your application to catch and manage these errors.