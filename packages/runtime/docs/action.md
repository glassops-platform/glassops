---
type: Documentation
domain: runtime
origin: packages/runtime/action.yml
last_modified: 2026-02-01
generated: true
source: packages/runtime/action.yml
generated_at: 2026-02-01T19:37:36.711062
hash: ec91c30a592eb0967662ac9901db54840aebc296fa1897a08af5535daa8d354e
---

# GlassOps Runtime Action Configuration

This document details the configuration for the GlassOps Runtime action. This action establishes a controlled Salesforce environment following the GlassOps Protocol. We designed this action to provide a secure and governed execution context for Salesforce development and deployment tasks.

## Structure

The configuration is structured using YAML and defines the action's metadata, inputs, outputs, and execution environment. It is divided into the following sections:

*   `name`: A human-readable name for the action.
*   `description`: A brief explanation of the action's purpose.
*   `author`: The creator of the action.
*   `branding`: Visual elements for the action's presentation.
*   `inputs`: Parameters that You provide to customize the action's behavior.
*   `outputs`: Values that the action returns after execution.
*   `runs`: Specifies how the action is executed.

## Key Details

### Metadata

*   `name`: `GlassOps Runtime`
    *   The name displayed for this action.
*   `description`: `Bootstraps a governed Salesforce execution environment adhering to the GlassOps Protocol.`
    *   A description of the action's function.
*   `author`: `Ryan Bumstead`
    *   The author responsible for the action.
*   `branding`:
    *   `icon`: `shield`
        *   An icon representing the action.
    *   `color`: `blue`
        *   The primary color associated with the action.

### Inputs

These are the parameters You can configure when using the action.

*   `jwt_key`:
    *   `description`: `Private Key for JWT flow`
    *   `required`: `true`
    *   The private key used for JSON Web Token (JWT) authentication. This is a mandatory input.
*   `client_id`:
    *   `description`: `External Client App Consumer Key`
    *   `required`: `true`
    *   The consumer key for the external client application. This is a mandatory input.
*   `username`:
    *   `description`: `Target Org Username`
    *   `required`: `true`
    *   The username of the Salesforce organization to connect to. This is a mandatory input.
*   `instance_url`:
    *   `description`: `Salesforce login URL (e.g., https://login.salesforce.com or https://test.salesforce.com)`
    *   `default`: `https://login.salesforce.com`
    *   `required`: `false`
    *   The Salesforce instance URL. Defaults to `https://login.salesforce.com` if not provided.
*   `enforce_policy`:
    *   `description`: `If true, scans devops-config.json before allowing CLI use`
    *   `default`: `true`
    *   `required`: `false`
    *   A boolean value indicating whether to enforce policy checks against the `devops-config.json` file before allowing CLI access. Defaults to `true`.
*   `test_results`:
    *   `description`: `JSON string with test results: {"total": 100, "passed": 95, "failed": 5}`
    *   `required`: `false`
    *   A JSON string containing test results.
*   `coverage_percentage`:
    *   `description`: `Code coverage percentage (0-100)`
    *   `required`: `false`
    *   The code coverage percentage achieved during testing.
*   `coverage_required`:
    *   `description`: `Required coverage percentage threshold`
    *   `default`: `80`
    *   `required`: `false`
    *   The minimum required code coverage percentage. Defaults to `80`.
*   `plugins`:
    *   `description`: `Comma-separated list of Salesforce CLI plugins to install (e.g., 'sfdx-hardis,@salesforce/plugin-deploy-retrieve')`
    *   `required`: `false`
    *   A comma-separated list of Salesforce CLI plugins to install.
*   `skip_auth`:
    *   `description`: `If true, skips Salesforce authentication (useful for CI/governance testing)`
    *   `default`: `false`
    *   `required`: `false`
    *   A boolean value indicating whether to skip Salesforce authentication. Defaults to `false`.

### Outputs

These are the values returned by the action after it completes.

*   `runtime_id`:
    *   `description`: `Unique ID for this execution session`
    *   A unique identifier for the runtime session.
*   `org_id`:
    *   `description`: `Authenticated Org ID`
    *   The ID of the authenticated Salesforce organization.
*   `is_locked`:
    *   `description`: `True if the environment is currently frozen by policy`
    *   A boolean value indicating whether the environment is locked due to policy enforcement.
*   `contract_path`:
    *   `description`: `Path to the generated deployment contract JSON file`
    *   The file path to the generated deployment contract.

### Execution

*   `runs`:
    *   `using`: `docker`
        *   Specifies that the action is executed within a Docker container.
    *   `image`: `Dockerfile`
        *   The name of the Dockerfile used to build the container image. We expect a Dockerfile to be present in the same directory as this configuration file.