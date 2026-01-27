# GlassOps Protocol Specification (`@glassops-platform/glassspec`)

> **The Universal Language of Governance.**
>
> This package defines the interfaces, schemas, and contracts that allow independent components to work together.

---

## 1. The Layered Contract Model

GlassOps uses different standards for different signal types to avoid "Square Peg, Round Hole" problems.

| Layer             | Standard            | Component     | Purpose                                              |
| ----------------- | ------------------- | ------------- | ---------------------------------------------------- |
| **1. Governance** | **SARIF 2.1.0**     | Policy Engine | Static analysis, findings, violations.               |
| **2. Telemetry**  | **OpenTelemetry**   | Observability | Metrics (CPU, Duration) and Traces.                  |
| **3. Transport**  | **CloudEvents**     | Event Bus     | Lifecycle notifications ("Deploy Started").          |
| **4. Native**     | **Native JSON/XML** | Edge Storage  | Raw tool data (AWS CloudTrail, Salesforce Metadata). |

### The Invariant

> **"SARIF is Authoritative."**
> If a finding is not in the SARIF contract, it does not exist for governance purposes.

---

## 2. Adapter Interface

GlassOps Adapters are stateless workers that translate Native output into the Layered Contract.

### 2.1 Command Line Contract

To be a compliant adapter, the container must accept standard arguments:

```bash
# INPUT: The Policy to enforce
--policy-in <path/to/policy.json>

# OUTPUT: The Contract to emit
--contract-out <path/to/contract.json>
```

### 2.2 Behavior Contract

1.  **Stateless**: Adapters must not store state between runs.
2.  **Swallow Crashes**: If the underlying tool crashes, the Adapter MUST catch it and emit a `Failed` Contract, not just exit 1.
3.  **No Bypass**: The Adapter must not offer flags to skip governance checks.
4.  **Atomic Write**: Write to a temp file and rename to avoid partial reads.

### 2.3 Schema Mapping (Native -> SARIF)

Adapters must map tool concepts to SARIF fields:

| Tool Concept     | SARIF Field            | Notes                              |
| ---------------- | ---------------------- | ---------------------------------- |
| **Violation ID** | `result.ruleId`        | Must be stable and queryable.      |
| **Severity**     | `result.level`         | Map to `error`, `warning`, `note`. |
| **Message**      | `result.message.text`  | Human-readable explanation.        |
| **File Path**    | `physicalLocation.uri` | Relative path from repo root.      |

---

## 3. Deployment Contract Schema

**File:** `.glassops/deployment-contract.json`

Every adapter determines the deployment status by emitting this contract.

```json
{
    "schemaVersion": "1.0",
    "meta": {
        "adapter": "glassops-native-adapter",
        "version": "1.0.0",
        "timestamp": "2026-01-25T12:00:00Z"
    },
    "status": "Succeeded",
    "quality": {
        "coverage": {
            "actual": 85,
            "required": 75,
            "met": true
        },
        "tests": {
            "total": 120,
            "failed": 0
        }
    },
    "policy": {
        "source": { "githubFloor": 75 },
        "effective": 75
    },
    "results": [
        // Array of SARIF result objects
        {
            "ruleId": "GOV-001",
            "level": "note",
            "message": { "text": "Deployment passed all checks." }
        }
    ]
}
```

---

## 4. Contributing

See `schemas/` directory for raw JSON Schema definitions.
