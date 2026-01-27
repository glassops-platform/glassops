---
type: Documentation
domain: runtime
origin: packages/runtime/src/services/identity.ts
last_modified: 2026-01-26
generated: true
source: packages/runtime/src/services/identity.ts
generated_at: 2026-01-26T14:21:30.363Z
hash: b2f611e98c51d3110eedded48910994c38f58824770583b1d699d496877ce3e5
---

## Identity Resolver Service Documentation

**Overview**

The Identity Resolver service manages authentication with an organization using Salesforce CLI (sf). It securely authenticates using a JSON Web Token (JWT) and retrieves the organization ID. This service is designed for automated processes requiring access to a Salesforce organization.

**Functionality**

The core function of this service is to authenticate against a Salesforce instance and obtain the organization ID. It achieves this by:

1.  **Secure Key Storage:**  A temporary, securely permissioned file is created to store the provided JWT key. This key is essential for authentication.
2.  **Salesforce CLI Interaction:** The service executes the `sf org login jwt` command with the provided credentials and key. This command handles the authentication process with Salesforce.
3.  **Response Parsing:** The output from the Salesforce CLI is parsed to extract the organization ID and access token.
4.  **Secure Key Deletion:** After authentication, the temporary JWT key file is immediately deleted to maintain security.
5.  **Error Handling:**  Robust error handling is implemented to catch authentication failures and provide informative error messages.

**Inputs**

The `authenticate` function accepts a single object, `AuthRequest`, with the following properties:

*   `clientId`: (String) The Salesforce Connected App Client ID. This identifies the application authenticating.
*   `jwtKey`: (String) The private key used to generate the JWT.  This key must be securely managed.
*   `username`: (String) The Salesforce username associated with the JWT.
*   `instanceUrl`: (Optional String) The Salesforce instance URL. If not provided, the service will use the default Salesforce instance.

**Outputs**

Upon successful authentication, the `authenticate` function returns a Promise that resolves with the Salesforce Organization ID (String).

**Error Handling**

If authentication fails, the `authenticate` function throws an Error with the message "Authentication Failed. Check Client ID and JWT Key."  This indicates a problem with the provided credentials or key.

**Usage**

To authenticate, you must first create an instance of the `IdentityResolver` class. Then, call the `authenticate` method with an `AuthRequest` object containing the necessary authentication details.

```typescript
import { IdentityResolver } from './identity';

const resolver = new IdentityResolver();

const authRequest = {
  clientId: 'your_client_id',
  jwtKey: 'your_jwt_key',
  username: 'your_username',
  instanceUrl: 'https://your-instance.salesforce.com' // Optional
};

resolver.authenticate(authRequest)
  .then(orgId => {
    console.log(`Successfully authenticated with org ID: ${orgId}`);
  })
  .catch(error => {
    console.error('Authentication error:', error.message);
  });
```

**Security Considerations**

*   The JWT key is stored in a temporary file with restricted permissions (0o600) to prevent unauthorized access.
*   The temporary key file is deleted immediately after authentication.
*   Ensure the JWT key is securely managed and never committed to source control.
*   Protect the `clientId` and `username` as sensitive information.