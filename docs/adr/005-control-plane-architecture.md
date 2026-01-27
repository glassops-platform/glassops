# ADR-005: Control Plane Architecture & Reference Console

**Date:** 2026-01-19
**Status:** Accepted
**Deciders:** Ryan Bumstead

---

## Overview

This ADR defines the decoupling of the GlassOps UI from the core protocol, establishing the "Reference Console" pattern.

## Components

1. **Protocol (Level 2):** The governance logic (policy, contracts).
2. **Reference Console (Level 3):** The optional visualization layer.

## Data Flow

Data flows from the Protocol execution (Github Actions) to the Console (Salesforce) via Platform Events, but not vice-versa for execution control.

## Context

The initial "Three-Level" architecture implied that "Level 3: GlassOps Reference Console" (the Salesforce UI) was the ultimate destination of the adoption journey. This created friction:

1.  **Implication of Necessity:** Teams felt they _had_ to install the package to get value.
2.  **Naming Confusion:** "Manager" implied it orchestrated deployments (it doesn't; GitHub Actions does).
3.  **Role Mismatch:** DevOps engineers don't want a Salesforce UI; Admins do.

### Decision

1.  **Decoupling:** The UI is explicitly defined as an **optional** "Level 3" component, purely for visualization.
2.  **Renaming:** "GlassOps Manager" is renamed to "**GlassOps Reference Console**".
3.  **Role:** The Salesforce UI is an **optional reference implementation** of a control plane consumer, not a required component.

## Consequences

**Positive:**

- **Psychological Safety:** Teams can adopt the Protocol (Level 2) without fearing vendor lock-in to the UI.
- **Ecosystem Friendly:** Invites 3rd-party vendors (e.g., sfdx-hardis, Copado) to plug into the event stream rather than competing for the glass.
- **Clear Boundaries:** Reinforces that GlassOps owns _State_ and _Policy_, but delegates _Execution_ and _UX_.

**Negative:**

- Documentation is slightly more abstract ("Protocol" vs "Product").

## Mitigation

- Use "Three-Level Adoption Model" to keep the simple "Walk, Run, Fly" narrative for beginners, while using "Control Plane" for architects.

## Alternatives

- None considered.
