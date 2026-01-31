---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/types/contract.d.ts
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/src/types/contract.d.ts
generated_at: 2026-01-31T10:15:54.968292
hash: 93fa4c433babe2f882aeffac091d68dc3f8ae11749a82816392fa518c93f9a44
---

## Deployment Contract Specification

This document details the structure of the Deployment Contract, a standardized format for communicating deployment results. It provides a comprehensive record of a deployment attempt, encompassing metadata, status, quality assessments, policy adherence, and audit information. This contract facilitates consistent interpretation of deployment outcomes across different tools and platforms.

### Overview

The Deployment Contract is a JSON object conforming to the `DeploymentContract` interface. It serves as a single source of truth for all aspects of a deployment, enabling automated analysis, reporting, and governance. You can use this contract to understand the success, quality, and compliance of your deployments.

### Contract Structure

The `DeploymentContract` interface is defined as follows:

```typescript
export interface DeploymentContract {
    schemaVersion: string;

    meta: {
        adapter: string;
        engine: string;
        engineVersion: string;
        timestamp: string;
        trigger: string;
    };

    status: "Succeeded" | "Failed" | "Canceled" | "Partial";

    deployment: {
        id: string;
        url?: string;
        mode: "validate" | "deploy" | "quick_deploy";
        validationId?: string;
        metrics: {
            componentsDeployed: number;
            componentsFailed: number;
            testsRun: number;
            durationMs: number;
        };
    };

    quality: {
        coverage: {
            actual: number;
            required: number;
            met: boolean;
            details?: {
                orgWideCoverage: number;
                packageCoverage?: number;
            };
        };

        staticAnalysis?: {
            toolsUsed: string[];
            score: number;
            criticalViolations: number;
            warningViolations: number;
            topIssues?: Array<{
                file: string;
                rule: string;
                severity: "Critical" | "Warning";
                line: number;
                message: string;
            }>;
        };
    };

    policy: {
        source: {
            githubFloor: number;
            configFile?: number;
            salesforceCmdt?: number;
        };
        effective: number;
        overrides?: Array<{
            reason: string;
            approver: string;
        }>;
    };

    audit: {
        triggeredBy: string;
        repository: string;
        ref: string;
        commit: string;
        runUrl: string;
    };

    errors?: Array<{
        code: string;
        message: string;
        severity: "Critical" | "Warning";
        component?: string;
    }>;

    extensions?: {
        [key: string]: any;
    };
}
```

### Key Fields

*   **`schemaVersion`**:  A string indicating the version of the contract schema used.
*   **`meta`**: Contains metadata about the deployment process.
    *   `adapter`: The name of the adapter used for deployment.
    *   `engine`: The name of the deployment engine.
    *   `engineVersion`: The version of the deployment engine.
    *   `timestamp`: The time the deployment was initiated.
    *   `trigger`: The event that initiated the deployment.
*   **`status`**:  The overall status of the deployment: `"Succeeded"`, `"Failed"`, `"Canceled"`, or `"Partial"`.
*   **`deployment`**: Details specific to the deployment itself.
    *   `id`: A unique identifier for the deployment.
    *   `url`:  A URL pointing to the deployment details (optional).
    *   `mode`: The deployment mode: `"validate"`, `"deploy"`, or `"quick_deploy"`.
    *   `validationId`:  The ID of the validation run associated with this deployment (optional).
    *   `metrics`: Quantitative data about the deployment.
        *   `componentsDeployed`: The number of components successfully deployed.
        *   `componentsFailed`: The number of components that failed to deploy.
        *   `testsRun`: The number of tests executed.
        *   `durationMs`: The total duration of the deployment in milliseconds.
*   **`quality`**:  Assessment of the deployment quality.
    *   `coverage`: Code coverage metrics.
        *   `actual`: The actual code coverage achieved.
        *   `required`: The required code coverage threshold.
        *   `met`:  A boolean indicating whether the coverage requirement was met.
        *   `details`:  Optional details about coverage breakdown.
            *   `orgWideCoverage`: Coverage across the entire organization.
            *   `packageCoverage`: Coverage for a specific package.
    *   `staticAnalysis`: Results from static analysis tools.
        *   `toolsUsed`: An array of tools used for static analysis.
        *   `score`: An overall health score.
        *   `criticalViolations`: The number of critical violations found.
        *   `warningViolations`: The number of warning violations found.
        *   `topIssues`: An array of the most significant issues identified (optional).
*   **`policy`**: Information related to policy enforcement.
    *   `source`:  The source of the policy rules.
        *   `githubFloor`: Policy level enforced by GitHub.
        *   `configFile`: Policy level enforced by a configuration file.
        *   `salesforceCmdt`: Policy level enforced by Salesforce CMT.
    *   `effective`: The effective policy level applied.
    *   `overrides`:  A list of policy overrides, if any.
        *   `reason`: The reason for the override.
        *   `approver`: The user who approved the override.
*   **`audit`**:  Audit trail information.
    *   `triggeredBy`: The user or system that triggered the deployment.
    *   `repository`: The repository associated with the deployment.
    *   `ref`: The branch or tag used for the deployment.
    *   `commit`: The commit hash used for the deployment.
    *   `runUrl`: A URL to the deployment run details.
*   **`errors`**: An array of errors encountered during the deployment (optional).
    *   `code`: An error code.
    *   `message`: A descriptive error message.
    *   `severity`: The severity of the error: `"Critical"` or `"Warning"`.
    *   `component`: The component associated with the error (optional).
*   **`extensions`**:  A flexible field for adding custom data (optional).  Allows for extensibility without modifying the core contract.

### Usage

I generate this contract after each deployment attempt. You can then parse and analyze this contract to gain insights into your deployments. We recommend integrating this contract into your CI/CD pipelines and reporting systems.