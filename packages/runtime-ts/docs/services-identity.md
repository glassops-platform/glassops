---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/services/identity.ts
last_modified: 2026-01-29
generated: true
source: packages/runtime-ts/src/services/identity.ts
generated_at: 2026-01-29T21:00:16.061344
hash: 7da4458a3bb94c021316e184cd6743876ac04c4f59fee21885b081066399ee60
---

## Identity Resolver Service Documentation

**Introduction**

This document details the Identity Resolver service, a component designed to authenticate with an organization using a JSON Web Token (JWT) and retrieve the organization’s ID. This service is intended for automated processes requiring secure access to an organization’s resources.

**Functionality**

The Identity Resolver service handles the authentication process against an organization’s security infrastructure. It accepts authentication credentials and securely obtains an organization ID, which can then be used for subsequent operations. The service incorporates retry logic to handle transient errors during authentication.

**Key Concepts**

* **JWT Authentication:** The service authenticates using a JWT, a standard for securely transmitting information between parties as a JSON object.
* **Organization ID:** A unique identifier for the target organization. This ID is the primary output of the authentication process.
* **Retry Logic:** The service automatically retries the authentication process a configurable number of times to mitigate intermittent failures.
* **Secure Key Handling:** The JWT key is written to a temporary file with restricted permissions and securely deleted after use.

**Inputs**

The `authenticate` function requires an `AuthRequest` object containing the following properties:

* **clientId:** (String) The client ID associated with the JWT.
* **jwtKey:** (String) The private key used to generate the JWT.
* **username:** (String) The username associated with the JWT.
* **instanceUrl:** (Optional String) The instance URL of the organization. If not provided, the service will use the default instance.

**Outputs**

The `authenticate` function returns a Promise that resolves with the organization ID (String) upon successful authentication.

**Error Handling**

If authentication fails, the `authenticate` function throws an error with a descriptive message. Common causes of failure include invalid client ID, an incorrect JWT key, or network connectivity issues.

**Security Considerations**

* The JWT key is written to a temporary file with permissions set to 600 (read/write for the owner only).
* Before deletion, the temporary file containing the JWT key is overwritten with zeros to prevent data recovery.
* All sensitive information is handled in memory whenever possible.

**Usage**

To authenticate and retrieve the organization ID, you must first create an `IdentityResolver` instance. Then, call the `authenticate` function with a valid `AuthRequest` object.

```typescript
import { IdentityResolver } from './identity';

const resolver = new IdentityResolver();

const authRequest = {
  clientId: 'your_client_id',
  jwtKey: 'your_jwt_key',
  username: 'your_username',
  instanceUrl: 'https://your-instance.salesforce.com' // Optional
};

try {
  const orgId = await resolver.authenticate(authRequest);
  console.log(`Successfully authenticated with Org ID: ${orgId}`);
} catch (error) {
  console.error('Authentication failed:', error);
}
```

**Configuration**

The number of retries and the backoff time between retries are configurable within the `executeWithRetry` function. Currently, the service is configured for a maximum of 3 retries with a 2-second backoff.

**Dependencies**

* `@actions/core`
* `@actions/exec`
* `fs`
* `path`
* `os`