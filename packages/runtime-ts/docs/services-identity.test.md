---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/services/identity.test.ts
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/src/services/identity.test.ts
generated_at: 2026-01-31T10:13:10.720873
hash: 918ca996fb332cea01787f5421099d52ba39b83780127de3f138ac7b0198480f
---

## Identity Resolver Service Documentation

This document details the functionality of the Identity Resolver service, responsible for authenticating with external systems and obtaining organizational identifiers.

**Overview**

The Identity Resolver service manages the authentication process using the Salesforce CLI (sf). It securely handles credentials and interacts with the CLI to retrieve an organization ID (org ID) and access token. The service is designed to be resilient, with built-in retry logic for handling transient authentication failures.

**Functionality**

The primary function of this service is the `authenticate` method. This method accepts authentication request parameters and returns the org ID upon successful authentication.

**`authenticate(authRequest)`**

This method performs the following actions:

1.  **Credential Storage:** Securely writes the provided JSON Web Token (JWT) key to a file named `glassops-jwt` with restricted permissions (mode 0o600).
2.  **Authentication Execution:** Executes the `sf` command with appropriate arguments, including the client ID and optionally the instance URL. The command's output, which contains the org ID and access token, is parsed.
3.  **Instance URL Handling:**  If an instance URL is provided in the `authRequest`, it is included as an argument to the `sf` command. If no instance URL is provided, it is omitted.
4.  **Success Handling:** Upon successful authentication, the method returns the retrieved org ID.
5.  **Failure Handling:** If authentication fails (e.g., due to invalid credentials or network issues), the method retries the authentication process multiple times with increasing backoff intervals.  If all retries fail, an "Authentication Failed" error is thrown.
6.  **Credential Cleanup:** Regardless of success or failure, the JWT key file (`glassops-jwt`) is deleted to maintain security.  The deletion is skipped if the file does not exist.

**Input Parameters (`authRequest`)**

The `authenticate` method accepts an object with the following properties:

*   `clientId`:  A string representing the client ID.
*   `jwtKey`: A string containing the private key in PEM format.
*   `username`: A string representing the username.
*   `instanceUrl` (Optional): A string representing the Salesforce instance URL.

**Error Handling**

The `authenticate` method throws an "Authentication Failed" error if authentication cannot be completed after multiple retries.

**Security Considerations**

*   The JWT key is stored securely with restricted file permissions (0o600).
*   The JWT key file is deleted immediately after use, regardless of authentication success or failure.
*   The service handles potential errors and retries authentication to improve resilience.

**Dependencies**

*   `@actions/exec`: Used for executing the `sf` command.
*   `fs`: Used for file system operations (writing and deleting the JWT key file).