# ADR 008: Federated Documentation Structure

**Status:** Accepted  
**Date:** 2026-01-24  
**Deciders:** Ryan Bumstead

---

## Context

As GlassOps evolves from a conceptual framework to a deployable platform, the documentation architecture must scale to support three distinct audiences: **Architects** (evaluating the protocol), **Implementers** (building adapters), and **Operators** (running the control plane).

A flat structure creates friction for these users, mixing high-level philosophy with low-level implementation details. To support the "Protocol Ecosystem" vision, the documentation structure must mirror the separation of concerns found in the code itself—creating clear, navigable paths for each stakeholder while maintaining a single source of truth for core principles.

---

## Decision

**Adopt a Federated Documentation Structure** with clear separation of concerns:

### 7-Tier Hierarchy

```
docs/
├── Tier 1: Entry Points (under 5 min)
│   ├── overview.md
│   └── README.md (root)
│
├── Tier 2: Philosophy & Story (15-20 min)
│   └── philosophy.md  [NEW - Merged vision_v1_legacy + executive-summary]
│
├── Tier 3: Technical Architecture (30-60 min)
│   └── architecture/
│       ├── protocol.md  [NEW - Links to glassspec/]
│       ├── current.md   [Future - extracted from vision.md]
│       └── future.md    [Moved from future-architecture.md]
│
├── Tier 4: Business Vision (30-45 min)
│   └── vision/
│       ├── technical.md  [Future - extracted from vision.md]
│       ├── business.md   [Future - extracted from vision.md]
│       └── roadmap.md    [Future - extracted from vision.md]
│
├── Tier 5: Implementation Guides (30+ min)
│   └── guides/
│       ├── getting-started.md
│       ├── adapter-development.md
│       ├── level-3-setup.md
│       └── comparison.md
│
├── Tier 6: Reference Material
│   └── reference/
│       ├── platform-reference.md
│       ├── troubleshooting.md
│       ├── notifications.md
│       └── governance-without-lockin.md
│
└── Tier 7: Meta Documentation
    ├── doc-map.md
    ├── adr/  (Architecture Decision Records)
    └── specs/  (Technical specifications)
```

---

## Rationale

### Why Federated Structure

1. **Clear Audience Targeting** - Different tiers serve different readers
    - Entry points for evaluators
    - Philosophy for architects
    - Guides for implementers
    - Reference for operators

2. **Single Source of Truth** - No duplicate content
    - One philosophy document (not three)
    - Clear delineation between guides and reference
    - Protocol authority reinforced

3. **Scalability** - Easy to add content without clutter
    - New guides go in guides/
    - New references go in reference/
    - No flat directory confusion

4. **Protocol Clarity** - glassspec/ remains authoritative
    - architecture/protocol.md links to glassspec/
    - Reinforces Law (Protocol) vs Manual (Docs) separation

### Why "Federated"

The term "Federated" reflects the separation of concerns:

- **Protocol (Law):** `packages/glassspec/` - Immutable, versioned, authoritative
- **Manual (Guide):** `docs/` - Tutorials, examples, implementation
- **Implementations:** `packages/runtime/`, `packages/control-plane/`, `packages/adapters/`

Each has its own documentation, but all defer to the Protocol.

---

## Implementation

### Phase 1: Core Restructure (Completed)

1. ✅ Created directory structure (`architecture/`, `vision/`, `guides/`, `reference/`)
2. ✅ Created `philosophy.md` by merging:
    - `vision_v1_legacy.md` (principles, 8/10 vision, adapter philosophy)
    - `executive-summary.md` (origin story, AI era positioning)
3. ✅ Moved files to logical locations:
    - 4 guides → `guides/`
    - 4 references → `reference/`
4. ✅ Created `architecture/protocol.md` linking to `glassspec/`
5. ✅ Moved `future-architecture.md` → `architecture/future.md`
6. ✅ Deleted legacy files (`vision_v1_legacy.md`, `executive-summary.md`)

### Phase 2: Large File Splitting (Future)

Large files still in root need splitting:

- `vision.md` (1201 lines) → `vision/technical.md`, `vision/business.md`, `vision/roadmap.md`
- `architecture.md` (1003 lines) → `architecture/current.md`
- `doc-map.md` → Complete rewrite for new structure

