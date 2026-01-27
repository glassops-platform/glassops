# ADR 009: The 8/10 vs 10/10 Bridge Strategy

**Status:** Accepted  
**Date:** 2026-01-24  
**Deciders:** Ryan Bumstead

---

## Context

GlassOps aims to be the "Universal Control Plane" (10/10 Vision), but immediate tailored value is required to drive adoption in the Salesforce ecosystem (8/10 Vision).

Rather than viewing these as conflicting goals, we recognize them as a **strategic progression**. The architecture must be designed to deliver immediate, concrete value for Salesforce teams today while laying the immutable foundation (SARIF, OpenTelemetry, CloudEvents) required for the future universal platform. This "Bridge Strategy" ensures that every artifact built for the current narrow use case is forward-compatible with the broader long-term vision.

### The Bifurcation

**8/10 Vision (Current - 18 months):**

- **Authority:** GitHub Actions
- **Model:** Event-driven (commit → webhook)
- **Focus:** Salesforce DevOps governance
- **Deployment:** Free/low-cost (Vercel, Neon, GitHub Actions)
- **Users:** 15-20 production deployments

**10/10 Vision (Future - 36+ months):**

- **Authority:** Kubernetes Operator
- **Model:** Level-triggered (controller reconciliation loop)
- **Focus:** Universal multi-substrate governance
- **Deployment:** Enterprise Kubernetes clusters
- **Users:** Multi-tenant SaaS platform

### The Conflicts Identified

1. **Authority Conflict:** GitHub vs Kubernetes as execution authority
2. **Identity Gap:** No unified identity contract across substrates
3. **Scaling Path:** No migration playbook from zero-dollar to enterprise
4. **Documentation vs Reality:** Roadmap inconsistencies

---

## Decision

**Accept the bifurcation as an intentional architectural evolution, with the Layered Contract Model as the immutable bridge.**

### Core Principle

> **The Contract is the Constant**

Regardless of which execution authority (GitHub or Kubernetes) orchestrates the deployment, the Layered Contract (SARIF + OTel + CloudEvents) remains the immutable governance record.

---

## Rationale

### Why Two Timelines Are Necessary

1.  **Market Reality** - Salesforce teams need GitHub Actions today, not K8s operators
2.  **Learning Curve** - Building a K8s operator without production usage is academic
3.  **Revenue Path** - 8/10 generates consulting revenue; 10/10 requires venture funding
4.  **Technical Debt** - Better to validate Protocol with real users before committing to K8s

### Why the Contract Bridges Them

The Layered Contract provides **temporal invariance:**

```
GitHub Actions (8/10)         Kubernetes Operator (10/10)
        │                              │
        ▼                              ▼
    [Adapter]                      [Adapter]
        │                              │
        ▼                              ▼
 Layered Contract  ════════════  Layered Contract
  (SARIF+OTel+CE)   [IDENTICAL]   (SARIF+OTel+CE)
        │                              │
        ▼                              ▼
  Policy Engine                  Policy Engine
```

**Key Insight:** If both emit identical Layered Contracts, the governance layer doesn't care which orchestrator triggered it.

---

## Implementation Strategy

### Phase 1: 8/10 (Current - Months 1-18)

**Goal:** Validate Protocol with real Salesforce users

**Architecture:**

```yaml
Execution Authority: GitHub Actions
Adapters: glassops-native, glassops-hardis
Contract Store: GitHub Artifacts (S3 later)
Policy Engine: JavaScript (serverless)
UI: Optional Salesforce Console
Deployment: Zero-dollar stack (Neon, Vercel)
```

**Outcomes:**

- 15-20 production deployments
- Validated SARIF schema
- Proven adapter interface
- Real-world policy patterns
- Consulting revenue

---

### Bridge Phase: Hybrid (Months 19-24)

**Goal:** Decouple Protocol from GitHub-specific assumptions

**Architecture:**

```yaml
Execution Authority: GitHub Actions OR Kubernetes
Adapters: Same adapters work in both modes
Contract Store: S3/GCS (cloud-agnostic)
Policy Engine: Go microservice (can run anywhere)
UI: API-first (UI is projection)
Deployment: Optional K8s for policy engine
```

**Key Migrations:**

1. **Policy Engine:** JavaScript → Go microservice
2. **Contract Store:** GitHub Artifacts → S3/GCS
3. **Identity:** Add Identity Contract abstraction
4. **Adapters:** Make substrate-agnostic (via env vars)

---

### Phase 2: 10/10 (Months 25+)

**Goal:** Universal governance control plane

**Architecture:**

```yaml
Execution Authority: Kubernetes Operator
Adapters: CRD-based adapter resources
Contract Store: TimescaleDB (queryable history)
Policy Engine: Operator controller
UI: Next.js dashboard (multi-tenant)
Deployment: Enterprise K8s (multi-cluster)
```

**New Capabilities:**

- Multi-substrate beyond Salesforce
- Level-triggered reconciliation
- Multi-tenant SaaS
- Advanced correlation engine
- Glass Language DSL

---

## Addressing the Conflicts

### 1. Authority Conflict (GitHub vs Kubernetes)

**Resolution:** The authority changes, but the contract doesn't.

**Implementation:**

```yaml
# Phase 8/10: GitHub Actions
on: [push]
steps:
    - name: Deploy via Adapter
      run: glassops-native-adapter deploy
    # Emits: .glassops/glassops-contract.sarif.json

    - name: Enforce Policy
      run: glassops policy enforce .glassops/glassops-contract.sarif.json
```

