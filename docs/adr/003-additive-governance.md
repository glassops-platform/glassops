# ADR-003: Additive Governance Model

**Date:** 2026-01-19
**Status:** Accepted
**Deciders:** Ryan Bumstead

---

## Context

We need to govern deployments (check coverage, freeze windows) without creating 'lowest common denominator' policies that block high-performing teams.

## Decision

We adopt an **Additive Governance** model. GitHub Environment config is the **Floor**. devops-config.json is the **Team Standard**. Salesforce CMDT is the **Org-Specific Ceiling/Addition**.

## Consequences

**Positive:**

- Governance can only RAISE quality standards, never lower them below the platform floor.

**Negative:**

- Runtime resolution complexity.

## Mitigation

- Clear documentation on precedence logic.

## Alternatives

- None considered.
