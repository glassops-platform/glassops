---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/index.ts
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/src/index.ts
generated_at: 2026-01-31T09:11:15.767380
hash: 54c1bd9aa48e0a943377186f288e629b527a9947f74efd9bc1bfb7ea4386f69a
---

## GlassOps Runtime Documentation

This document details the functionality and operation of the GlassOps Runtime. It provides a comprehensive overview for both technical and non-technical users.

**Overview**

The GlassOps Runtime is a tool designed to provide a governed and secure environment for interacting with Salesforce. It automates key processes including policy enforcement, environment bootstrapping, identity management, and contract generation. The runtime facilitates controlled deployments and operations, ensuring adherence to organizational standards and security protocols.

**Key Features**

*   **Policy Enforcement:** Evaluates and enforces pre-defined governance policies, including freeze windows and static code analysis, to prevent unauthorized or risky operations.
*   **Environment Bootstrapping:** Automatically installs and configures the Salesforce CLI and necessary plugins, ensuring a consistent and reliable environment.
*   **Identity Management:** Securely authenticates with Salesforce using client credentials and JWT, managing access control.
*   **Contract Generation:** Creates a verifiable deployment contract outlining the details of the operation, including quality metrics and audit information.
*   **Telemetry:** Integrates with OpenTelemetry for monitoring and tracing, providing insights into runtime behavior.
*   **Caching:** Attempts to restore cached dependencies to accelerate execution, though functionality may be limited in ephemeral environments like Docker.

**Error Handling**

The runtime employs a structured error handling system with specific error types:

*   **GlassOpsError:** Base class for all runtime errors.
*   **PolicyError:** Indicates a violation of a defined governance policy.
*   **BootstrapError:** Signals a failure during the environment bootstrapping process.
*   **IdentityError:** Indicates an issue with Salesforce authentication.
*   **ContractError:** Signals a failure during contract generation.

**Input Parameters**

The runtime accepts input via GitHub Actions core inputs or environment variables. Priority is given to GitHub Actions inputs, followed by Docker environment variables (prefixed with `INPUT_`), and then `GLASSOPS_` prefixed environment variables.  Required inputs include:

*   `client_id`: Salesforce Connected App Client ID.
*   `jwt_key`: Salesforce JWT Key.  This value will be masked in logs.
*   `username`: Salesforce Username.

Optional inputs include:

*   `instance_url`: Salesforce instance URL (defaults to `https://login.salesforce.com`).
*   `otel_service_name`: OpenTelemetry service name (defaults to `glassops-runtime`).
*   `plugins`: Comma-separated list of Salesforce CLI plugins to install.
*   `test_results`: JSON string representing test results (total, passed, failed).
*   `coverage_percentage`: Code coverage percentage.
*   `coverage_required`: Minimum required code coverage percentage (defaults to 80).
*   `skip_auth`:  If set to "true", skips Salesforce authentication.
*   `enforce_policy`: If set to "true", enforces governance policies.

**Workflow**

1.  **Environment Validation:** Checks for required environment variables (e.g., `GITHUB_WORKSPACE`, `GITHUB_ACTOR`, `GITHUB_REPOSITORY`).
2.  **Input Validation:** Validates required inputs and performs basic format checks (e.g., JWT key format, instance URL).
3.  **Policy Evaluation:** Loads and evaluates governance policies, including static code analysis and freeze window checks.
4.  **Bootstrap:** Installs and configures the Salesforce CLI and plugins.
5.  **Identity Resolution:** Authenticates with Salesforce using provided credentials.
6.  **Contract Generation:** Creates a deployment contract containing metadata about the operation, quality metrics, and audit information.
7.  **Output:** Sets output variables for use in subsequent workflow steps (e.g., `org_id`, `contract_path`).

**Outputs**

The runtime provides the following outputs:

*   `runtime_id`: A unique identifier for the runtime session.
*   `org_id`: The Salesforce organization ID.
*   `contract_path`: The path to the generated deployment contract file.
*   `is_locked`: Indicates whether a freeze window is active.
*   `glassops_ready`: Indicates successful runtime initialization.

**Security Considerations**

*   The JWT key is treated as a secret and masked in logs.
*   The runtime performs basic security checks, such as validating the repository format and providing warnings for forked repositories.
*   Policy enforcement helps prevent unauthorized operations.

**Usage**

You can integrate the GlassOps Runtime into your CI/CD pipelines by calling the `run` function. The runtime is designed to be used within a GitHub Actions workflow or a similar environment.

**Future Enhancements**

*   Implementation of rate limiting and concurrency controls.
*   Expanded caching strategies for improved performance.
*   Enhanced telemetry and monitoring capabilities.
*   Support for additional policy types and governance controls.