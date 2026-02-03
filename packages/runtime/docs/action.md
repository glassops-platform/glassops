---
type: Documentation
domain: runtime
last_modified: 2026-02-02
generated: true
source: packages/runtime/action.yml
generated_at: 2026-02-02T22:33:51.940703
hash: ec91c30a592eb0967662ac9901db54840aebc296fa1897a08af5535daa8d354e
---

# GlassOps Runtime Action Documentation

This document details the configuration for the GlassOps Runtime action. This action bootstraps a governed Salesforce execution environment, ensuring adherence to the GlassOps Protocol.

## Purpose

The primary purpose of this action is to set up a secure and controlled environment for interacting with a Salesforce organization. It handles authentication, policy enforcement, and provides outputs for tracking the execution session.

## Structure

The configuration is structured as a YAML file with the following top-level keys:

*   `name`: A descriptive name for the action.
*   `description`: A detailed explanation of the action's purpose.
*   `author`: The creator of the action.
*   `branding`: Visual elements for the action.
*   `inputs`: Parameters that You provide to customize the action's behavior.
*   `outputs`: Values that the action produces and makes available for subsequent steps.
*   `runs`: Specifies how the action is executed.

## Key Details

### `name`

*   Type: String
*   Value: `'GlassOps Runtime'`
*   Purpose:  Identifies the action.

### `description`

*   Type: String
*   Value: `'Bootstraps a governed Salesforce execution environment adhering to the GlassOps Protocol.'`
*   Purpose: Provides a human-readable explanation of the action.

### `author`

*   Type: String
*   Value: `'Ryan Bumstead'`
*   Purpose:  Identifies the action's creator.

### `branding`

*   Type: Object
    *   `icon`: String, Value: `'shield'`, Purpose:  Icon to represent the action.
    *   `color`: String, Value: `'blue'`, Purpose: Color associated with the action.

### `inputs`

This section defines the input parameters for the action.

*   `jwt_key`
    *   Type: String
    *   Required: `true`
    *   Purpose: The private key used for JSON Web Token (JWT) based authentication.
*   `client_id`
    *   Type: String
    *   Required: `true`
    *   Purpose: The consumer key for the external client application.
*   `username`
    *   Type: String
    *   Required: `true`
    *   Purpose: The username of the target Salesforce organization.
*   `instance_url`
    *   Type: String
    *   Required: `false`
    *   Default: `'https://login.salesforce.com'`
    *   Purpose: The Salesforce login URL.  You can specify `https://test.salesforce.com` for sandbox environments.
*   `enforce_policy`
    *   Type: Boolean
    *   Required: `false`
    *   Default: `'true'`
    *   Purpose:  Determines whether to scan the `devops-config.json` file for policy violations before allowing CLI access.
*   `test_results`
    *   Type: String
    *   Required: `false`
    *   Purpose: A JSON string containing test results in the format `{"total": 100, "passed": 95, "failed": 5}`.
*   `coverage_percentage`
    *   Type: Integer
    *   Required: `false`
    *   Purpose: The code coverage percentage achieved during testing (a value between 0 and 100).
*   `coverage_required`
    *   Type: Integer
    *   Required: `false`
    *   Default: `'80'`
    *   Purpose: The minimum required code coverage percentage.
*   `plugins`
    *   Type: String
    *   Required: `false`
    *   Purpose: A comma-separated list of Salesforce CLI plugins to install. Example: `'sfdx-hardis,@salesforce/plugin-deploy-retrieve'`.
*   `skip_auth`
    *   Type: Boolean
    *   Required: `false`
    *   Default: `'false'`
    *   Purpose:  If set to `true`, skips Salesforce authentication. This is useful for CI/governance testing scenarios.

### `outputs`

This section defines the outputs produced by the action.

*   `runtime_id`
    *   Type: String
    *   Purpose: A unique identifier for the execution session.
*   `org_id`
    *   Type: String
    *   Purpose: The ID of the authenticated Salesforce organization.
*   `is_locked`
    *   Type: Boolean
    *   Purpose: Indicates whether the environment is currently frozen due to policy restrictions.
*   `contract_path`
    *   Type: String
    *   Purpose: The file path to the generated deployment contract JSON file.

### `runs`

*   `using`: String, Value: `'docker'`, Purpose: Specifies that the action is executed within a Docker container.
*   `image`: String, Value: `'Dockerfile'`, Purpose:  Specifies the Dockerfile used to build the container image. We maintain this Dockerfile within the repository.