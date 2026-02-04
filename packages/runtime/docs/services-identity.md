---
type: Documentation
domain: runtime
last_modified: 2026-02-03
generated: true
source: packages/runtime/internal/services/identity.go
generated_at: 2026-02-03T18:08:40.579277
hash: 9863a8874eea37845219620cdfa638d85ffbc79826a153f9950a5013b8567655
---

## Identity Service Documentation

This document describes the Identity Service, responsible for handling authentication with Salesforce. It provides functionality for authenticating using both JWT (JSON Web Token) and SFDX (Salesforce DX) authentication URLs.

### Package Responsibilities

The `services` package contains the core logic for interacting with Salesforce identity management. Specifically, this module focuses on:

*   Authenticating with Salesforce using a JWT key.
*   Parsing and validating Salesforce SFDX authentication URLs.
*   Authenticating with Salesforce using an SFDX authentication URL.
*   Retrieving the Salesforce Organization ID after successful authentication.

### Key Types and Interfaces

*   **`AuthRequest`**: This structure encapsulates the parameters required for JWT-based authentication. It contains the following fields:
    *   `ClientID`: The Salesforce Connected App Client ID.
    *   `JWTKey`: The private key used to generate the JWT.
    *   `Username`: The Salesforce username.
    *   `InstanceURL`: The Salesforce instance URL (optional).

*   **`IdentityResolver`**: This type provides methods for JWT authentication. It interacts directly with the Salesforce CLI (`sf`) to perform the authentication process.

*   **`Identity`**: This type provides methods for authenticating using SFDX auth URLs and parsing those URLs. It also interacts with the Salesforce CLI.

### Important Functions

**`NewIdentityResolver()`**:
This function creates and returns a new instance of the `IdentityResolver` type. You can use this to obtain an instance ready for JWT authentication.

**`Authenticate(req AuthRequest)`**: (Method of `IdentityResolver`)
This function performs JWT-based authentication with Salesforce. It takes an `AuthRequest` as input and returns the Salesforce Organization ID upon successful authentication, along with any potential error. The process involves:

1.  Sanitizing the JWT key to handle escaped newline characters.
2.  Writing the JWT key to a temporary file with restricted permissions (0600).
3.  Executing the `sf org login jwt` command with the provided parameters.
4.  Retrying the command up to three times for transient Salesforce API failures, with exponential backoff.
5.  Parsing the JSON response from the Salesforce CLI to extract the Organization ID.
6.  Securely deleting the temporary JWT key file by overwriting it with zeros before removing it.

**`NewIdentity()`**:
This function creates and returns a new instance of the `Identity` type. You can use this to obtain an instance ready for SFDX URL authentication.

**`ParseAuthURL(authURL string)`**: (Method of `Identity`)
This function parses and validates a Salesforce SFDX authentication URL. It checks the URL format and validates the instance URL against a regular expression to ensure it is a valid Salesforce domain. It returns the instance URL if the validation is successful.

**`AuthenticateWithURL(authURL string)`**: (Method of `Identity`)
This function authenticates with Salesforce using an SFDX authentication URL. It first validates the URL using `ParseAuthURL`. Then, it writes the URL to a temporary file, executes the `sf org login sfdx-url` command, parses the JSON response to extract the Organization ID, and securely deletes the temporary file.

### Error Handling

The service employs robust error handling:

*   Functions return both a result and an error value.
*   Errors are wrapped using `fmt.Errorf` to provide context and preserve the original error.
*   Transient errors (Salesforce API failures) are handled with retries and exponential backoff.
*   Detailed error messages from the Salesforce CLI are logged for debugging purposes.
*   Temporary files are securely cleaned up even in the event of errors.

### Concurrency

This service does not explicitly use goroutines or channels. However, the `exec.Command` function used to interact with the Salesforce CLI may execute in a separate process.

### Design Decisions

*   **Temporary Files:** The use of temporary files for storing the JWT key and SFDX auth URL is a security measure to avoid exposing sensitive information in memory or process listings.
*   **Secure File Deletion:** Overwriting temporary files with zeros before deletion ensures that the data is unrecoverable.
*   **Salesforce CLI Dependency:** The service relies on the Salesforce CLI (`sf`) being installed and configured on the system.
*   **Retry Mechanism:** The retry mechanism for JWT authentication improves resilience to transient Salesforce API failures.
*   **Error Wrapping:** Wrapping errors provides valuable context for debugging and troubleshooting.