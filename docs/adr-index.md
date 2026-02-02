# GlassOps Architecture Decision Records (ADR) Index

This document provides a centralized index of all architectural decisions across the GlassOps project.

---

## What are ADRs?

Architecture Decision Records (ADRs) capture important architectural decisions along with their context and consequences.

**Why ADRs Matter:**

- 📝 **Document Decisions** - Capture the "why" behind technical choices
- 🔍 **Historical Context** - Understand past decisions when making new ones
- 🤝 **Team Alignment** - Ensure everyone understands the architecture
- ⚖️ **Trade-offs** - Explicitly document alternatives and consequences

---

## ADR Locations by Project

### 1. GlassOps Core (`docs/adr/`)

**Platform-level decisions affecting the entire GlassOps ecosystem**

| ADR                                                  | Title                              | Status   | Date       |
| ---------------------------------------------------- | ---------------------------------- | -------- | ---------- |
| [001](adr/001-system-api-primitive.md)               | System API Primitive               | Accepted | 2026-01-19 |
| [002](adr/002-github-execution-authority.md)         | GitHub Execution Authority         | Accepted | 2026-01-19 |
| [003](adr/003-additive-governance.md)                | Additive Governance                | Accepted | 2026-01-19 |
| [004](adr/004-schema-contract-config.md)             | Schema Contract Config             | Accepted | 2026-01-19 |
| [005](adr/005-control-plane-architecture.md)         | Control Plane Architecture         | Accepted | 2026-01-19 |
| [006](adr/006-documentation-as-governed-artifact.md) | Documentation as Governed Artifact | Proposed | 2026-01-24 |
| [007](adr/007-protocol-supremacy-enforcement.md)     | Protocol Supremacy Enforcement     | Accepted | 2026-01-24 |
| [008](adr/008-federated-documentation-structure.md)  | Federated Documentation Structure  | Accepted | 2026-01-24 |
| [009](adr/009-8-10-vs-10-10-bridge-strategy.md)      | The 8/10 vs 10/10 Bridge Strategy  | Accepted | 2026-01-24 |
| [010](adr/010-identity-contract.md)                  | Identity Contract                  | Accepted | 2026-01-24 |
| [011](adr/011-monorepo-strategy.md)                  | Monorepo Strategy                  | Accepted | 2026-01-24 |

---

### 2. GlassOps Runtime (`glassops-runtime/docs/adr/`)

**Runtime execution decisions for GitHub Actions adapter**

| ADR                                                           | Title                              | Status   | Date       |
| ------------------------------------------------------------- | ---------------------------------- | -------- | ---------- |
| [001](../packages/runtime/adr/001-six-phase-execution.md)     | 6-Phase Execution Model            | Accepted | 2026-01-24 |
| [002](../packages/runtime/adr/002-docker-caching-strategy.md) | Caching Strategy in Docker Runtime | Accepted | 2026-02-01 |

---

### 3. GlassOps Control Plane (`glassops-control-plane/docs/adr/`)

**Control plane and operator architecture decisions**

| ADR                                                                     | Title                       | Status   | Date       |
| ----------------------------------------------------------------------- | --------------------------- | -------- | ---------- |
| [001](../packages/control-plane/adr/001-kubernetes-operator-pattern.md) | Kubernetes Operator Pattern | Accepted | 2026-01-24 |

---

### 4. GlassSpec Protocol (`glassspec/docs/adr/`)

**Protocol specification and standard decisions**

| ADR                                                            | Title                  | Status   | Date       |
| -------------------------------------------------------------- | ---------------------- | -------- | ---------- |
| [001](../packages/glassspec/adr/001-layered-contract-model.md) | Layered Contract Model | Accepted | 2026-01-24 |

---

### 5. GlassOps Native Adapter (`packages/adapters/native/docs/adr/`)

**Native Salesforce adapter decisions**

| ADR | Title         | Status | Date |
| --- | ------------- | ------ | ---- |
| -   | _No ADRs yet_ | -      | -    |

**Proposed ADRs:**

- sf CLI vs @salesforce/core Integration
- JWT Key Management Strategy
- SARIF Mapping from Deployment Results
- Coverage Calculation Methodology

---

### 6. GlassOps SFDX-Hardis Adapter (`packages/adapters/hardis/docs/adr/`)

**SFDX-Hardis quality gate adapter decisions**

| ADR | Title         | Status | Date |
| --- | ------------- | ------ | ---- |
| -   | _No ADRs yet_ | -      | -    |

**Proposed ADRs:**

- Hardis Quality Gates to SARIF Mapping
- Branch Protection Enforcement Strategy
- Integration with Native Adapter Results

---

