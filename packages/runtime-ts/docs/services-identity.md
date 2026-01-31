---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/services/identity.ts
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/src/services/identity.ts
generated_at: 2026-01-31T09:17:07.844370
hash: 7da4458a3bb94c021316e184cd6743876ac04c4f59fee21885b081066399ee60
---

## Identity Resolver Service Documentation

**Introduction**

This document details the Identity Resolver service, a component designed to authenticate with an organization using a JSON Web Token (JWT) and retrieve the organization’s ID. This service is intended for automated processes requiring secure access to an organization’s resources.

**Functionality**

The Identity Resolver service handles the authentication process against an organization’s security infrastructure. It accepts authentication credentials and securely obtains an organization ID, which can then be used for subsequent operations. The service incorporates retry logic to handle transient API failures.

**Key Concepts**

* **JWT Authentication:** The service authenticates using a JWT, a standard for securely transmitting information between parties as a JSON object.
* **Organization ID:** A unique identifier for the target organization. This ID is the primary output of the authentication process.
* **Transient Errors:** Temporary issues encountered when communicating with the organization’s API. The service is designed to automatically retry these operations.

**Inputs**

The `authenticate` function requires an `AuthRequest` object containing the following properties:

* `clientId`: The client ID associated with the JWT. (String)
* `jwtKey`: The private key used to generate the JWT. (String)
* `username`: The username associated with the JWT. (String)
* `instanceUrl` (Optional): The URL of the organization’s instance. If not provided, the service will attempt to connect to the default instance. (String)

**Outputs**

Upon successful authentication, the `authenticate` function returns the organization ID as a string.

**Error Handling**

If authentication fails, the `authenticate` function throws an error with a descriptive message. Common causes of failure include invalid client IDs or JWT keys.

**Security Considerations**

* **JWT Key Management:** The JWT key is written to a temporary file during the authentication process.
* **Secure File Handling:** Before deletion, the temporary file containing the JWT key is overwritten with zeros to prevent data recovery.
* **Permissions:** The temporary file is created with restricted permissions (mode 0o600) to limit access.

**Usage**

To authenticate and retrieve the organization ID, you must:

1.  Create an `AuthRequest` object with the necessary credentials.
2.  Call the `authenticate` function with the `AuthRequest` object.
3.  Handle the returned organization ID or any potential errors.

**Implementation Details**

The service leverages the following dependencies:

* `@actions/core`: For logging and managing environment variables.
* `@actions/exec`: For executing external commands (specifically, the `sf` CLI).
* `fs`: For file system operations (writing and deleting the temporary JWT key file).
* `path`: For constructing file paths.
* `os`: For accessing operating system-specific information (like the temporary directory).

The service uses the `sf` command-line interface (CLI) to perform the actual authentication. It constructs the appropriate command arguments based on the provided input and parses the JSON output from the CLI.

**Retry Mechanism**

The service incorporates a retry mechanism to handle transient errors during the authentication process. It will attempt the authentication up to three times, with a two-second delay between each attempt. This improves the reliability of the service in environments with intermittent network connectivity or API availability issues.