```yaml
# Phase 10/10: Kubernetes Operator
apiVersion: governance.glassops.io/v1
kind: Deployment
spec:
    adapter: glassops-native-adapter
    target: salesforce-prod
# Controller reconciles, emits SARIF to TimescaleDB
```

**Bridge:** Both emit identical SARIF. Policy engine doesn't know (or care) which triggered it.

---

### 2. Identity Gap

**Problem:** "GitHub User" ≠ "Salesforce User" ≠ "K8s Service Account"

**Resolution:** Create Identity Contract (ADR-010 to be written)

**Proposal:**

```json
{
    "identity": {
        "subject": "alice@company.com",
        "provider": "github|salesforce|kubernetes",
        "provider_id": "github:alice|sf:005...|k8s:sa:...",
        "roles": ["developer", "deployer"],
        "authorization": {
            "can_deploy": true,
            "can_override": false
        }
    }
}
```

**Bridge:** SARIF `invocations[].properties.glassops.identity` field works in both modes.

---

### 3. Scaling Path (Zero-Dollar → Enterprise)

**Problem:** No migration playbook from Neon/Vercel to K8s/TimescaleDB.

**Resolution:** Document explicit migration path

**Path:**

```
Zero-Dollar (8/10)
  ├─ GitHub Actions (free tier)
  ├─ Neon PostgreSQL (free tier)
  ├─ Vercel (hobbyist plan)
  └─ S3 artifacts (AWS free tier)
        │
        ▼ [Migrate when: >20 deploys/day OR >5 orgs]
        │
Low-Cost (8.5/10)
  ├─ GitHub Actions (Team plan $4/user)
  ├─ AWS RDS PostgreSQL ($50/mo)
  ├─ Vercel Pro ($20/mo)
  └─ S3 artifacts ($10/mo)
        │
        ▼ [Migrate when: Need multi-tenant OR >100 users]
        │
Enterprise (10/10)
  ├─ Kubernetes cluster ($500-2000/mo)
  ├─ TimescaleDB (managed $300/mo)
  ├─ Multi-region S3 ($100/mo)
  └─ Load balancer + monitoring ($200/mo)
```

**Key Insight:** Each phase is fully functional, not a "prototype."

---

### 4. Documentation vs Reality

**Problem:** Some docs describe 8/10 (GitHub), others describe 10/10 (K8s).

**Resolution:** Use `[Phase 8/10]` and `[Phase 10/10]` prefixes

**Example:**

```markdown
## Deployment

**[Phase 8/10]** Use GitHub Actions workflow:
...

**[Phase 10/10]** Use Kubernetes CRD:
...

**Bridge:** Both emit SARIF contracts to the Policy Engine.
```

---

## Invariants (What Never Changes)

1. **SARIF 2.1.0** - Canonical governance format across all phases
2. **Layered Contract Model** - SARIF + OTel + CloudEvents + Native
3. **Protocol Supremacy** - glassspec/ is authoritative
4. **Adapter Interface** - Same interface works in GitHub and K8s modes
5. **Stateless Adapters** - Never store state, always emit contracts

---

## Consequences

### Positive

- ✅ **Validates before scaling** - 8/10 proves market fit before K8s investment
- ✅ **Revenue path** - Consulting revenue funds 10/10 development
- ✅ **User-driven evolution** - Real production usage informs K8s design
- ✅ **Explicit migration** - No hidden technical debt
- ✅ **Contract immutability** - Governance records survive authority transition

### Negative

- ⚠️ **Two codebases** - Policy engine exists in JS (8/10) and Go (10/10)
- ⚠️ **Documentation burden** - Must maintain both timelines
- ⚠️ **User confusion** - Some users on 8/10, some on 10/10
- ⚠️ **Feature parity lag** - 10/10 features not available in 8/10

### Neutral

- 📝 Contract is constant, execution varies
- 📝 8/10 is not a "prototype" - it's a deliberate constraint
- 📝 10/10 is not speculative - it's informed by 8/10 usage

---

## Alternatives Considered

### Alternative 1: Build Only 8/10 (GitHub Forever)

**Rejected:** Limits addressable market to Salesforce-only teams

### Alternative 2: Build Only 10/10 (K8s From Day 1)

**Rejected:** Too complex for initial users, burns runway on speculation

### Alternative 3: Build Both Simultaneously

**Rejected:** Splits focus, neither gets production validation

### Alternative 4: Hide the Bifurcation

**Rejected:** Users deserve transparency about roadmap

---

## Migration Triggers

**When to move from 8/10 → 10/10:**

1. **Volume:** >100 deployments/day across all users
2. **Multi-tenancy:** >10 distinct organizations
3. **Multi-substrate:** Need to govern beyond Salesforce (AWS, K8s, etc.)
4. **Funding:** Secured seed round to invest in K8s infrastructure
5. **Team:** Hired K8s/Go engineers (can't build 10/10 solo)

**Until then:** 8/10 is the recommended path.

---

## Related ADRs

- [ADR-001: System API Primitive](001-system-api-primitive.md)
- [ADR-002: GitHub Execution Authority](002-github-execution-authority.md) ← 8/10 decision
- [ADR-005: Control Plane Architecture](005-control-plane-architecture.md) ← 10/10 vision
- [ADR-007: Protocol Supremacy Enforcement](007-protocol-supremacy-enforcement.md)
- [ADR-008: Federated Documentation Structure](008-federated-documentation-structure.md)
- ADR-010: Identity Contract (Future - addresses identity gap)

---

## References

- [Philosophy](../vision.md)
- [Architecture Future](../architecture/overview.md)

---

**Author:** Ryan Bumstead  
**Implemented:** 2026-01-24  
**Status:** Active - Defines long-term strategy

## Alternatives

- None considered.
