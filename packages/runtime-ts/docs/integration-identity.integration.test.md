---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/integration/identity.integration.test.ts
last_modified: 2026-01-29
generated: true
source: packages/runtime-ts/src/integration/identity.integration.test.ts
generated_at: 2026-01-29T20:55:51.673333
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
*   `instanceUrl` (Optional): The Salesforce instance URL (e.g., `https://login.salesforce.com` or a sandbox URL). If not provided, the resolver will attempt authentication without it.

Upon successful authentication, the method returns the Salesforce Organization ID.

**Test Methodology**

These tests employ mocking of the `@actions/exec` module to simulate interactions with the Salesforce CLI (`sf`).  This allows for controlled testing without requiring a live Salesforce environment.  The tests verify:

*   The correct CLI command and arguments are invoked.
*   The expected output is processed correctly.
*   Error conditions are handled gracefully.
*   Temporary files are cleaned up as expected.

**Error Handling**

The `authenticate` method throws an error with a descriptive message ("❌ Authentication Failed. Check Client ID and JWT Key.") in the following cases:

*   The Salesforce CLI command fails.
*   The CLI output is not valid JSON.

**Cleanup**

The Identity Resolver ensures that any temporary JWT key files created during the authentication process are removed, regardless of whether authentication succeeds or fails. This prevents sensitive information from persisting on the system.

**Silent Execution**

The `authenticate` method executes the Salesforce CLI commands in silent mode, suppressing standard output.  Stdout is captured via listeners for processing.

**Environments**

The Identity Resolver supports authentication against:

*   Production Salesforce instances.
*   Sandbox Salesforce instances.
*   Salesforce instances with custom domains.