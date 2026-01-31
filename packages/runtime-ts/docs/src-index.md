---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/index.ts
last_modified: 2026-01-29
generated: true
source: packages/runtime-ts/src/index.ts
generated_at: 2026-01-29T20:54:25.010171
hash: 54c1bd9aa48e0a943377186f288e629b527a9947f74efd9bc1bfb7ea4386f69a
---

## GlassOps Runtime Documentation

This document details the functionality and operation of the GlassOps Runtime. It provides a comprehensive overview for both technical and non-technical users.

**Overview**

The GlassOps Runtime is a tool designed to provide a governed and secure environment for interacting with Salesforce. It automates key processes including policy enforcement, environment bootstrapping, authentication, and contract generation. The runtime facilitates controlled deployments and operations, ensuring adherence to organizational standards and compliance requirements.

**Key Features**

*   **Policy Enforcement:** Evaluates pre-defined governance policies to prevent unauthorized or risky actions. Includes static code analysis and freeze window checks.
*   **Environment Bootstrapping:** Automatically installs and configures the Salesforce CLI and necessary plugins.
*   **Secure Authentication:** Manages authentication with Salesforce using JWT (JSON Web Token) based credentials. Supports skipping authentication for testing purposes.
*   **Deployment Contract Generation:** Creates a verifiable record of deployment metadata, including quality metrics and audit information.
*   **Telemetry Integration:** Provides optional integration with OpenTelemetry for monitoring and tracing.
*   **Caching:** Attempts to restore cached dependencies to accelerate execution.

**Error Handling**

The runtime defines specific error types to categorize issues:

*   **GlassOpsError:** Base class for all runtime errors.
*   **PolicyError:** Indicates a violation of a defined governance policy.
*   **BootstrapError:** Signals a failure during the environment bootstrapping process.
*   **IdentityError:** Indicates an authentication failure.
*   **ContractError:** Signals an issue during deployment contract generation.

**Input Parameters**

The runtime accepts input via GitHub Actions core inputs or environment variables. Priority is given to GitHub Actions inputs, followed by Docker environment variables (prefixed with `INPUT_`), and then environment variables prefixed with `GLASSOPS_`. 

Required inputs include:

*   `client_id`: Salesforce Client ID.
*   `jwt_key`: JWT key for authentication.  This value will be masked in logs.
*   `username`: Salesforce username.

Optional inputs include:

*   `instance_url`: Salesforce instance URL (defaults to `https://login.salesforce.com`).
*   `otel_service_name`: Name of the OpenTelemetry service (defaults to `glassops-runtime`).
*   `plugins`: Comma-separated list of Salesforce CLI plugins to install.
*   `test_results`: JSON string representing test results (total, passed, failed).
*   `coverage_percentage`: Code coverage percentage.
*   `coverage_required`: Minimum required code coverage percentage (defaults to 80).
*   `enforce_policy`: Boolean flag to enable/disable policy enforcement.
*   `skip_auth`: Boolean flag to skip authentication.

**Workflow**

1.  **Environment Validation:** Checks for required environment variables (e.g., `GITHUB_WORKSPACE`, `GITHUB_ACTOR`, `GITHUB_REPOSITORY`).
2.  **Input Validation:** Validates required inputs and performs basic format checks (e.g., JWT key format, instance URL).
3.  **Policy Evaluation:** Loads and evaluates governance policies, including static code analysis (if enabled).
4.  **Bootstrap:** Installs and configures the Salesforce CLI and plugins.
5.  **Authentication:** Authenticates with Salesforce using the provided credentials.
6.  **Contract Generation:** Creates a deployment contract containing metadata about the operation.
7.  **Output:** Sets output variables for use in subsequent workflow steps (e.g., `org_id`, `contract_path`).

**Outputs**

The runtime sets the following output variables:

*   `runtime_id`: A unique identifier for the runtime session.
*   `org_id`: The Salesforce organization ID.
*   `contract_path`: The path to the generated deployment contract file.
*   `is_locked`: Indicates whether a policy freeze window is active.
*   `glassops_ready`: Indicates successful runtime initialization.

**Security Considerations**

*   The `jwt_key` input is treated as a secret and will be masked in logs.
*   Running on forked repositories may require additional security validations.
*   The runtime includes data integrity checks and compliance validations.

**Usage**

You can integrate the GlassOps Runtime into your CI/CD pipelines by calling the `run` function. The runtime is designed to be used within a GitHub Actions workflow or similar environment.

**Future Enhancements**

*   Implementation of rate limiting and concurrency controls.
*   Improved caching strategies for Docker environments.
*   Expanded policy engine capabilities.
*   Enhanced telemetry and reporting.