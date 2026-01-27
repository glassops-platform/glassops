# ADR-001: glassops-runtime as System API Primitive

**Date:** 2026-01-19
**Status:** Accepted
**Deciders:** Ryan Bumstead

---

## Context

Salesforce Actions on the GitHub Marketplace vary wildly in scope, often mixing CLI installation, authentication, and deployment logic. This creates coupling and maintenance burdens.

## Decision

We define `glassops-runtime` as a 'System API Primitive' - a foundational layer with a strictly limited scope:

1. Install CLI
2. Authenticate (JWT/SfdxUrl)
3. Validate Environment (Invariants)

It provides **stable outputs** (org_id, instance_url) but enforces **zero policy**.

## Consequences

**Positive:**

- Reusable workflows (Level 2) can build ANY policy on top of this stable foundation.

**Negative:**

- Requires a second layer (workflows) to do anything useful (like deploy).

## Mitigation

- Provide robust reusable workflow templates.

## Future Evolution (v4 Roadmap)

**Concept:** The "SAPI Consolidator" Pattern

As the platform matures, `glassops-runtime@v4` will evolve from a monolithic JavaScript action into a **Composite Action** that orchestrates smaller, atomic primitives.

**Proposed Structure:**

- **Wrapper:** `glassops-platform/glassops-runtime@v1` (Maintains the stable API Contract)
- **Atomic Primitive A:** `glassops-platform/action-install-cli` (Binary management)
- **Atomic Primitive B:** `glassops-platform/action-auth-jwt` (Identity management)
- **Atomic Primitive C:** `glassops-platform/action-resolve-context` (Org introspection)

**Benefit:** This allows us to patch authentication logic (e.g., if Salesforce changes `auth:jwt:grant`) without risking regressions in CLI installation logic, while consumers (Adapters) see no breaking changes.

## Alternatives

- None considered.
