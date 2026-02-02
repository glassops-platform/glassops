---
type: Documentation
domain: runtime
origin: packages/runtime/internal/services/identity.go
last_modified: 2026-02-01
generated: true
source: packages/runtime/internal/services/identity.go
generated_at: 2026-02-01T19:44:42.253696
hash: 3940eb183960a36e834ef630b97962783208d0a0acd53fd5522e36dcb1c8a15d
---

## Identity Service Documentation

This document describes the Identity Service, responsible for handling authentication with Salesforce. It provides methods for authenticating using both JWT (JSON Web Token) and SFDX (Salesforce DX) authentication URLs.

### Package Responsibilities

The `services` package contains the `IdentityResolver` and `Identity` types, which encapsulate the logic for interacting with Salesforce authentication mechanisms. The primary goal is to securely obtain a Salesforce Organization ID (OrgID) for use by other components.

### Key Types and Interfaces

*   **`AuthRequest`**: A structure that holds the parameters required for JWT-based authentication. It contains the following fields:
    *   `ClientID`: The Salesforce Client ID.
    *   `JWTKey`: The private key used to generate the JWT.
    *   `Username`: The Salesforce username.
    *   `InstanceURL`: The Salesforce instance URL (optional).

*   **`IdentityResolver`**: This type handles JWT authentication. It provides a method to authenticate with Salesforce using a JWT key.

*   **`Identity`**: This type handles authentication using SFDX auth URLs. It provides methods to parse and authenticate using these URLs.

### Important Functions

**`IdentityResolver.Authenticate(req AuthRequest) (string, error)`**:

This function performs JWT-based authentication with Salesforce. It takes an `AuthRequest` as input and returns the Salesforce OrgID upon successful authentication, along with a potential error. The process involves:

1.  Writing the JWT key to a temporary file with restricted permissions (0600).
2.  Executing the `sf org login jwt` command using the Salesforce CLI (SFDX).
3.  Retrying the command up to three times with exponential backoff in case of transient Salesforce API failures.
4.  Parsing the JSON response from the SFDX command to extract the OrgID.
5.  Securely deleting the temporary JWT key file by overwriting it with zeros before unlinking.

**`Identity.ParseAuthURL(authURL string) (string, error)`**:

This function parses a Salesforce SFDX auth URL and validates its format. It checks if the URL starts with "force://" and if it contains a valid Salesforce instance URL (ending with ".salesforce.com"). It returns the instance URL if the format is valid.

**`Identity.AuthenticateWithURL(authURL string) (string, error)`**:

This function authenticates with Salesforce using an SFDX auth URL. It takes the auth URL as input and returns the Salesforce OrgID upon successful authentication, along with a potential error. The process involves:

1.  Validating the auth URL using `ParseAuthURL`.
2.  Writing the auth URL to a temporary file with restricted permissions (0600).
3.  Executing the `sf org login sfdx-url` command using the Salesforce CLI (SFDX).
4.  Parsing the JSON response from the SFDX command to extract the OrgID.
5.  Securely deleting the temporary auth URL file by overwriting it with zeros before unlinking.

**`NewIdentityResolver() *IdentityResolver`**:

This function creates and returns a new instance of the `IdentityResolver` type.

**`NewIdentity() *Identity`**:

This function creates and returns a new instance of the `Identity` type.

### Error Handling

The service employs robust error handling. Functions return both a result (OrgID) and an error value. Errors are wrapped using `fmt.Errorf` to provide context and preserve the original error for debugging. Specific error messages are provided for common issues like invalid URL formats, file write failures, and authentication failures.

### Concurrency

This service does not explicitly use goroutines or channels. However, the `Authenticate` function incorporates a retry mechanism with a `time.Sleep` to handle transient errors, which could be considered a form of implicit concurrency management.

### Design Decisions

*   **Secure Key Management**: The JWT key and auth URL are written to temporary files with restricted permissions (0600) and securely deleted after use by overwriting the file contents with zeros before removal. This minimizes the risk of sensitive information being exposed.
*   **SFDX CLI Dependency**: The service relies on the Salesforce CLI (SFDX) being installed and configured on the system.
*   **Retry Mechanism**: The `Authenticate` function includes a retry mechanism to handle transient Salesforce API failures, improving reliability.
*   **Clear Error Messages**:  The service provides informative error messages to aid in troubleshooting.