---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/services/identity.test.ts
last_modified: 2026-01-29
generated: true
source: packages/runtime-ts/src/services/identity.test.ts
generated_at: 2026-01-29T20:59:56.058102
hash: 918ca996fb332cea01787f5421099d52ba39b83780127de3f138ac7b0198480f
---

## Identity Resolver Service Documentation

This document details the functionality of the Identity Resolver service, responsible for authenticating with external systems and obtaining organizational identifiers.

**Overview**

The Identity Resolver service manages authentication processes. It securely handles credentials and interacts with external authentication tools to retrieve an organization ID (Org ID) and access token.  The service is designed to be resilient to temporary authentication failures through a retry mechanism.

**Functionality**

The primary function of this service is the `authenticate` method. This method accepts authentication request parameters and returns the associated Org ID upon successful authentication.

**Authentication Request Parameters:**

*   `clientId`: A unique identifier for the client application.
*   `jwtKey`: The JSON Web Token (JWT) private key used for authentication. This key is written to disk securely with restricted permissions (mode 0o600).
*   `username`: The username associated with the identity.
*   `instanceUrl` (Optional): The URL of the instance to authenticate against. If provided, it is included in the authentication request.

**Authentication Process:**

1.  **JWT Key Storage:** Upon receiving an authentication request, the service securely writes the provided `jwtKey` to a file on disk.
2.  **External Tool Execution:** The service executes an external tool ("sf") with appropriate arguments, including the `clientId` and optionally the `instanceUrl`.
3.  **Response Handling:** The service parses the response from the external tool, expecting a JSON object containing the `orgId` and `accessToken`.
4.  **Org ID Return:** If the authentication is successful, the service returns the extracted `orgId`.
5.  **Error Handling:** If the external tool returns an error, the service retries the authentication process. If retries are exhausted, an "Authentication Failed" error is thrown.
6.  **Cleanup:** Regardless of success or failure, the service attempts to delete the JWT key file to maintain security.  Deletion is skipped if the file does not exist.

**Dependencies:**

*   `@actions/exec`: Used for executing external processes.
*   `fs`: Used for file system operations (writing and deleting the JWT key file).

**Security Considerations:**

*   The JWT key is written to disk with restricted permissions (mode 0o600) to prevent unauthorized access.
*   The JWT key file is deleted after authentication, regardless of success or failure, to minimize the risk of compromise.

**Usage:**

You should instantiate the `IdentityResolver` class and call the `authenticate` method with the appropriate authentication request parameters.  Ensure you handle potential errors, such as "Authentication Failed", and implement appropriate logging and error reporting.