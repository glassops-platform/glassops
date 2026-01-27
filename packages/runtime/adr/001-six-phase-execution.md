# ADR 001: 6-Phase Runtime Execution Model

**Status:** Accepted  
**Date:** 2026-01-24  
**Context:** GlassOps Runtime Architecture

---

## Context

Running ad-hoc scripts in CI pipelines creates a "governance gap." We need a deterministic, repeatable way to execute Salesforce deployments that ensures policy is checked **before** execution and that outcomes are recorded **after**.

## Decision

GlassOps Runtime enforces a strict **6-Phase Execution Model** for every operation:

1.  **Phase 0: Cache Retrieval** - Restore environment from Protocol-Linked Cache.
2.  **Phase 1: Policy Evaluation** - Check Freeze Windows and effective policy rules.
3.  **Phase 2: Bootstrap** - Install exact CLI versions defined in policy.
4.  **Phase 3: Static Analysis** (Optional) - Run scanners and validate ADRs.
5.  **Phase 4: Identity & Execution** - Authenticate (JWT) and execute the Adapter.
6.  **Phase 5: Contract Validation** - Emit and validate the SARIF contract.

## Rationale

1.  **Fail Fast:** Policy (Phase 1) runs before Bootstrap (Phase 2), saving time if a Freeze Window is active.
2.  **Determinism:** Bootstrap (Phase 2) ensures the CLI version is explicitly managed, preventing "it worked on my machine" issues.
3.  **Security:** Identity (Phase 4) is JIT-injected only when needed, reducing surface area.
4.  **Traceability:** Every execution is guaranteed to produce a Contract (Phase 5).

## Consequences

- **Positive:** Guaranteed consistency across all deployments. No "hidden scripts."
- **Negative:** Adds slight overhead (seconds) to small jobs compared to raw CLI.
- **Compliance:** This model satisfies SOC 2 "Change Management" controls by design.

---

**Author:** Ryan Bumstead

## Alternatives

- None considered.
