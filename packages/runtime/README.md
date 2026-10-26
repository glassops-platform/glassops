# GlassOps Runtime

> **A Go GitHub Action that checks policy, authenticates to Salesforce, and writes a permit and a contract file.**

**Status:** Early, unreleased. No tag or release exists.

The runtime is the only working component in this repository. The project is being rebuilt as a CI-agnostic command line tool, and this action is the starting point for that work.

---

## What It Does

The runtime runs these steps in order:

1.  **Policy.** Loads `devops-config.json` (by default `config/devops-config.json` under `GITHUB_WORKSPACE`). Checks freeze windows when `enforce_policy` is `true`. Runs the Salesforce CLI code analyzer when the config enables it.
2.  **Identity.** Authenticates to the target org with the JWT flow. The private key is only read inside the action container.
3.  **Permit.** Writes `.glassops/glassops-permit.json`. The file records the actor, the permit ID and the policy checks that were listed. It is unsigned and has no expiry, so it is not proof of anything yet.
4.  **Contract.** Writes `.glassops/glassops-contract.json`, a deployment contract record.

The runtime does not deploy. Authentication happens inside the action container, so later workflow steps do not receive a session or any credential.

---

## Contract Artifacts

The runtime writes these files to the `.glassops/` directory (gitignored):

- **`glassops-permit.json`**: Actor, permit ID and policy evaluation record. Unsigned.
- **`glassops-contract.json`**: Deployment contract record (status, coverage, tests, audit fields).

---

## Usage (GitHub Actions)

Not yet released. There is no tag to pin, so use the action from a local checkout of this repository:

```yaml
- name: GlassOps Runtime
  uses: ./packages/runtime
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
| `enforce_policy` | Default `true`. Checks freeze windows             | No       |
| `test_results` | JSON string: `{"total": 100, "passed": 95, "failed": 5}` | No |
| `coverage_percentage` | Code coverage, 0 to 100                      | No       |
| `coverage_required` | Required coverage, default `80`                | No       |
| `skip_auth`    | Set to `true` to skip Salesforce auth (for testing) | No       |

### Outputs

| Output           | Description                             |
| :--------------- | :-------------------------------------- |
| `glassops_ready` | `true` if the runtime succeeded.        |
| `org_id`         | The authenticated Organization ID.      |
| `runtime_id`     | Unique UUID for this execution session. |
| `is_locked`      | `true` if a freeze window is active.    |
| `contract_path`  | Path to the generated contract file.    |

---

## Local Development

We provide first-class support for local debugging and testing.

### Prerequisites

- Go 1.25 (see `go.mod`)
- Salesforce CLI (`sf`)

### Helper Scripts (Root)

- **Build**: `npm run runtime:build` (Builds to `dist/glassops.exe` at the repo root)
- **Test**: `npm run runtime:test` (Runs Go unit tests)

### debugging in VS Code

1.  Create a `.env` file in the root (see `.env.example` logic).
2.  **Key File Support**: To avoid multiline env var issues, point `INPUT_JWT_KEY_FILE` to your local `secrets/server.key`.
3.  Press **F5** (Debug GlassOps) to run the runtime with breakpoints.
