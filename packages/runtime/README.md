# GlassOps Runtime

> [!CAUTION]
> **DEPRECATED:** The TypeScript implementation of the runtime is now fully deprecated.
> All new development has moved to the **Go-based architecture**.
>
> For the legacy TypeScript implementation, see [runtime-ts](../runtime-ts).

> **The Bootstrapping Primitive for Secure, Governed Execution.**

**Status:** Alpha (Re-implemented in Go)
**Version:** 2.0.0

The Runtime is designed as a trusted primitive in the ecosystem. It is responsible for bootstrapping a secure environment where untrusted adapters can run.

---

## Architecture

The Runtime executes in **4 Strictly Defined Phases**:

1.  **Phase 1: Identity Resolution**
    - Exchanges GitHub Secrets (`JWT_KEY`) for a session token.
    - **Security:** Secrets never leave this phase. Downstream adapters only receive a masked session ID.

2.  **Phase 2: Policy Enforcement**
    - Checks "Freeze Windows" (e.g., No deploys on Friday).
    - Checks "Environment Type" (e.g., Is this Production?).

3.  **Phase 3: Tool Bootstrap**
    - Installs the exact version of Salesforce CLI required by policy.
    - Installs `sfdx-git-delta` or other core dependencies.

4.  **Phase 4: Handoff**
    - Emits the `glassops_context` object to the GitHub Runner.
    - Ready for Adapter Execution.

---

## Usage

```yaml
- name: Boostrap Runtime
  uses: glassops-platform/glassops/packages/runtime@v1
  with:
      client_id: ${{ secrets.SF_CLIENT_ID }}
      jwt_key: ${{ secrets.SF_JWT_KEY }}
      username: ${{ vars.SF_USERNAME }}
```

---

## Inputs & Outputs

| Input       | Required | Description                       |
| ----------- | -------- | --------------------------------- |
| `jwt_key`   | Yes      | The PEM-encoded private key.      |
| `client_id` | Yes      | Connected App Consumer Key.       |
| `strict`    | No       | Fail on warnings? (Default: true) |

| Output          | Description                                  |
| --------------- | -------------------------------------------- |
| `org_id`        | The 18-char Organization ID.                 |
| `is_production` | `true` if this org is flagged as Production. |

---

## Security Model

- **Zero-Trust:** The Runtime assumes the environment is hostile. It validates every input.
- **Secret Isolation:** This Action is the _only_ one that requires the Private Key. Adapters generally do not need the private key if they reuse the session.
