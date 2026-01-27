# GlassOps Configuration Guide

The `config/devops-config.json` file is the central nervous system of GlassOps. It defines the rules, thresholds, and behaviors that the Governance Protocol enforces.

## Structure

The configuration is divided into 4 main sections:

1.  `execution`: Defines which engine (Native/Hardis) runs the deployment.
2.  `governance`: Sets the baseline quality standards (e.g., Min Coverage).
3.  `environments`: Maps git branches to Salesforce orgs and sets per-environment rules.
4.  `notifications`: Controls Slack/Email alerting behavior.

## Schema Reference

The configuration is formally defined in two places:

1.  **Validation (JSON Schema):** [`devops-config.schema.json`](devops-config.schema.json) - Direct validation for your IDE.
2.  **Type Definition (TypeScript):** [`types/config.d.ts`](../packages/runtime/src/types/config.d.ts) - The authoratative interface for developers.

## Example Configuration

```json
{
    "version": "1.0",
    "execution": {
        "engine": "native",
        "fallback": "none"
    },
    "governance": {
        "minCoverage": 75,
        "requireTests": true
    },
    "environments": {
        "prod": {
            "branch_mapping": "main",
            "quality_gates": {
                "minCoverage": 85
            }
        }
    }
}
```
