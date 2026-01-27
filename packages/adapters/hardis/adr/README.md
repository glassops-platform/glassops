# GlassOps SFDX-Hardis Adapter - Architecture Decision Records

This directory contains Architecture Decision Records (ADRs) specific to the SFDX-Hardis Adapter.

## Scope

ADRs in this directory cover decisions related to:

- sfdx-hardis CLI integration patterns
- Quality gate mappings to SARIF
- Branch protection policy enforcement
- Metadata quality checks
- SARIF normalization from hardis outputs

## Current ADRs

| Number | Title         | Status | Date |
| ------ | ------------- | ------ | ---- |
| -      | _No ADRs yet_ | -      | -    |

## Proposed Decisions to Document

1. **Hardis Quality Gates → SARIF Mapping** - How to map hardis checks to SARIF results
2. **Branch Protection Enforcement** - When to block vs warn
3. **Metadata Quality Thresholds** - What constitutes a violation
4. **Integration with Native Adapter** - How to combine hardis + native results
5. **Performance vs Coverage Trade-off** - Skip checks for speed?

## Adapter-Specific Context

**Purpose:** Extends native Salesforce adapter with advanced quality gates via sfdx-hardis

**Key Characteristics:**

- Wraps sfdx-hardis quality commands
- Emits SARIF 2.1.0 contracts
- Enforces metadata quality standards
- Supports branch protection policies

**Integration Points:**

- Consumes: sfdx-hardis CLI
- Emits: SARIF 2.1.0 contracts to `.glassops/glassops-contract.sarif.json`
- Can combine with: Native adapter results

## Creating a New ADR

See the [Master ADR Index](../../../../docs/adr-index.md) for the template and guidelines.

---

**Related Documentation:**

- [Master ADR Index](../../../../docs/adr-index.md)
- [Platform ADRs](../../../../docs/adr)
- [Adapter Development Guide](../../docs/adapter-development.md)
- [SFDX-Hardis Adapter Package README](../README.md)
- [SFDX-Hardis Documentation](https://sfdx-hardis.cloudity.com/)
