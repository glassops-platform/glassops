---
type: Documentation
domain: runtime
last_modified: 2026-02-02
generated: true
source: packages/runtime/internal/services/identity.go
generated_at: 2026-02-02T22:40:19.610399
hash: 3940eb183960a36e834ef630b97962783208d0a0acd53fd5522e36dcb1c8a15d
---

## Identity Service Documentation

This document describes the Identity Service, responsible for handling authentication with Salesforce. It provides methods for authenticating using both JWT (JSON Web Token) and SFDX (Salesforce DX) authentication URLs.

### Package Responsibilities

The `services` package contains the `IdentityResolver` and `Identity` types, which encapsulate the logic for interacting with Salesforce authentication mechanisms. The primary responsibilities are:

*   Performing JWT-based authentication with Salesforce.
*   Parsing and validating Salesforce SFDX authentication URLs.
*   Authenticating using SFDX authentication URLs.
*   Retrieving the Salesforce Organization ID after successful authentication.

### Key Types and Interfaces

*   **`AuthRequest`**: This structure holds the parameters required for JWT authentication. It contains:
    *   `ClientID`: The Salesforce Client ID.
    *   `JWTKey`: The JWT key used for authentication.
    *   `Username`: The Salesforce username.
    *   `InstanceURL`: The Salesforce instance URL (optional).

*   **`IdentityResolver`**: This type handles JWT authentication. It provides a method to authenticate with Salesforce using a JWT key.

*   **`Identity`**: This type handles authentication via SFDX auth URLs. It provides methods to parse and authenticate using these URLs.

### Important Functions

**`NewIdentityResolver()`**: This function returns a new instance of the `IdentityResolver` type.

**`Authenticate(req AuthRequest)` (method of `IdentityResolver`)**: This function performs JWT-based authentication with Salesforce. It takes an `AuthRequest` as input and returns the Salesforce Organization ID and an error, if any. The function:

1.  Writes the JWT key to a temporary file with restricted permissions (0600).
2.  Executes the `sf org login jwt` command using the Salesforce CLI (SFDX).
3.  Retries the command up to three times in case of transient Salesforce API failures, with exponential backoff.
4.  Parses the JSON response from the SFDX command to extract the Organization ID.
5.  Securely cleans up the temporary JWT key file by overwriting it with zeros before deleting it.

**`NewIdentity()`**: This function returns a new instance of the `Identity` type.

**`ParseAuthURL(authURL string)` (method of `Identity`)**: This function parses and validates a Salesforce SFDX authentication URL. It checks the URL format and validates the instance URL domain. It returns the instance URL if the validation is successful, or an error if the URL is invalid.

**`AuthenticateWithURL(authURL string)` (method of `Identity`)**: This function authenticates using an SFDX authentication URL. It:

1.  Validates the provided auth URL using `ParseAuthURL`.
2.  Writes the auth URL to a temporary file with restricted permissions (0600).
3.  Executes the `sf org login sfdx-url` command using the Salesforce CLI (SFDX).
4.  Parses the JSON response from the SFDX command to extract the Organization ID.
5.  Securely cleans up the temporary auth URL file by overwriting it with zeros before deleting it.

### Error Handling

The service employs robust error handling:

*   Functions return an error value to indicate failure.
*   Errors are wrapped using `fmt.Errorf` to provide context and preserve the original error.
*   Transient errors (e.g., Salesforce API failures) are handled with retries and exponential backoff.
*   Input validation is performed to prevent invalid data from being processed.

### Concurrency

This service does not explicitly use goroutines or channels. However, the `exec.Command` function used to interact with the Salesforce CLI may internally use concurrency.

### Design Decisions

*   **Secure File Handling:** The JWT key and authentication URL are written to temporary files with restricted permissions (0600) to prevent unauthorized access. These files are securely cleaned up after use by overwriting their contents with zeros before deletion.
*   **SFDX CLI Dependency:** The service relies on the Salesforce CLI (SFDX) being installed and configured on the system.
*   **Retry Mechanism:** A retry mechanism with exponential backoff is implemented for the JWT authentication process to handle transient Salesforce API failures.
*   **Temporary Files:** The use of temporary files is a design choice to avoid storing sensitive information in memory for extended periods.
*   **Instance URL Handling:** The `InstanceURL` is optional in the `AuthRequest` structure, allowing for flexibility in authentication scenarios.