# GlassOps Runtime

> **The Bootstrapping Primitive for Secure, Governed Execution.**

**Status:** Alpha (Go Implementation)
**Version:** 1.0.0

The Runtime is the trusted core of the GlassOps ecosystem. It bootstraps a secure environment, validates identity, enforces policy, and hands off a signed **Permit** to downstream adapters.

---

## Architecture

The Runtime executes in **3 Strictly Defined Phases**:

1.  **Phase 1: Policy Enforcement**
    - Loads `devops-config.json`.
    - Enforces "Freeze Windows" (e.g., No deploys on Friday).
    - Validates static analysis results (if enabled).

2.  **Phase 2: Identity Resolution**
    - Exchanges GitHub Secrets (`JWT_KEY`) for a Salesforce session token.
    - **Security:** Secrets never leave this phase. Downstream adapters receiving the Permit get a verified session ID, not the private key.

3.  **Phase 3: Context Handoff (Permit Generation)**
    - Generates the **Permit Contract** (`.glassops/glassops-permit.json`).
    - This contract contains the **Verified Identity** and **Policy Evaluation Results**.

---

## Contract Artifacts

The Runtime produces ephemeral artifacts in the `.glassops/` directory (gitignored):

- **`glassops-permit.json`**: The "ticket" that authorizes an Adapter to perform work.
- **`glassops-contract.json`**: The initial compliance contract (deployment status, audit trail).

---

## Usage (GitHub Actions)

```yaml
- name: GlassOps Runtime
  uses: glassops-platform/glassops/packages/runtime@v1
  with:
      client_id: ${{ secrets.SF_CLIENT_ID }}
      jwt_key: ${{ secrets.SF_JWT_KEY }}
      username: ${{ vars.SF_USERNAME }}
```

### Inputs

| Input          | Description                                         | Required |
| :------------- | :-------------------------------------------------- | :------- |
| `jwt_key`      | PEM-encoded private key                             | Yes      |
| `client_id`    | Connected App Consumer Key                          | Yes      |
| `username`     | Salesforce Username                                 | Yes      |
| `instance_url` | Login URL (default: `https://login.salesforce.com`) | No       |
| `skip_auth`    | Set to `true` to skip Salesforce auth (for testing) | No       |

### Outputs

| Output           | Description                             |
| :--------------- | :-------------------------------------- |
| `glassops_ready` | `true` if runtime suceeded.             |
| `org_id`         | The authenticated Organization ID.      |
| `runtime_id`     | Unique UUID for this execution session. |
| `is_locked`      | `true` if a freeze window is active.    |

---

## Local Development

We provide first-class support for local debugging and testing.

### Prerequisites

- Go 1.21+
- Salesforce CLI (`sf`)

### Helper Scripts (Root)

- **Build**: `npm run runtime:build` (Builds to `dist/glassops.exe`)
- **Test**: `npm run runtime:test` (Runs Go unit tests)

### debugging in VS Code

1.  Create a `.env` file in the root (see `.env.example` logic).
2.  **Key File Support**: To avoid multiline env var issues, point `INPUT_JWT_KEY_FILE` to your local `secrets/server.key`.
3.  Press **F5** (Debug GlassOps) to run the runtime with breakpoints.
