---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/types/contract.d.ts
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/src/types/contract.d.ts
generated_at: 2026-01-31T09:19:08.733627
hash: 93fa4c433babe2f882aeffac091d68dc3f8ae11749a82816392fa518c93f9a44
---

## Deployment Contract Specification

This document details the structure of the Deployment Contract, a standardized format for reporting the results of deployments and quality checks. It provides a comprehensive record of a deployment attempt, including its status, quality metrics, policy adherence, and audit information. You can use this contract to understand the outcome of a deployment process and identify areas for improvement.

**1. Overview**

The Deployment Contract is a JSON-compatible object defining the results of a deployment operation. It is designed to be machine-readable and human-understandable, facilitating integration with various tools and systems.

**2. Contract Structure**

The contract consists of the following top-level properties:

*   **schemaVersion:** (String) Indicates the version of the contract schema used. This allows for future evolution of the contract format.
*   **meta:** (Object) Contains metadata about the deployment process itself.
    *   **adapter:** (String) Identifies the adapter used for deployment.
    *   **engine:** (String) Specifies the deployment engine used.
    *   **engineVersion:** (String) The version of the deployment engine.
    *   **timestamp:** (String) The date and time of the deployment.
    *   **trigger:** (String) The event that initiated the deployment.
*   **status:** (String) Represents the overall status of the deployment. Possible values are: "Succeeded", "Failed", "Canceled", or "Partial".
*   **deployment:** (Object) Details about the deployment action.
    *   **id:** (String) A unique identifier for the deployment.
    *   **url:** (String, Optional) A URL pointing to the deployment details.
    *   **mode:** (String) The deployment mode used ("validate", "deploy", or "quick\_deploy").
    *   **validationId:** (String, Optional) The ID of the validation run, if applicable.
    *   **metrics:** (Object) Quantitative data about the deployment.
        *   **componentsDeployed:** (Number) The number of components successfully deployed.
        *   **componentsFailed:** (Number) The number of components that failed to deploy.
        *   **testsRun:** (Number) The number of tests executed.
        *   **durationMs:** (Number) The total duration of the deployment in milliseconds.
*   **quality:** (Object) Contains information about the quality of the deployed code.
    *   **coverage:** (Object) Details about code coverage.
        *   **actual:** (Number) The actual code coverage percentage.
        *   **required:** (Number) The required code coverage percentage.
        *   **met:** (Boolean) Indicates whether the required coverage was met.
        *   **details:** (Object, Optional) Provides more granular coverage details.
            *   **orgWideCoverage:** (Number) Code coverage across the entire organization.
            *   **packageCoverage:** (Number, Optional) Code coverage for a specific package.
    *   **staticAnalysis:** (Object, Optional) Results from static analysis tools.
        *   **toolsUsed:** (Array of Strings) The tools used for static analysis.
        *   **score:** (Number) An overall health score from the static analysis.
        *   **criticalViolations:** (Number) The number of critical violations found.
        *   **warningViolations:** (Number) The number of warning violations found.
        *   **topIssues:** (Array of Objects, Optional) A list of the most significant issues identified. Each issue includes:
            *   **file:** (String) The file where the issue was found.
            *   **rule:** (String) The rule that was violated.
            *   **severity:** (String) The severity of the issue ("Critical" or "Warning").
            *   **line:** (Number) The line number where the issue was found.
            *   **message:** (String) A description of the issue.
*   **policy:** (Object) Information about policy adherence.
    *   **source:** (Object) Details about the source of the policy.
        *   **githubFloor:** (Number) Policy level from GitHub.
        *   **configFile:** (Number, Optional) Policy level from configuration files.
        *   **salesforceCmdt:** (Number, Optional) Policy level from Salesforce CMT.
    *   **effective:** (Number) The effective policy level applied.
    *   **overrides:** (Array of Objects, Optional) A list of policy overrides. Each override includes:
        *   **reason:** (String) The reason for the override.
        *   **approver:** (String) The user who approved the override.
*   **audit:** (Object) Audit information about the deployment.
    *   **triggeredBy:** (String) The user or system that triggered the deployment.
    *   **repository:** (String) The repository where the code is stored.
    *   **ref:** (String) The branch or tag that was deployed.
    *   **commit:** (String) The commit hash that was deployed.
    *   **runUrl:** (String) A URL to the deployment run details.
*   **errors:** (Array of Objects, Optional) A list of errors that occurred during the deployment. Each error includes:
    *   **code:** (String) An error code.
    *   **message:** (String) A descriptive error message.
    *   **severity:** (String) The severity of the error ("Critical" or "Warning").
    *   **component:** (String, Optional) The component associated with the error.
*   **extensions:** (Object, Optional) A flexible container for adding custom data to the contract.

**3. Usage**

I generate this contract after each deployment attempt. You can parse and analyze this contract to gain insights into the deployment process, identify potential issues, and track quality metrics over time. We intend for this contract to be a central source of truth for deployment information.