# GlassOps Runtime - Architecture Decision Records

This directory contains Architecture Decision Records (ADRs) specific to the GlassOps Runtime execution primitive.

## Scope

ADRs in this directory cover decisions related to:

- Runtime phase execution model
- CLI wrapper patterns
- Contract generation and atomicity
- Authentication strategies
- Error handling and resilience

## Current ADRs

| Number | Title         | Status | Date |
| ------ | ------------- | ------ | ---- |
| -      | _No ADRs yet_ | -      | -    |

## Proposed Decisions to Document

1. **Runtime Phase Execution Model** - The 6-phase execution pattern
2. **CLI Wrapper vs Direct Commands** - Using sf CLI vs @salesforce packages
3. **Contract Atomicity Strategy** - Temp file + rename pattern
4. **JWT-OAuth Authentication** - Strict identity enforcement
5. **Cache Strategy** - Protocol-linked cache design

## Creating a New ADR

See the [Master ADR Index](../../../docs/adr-index.md) for the template and guidelines.

**Quick start:**

```bash
cp adr-template.md 001-your-decision-title.md
# Edit the file
# Update this README
# Update ../../../../docs/adr-index.md
```

---

**Related Documentation:**

- [Master ADR Index](../../../docs/adr-index.md)
- [GlassOps ADRs](../../../docs/adr)
- [Runtime Package README](../README.md)
