# ADR 001: Layered Contract Model

**Status:** Accepted  
**Date:** 2026-01-24  
**Context:** GlassSpec Protocol Definition

---

## Context

A core challenge in creating a "Universal Governance Protocol" is that no single data format fits all use cases.

- **SARIF** is excellent for static analysis and policy decisions but poor for time-series metrics.
- **OpenTelemetry** is the standard for metrics and traces but lacks the semantic richness for governance policy.
- **CloudEvents** is perfect for transport but doesn't define payload schemas.
- **Native Logs** (e.g., AWS CloudTrail types) are necessary for audit fidelity but impossible to govern consistently.

Trying to force all these into one schema (e.g., "Putting CPU metrics into SARIF properties") leads to bloated, unqueryable contracts.

## Decision

We will adopt a **Layered Contract Model** that strictly separates concerns into four layers:

| Layer             | Purpose                         | Standard           |
| :---------------- | :------------------------------ | :----------------- |
| **1. Governance** | Decisions, Violations, Findings | **SARIF 2.1.0**    |
| **2. Telemetry**  | Metrics, Traces, Signals        | **OpenTelemetry**  |
| **3. Transport**  | Events, Correlation             | **CloudEvents**    |
| **4. Native**     | Raw edge data                   | **Native Schemas** |

All layers are linked by a common `correlation_id`.

## Rationale

1.  **Separation of Concerns:** Policies only evaluate Layer 1 (SARIF). Use dashboards for Layer 2 (OTel).
2.  **Ecosystem Compatibility:**
    - Security tools output SARIF natively.
    - Observability tools consume OTel natively.
    - Event buses route CloudEvents natively.
3.  **Governance Stability:** The Governance Layer (Layer 1) changes slowly (human time). The Telemetry Layer (Layer 2) changes rapidly (machine time). Decoupling them prevents "audit log noise."

## Consequences

- **Positive:** We stop fighting file formats. Metrics go where metrics belong.
- **Negative:** Requires correlation logic to link SARIF findings to OTel traces.
- **Neutral:** Adapters must be disciplined about what goes where (see- [Original Protocol Spec](../README.md)).

---

**Author:** Ryan Bumstead

## Alternatives

- None considered.
