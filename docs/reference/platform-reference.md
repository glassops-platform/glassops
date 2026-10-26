# Platform Reference & API

> **Reference for the one component that exists today: the runtime.**

This document lists the inputs and outputs of the runtime action. The contract, policy and adapter ideas from earlier drafts are not implemented and are not specified here. For architecture, see [overview.md](../architecture/overview.md).

---

## 1. Runtime (`packages/runtime`)

**Location:** `packages/runtime`
**Purpose:** Checks policy, authenticates to Salesforce, and writes a permit and a contract file.

### Inputs

| Input          | Required | Description                   |
| -------------- | -------- | ----------------------------- |
| `jwt_key`      | **Yes**  | PEM-encoded private key       |
| `client_id`    | **Yes**  | Connected App Consumer Key    |
| `username`     | **Yes**  | Target Username               |
| `instance_url` | No       | Default: login.salesforce.com |
| `enforce_policy` | No     | Default `true`. Checks freeze windows |
| `test_results` | No       | JSON string of test results   |
| `coverage_percentage` | No | Code coverage, 0 to 100      |
| `coverage_required` | No  | Default `80`                  |
| `skip_auth`    | No       | Default `false`. Skips Salesforce auth |

### Outputs

| Output           | Description                                  |
| ---------------- | -------------------------------------------- |
| `runtime_id`     | Unique ID for this run                       |
| `org_id`         | Authenticated Organization ID                |
| `is_locked`      | `true` if a freeze window is active          |
| `contract_path`  | Path to the generated contract JSON file     |
| `glassops_ready` | `true` when the runtime finished successfully |

The runtime does not output a session ID or a production flag. Authentication happens inside the action container, so no credential is handed to later steps.

---
