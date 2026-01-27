# ADR-002: GitHub as Execution Authority

**Date:** 2026-01-19
**Status:** Accepted
**Deciders:** Ryan Bumstead

---

## Context

Traditional Salesforce DevOps tools (Gearset, Copado) use Salesforce as the execution authority. This creates a single point of failure.

## Decision

GitHub will be the final execution authority. Salesforce provides governance but cannot block GitHub from executing workflows.

## Consequences

**Positive:**

- Deployments continue during Salesforce outages
- Break-glass operations don't require Salesforce access
- Higher uptime (GitHub Actions 99.95% vs Salesforce 99.9%)

**Negative:**

- GitHub outage = deployment outage
- Requires DevOps engineers to have GitHub UI access

## Mitigation

- Maintain local Salesforce CLI access
- Document break-glass runbooks
- Cross-train team on GitHub Actions UI

## Alternatives

- None considered.
