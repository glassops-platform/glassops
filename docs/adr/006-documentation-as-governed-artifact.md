# ADR 006: Documentation as a Governed Artifact

**Status:** Proposed  
**Date:** 2026-01-24  
**Deciders:** Ryan Bumstead

---

## Context

GlassOps currently governs:

- **Code Quality** (SARIF findings from static analysis)
- **Deployment Safety** (test coverage, validation results)
- **Architecture Decisions** (ADR compliance)
- **Runtime Risk** (OpenTelemetry correlation)

**Missing:** Documentation itself is not treated as a first-class governed signal.

### The Reframed Question

We're not asking: "Should GlassOps generate documentation?"

We're asking: **"Should documentation be a governed, enforceable, first-class signal in GlassOps?"**

Once you see it that way, GlassOps Knowledge + MkDocs aren't tooling choices — they're reference implementations of a capability.

---

## Decision

**GlassOps will treat documentation as a governed artifact**, subject to the same protocol enforcement as code quality, deployments, and architecture.

### Core Principle

> **Most platforms consume documentation. GlassOps governs it.**

Documentation becomes the **9th substrate** (joining Code, Cloud, Network, Analytics, IaC, Databases, Monitoring, Deployments).

---

## Architecture: Docs as a First-Class Signal

```
Code / IaC / Config
        │
        ▼
GlassOps Knowledge (AST + Auto Docs)
        │
        ▼
Structured Doc Artifacts
        │
        ▼
GlassOps Documentation Adapter
        │
        ▼
Protocol Engine
        │
        ├─ Enforce coverage
        ├─ Detect drift
        ├─ Gate deployments
        └─ Correlate with runtime risk
        │
        ▼
MkDocs (Curated Projection)
```

**Key Insight:** MkDocs is not "the docs site". It's the **compiled output of governed documentation**.

---

## Concrete Product Features

### 1. Documentation Coverage Enforcement

GlassOps enforces rules like:

**Every deployable unit must have:**

- Architecture overview
- Ownership
- Risk classification

**Public APIs require:**

- Versioned contract docs
- Breaking-change notices

**Regulated repos require:**

- ADRs
- Data handling docs

**Flow:**

- RepoAgent → detects
- GlassOps → enforces
- MkDocs → proves compliance

---

### 2. Documentation Drift Detection

**Problem:** This is extremely under-served in tooling.

**Examples GlassOps can detect:**

- ❌ Code changed, docs didn't
- ❌ Architecture diagram references deleted modules
- ❌ README says "sync", code is async
- ❌ Maturity level claims don't match repo reality

**Output:**

> "This system claims Level 3 governance but behaves like Level 1."

**That's gold.**

---

### 3. Governance-Aware Documentation Views

MkDocs becomes **dynamic by governance state:**

- Show warnings for non-compliant sections
- Highlight "out of protocol" components
- Link runtime incidents → docs sections
- Surface risk levels inline

**Nobody else does this.**

---

### 4. Audit & Compliance Without the Theater

**Instead of:**

- Word docs
- Confluence sprawl
- Manual audits

**You get:**

- Versioned docs
- Deterministic builds
- Policy-backed guarantees

**GlassOps can say:**

> "This documentation set passed governance checks at commit X."

**That's audit-grade.**

---

### 5. Documentation as an Execution Gate

**This is where it gets spicy 🔥**

**Examples:**

- ❌ No docs → no deploy
- ⚠️ Docs incomplete → deploy to lower environment only
- ⚠️ Docs stale → require override
- ✅ Docs updated → unlock fast path

**Documentation literally controls velocity.**

---

## Why RepoAgent + MkDocs Specifically

### RepoAgent

- Objective signal generator
- AST-aware
- Language-agnostic trajectory
- Automatable

### MkDocs

- Deterministic
- Static
- Versionable
- CI-native
- No runtime dependencies

**Together they form: The reference pipeline for governable documentation.**

You're not locking users in — you're showing them what "good" looks like.

---

## Positioning

### ❌ Don't Say:

"GlassOps uses RepoAgent and MkDocs"

### ✅ Do Say:

"GlassOps governs documentation pipelines. RepoAgent and MkDocs are supported reference implementations."

