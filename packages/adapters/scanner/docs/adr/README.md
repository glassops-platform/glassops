# GlassOps Scanner Adapter - Architecture Decision Records

This directory contains Architecture Decision Records (ADRs) specific to the Scanner Adapter.

## Scope

ADRs in this directory cover decisions related to:

- Static analysis tool integration (MegaLinter, Code Analyzer, etc.)
- SARIF aggregation from multiple scanners
- Scanner selection and prioritization
- Performance optimization (parallel execution)
- False positive filtering

## Current ADRs

| Number | Title         | Status | Date |
| ------ | ------------- | ------ | ---- |
| -      | _No ADRs yet_ | -      | -    |

## Proposed Decisions to Document

1. **MegaLinter vs Individual Linters** - Orchestrator vs direct integration
2. **SARIF Aggregation Strategy** - How to merge multiple SARIF outputs
3. **Scanner Priority/Weighting** - Which findings take precedence
4. **Performance Optimization** - Parallel vs sequential execution
5. **Baseline/Suppression Strategy** - How to handle false positives

## Adapter-Specific Context

**Purpose:** Provides static analysis governance by orchestrating code quality scanners

**Key Characteristics:**

- Orchestrates multiple static analysis tools
- Aggregates SARIF outputs from scanners
- Normalizes findings to common severity levels
- Supports incremental/differential scanning

**Integration Points:**

- Consumes: MegaLinter, Salesforce Code Analyzer, ESLint, PMD, etc.
- Emits: Aggregated SARIF 2.1.0 contracts to `.glassops/glassops-contract.sarif.json`
- Can filter: Based on baseline or suppression rules

## Creating a New ADR

See the [Master ADR Index](../../../../../docs/adr-index.md) for the template and guidelines.

---

**Related Documentation:**

- [Master ADR Index](../../../../../docs/adr-index.md)
- [Platform ADRs](../../../../../docs/adr)
- [Adapter Development Guide](../../../docs/adapter-development.md)
- [Scanner Adapter README](../../README.md)
- [MegaLinter Adapter Spec](../../../megalinter/README.md)