---

## Consequences

### Positive

- ✅ No duplicate content in reorganized files
- ✅ Clear purpose for each document
- ✅ Logical grouping by audience and type
- ✅ Easy navigation with subdirectories
- ✅ Protocol authority reinforced
- ✅ Scalable for future growth

### Negative

- ⚠️ Existing links may break (need doc-map.md update)
- ⚠️ Learning curve for new structure
- ⚠️ Large files still need splitting (Phase 2 work)

### Neutral

- 📝 More directories but clearer organization
- 📝 Longer file paths but more intuitive
- 📝 Requires maintenance of directory structure

---

## Design Principles

### 1. Clear Ownership

| Document Type | Owner              | Purpose                             |
| ------------- | ------------------ | ----------------------------------- |
| Protocol      | glassspec/         | Law - What MUST be                  |
| Philosophy    | docs/philosophy.md | Why - Core principles               |
| Architecture  | docs/architecture/ | How - Technical implementation      |
| Guides        | docs/guides/       | Do - Step-by-step tutorials         |
| Reference     | docs/reference/    | Lookup - API specs, troubleshooting |

### 2. No Duplication

- Each concept documented once
- Cross-links when referencing other docs
- Philosophy document is single source for principles

### 3. Protocol Supremacy

- `glassspec/` is always authoritative
- `docs/` defers to Protocol
- architecture/protocol.md reinforces this hierarchy

### 4. Audience-First

- Tier 1: Quick evaluation (3-5 min)
- Tier 2: Understand philosophy (15-20 min)
- Tier 3-4: Technical depth (30-60 min)
- Tier 5: Implement (30+ min)
- Tier 6: Operate (reference)

---

## Alternatives Considered

### Alternative 1: Keep Flat Structure

**Rejected:** 17 files in single directory is unmanageable and will only grow

### Alternative 2: Split by Topic (e.g., governance/, deployment/, monitoring/)

**Rejected:** Doesn't serve different audiences; creates artificial topic boundaries

### Alternative 3: Single Mega-Document

**Rejected:** vision.md at 3000 lines proves this doesn't scale

### Alternative 4: Wiki-Style with Heavy Cross-Linking

**Rejected:** Requires maintenance of link graph; less discoverable than directory structure

---

## Validation Checklist

- [x] No duplicate content between moved files
- [x] Each file has clear, single purpose
- [x] Reading paths make sense for different audiences
- [x] Protocol files (glassspec/) are authoritative and referenced
- [x] Legacy files deleted
- [ ] All links updated (Phase 2 - doc-map.md rewrite)
- [ ] Large files split (Phase 2 - optional)

---

## Migration Path

For teams with existing documentation:

1. **Audit** - Identify overlap and legacy content
2. **Categorize** - Sort docs into Entry/Philosophy/Architecture/Vision/Guides/Reference
3. **Consolidate** - Merge duplicate content
4. **Move** - Relocate to appropriate subdirectories
5. **Link** - Update cross-references
6. **Delete** - Remove legacy files
7. **Validate** - Ensure no broken links

---

## Future Enhancements

### Potential Additions

- **docs/tutorials/** - Separate from guides/ for hands-on workshops
- **docs/examples/** - Sample implementations
- **docs/api/** - Auto-generated API documentation
- **docs/diagrams/** - Centralized architecture diagrams

### Automation Opportunities

- **Link checker** - Validate all cross-references
- **Doc linter** - Enforce structure and formatting
- **Auto-generated indexes** - Build doc-map.md from metadata
- **Versioning** - Tag docs with Protocol version compatibility

---

## Related ADRs

- [ADR-006: Documentation as Governed Artifact](006-documentation-as-governed-artifact.md)
- [ADR-007: Protocol Supremacy Enforcement](007-protocol-supremacy-enforcement.md)

---

## References

- [Divio Documentation System](https://documentation.divio.com/) - Inspiration for four-quadrant structure
- [GlassOps Protocol](../../packages/glassspec/README.md)

---

**Author:** Ryan Bumstead  
**Implemented:** 2026-01-24  
**Status:** Active (Phase 1 Complete)

## Alternatives

- None considered.
