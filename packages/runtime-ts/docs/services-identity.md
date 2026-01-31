---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/services/identity.ts
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/src/services/identity.ts
generated_at: 2026-01-31T10:13:31.716537
hash: 7da4458a3bb94c021316e184cd6743876ac04c4f59fee21885b081066399ee60
---

## Identity Resolver Service Documentation

This document details the functionality of the Identity Resolver service, responsible for authenticating with an external organization using a JSON Web Token (JWT).

**Overview**

The Identity Resolver service provides a method to authenticate against an organization, obtaining an organization ID and access token. It is designed for use in automated environments where interactive login is not possible. The service handles temporary file management for the JWT key and incorporates retry logic for transient API failures.

**Key Concepts**

* **JWT Authentication:** This service supports authentication via JWT, a standard for securely transmitting information between parties as a JSON object.
* **Salesforce CLI (sf):** The service leverages the Salesforce CLI (`sf`) to perform the authentication process.  Ensure the Salesforce CLI is installed and configured in the execution environment.
* **Retry Logic:**  Transient errors during authentication are automatically retried up to three times with exponential backoff.

**Data Structures**

* **AuthRequest:**  An interface defining the input parameters for the authentication process.
    * `clientId`:  The client ID associated with the JWT. (string)
    * `jwtKey`: The private key used to generate the JWT. (string)
    * `username`: The username for the organization. (string)
    * `instanceUrl`: (Optional) The instance URL of the organization. (string)

**Functions**

* **`authenticate(req: AuthRequest): Promise<string>`**

    This asynchronous function performs the authentication process.

    **Parameters:**

    * `req`: An `AuthRequest` object containing the necessary authentication details.

    **Return Value:**

    * A `Promise` that resolves with the organization ID (orgId) as a string upon successful authentication.

    **Behavior:**

    1.  Creates a temporary file to store the JWT key securely. The file is created with restricted permissions (mode 0o600).
    2.  Constructs the command-line arguments for the `sf org login jwt` command.
    3.  Executes the `sf org login jwt` command using the provided credentials and arguments.
    4.  Parses the JSON output from the `sf` command to extract the organization ID and access token.
    5.  Logs a success message including the authenticated username and organization ID.
    6.  Returns the organization ID.
    7.  In case of failure, throws an error indicating authentication failure.
    8.  In the `finally` block, securely deletes the temporary JWT key file by first overwriting it with zeros and then unlinking it.

    **Example Usage:**

    ```typescript
    const identityResolver = new IdentityResolver();
    const authRequest: AuthRequest = {
      clientId: "your_client_id",
      jwtKey: "your_jwt_key",
      username: "your_username",
      instanceUrl: "your_instance_url",
    };

    try {
      const orgId = await identityResolver.authenticate(authRequest);
      console.log(`Successfully authenticated. Org ID: ${orgId}`);
    } catch (error) {
      console.error("Authentication failed:", error);
    }
    ```

**Security Considerations**

*   The JWT key is stored in a temporary file with restricted permissions (0o600).
*   The temporary file is securely overwritten with zeros before being deleted to prevent data recovery.
*   Ensure the environment where this service runs is secure to protect the JWT key.

**Error Handling**

*   The `authenticate` function throws an error if authentication fails. The error message provides guidance on checking the client ID and JWT key.
*   Transient errors during the `sf` command execution are handled with retry logic.
*   File system operations include error handling to ensure best-effort cleanup of the temporary JWT key file.

**Dependencies**

*   `@actions/core`: For logging and environment variable access.
*   `@actions/exec`: For executing shell commands.
*   `fs`: For file system operations.
*   `path`: For constructing file paths.
*   `os`: For accessing operating system-specific information (e.g., temporary directory).
*   `./retry`: For implementing retry logic.