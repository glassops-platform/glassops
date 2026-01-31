---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/action.yml
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/action.yml
generated_at: 2026-01-31T11:06:43.771204
hash: 51105da1b20fc22775e94d2a5b8f0b93a942c069a628f45b47f33c6d494426a3
---

# GlassOps Runtime Action Configuration

This document details the configuration for the GlassOps Runtime action. This action bootstraps a governed Salesforce execution environment following the GlassOps Protocol.

## Purpose

The primary purpose of this action is to set up a secure and controlled environment for interacting with a Salesforce organization. It handles authentication, policy enforcement, and telemetry integration.

## Structure

The configuration is structured using YAML and defines the action's metadata, inputs, outputs, and execution details.

### `name`

*   **Type:** String
*   **Description:** The name of the action.
*   **Value:** `GlassOps Runtime`

### `description`

*   **Type:** String
*   **Description:** A brief description of the action's purpose.
*   **Value:** `Bootstraps a governed Salesforce execution environment adhering to the GlassOps Protocol.`

### `author`

*   **Type:** String
*   **Description:** The author of the action.
*   **Value:** `Ryan Bumstead`

### `branding`

*   **Type:** Object
*   **Description:** Defines the visual branding for the action.
    *   `icon`: The icon to display.
    *   `color`: The color to use for the icon and other elements.
*   **Value:**
    ```yaml
    icon: 'shield'
    color: 'blue'
    ```

### `inputs`

This section defines the input parameters that You can provide to the action.

*   `jwt_key`:
    *   **Type:** String
    *   **Description:** Private Key for JWT flow.
    *   **Required:** True
*   `client_id`:
    *   **Type:** String
    *   **Description:** External Client App Consumer Key.
    *   **Required:** True
*   `username`:
    *   **Type:** String
    *   **Description:** Target Org Username.
    *   **Required:** True
*   `instance_url`:
    *   **Type:** String
    *   **Description:** Salesforce login URL (e.g., https://login.salesforce.com or https://test.salesforce.com).
    *   **Default:** `https://login.salesforce.com`
    *   **Required:** False
*   `enforce_policy`:
    *   **Type:** Boolean (String representation)
    *   **Description:** If true, scans `devops-config.json` before allowing CLI use.
    *   **Default:** `true`
    *   **Required:** False
*   `test_results`:
    *   **Type:** String (JSON)
    *   **Description:** JSON string with test results: `{"total": 100, "passed": 95, "failed": 5}`.
    *   **Required:** False
*   `coverage_percentage`:
    *   **Type:** String
    *   **Description:** Code coverage percentage (0-100).
    *   **Required:** False
*   `coverage_required`:
    *   **Type:** String
    *   **Description:** Required coverage percentage threshold.
    *   **Default:** `80`
    *   **Required:** False
*   `plugins`:
    *   **Type:** String
    *   **Description:** Comma-separated list of Salesforce CLI plugins to install (e.g., `sfdx-hardis,@salesforce/plugin-deploy-retrieve`).
    *   **Required:** False
*   `skip_auth`:
    *   **Type:** Boolean (String representation)
    *   **Description:** If true, skips Salesforce authentication (useful for CI/governance testing).
    *   **Default:** `false`
    *   **Required:** False
*   `config_path`:
    *   **Type:** String
    *   **Description:** Path to `devops-config.json` (relative to workspace root).
    *   **Default:** `config/devops-config.json`
    *   **Required:** False
*   `otel_endpoint`:
    *   **Type:** String
    *   **Description:** OTLP endpoint for telemetry export (optional, e.g., `http://localhost:4318`).
    *   **Required:** False
*   `otel_service_name`:
    *   **Type:** String
    *   **Description:** Service name for OpenTelemetry traces.
    *   **Default:** `glassops-runtime`
    *   **Required:** False
*   `otel_headers`:
    *   **Type:** String
    *   **Description:** Key-value pairs for OTLP exporter headers (e.g., `"Authorization=Basic ..."`).
    *   **Required:** False

### `outputs`

This section defines the output variables that the action produces.

*   `runtime_id`:
    *   **Type:** String
    *   **Description:** Unique ID for this execution session.
*   `org_id`:
    *   **Type:** String
    *   **Description:** Authenticated Org ID.
*   `is_locked`:
    *   **Type:** Boolean (String representation)
    *   **Description:** True if the environment is currently frozen by policy.
*   `contract_path`:
    *   **Type:** String
    *   **Description:** Path to the generated deployment contract JSON file.
*   `glassops_ready`:
    *   **Type:** Boolean (String representation)
    *   **Description:** Indicates if runtime is ready for execution.

### `runs`

This section defines how the action is executed.

*   `using`:
    *   **Type:** String
    *   **Description:** Specifies the runner type.
    *   **Value:** `docker`
*   `image`:
    *   **Type:** String
    *   **Description:** The Docker image to use.
    *   **Value:** `Dockerfile`
*   `env`:
    *   **Type:** Object
    *   **Description:** Defines environment variables to be set during execution.  These variables are populated from the action's inputs.
    *   **Value:**
        ```yaml
        GLASSOPS_JWT_KEY: ${{ inputs.jwt_key }}
        GLASSOPS_CLIENT_ID: ${{ inputs.client_id }}
        GLASSOPS_USERNAME: ${{ inputs.username }}
        GLASSOPS_INSTANCE_URL: ${{ inputs.instance_url }}
        GLASSOPS_ENFORCE_POLICY: ${{ inputs.enforce_policy }}
        GLASSOPS_TEST_RESULTS: ${{ inputs.test_results }}
        GLASSOPS_COVERAGE_PERCENTAGE: ${{ inputs.coverage_percentage }}
        GLASSOPS_COVERAGE_REQUIRED: ${{ inputs.coverage_required }}
        GLASSOPS_PLUGINS: ${{ inputs.plugins }}
        GLASSOPS_SKIP_AUTH: ${{ inputs.skip_auth }}
        GLASSOPS_CONFIG_PATH: ${{ inputs.config_path }}
        OTEL_EXPORTER_OTLP_ENDPOINT: ${{ inputs.otel_endpoint }}
        GLASSOPS_OTEL_SERVICE_NAME: ${{ inputs.otel_service_name }}
        OTEL_EXPORTER_OTLP_HEADERS: ${{ inputs.otel_headers }}