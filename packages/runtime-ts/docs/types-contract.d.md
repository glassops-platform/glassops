---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/types/contract.d.ts
last_modified: 2026-01-29
generated: true
source: packages/runtime-ts/src/types/contract.d.ts
generated_at: 2026-01-29T21:02:50.896652
hash: 93fa4c433babe2f882aeffac091d68dc3f8ae11749a82816392fa518c93f9a44
---

## Deployment Contract Specification

This document details the structure of the Deployment Contract, a standardized format for reporting the results of deployments and quality checks. It provides a comprehensive record of the process, enabling clear communication and informed decision-making.

**Purpose**

The Deployment Contract serves as a single source of truth regarding a deployment’s execution, quality, policy compliance, and audit trail. You can use this contract to understand the outcome of a deployment, identify potential issues, and track improvements over time.

**Structure**

The contract is a JSON object with the following key sections:

**1. `schemaVersion`**:  A string indicating the version of the contract schema used. This allows for future evolution of the contract format.

**2. `meta`**: Metadata about the deployment process itself.
    *   `adapter`: The name of the adapter used for deployment.
    *   `engine`: The name of the deployment engine.
    *   `engineVersion`: The version of the deployment engine.
    *   `timestamp`: The date and time of the deployment.
    *   `trigger`: The event that initiated the deployment.

**3. `status`**: The overall status of the deployment. Possible values are: `Succeeded`, `Failed`, `Canceled`, or `Partial`.

**4. `deployment`**: Details specific to the deployment action.
    *   `id`: A unique identifier for the deployment.
    *   `url`: (Optional) A URL pointing to the deployment details.
    *   `mode`: The deployment mode: `validate`, `deploy`, or `quick_deploy`.
    *   `validationId`: (Optional) An identifier for a preceding validation run.
    *   `metrics`: Quantitative data about the deployment.
        *   `componentsDeployed`: The number of components successfully deployed.
        *   `componentsFailed`: The number of components that failed to deploy.
        *   `testsRun`: The number of tests executed.
        *   `durationMs`: The total duration of the deployment in milliseconds.

**5. `quality`**:  Information about the quality of the deployed code.
    *   `coverage`: Code coverage metrics.
        *   `actual`: The actual code coverage achieved.
        *   `required`: The required code coverage threshold.
        *   `met`: A boolean indicating whether the required coverage was met.
        *   `details`: (Optional) Further breakdown of coverage.
            *   `orgWideCoverage`: Coverage across the entire organization.
            *   `packageCoverage`: Coverage for a specific package.
    *   `staticAnalysis`: Results from static code analysis tools. (Optional)
        *   `toolsUsed`: An array of tools used for static analysis.
        *   `score`: An overall health score.
        *   `criticalViolations`: The number of critical violations found.
        *   `warningViolations`: The number of warning violations found.
        *   `topIssues`: (Optional) An array of the most significant issues identified, including file, rule, severity, line number, and message.

**6. `policy`**:  Information about policy compliance.
    *   `source`: Where policy definitions were sourced from.
        *   `githubFloor`: Policy version from the GitHub floor.
        *   `configFile`: Policy version from a configuration file.
        *   `salesforceCmdt`: Policy version from Salesforce CMT.
    *   `effective`: The effective policy version used during the deployment.
    *   `overrides`: (Optional) An array of policy overrides, including the reason and approver.

**7. `audit`**:  Audit trail information.
    *   `triggeredBy`: The user or system that triggered the deployment.
    *   `repository`: The repository where the code resides.
    *   `ref`: The branch or tag that was deployed.
    *   `commit`: The commit hash of the deployed code.
    *   `runUrl`: A URL to the execution run details.

**8. `errors`**: (Optional) An array of errors encountered during the deployment.
    *   `code`: An error code.
    *   `message`: A descriptive error message.
    *   `severity`: The severity of the error: `Critical` or `Warning`.
    *   `component`: (Optional) The component associated with the error.

**9. `extensions`**: (Optional) A flexible section for adding custom data. This allows for extending the contract with additional information specific to your environment. It is a key-value store where keys are strings and values can be of any type.

I maintain this contract to ensure consistency and clarity in deployment reporting. We aim to provide a robust and extensible framework for managing deployments and improving software quality.