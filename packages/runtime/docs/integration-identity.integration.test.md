---
type: Documentation
domain: runtime
origin: packages/runtime/src/integration/identity.integration.test.ts
last_modified: 2026-01-26
generated: true
source: packages/runtime/src/integration/identity.integration.test.ts
generated_at: 2026-01-26T14:15:25.233Z
hash: 9eb477ecf01bd9d842f7527397bdb0125870e7f174a9d8099cb1bce7ae65086e
---

## Identity Resolver Integration Test Documentation

This document details the integration tests for the Identity Resolver component. These tests ensure proper functionality when interacting with Salesforce authentication processes via the Salesforce CLI.

**Purpose**

The Identity Resolver facilitates authentication with Salesforce using JWT (JSON Web Token) authentication flows. These tests verify the resolver’s ability to successfully authenticate, handle failures, and manage temporary key files.

**Scope**

The tests cover the following scenarios:

*   Successful authentication with valid credentials, including optional instance URL specification.
*   Handling authentication failures due to CLI command errors or invalid JSON responses.
*   Proper cleanup of temporary JWT key files, both on success and failure.
*   Authentication against different Salesforce environments (Production, Sandbox, Custom Domains).
*   Silent execution of CLI commands and stdout capture.

**Functionality**

The `IdentityResolver` class interacts with the Salesforce CLI (`sf`) to perform authentication.  It accepts the following inputs:

*   `clientId`: The client ID for the Salesforce connected app.
*   `jwtKey`: The private key associated with the client ID.
*   `username`: The Salesforce username.
*   `instanceUrl` (Optional): The Salesforce instance URL (e.g., `https://login.salesforce.com` or a sandbox URL). If not provided, the resolver attempts authentication without it.

Upon successful authentication, the resolver returns the Salesforce Organization ID.

**Test Methodology**

These tests employ mocking to simulate the Salesforce CLI execution environment. Specifically, the `@actions/exec` module is mocked to control the behavior of the `sf` command.  This allows for testing various success and failure scenarios without requiring a live Salesforce environment.  The tests verify:

*   The correct `sf` command and arguments are invoked.
*   The expected output is processed correctly.
*   Error conditions are handled gracefully.
*   Temporary files are created and deleted as expected.

**Error Handling**

The Identity Resolver is designed to handle the following error conditions:

*   **CLI Command Failure:** If the `sf` command fails to execute, the resolver throws an error message: "❌ Authentication Failed. Check Client ID and JWT Key."
*   **Invalid JSON Response:** If the `sf` command returns an invalid JSON response, the resolver throws an error message: "❌ Authentication Failed. Check Client ID and JWT Key."

**Temporary File Management**

The resolver creates temporary files to store the JWT key during the authentication process. These files are automatically cleaned up after authentication, regardless of success or failure. The files are named with a prefix of "glassops-jwt-".

**Configuration**

You can configure the authentication process by providing the necessary credentials (clientId, jwtKey, username, instanceUrl) to the `authenticate` method. The `instanceUrl` parameter is optional.

**Dependencies**

*   Salesforce CLI (`sf`) – Mocked in the test environment.
*   Node.js file system (`fs`) and operating system (`os`) modules.