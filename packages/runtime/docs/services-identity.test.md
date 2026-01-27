---
type: Documentation
domain: runtime
origin: packages/runtime/src/services/identity.test.ts
last_modified: 2026-01-26
generated: true
source: packages/runtime/src/services/identity.test.ts
generated_at: 2026-01-26T14:21:12.145Z
hash: 91451b37ca26c42e1476cb4d783ab2992407250fda23a53ce8a7fcc9feb6a646
---

## IdentityResolver Service Documentation

This document details the functionality of the `IdentityResolver` service, responsible for authenticating with Salesforce and retrieving organization (org) information.

**Overview**

The `IdentityResolver` service manages the authentication process against a Salesforce instance. It leverages the Salesforce CLI (`sf`) to perform the authentication and obtain the org ID and access token.  The service handles secure storage of the JWT key used for authentication and ensures cleanup of this key file, even in the event of authentication failures.

**Functionality**

The primary function of this service is the `authenticate` method. This method accepts authentication request parameters and returns the Salesforce org ID upon successful authentication.

**Authentication Request Parameters:**

*   `clientId`: A unique identifier for the client application.
*   `jwtKey`: The JSON Web Token (JWT) private key used for authentication. This key is written to a file with restricted permissions (mode 0o600) for security.
*   `username`: The Salesforce username associated with the authentication.
*   `instanceUrl` (Optional): The Salesforce instance URL. If provided, it is included in the authentication command.  If not provided, the service will proceed without it.

**Workflow:**

1.  **JWT Key Storage:** Upon receiving an authentication request, the service securely writes the provided `jwtKey` to a file.
2.  **Salesforce CLI Execution:** The service executes the `sf` command with appropriate arguments, including the username and optionally the instance URL.
3.  **Response Parsing:** The service parses the JSON response from the `sf` command, extracting the org ID and access token.
4.  **Org ID Return:**  If authentication is successful, the service returns the extracted org ID.
5.  **JWT Key Cleanup:** Regardless of success or failure, the service attempts to delete the JWT key file to maintain security.  Deletion is skipped if the file does not exist.
6.  **Error Handling:** If the `sf` command fails, the service throws an "Authentication Failed" error.

**Dependencies:**

*   `@actions/exec`: Used for executing the Salesforce CLI (`sf`) command.
*   `fs`: Used for file system operations (writing and deleting the JWT key file).

**Testing Considerations:**

The provided tests verify the following:

*   Successful authentication and org ID retrieval.
*   Correct inclusion of the instance URL in the `sf` command when provided.
*   Exclusion of the instance URL when not provided.
*   Error handling for authentication failures.
*   Proper cleanup of the JWT key file in both success and failure scenarios.
*   Prevention of deletion attempts on non-existent key files.

**Configuration:**

You can configure the Salesforce CLI path if it is not in your system's PATH environment variable.  The service relies on the `sf` command being accessible in the execution environment.