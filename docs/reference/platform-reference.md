# Platform Reference & API

> **The Normative Specification for GlassOps Components.**

This document defines the inputs, outputs, and interfaces for the core system. For architecture, see [overview.md](../architecture/overview.md).

---

## 1. Runtime Primitive (`@glassops-platform/runtime`)

**Location:** `packages/runtime`
**Purpose:** Bootstraps the secure environment.

### Inputs

| Input          | Required | Description                   |
| -------------- | -------- | ----------------------------- |
| `jwt_key`      | **Yes**  | PEM Encrypted Private Key     |
| `client_id`    | **Yes**  | Connected App Consumer Key    |
| `username`     | **Yes**  | Target Username               |
| `instance_url` | No       | Default: login.salesforce.com |

### Outputs

| Output          | Description                   |
| --------------- | ----------------------------- |
| `org_id`        | 18-char Organization ID       |
| `is_production` | Boolean flag for safety gates |
| `session_id`    | (Masked) Access Token         |

---

## 2. Deployment Contract Schema

**Authority:** `packages/glassspec/README.md`

Every adapter MUST emit a contract matching this schema (simplified):

```json
{
    "schemaVersion": "1.0",
    "meta": { "adapter": "native", "timestamp": "..." },
    "status": "Succeeded",
    "policy": {
        "effective": { "minCoverage": 75 },
        "met": true
    },
    "results": [{ "ruleId": "COVERAGE", "level": "error", "message": "72% < 75%" }]
}
```

---

## 3. Policy Resolution API

**Location:** `packages/control-plane`

### Resolution Logic

```javascript
EffectivePolicy = MAX(GitHub_Env_Var(Floor), Repo_Config_File(Team), Salesforce_CMDT(Organization));
```

**Invariant:** No source can lower the GitHub Floor.

---

## 4. Adapter Interface

To be a compliant adapter, a container must:

1. Accept `--policy-in` (JSON path).
2. Emit `--contract-out` (JSON path).
3. Exit `0` for success, `1` for crash (System Error).
4. Exit `0` for Policy Failure (Logic Error) but mark Contract status as `Failed`.
