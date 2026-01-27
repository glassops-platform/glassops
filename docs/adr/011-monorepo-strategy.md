# ADR 011: Monorepo Strategy for Protocol Ecosystem

**Status:** Accepted  
**Date:** 2026-01-24  
**Deciders:** Ryan Bumstead

---

## Context

GlassOps is a "Protocol Ecosystem" consisting of:

1.  **The Law** (`glassspec`) - The Interface
2.  **The Manual** (`glassops/docs`) - The Documentation
3.  **The Implementations** (`runtime`, `control-plane`, `adapters`) - The Code

The development team is currently **one person**.

### The Problem with Polyrepos (Separate Repos)

If we split these into 5+ repositories (`glass-protocol`, `glass-runtime`, `glass-adapter-native`, etc.):

1.  **Context Switching Cost:** Updating a standard requires 5 PRs across 5 repos.
2.  **Versioning Hell:** "Runtime v1.2 depends on Protocol v1.1 but imports Adapter v1.0 which expects Protocol v1.0."
3.  **Synchronization Lag:** The docs will inevitably drift from the code because they live in different universes.

For a solo developer, this friction is fatal.

## Decision

**We will use a Single Governance Monorepo (`glassops-platform/glassops`) for all core components.**

### The Structure

```text
glassops/
├── packages/
│   ├── glassspec/              # The constraints (Sarif schemas, Protocol.md)
│   ├── runtime/                # The machinery (@glassops/runtime)
│   ├── control-plane/          # The brains (@glassops/control-plane)
│   └── adapters/               # The hands (@glassops/native-adapter, etc)
├── docs/                       # The Manual
└── config/                     # Shared configuration
```

### The "Publisher Pattern" for Distribution

While development happens in the Monorepo, consumption may require separate repos (e.g., GitHub Actions marketplace limitations).

**Strategy:**

1.  **Develop** in `glassops`.
2.  **Release** via automated pipelines that copy folders to read-only public mirrors (e.g., `glassops-platform/glassops-runtime`).
3.  **Never** write code in the mirrors.

## Rationale

1.  **Atomic Protocol Updates:** A single commit can update the SARIF spec in `glassspec` AND the implementation in `glassops-runtime`. Drift is impossible.
2.  **Solo Velocity:** Zero context switching. grep across the entire universe. Refactor global names in one pass.
3.  **Coherent Versioning:** The ecosystem moves forward together.

### Strategic Extensibility

The monorepo structure allows for seamless expansion. **we can add `glassops-cli`, `glassops-docs`, `glassops-policy-engine`, or even a second platform later without renaming anything.** The structure is designed to accept new domains as top-level directories without disrupting the core.

## Consequences

- **Positive:** Maximum velocity for 1-3 developers. Guaranteed consistency.
- **Negative:** CI pipelines are complex (need to detect which folder changed).
- **Mitigation:** Use tools like `turborepo` or simple path-based GitHub Action triggers.

---

**Author:** Ryan Bumstead (Population: 1)

## Alternatives

- None considered.
