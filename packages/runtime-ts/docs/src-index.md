---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/index.ts
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/src/index.ts
generated_at: 2026-01-31T10:07:23.727444
hash: 54c1bd9aa48e0a943377186f288e629b527a9947f74efd9bc1bfb7ea4386f69a
---

## GlassOps Runtime Documentation

This document details the functionality and operation of the GlassOps Runtime. It provides a comprehensive overview for both technical and non-technical users.

**Overview**

The GlassOps Runtime is a tool designed to provide a governed and secure environment for interacting with Salesforce. It automates key processes including policy enforcement, environment bootstrapping, identity management, and contract generation. The runtime facilitates controlled deployments and operations, ensuring compliance and reducing risk.

**Key Features**

*   **Policy Enforcement:** Evaluates pre-defined governance policies to prevent unauthorized or risky actions. Includes static code analysis and freeze window checks.
*   **Environment Bootstrapping:** Automatically installs and configures the Salesforce CLI and necessary plugins.
*   **Identity Management:** Securely authenticates with Salesforce using JWT authentication. Supports skipping authentication for testing purposes.
*   **Contract Generation:** Creates a deployment contract outlining the details of the operation, including quality metrics and audit information.
*   **Telemetry:** Integrates with OpenTelemetry for monitoring and tracing (optional configuration required).
*   **Caching:** Attempts to restore cached dependencies to accelerate execution (behavior may vary in Docker environments).

**Error Handling**

The runtime defines specific error types for improved clarity and handling:

*   `GlassOpsError`: Base class for all runtime errors.
*   `PolicyError`: Indicates a violation of a governance policy.
*   `BootstrapError`: Indicates a failure during the environment bootstrapping process.
*   `IdentityError`: Indicates an authentication failure.
*   `ContractError`: Indicates a failure during contract generation.

**Input Parameters**

The runtime accepts input via GitHub Actions core inputs or environment variables. The following inputs are required:

*   `client_id`: The Salesforce Connected App client ID.
*   `jwt_key`: The private key for JWT authentication (must include BEGIN and END markers). This value is automatically masked in logs.
*   `username`: The Salesforce username.

The following inputs are optional:

*   `instance_url`: The Salesforce instance URL (defaults to `https://login.salesforce.com`).
*   `otel_service_name`: The name of the OpenTelemetry service (defaults to `glassops-runtime`).
*   `plugins`: A comma-separated list of Salesforce CLI plugins to install.
*   `test_results`: A JSON string representing test results (format: `{ "total": number, "passed": number, "failed": number }`).
*   `coverage_percentage`: The actual code coverage percentage.
*   `coverage_required`: The required code coverage percentage (defaults to 80).
*   `enforce_policy`: A boolean value indicating whether to enforce governance policies (defaults to `false`).
*   `skip_auth`: A boolean value indicating whether to skip Salesforce authentication.

**Workflow**

1.  **Environment Validation:** Checks for required environment variables (e.g., `GITHUB_WORKSPACE`, `GITHUB_ACTOR`, `GITHUB_REPOSITORY`).
2.  **Input Validation:** Validates required inputs and performs basic format checks (e.g., JWT key format, instance URL).
3.  **Policy Evaluation:** Loads and evaluates governance policies, including static code analysis (if enabled).
4.  **Bootstrap:** Installs the Salesforce CLI and specified plugins.
5.  **Identity Resolution:** Authenticates with Salesforce using the provided credentials.
6.  **Contract Generation:** Creates a deployment contract containing metadata about the operation.
7.  **Output:** Sets output variables for use in subsequent workflow steps (e.g., `org_id`, `contract_path`).

**Outputs**

The runtime sets the following output variables:

*   `runtime_id`: A unique identifier for the runtime session.
*   `org_id`: The Salesforce organization ID.
*   `contract_path`: The path to the generated deployment contract file.
*   `is_locked`: Indicates whether a freeze window is active.
*   `glassops_ready`: Indicates that the runtime is ready for execution.

**Important Considerations**

*   **Docker Environments:** Caching behavior may be limited in Docker environments due to their ephemeral nature.
*   **Security:** Protect the `jwt_key` input. It is automatically masked in logs, but should not be exposed in other ways.
*   **Forked Repositories:** Running in forked repositories may require additional security validations.
*   **Execution Timeout:** A maximum execution time of 30 minutes is enforced.

**Running the Runtime**

You can run the runtime directly using Node.js:

```bash
node index.js
```

However, it is primarily intended to be used as part of a GitHub Actions workflow.