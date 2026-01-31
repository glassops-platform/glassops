---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/integration/identity.integration.test.ts
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/src/integration/identity.integration.test.ts
generated_at: 2026-01-31T10:08:52.525436
hash: 08ee7d731762dc4c1479b9f5d75b6fe4580c23a9416e0cec0c9778be60362917
---

## Identity Resolver Integration Test Documentation

This document details the integration tests for the Identity Resolver component. These tests ensure proper functionality when interacting with Salesforce CLI authentication processes.

**Purpose**

The Identity Resolver facilitates authentication with Salesforce using JWT (JSON Web Token) flows. These tests verify successful authentication, error handling, and proper cleanup procedures.

**Functionality**

The `authenticate` function is the primary interface. It accepts the following parameters:

*   `clientId`:  The Salesforce Connected App Client ID (string).
*   `jwtKey`: The RSA private key used for JWT generation (string).
*   `username`: The Salesforce username (string).
*   `instanceUrl`: (Optional) The Salesforce instance URL (string). If not provided, the default Salesforce production URL is used.

The function returns the Salesforce Organization ID (Org ID) upon successful authentication.

**Test Coverage**

The integration tests cover the following scenarios:

*   **Successful Authentication:** Validates successful authentication with both production and sandbox instances, including scenarios with and without a specified `instanceUrl`.
*   **Authentication Failure:**
    *   Handles failures in the Salesforce CLI command execution.
    *   Handles errors during JSON response parsing.
*   **JWT Key Management:** Ensures temporary JWT key files are created and removed correctly, even in failure scenarios.
*   **Silent Execution:** Confirms the Salesforce CLI commands are executed silently, and standard output is captured via listeners.
*   **Environment Support:** Validates authentication against different Salesforce environments, including custom domains and sandboxes.

**Behavior**

*   Upon successful authentication, the `authenticate` function returns the Salesforce Organization ID.
*   The component leverages the Salesforce CLI (`sf`) for authentication.
*   Temporary files containing the JWT key are created during the authentication process and are automatically deleted afterward, regardless of success or failure.
*   The component defaults to a silent execution mode for CLI commands.
*   Error messages are provided to indicate authentication failures, guiding the user to check their Client ID and JWT Key.

**Dependencies**

*   Salesforce CLI (`sf`) must be installed and configured.
*   The `@actions/exec` and `@actions/core` modules are mocked during testing to isolate the component and simulate CLI interactions.
*   Node.js file system (`fs`) and operating system (`os`) modules are used for temporary file management.

**Usage**

You must provide a valid `clientId`, `jwtKey`, and `username` to authenticate.  Optionally, you can specify the `instanceUrl` for sandbox or custom domain environments.