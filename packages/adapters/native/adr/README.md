# GlassOps Native Adapter - Architecture Decision Records

This directory contains Architecture Decision Records (ADRs) specific to the Native Salesforce Adapter.

## Scope

ADRs in this directory cover decisions related to:

- Salesforce CLI integration patterns
- JWT authentication strategy
- Deployment metadata extraction
- Test result normalization to SARIF
- Coverage threshold enforcement
- Error handling and retry logic

## Current ADRs

| Number | Title         | Status | Date |
| ------ | ------------- | ------ | ---- |
| -      | _No ADRs yet_ | -      | -    |

## Proposed Decisions to Document

1. **sf CLI vs @salesforce/core** - Direct CLI wrapper vs library usage
2. **JWT Key Management** - How to securely handle server.key
3. **SARIF Mapping Strategy** - Deployment results → SARIF structure
4. **Coverage Calculation** - Org-wide vs package-specific
5. **Deployment Timeout Handling** - Async deployments and status polling

## Adapter-Specific Context

**Purpose:** Provides native Salesforce deployment governance via `sf project deploy` commands

**Key Characteristics:**

- Protocol-compliant SARIF emission
- Salesforce-native authentication (JWT-OAuth)
- Supports both mdapi and source formats
- Integrates with Salesforce test execution

**Integration Points:**

- Consumes: Salesforce CLI (`sf`)
- Emits: SARIF 2.1.0 contracts to `.glassops/glassops-contract.sarif.json`
- Authenticates: Via JWT (server.key + client ID)

## Creating a New ADR

See the [Master ADR Index](../../../../docs/adr-index.md) for the template and guidelines.

**Quick start:**

```bash
cd glassops-adapters/glassops-native-adapter/docs/adr
# Create new ADR
cp ../../../../docs/adr/adr-template.md 001-your-decision-title.md
# Edit the file
# Update this README
# Update ../../../../../docs/adr-index.md
```

---

**Related Documentation:**

- [Master ADR Index](../../../../docs/adr-index.md)
- [Platform ADRs](../../../../docs/adr)
- [Adapter Development Guide](../../docs/adapter-development.md)
- [Native Adapter Package README](../README.md)
