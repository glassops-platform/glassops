---
type: Documentation
domain: runtime
origin: packages/runtime/internal/services/identity.go
last_modified: 2026-01-31
generated: true
source: packages/runtime/internal/services/identity.go
generated_at: 2026-01-31T10:05:11.418403
hash: 3940eb183960a36e834ef630b97962783208d0a0acd53fd5522e36dcb1c8a15d
---

## Identity Service Documentation

This document describes the Identity Service, responsible for handling authentication with Salesforce. It provides methods for authenticating using both JWT (JSON Web Token) and SFDX (Salesforce DX) authentication URLs.

### Package Responsibilities

The `services` package contains the `IdentityResolver` and `Identity` types, which encapsulate the logic for authenticating with Salesforce. The service supports two primary authentication methods: JWT-based authentication using a client ID and JWT key, and authentication via SFDX auth URLs.  It aims to securely manage credentials during the authentication process and provide a reliable interface for obtaining Salesforce organization IDs.

### Key Types and Interfaces

*   **`AuthRequest`**: A structure that holds the parameters required for JWT-based authentication. It contains the following fields:
    *   `ClientID`: The Salesforce Client ID.
    *   `JWTKey`: The JWT key used for authentication.
    *   `Username`: The Salesforce username.
    *   `InstanceURL`: The Salesforce instance URL (optional).

*   **`IdentityResolver`**: This type handles JWT-based authentication. It provides a method to authenticate with Salesforce using the `AuthRequest` parameters.

*   **`Identity`**: This type handles authentication using SFDX auth URLs. It provides methods to parse and authenticate using these URLs.

### Important Functions

**`NewIdentityResolver()`**:
This function creates and returns a new instance of the `IdentityResolver` type. You can use this to initialize the JWT authentication functionality.

**`Authenticate(req AuthRequest)` (method of `IdentityResolver`)**:
This function performs JWT-based authentication with Salesforce. It takes an `AuthRequest` as input and returns the Salesforce organization ID and an error, if any. The function securely writes the JWT key to a temporary file, executes the `sf org login jwt` command, parses the response, and returns the organization ID. It includes retry logic for transient Salesforce API failures.  The JWT key file is securely overwritten with zeros before deletion.

**`NewIdentity()`**:
This function creates and returns a new instance of the `Identity` type. You can use this to initialize the SFDX auth URL authentication functionality.

**`ParseAuthURL(authURL string)` (method of `Identity`)**:
This function parses and validates a Salesforce SFDX auth URL. It checks the URL format and validates the instance URL against a regular expression to ensure it is a valid Salesforce domain. It returns the instance URL if the validation is successful, or an error if the URL is invalid.

**`AuthenticateWithURL(authURL string)` (method of `Identity`)**:
This function authenticates with Salesforce using an SFDX auth URL. It first validates the URL using `ParseAuthURL`. It then writes the auth URL to a temporary file, executes the `sf org login sfdx-url` command, parses the response, and returns the Salesforce organization ID. The auth URL file is securely overwritten with zeros before deletion.

### Error Handling

The service employs robust error handling. Functions return an error value alongside their primary return value. Errors are often wrapped using `fmt.Errorf` with `%w` to preserve the original error context, aiding in debugging. Specific error messages are provided for common issues like invalid URL formats, file write failures, and authentication failures.

### Concurrency

This service does not explicitly use goroutines or channels. However, the `Authenticate` function incorporates a retry mechanism with exponential backoff using `time.Sleep`, which can be considered a form of implicit concurrency management to handle transient errors.

### Design Decisions

*   **Secure Credential Handling**: The service writes sensitive credentials (JWT key and auth URL) to temporary files with restricted permissions (0600) and securely overwrites the file contents with zeros before deleting them. This minimizes the risk of credential exposure.
*   **SFDX CLI Dependency**: The service relies on the Salesforce CLI (SFDX) being installed and configured on the system. This allows it to leverage the SFDX tooling for authentication.
*   **Retry Logic**: The `Authenticate` function includes retry logic to handle transient Salesforce API failures, improving the reliability of the authentication process.
*   **Temporary Files**: The use of temporary files is essential for securely passing credentials to the SFDX CLI without exposing them in process memory or command-line arguments.