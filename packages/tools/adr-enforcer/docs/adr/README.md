# GlassOps ADR Enforcement Adapter - Architecture Decision Records

This directory contains Architecture Decision Records (ADRs) specific to the ADR Enforcement Adapter.

## Scope

ADRs in this directory cover decisions related to:

- ADR compliance detection strategies
- Decision drift identification
- SARIF mapping for governance violations
- ADR coverage metrics
- Enforcement vs advisory modes

## Current ADRs

| Number | Title         | Status | Date |
| ------ | ------------- | ------ | ---- |
| -      | _No ADRs yet_ | -      | -    |

## Proposed Decisions to Document

1. **ADR Detection Strategy** - How to identify when ADRs should exist but don't
2. **Decision Drift Detection** - Validate implementation matches documented decision
3. **Enforcement Mode** - Block vs warn on violations
4. **ADR Coverage Metrics** - What percentage of architecture is documented
5. **Meta-Governance** - How does the ADR adapter govern itself?

## Adapter-Specific Context

**Purpose:** Governance of governance - ensures architectural decisions are documented and followed

**Key Characteristics:**

- **Meta-Adapter** - Governs the governance process itself
- Scans codebase for architectural patterns
- Validates implementations match ADRs
- Emits SARIF findings for missing or violated ADRs
- Self-referential (this adapter needs its own ADRs!)

**Integration Points:**

- Consumes: ADR markdown files across all projects
- Analyzes: Codebase for architectural patterns
- Emits: SARIF 2.1.0 contracts to `.glassops/glassops-contract.sarif.json`
- Detects: Missing ADRs, decision drift, violated constraints

## Example Violations This Adapter Would Detect

### Missing ADR

```
SARIF Finding:
  - ruleId: ADR_MISSING
  - message: "Database technology changed from PostgreSQL to MongoDB but no ADR exists"
  - level: warning
  - location: db/connection.ts
```

### Decision Drift

```
SARIF Finding:
  - ruleId: ADR_VIOLATED
  - message: "ADR-005 requires Kubernetes operator pattern but code uses polling"
  - level: error
  - location: operator/controller.go
  - relatedLocations: [ADR-005.md]
```

### Outdated ADR

```
SARIF Finding:
  - ruleId: ADR_STALE
  - message: "ADR-003 status is 'Proposed' but implementation exists"
  - level: warning
  - location: docs/adr/003-additive-governance.md
```

## Detection Strategies (Proposed)

1. **Pattern Matching** - Regex/AST analysis for architectural patterns
2. **Git History** - Detect significant changes without corresponding ADRs
3. **Dependency Analysis** - New major dependencies should have ADRs
4. **Contract Validation** - SARIF contracts should match documented patterns
5. **Status Checks** - ADR status (Proposed/Accepted/Deprecated) vs implementation state

## The Meta-Governance Challenge

**This adapter faces a unique challenge:** It enforces ADRs, but it also IS an adapter that needs ADRs.

**Questions:**

- Who governs the governance adapter?
- Do we need ADR-001 for this adapter before writing any code?
- What if the ADR enforcement logic violates its own ADRs?

**Resolution:** Bootstrap with ADR-001 defining the enforcement strategy, then self-enforce.

## Creating a New ADR

See the [Master ADR Index](../../../../../docs/adr-index.md) for the template and guidelines.

---

**Related Documentation:**

- [Master ADR Index](../../../../../docs/adr-index.md)
- [Platform ADRs](../../../../../docs/adr)
- [ADR-006: Documentation as Governed Artifact](../../../../../docs/adr/006-documentation-as-governed-artifact.md)
- [ADR- [Control Plane Architecture](../../../../control-plane/README.md)
