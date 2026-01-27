---
type: Documentation
domain: runtime
origin: packages/runtime/src/index.ts
last_modified: 2026-01-26
generated: true
source: packages/runtime/src/index.ts
generated_at: 2026-01-26T14:07:56.057Z
hash: a32b404252a192b420cc448b86d18823df3271776fe067058010db51a148e646
---

## GlassOps Runtime – Architecture Overview

This document details the architecture and functionality of the GlassOps Runtime, a tool designed to provide a governed and secure execution environment for Salesforce development operations. It outlines the core components, operational flow, and key considerations for its use.

**1. Introduction**

The GlassOps Runtime establishes a framework for automating Salesforce deployments and changes while enforcing organizational governance policies. It focuses on security, compliance, and operational integrity throughout the entire process.

**2. Core Components**

*   **Policy Engine:** Responsible for evaluating and enforcing pre-defined governance rules. This includes freeze windows, code quality standards, and access controls.
*   **Runtime Environment:** Manages the Salesforce CLI (SFDX) installation and plugin dependencies, ensuring a consistent and reliable execution environment.
*   **Identity Resolver:** Handles authentication with Salesforce organizations, securely managing credentials and establishing trust.
*   **Analyzer:** Performs static code analysis to identify potential vulnerabilities and ensure adherence to coding best practices.
*   **Deployment Contract:** A structured document that captures the state of the deployment, including quality metrics, audit information, and compliance status.
*   **Error Handling:** A custom error hierarchy (GlassOpsError, PolicyError, BootstrapError, IdentityError, ContractError) provides specific categorization for improved debugging and issue resolution.

**3. Operational Flow**

The runtime operates through a series of phases:

1.  **Environment Validation:** Checks for the presence of required environment variables (e.g., `GITHUB_WORKSPACE`, `GITHUB_ACTOR`, `GITHUB_REPOSITORY`) and input parameters (e.g., `client_id`, `jwt_key`, `username`).
2.  **Policy Evaluation:** Loads and evaluates governance policies, potentially halting execution if violations are detected. Static code analysis is performed if enabled.
3.  **Bootstrap:** Installs or updates the Salesforce CLI and any required plugins. Caching mechanisms are used to optimize performance.
4.  **Identity Resolution:** Authenticates with the target Salesforce organization using provided credentials. Authentication can be skipped for testing purposes.
5.  **Contract Generation:** Creates a Deployment Contract containing metadata about the deployment, quality metrics (code coverage, test results), and audit information.
6.  **Output:** Sets output variables for downstream processes, including the generated contract path and the Salesforce organization ID.

**4. Key Features & Considerations**

*   **Security:** Emphasizes secure credential management, input validation, and compliance checks.
*   **Governance:** Enforces organizational policies through the Policy Engine, preventing unauthorized or risky deployments.
*   **Caching:** Improves performance by caching dependencies and reducing redundant downloads.
*   **Extensibility:** Designed to be extensible through plugins and configurable policies.
*   **Error Handling:** Provides detailed error messages and categorization for easier troubleshooting.
*   **Rate Limiting:** Includes a placeholder for future implementation of rate limiting and concurrency controls.
*   **Timeout Mechanism:** Implements a safety timeout to prevent indefinite execution.
*   **Forked Repository Handling:** Provides a warning when running in forked repositories, recommending additional security validations.

**5. Configuration**

You can configure the runtime using input parameters:

*   `client_id`: Salesforce Connected App Client ID.
*   `jwt_key`: Salesforce JWT Key.
*   `username`: Salesforce Username.
*   `instance_url`: Salesforce instance URL (defaults to `https://login.salesforce.com`).
*   `enforce_policy`: Enables or disables policy enforcement (defaults to enabled).
*   `plugins`: Comma-separated list of Salesforce CLI plugins to install.
*   `test_results`: JSON string containing test results data.
*   `coverage_percentage`: Code coverage percentage.
*   `coverage_required`: Required code coverage percentage.
*   `skip_auth`: Skips Salesforce authentication (for testing only).

**6. Outputs**

The runtime provides the following outputs:

*   `runtime_id`: A unique identifier for the execution session.
*   `org_id`: The Salesforce organization ID.
*   `contract_path`: The path to the generated Deployment Contract.
*   `is_locked`: Indicates whether a policy freeze window is active.
*   `glassops_ready`: Indicates successful runtime initialization.