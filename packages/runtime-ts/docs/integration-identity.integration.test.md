---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/integration/identity.integration.test.ts
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/src/integration/identity.integration.test.ts
generated_at: 2026-01-31T09:12:29.782438
hash: 08ee7d731762dc4c1479b9f5d75b6fe4580c23a9416e0cec0c9778be60362917
---

## Identity Resolver Integration Test Documentation

This document details the integration tests for the Identity Resolver component. These tests ensure proper functionality when interacting with Salesforce authentication processes via the Salesforce CLI.

**Purpose**

The Identity Resolver facilitates authentication with Salesforce using JWT (JSON Web Token) flows. These tests verify successful authentication, error handling, and proper cleanup procedures.

**Scope**

The tests cover the following scenarios:

*   Successful authentication with valid credentials, including optional instance URL specification.
*   Handling of authentication failures due to CLI command errors or invalid JSON responses.
*   Cleanup of temporary JWT key files, both on successful authentication and failure.
*   Authentication against different Salesforce environments (Production, Sandbox, Custom Domains).
*   Silent execution of CLI commands and proper stdout capture.

**Functionality**

The Identity Resolver’s `authenticate` method accepts the following parameters:

*   `clientId`: The Salesforce Connected App Client ID.
*   `jwtKey`: The private key associated with the Connected App.
*   `username`: The Salesforce username.
*   `instanceUrl` (Optional): The Salesforce instance URL. If not provided, the default Salesforce production instance is used.

Upon successful authentication, the method returns the Salesforce Organization ID.

**Test Methodology**

These tests employ mocking of the `@actions/exec` module to simulate Salesforce CLI interactions. This allows for controlled testing without requiring a live Salesforce environment.  The tests verify:

*   Correct CLI command construction and execution.
*   Expected return values from the CLI.
*   Proper error handling and exception throwing.
*   File system operations related to JWT key management.

**Error Handling**

The Identity Resolver is designed to handle the following error conditions:

*   Failure of the Salesforce CLI command.  In this case, an error message "❌ Authentication Failed. Check Client ID and JWT Key." is thrown.
*   Invalid JSON response from the Salesforce CLI.  The same error message is thrown.

**Cleanup**

The Identity Resolver ensures that any temporary JWT key files created during the authentication process are removed, regardless of whether authentication succeeds or fails. This prevents sensitive information from persisting on the system.

**Environment Considerations**

The tests utilize the operating system’s temporary directory (`os.tmpdir()`) for JWT key file storage.  The tests verify that the correct files are created and deleted within this directory.

**Silent Execution**

The `authenticate` method executes Salesforce CLI commands in silent mode, suppressing standard output.  Stdout is captured via listeners for processing.