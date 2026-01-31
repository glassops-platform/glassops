---
type: Documentation
domain: runtime
origin: packages/runtime/internal/services/identity.go
last_modified: 2026-01-31
generated: true
source: packages/runtime/internal/services/identity.go
generated_at: 2026-01-31T09:09:23.437771
hash: 3940eb183960a36e834ef630b97962783208d0a0acd53fd5522e36dcb1c8a15d
---

## Identity Service Documentation

This document details the Identity Service, responsible for handling authentication with Salesforce. It provides functionality for authenticating using both JWT (JSON Web Token) and SFDX Auth URLs.

**Package Purpose:**

The `services` package contains core services used within the platform. The Identity Service specifically manages the process of authenticating with Salesforce, obtaining an Organization ID (OrgID) which is then used for subsequent operations.

**Key Types and Interfaces:**

*   **`AuthRequest`:** A structure that encapsulates the parameters required for JWT-based authentication. It contains the `ClientID`, `JWTKey`, `Username`, and `InstanceURL` associated with the Salesforce connection.
*   **`IdentityResolver`:** This type handles JWT authentication. It provides a method to authenticate against Salesforce using the provided credentials.
*   **`Identity`:** This type handles authentication via SFDX Auth URLs. It includes methods for parsing and authenticating using these URLs.
*   **No Interfaces:** The service does not currently define any interfaces for abstraction.

**Important Functions and Behavior:**

*   **`NewIdentityResolver()`:**  A constructor function that returns a new instance of the `IdentityResolver`.
*   **`Authenticate(req AuthRequest) (string, error)` (on `IdentityResolver`):** This function performs JWT-based authentication with Salesforce. It takes an `AuthRequest` as input, writes the JWT key to a temporary file with restricted permissions, and then executes the `sf org login jwt` command. It retries the operation up to three times in case of transient Salesforce API errors. Upon successful authentication, it parses the JSON response to extract the OrgID and returns it.  It includes secure cleanup of the temporary JWT key file by overwriting it with zeros before deletion.
*   **`NewIdentity()`:** A constructor function that returns a new instance of the `Identity`.
*   **`ParseAuthURL(authURL string) (string, error)` (on `Identity`):** This function validates and parses a Salesforce SFDX Auth URL. It checks for the correct format (`force://...`) and verifies that the instance URL is a valid Salesforce domain. It returns the instance URL if the validation is successful.
*   **`AuthenticateWithURL(authURL string) (string, error)` (on `Identity`):** This function authenticates with Salesforce using an SFDX Auth URL. It first validates the URL using `ParseAuthURL`. It then writes the URL to a temporary file, executes the `sf org login sfdx-url` command, and parses the JSON response to extract the OrgID. It also includes secure cleanup of the temporary Auth URL file.

**Error Handling:**

The service employs robust error handling. Functions return both a result (OrgID) and an error value. Errors are wrapped using `fmt.Errorf` with `%w` to preserve the original error context, allowing for easier debugging and error propagation. Specific error messages are provided for common issues like invalid URL formats, file write failures, and authentication failures.

**Concurrency:**

The service does not explicitly use goroutines or channels for concurrent operations. However, the `Authenticate` function incorporates a retry mechanism with exponential backoff, which implicitly introduces a form of controlled concurrency in handling transient errors.

**Notable Design Decisions:**

*   **Temporary File Usage:** The service writes sensitive information (JWT key and Auth URL) to temporary files to interact with the Salesforce CLI (`sf`). This approach allows the CLI to handle the authentication process securely.
*   **Secure File Cleanup:**  After use, temporary files containing sensitive data are securely cleaned up by first overwriting their contents with zeros before deleting them, mitigating the risk of data leakage.
*   **Salesforce CLI Dependency:** The service relies on the Salesforce CLI (`sf`) being installed and configured on the system where it is running.
*   **Retry Mechanism:** The `Authenticate` function includes a retry mechanism to handle transient errors from the Salesforce API, improving the reliability of the authentication process.
*   **JSON Parsing:** The service uses `json.Unmarshal` to parse the output from the Salesforce CLI, ensuring structured access to the authentication results.