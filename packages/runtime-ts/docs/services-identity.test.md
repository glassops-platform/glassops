---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/services/identity.test.ts
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/src/services/identity.test.ts
generated_at: 2026-01-31T09:16:47.470536
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

**Workflow:**

1.  **JWT Key Storage:** Upon receiving an authentication request, the service securely writes the provided `jwtKey` to a file on disk.
2.  **Authentication Execution:** The service executes an external authentication command ("sf") with the provided parameters. The command's output is parsed to extract the Org ID and access token.
3.  **Instance URL Handling:** If an `instanceUrl` is provided in the request, it is included as a parameter in the authentication command.
4.  **Success Handling:** If authentication is successful, the Org ID is returned.
5.  **Failure Handling:** If authentication fails, the service retries the operation. After exhausting retry attempts, an "Authentication Failed" error is thrown.
6.  **Cleanup:** Regardless of success or failure, the service attempts to delete the JWT key file to maintain security.  Deletion is skipped if the file does not exist.

**Error Handling:**

*   The `authenticate` method throws an "Authentication Failed" error if authentication cannot be completed after multiple retries.

**Security Considerations:**

*   The JWT key is written to disk with restricted permissions (mode 0o600) to prevent unauthorized access.
*   The JWT key file is deleted after authentication, regardless of success or failure, to minimize the risk of compromise.

**Dependencies:**

*   `@actions/exec`: Used to execute external commands.
*   `fs`: Used for file system operations (writing and deleting the JWT key file).