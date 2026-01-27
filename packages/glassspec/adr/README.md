# GlassSpec Protocol - Architecture Decision Records

This directory contains Architecture Decision Records (ADRs) specific to the GlassSpec Protocol specification.

## Scope

ADRs in this directory cover decisions related to:

- Protocol specification and standards
- SARIF 2.1.0 adoption and usage
- Layered Contract Model design
- Anti-pattern definitions
- Protocol versioning strategy

## Current ADRs

| Number | Title                  | Status   | Date       |
| ------ | ---------------------- | -------- | ---------- |
| 001    | Layered Contract Model | Accepted | 2026-01-26 |

## Proposed Decisions to Document

1. **SARIF 2.1.0 as Canonical Format** - Why SARIF over custom schemas
2. **Layered Contract Model** - SARIF + OTel + CloudEvents + Native
3. **Anti-Pattern Scope Boundaries** - What NOT to normalize into SARIF
4. **Protocol Versioning Strategy** - Semantic versioning for Protocol changes
5. **Adapter Interface Stability** - Backward compatibility guarantees

## Protocol Authority

> [!IMPORTANT]
> ADRs in this directory define the **Protocol (Law)**, not the implementation.

- **glassspec/protocol.md** is authoritative
- **glassspec/adapter-interface.md** is canonical
- All implementations (glassops/, glassops-runtime/, glassops-control-plane/) must defer to these specifications

## Creating a New ADR

See the [Master ADR Index](../../../docs/adr-index.md) for the template and guidelines.

> [!NOTE]
> Protocol ADRs require careful consideration as they affect the entire ecosystem.

---

**Related Documentation:**

- [Master ADR Index](../../../docs/adr-index.md)
- [GlassOps ADRs](../../../docs/adr)
- [GlassSpec Protocol](../README.md)
- [Adapter Interface](../README.md)