This keeps you:

- Vendor-neutral
- Extensible
- Architecturally credible

---

## Implementation Strategy

### Phase 1: Define the Documentation Adapter Interface

Create `DOCUMENTATION_ADAPTER.md` in `glassspec/`:

```yaml
apiVersion: glassops.io/v1alpha1
kind: DocumentationAdapter
spec:
    input:
        - type: markdown
          location: docs/
        - type: auto-generated
          generator: repoagent
          source: src/

    validation:
        - coverage: required
          sections: [architecture, ownership, api]
        - drift_detection: true
          compare_with: code_ast
        - adrs: required
          when: architectural_changes

    output:
        format: sarif
        findings:
            - rule: DOC_COVERAGE
            - rule: DOC_DRIFT
            - rule: DOC_STALE
```

### Phase 2: Build Reference Implementation

**Repo:** `glassops-documentation-adapter`

- Integrates RepoAgent for auto-generation
- Validates against documentation policies
- Emits SARIF contract
- Optionally builds MkDocs site

### Phase 3: Documentation Governance Policies

Add to Policy Store:

```yaml
documentation:
    coverage:
        required_sections:
            - architecture
            - ownership
            - risk_classification
        public_apis: versioned_contracts

    drift:
        max_staleness_days: 30
        require_update_on_code_change: true

    gates:
        no_docs: block_deployment
        incomplete_docs: restrict_to_dev
        stale_docs: require_override
```

---

## Benefits

1. **Governance Consistency** - Docs treated like code
2. **Audit Trail** - Provable compliance
3. **Velocity Control** - Bad docs = blocked deploys
4. **Drift Detection** - Code/docs sync enforced
5. **Competitive Differentiation** - Nobody else governs docs at protocol level

---

## Risks & Mitigations

### Risk 1: Documentation Feels Like Busywork

**Mitigation:** Auto-generate baseline with RepoAgent, humans curate

### Risk 2: Hard-Coded Tool Assumptions

**Mitigation:** Define adapter interface, support multiple generators

### Risk 3: Overhead Without Value

**Mitigation:** Start with optional enforcement, prove value before mandating

---

## Alternatives Considered

### Alternative 1: Docs Stay Outside Governance

**Rejected:** Misses huge value opportunity, docs drift leads to incidents

### Alternative 2: Hard-Code MkDocs

**Rejected:** Violates adapter philosophy, creates vendor lock-in

### Alternative 3: Manual Doc Review Only

**Rejected:** Doesn't scale, loses automation benefits

---

## Consequences

### Positive

- ✅ Documentation becomes enforceable
- ✅ Audit compliance provable
- ✅ Drift detection automated
- ✅ 9th substrate adds strategic value

### Negative

- ⚠️ Requires Documentation Adapter development
- ⚠️ Teams must adapt to doc-gated deployments
- ⚠️ Tooling ecosystem expansion

### Neutral

- 📝 RepoAgent and MkDocs become reference implementations
- 📝 Documentation governance optional but recommended

---

## Next Steps

1. **Create `glassspec/DOCUMENTATION_ADAPTER.md`** - Define the interface
2. **Build POC** - RepoAgent → SARIF → Policy Engine
3. **Add Documentation Policies** - Coverage, drift, staleness rules
4. **Document in Vision** - Add docs governance to vision/technical.md
5. **Reference Implementation** - Build glassops-documentation-adapter

---

## Related ADRs

- [ADR-001: System API Primitive](001-system-api-primitive.md)
- [ADR-003: Additive Governance](003-additive-governance.md)
- [ADR-004: Schema Contract Config](004-schema-contract-config.md)

---

## References

- [OASIS SARIF 2.1.0](https://docs.oasis-open.org/sarif/sarif/v2.1.0/)
- [RepoAgent Documentation](https://github.com/OpenBMB/RepoAgent)
- [MkDocs](https://www.mkdocs.org/)
- [GlassOps Protocol](../../packages/glassspec/README.md)

---

**Author:** Ryan Bumstead  
**Reviewers:** TBD  
**Implementation:** Proposed for v2.0

## Alternatives

- None considered.