### 7. GlassOps Scanner Adapter (`packages/adapters/scanner/docs/adr/`)

**Static analysis and code scanner adapter decisions**

| ADR | Title         | Status | Date |
| --- | ------------- | ------ | ---- |
| -   | _No ADRs yet_ | -      | -    |

**Proposed ADRs:**

- MegaLinter vs Individual Linters
- SARIF Aggregation from Multiple Scanners
- Scanner Priority and Weighting
- Baseline/Suppression Strategy

---

### 8. GlassOps ADR Enforcement Adapter (`packages/tools/adr-enforcer/docs/adr/`)

**ADR governance and compliance adapter decisions (Meta-Adapter)**

| ADR | Title         | Status | Date |
| --- | ------------- | ------ | ---- |
| -   | _No ADRs yet_ | -      | -    |

**Proposed ADRs:**

- ADR Detection Strategy
- Decision Drift Detection
- Enforcement Mode (Block vs Warn)
- ADR Coverage Metrics
- Meta-Governance (Self-Enforcement)

**Special Note:** This is a **meta-adapter** that governs the governance process itself. It enforces ADR compliance and detects when architectural decisions are missing or violated.

---

## ADR Lifecycle

```
Proposed → Accepted → Deprecated → Superseded
```

- **Proposed** - Under review, seeking feedback
- **Accepted** - Approved and actively implemented
- **Deprecated** - No longer recommended, but not yet replaced
- **Superseded** - Replaced by a newer ADR

---

## Creating a New ADR

### 1. Choose the Right Location

- **Platform-wide** → `glassops/docs/adr/`
- **Runtime-specific** → `glassops-runtime/docs/adr/`
- **Control plane** → `glassops-control-plane/docs/adr/`
- **Protocol changes** → `glassspec/docs/adr/`

### 2. Use the Template

```markdown
# ADR [NUMBER]: [Title]

**Status:** [Proposed|Accepted|Deprecated|Superseded]
**Date:** YYYY-MM-DD
**Context:** [What triggered this decision]

## Context

[Describe the forces at play: technical, political, social, project]

## Decision

[State the decision clearly]

## Rationale

[Explain why this decision was made]

## Consequences

### Positive

- [Benefits]

### Negative

- [Costs, risks]

### Neutral

- [Side effects]

## Alternatives Considered

[What other options were evaluated and why they were rejected]

## Related ADRs

- [Links to related decisions]

---

**Author:** [Name]
**Status:** [Current status]
```

### 3. Number Sequentially

- Each project maintains its own sequence
- Use leading zeros (001, 002, etc.)
- Never reuse numbers

### 4. Update This Index

After creating an ADR, add it to the appropriate section above.

---

## Key Decisions by Theme

### Governance & Protocol

- [ADR-002: GitHub Execution Authority](adr/002-github-execution-authority.md)
- [ADR-003: Additive Governance](adr/003-additive-governance.md)
- [ADR-007: Protocol Supremacy Enforcement](adr/007-protocol-supremacy-enforcement.md)

### Architecture & Structure

- [ADR-008: Federated Documentation Structure](adr/008-federated-documentation-structure.md)
- [ADR-011: Monorepo Strategy](adr/011-monorepo-strategy.md)
- [ADR-009: The 8/10 vs 10/10 Bridge Strategy](adr/009-8-10-vs-10-10-bridge-strategy.md)

### Contracts & Standards

- [ADR-004: Schema Contract Config](adr/004-schema-contract-config.md)
- [ADR-006: Documentation as Governed Artifact](adr/006-documentation-as-governed-artifact.md)
- [ADR-010: Identity Contract](adr/010-identity-contract.md)

---

## ADR Best Practices

### Do

✅ **Write ADRs when making significant decisions** - Architectural impact, costly to change, affects multiple teams

✅ **Keep them concise** - 1-2 pages maximum

✅ **Document alternatives** - Show you considered trade-offs

✅ **Update status** - Mark as Deprecated/Superseded when appropriate

✅ **Link related ADRs** - Show evolution of thinking

### Don't

❌ **Don't document implementation details** - ADRs are decisions, not code

❌ **Don't rewrite history** - Keep old ADRs as historical record

❌ **Don't delete ADRs** - Mark as Superseded instead

❌ **Don't wait too long** - Write ADR when decision is fresh

---

## Questions?

- **Platform ADRs:** See [docs/adr/](adr/README.md)
- **Contributing:** See [CONTRIBUTING.md](../CONTRIBUTING.md)
- **Discussion:** [GitHub Discussions](https://github.com/glassops-platform/glassops/discussions)

---

**Last Updated:** 2026-01-24  
**Maintainer:** Ryan Bumstead ([@rdbumstead](https://github.com/rdbumstead))
