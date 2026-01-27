# ADR 001: Kubernetes Operator Pattern

**Status:** Accepted  
**Date:** 2026-01-24  
**Context:** GlassOps Control Plane Architecture

---

## Context

GlassOps needs a way to "continuously reconcile" the state of infrastructure against governance policy. We moved away from "CI-only" governance because CI is ephemeral. We need a persistent control plane.

## Decision

We will implement the GlassOps Control Plane using the **Kubernetes Operator Pattern**.

- **Policy as Data:** Governance rules will be defined as Custom Resource Definitions (CRDs).
- **Continuous Reconciliation:** A custom Controller will watch these CRDs and enforce state.
- **Event-Driven:** The Operator will react to CloudEvents (Layer 3) to trigger reconciliation loops.

## Rationale

1.  **Drift Detection:** Operators run in a loop. If someone manually changes a setting (drift), the Operator sees it and alerts/reverts. CI pipelines only run on commit.
2.  **Kubernetes Native:** Inherits K8s RBAC, audit logs, and scalability.
3.  **Declarative Governance:** Allows us to define `GovernancePolicy` objects that look like K8s manifests.

## Structure

```yaml
apiVersion: glassops.io/v1alpha1
kind: GovernancePolicy
metadata:
    name: no-dml-in-loops
spec:
    enforcement: block
    rules:
        - engine: pmd
          ruleId: 'OperationWithLimitsInLoop'
```

## Consequences

- **Positive:** Real-time governance, drift handling, standard API.
- **Negative:** Complexity. Requires a running K8s cluster (even if local via kind).
- **Mitigation:** The "GlassOps Runtime" (GitHub Action) can run without the Operator for simpler use cases (Level 1/2 adoption).

---

**Author:** Ryan Bumstead

## Alternatives

- None considered.
