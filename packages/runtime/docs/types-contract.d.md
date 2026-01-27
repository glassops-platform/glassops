---
type: Documentation
domain: runtime
origin: packages/runtime/src/types/contract.d.ts
last_modified: 2026-01-26
generated: true
source: packages/runtime/src/types/contract.d.ts
generated_at: 2026-01-26T14:17:45.287Z
hash: bb0bd6e93d2b5187e0c3af6b299beb7ca645c62bf63eeb52deedefe1d7082034
---

## Deployment Contract Specification

This document details the structure of the Deployment Contract, a central data object representing the outcome of a deployment process. It provides a comprehensive record of the deployment’s execution, quality, policy compliance, and audit information. This contract is designed for both automated systems and human review.

**1. Overview**

The Deployment Contract encapsulates all relevant information pertaining to a deployment, from initial trigger to final status. It facilitates transparency, accountability, and informed decision-making throughout the software delivery lifecycle.

**2. Contract Structure**

The contract is organized into the following key sections:

**2.1. meta**

Contains metadata about the deployment process itself.

*   `adapter`: The name of the adapter used for deployment.
*   `engine`: The deployment engine used.
*   `engineVersion`: The version of the deployment engine.
*   `timestamp`: The date and time of the deployment.
*   `trigger`: The event that initiated the deployment.

**2.2. status**

Indicates the overall outcome of the deployment. Possible values:

*   `Succeeded`: The deployment completed successfully.
*   `Failed`: The deployment failed.
*   `Canceled`: The deployment was canceled.
*   `Partial`: The deployment completed with some failures.

**2.3. deployment**

Details specific to the deployment execution.

*   `id`: A unique identifier for the deployment.
*   `url`: (Optional) A URL pointing to the deployment details.
*   `mode`: The deployment mode used: `validate`, `deploy`, or `quick_deploy`.
*   `validationId`: (Optional) The ID of the validation run associated with this deployment.
*   `metrics`: Quantitative data about the deployment.
    *   `componentsDeployed`: The number of components successfully deployed.
    *   `componentsFailed`: The number of components that failed to deploy.
    *   `testsRun`: The number of tests executed.
    *   `durationMs`: The total duration of the deployment in milliseconds.

**2.4. quality**

Assesses the quality of the deployed code.

*   `coverage`: Code coverage metrics.
    *   `actual`: The actual code coverage achieved.
    *   `required`: The required code coverage threshold.
    *   `met`: A boolean indicating whether the required coverage was met.
    *   `details`: (Optional) Granular coverage details.
        *   `orgWideCoverage`: Coverage across the entire organization.
        *   `packageCoverage`: Coverage for a specific package.
*   `staticAnalysis`: Results from static code analysis tools.
    *   `toolsUsed`: An array of tools used for static analysis.
    *   `score`: An overall health score from the analysis.
    *   `criticalViolations`: The number of critical violations found (e.g., security risks).
    *   `warningViolations`: The number of warning violations found (e.g., potential technical debt).
    *   `topIssues`: (Optional) An array of the most significant issues identified, including file, rule, severity, line number, and message.

**2.5. policy**

Information related to policy enforcement.

*   `source`: Details about the source of the policy rules.
    *   `githubFloor`: Policy configuration from GitHub.
    *   `configFile`: Policy configuration from a configuration file.
    *   `salesforceCmdt`: Policy configuration from Salesforce CMT.
*   `effective`: The effective policy version applied during deployment.
*   `overrides`: (Optional) An array of policy overrides, including the reason and approver.

**2.6. audit**

Provides audit trail information.

*   `triggeredBy`: The user or system that triggered the deployment.
*   `repository`: The repository containing the code.
*   `ref`: The branch or tag used for the deployment.
*   `commit`: The commit hash of the deployed code.
*   `runUrl`: A URL to the deployment run details.

**2.7. errors**

(Optional) An array of errors encountered during the deployment.

*   `code`: An error code.
*   `message`: A descriptive error message.
*   `severity`: The severity of the error (`Critical` or `Warning`).
*   `component`: (Optional) The component associated with the error.

**2.8. extensions**

(Optional) A flexible section for adding custom data. This allows for extending the contract with additional information specific to your environment. You can add any key-value pair here.

**3. Usage**

I generate this contract after each deployment attempt. You can use this data for reporting, analysis, and automated remediation. The contract’s standardized format enables integration with various tools and systems.