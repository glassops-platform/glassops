# GlassOps Control Plane - Architecture Decision Records

This directory contains Architecture Decision Records (ADRs) specific to the GlassOps Control Plane.

## Scope

ADRs in this directory cover decisions related to:

- Control plane architecture and technology choices
- Policy store implementation
- Kubernetes operator patterns
- Multi-tenant isolation
- API design and versioning

## Current ADRs

| Number | Title         | Status | Date |
| ------ | ------------- | ------ | ---- |
| -      | _No ADRs yet_ | -      | -    |

## Proposed Decisions to Document

1. **Go vs Python for Control Plane** - Language choice rationale
2. **PostgreSQL vs TimescaleDB** - Policy store database selection
3. **Kubernetes Operator Pattern** - CRD and controller design
4. **Multi-Tenant Isolation Strategy** - Namespace vs cluster isolation
5. **Policy Resolution Algorithm** - How effective policy is calculated
6. **API Versioning Strategy** - v1alpha1, v1beta1, v1 progression

## Creating a New ADR

See the [Master ADR Index](../../../docs/adr-index.md) for the template and guidelines.

---

**Related Documentation:**

- [Master ADR Index](../../../docs/adr-index.md)
- [Platform ADRs](../../../docs/adr)
- [Control Plane Docs](../../../README.md)
