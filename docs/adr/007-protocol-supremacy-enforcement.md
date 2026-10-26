# ADR 007: Protocol Supremacy Enforcement in Adapter Development

**Status:** Accepted  
**Date:** 2026-01-24  
**Deciders:** Ryan Bumstead

---

> **Note:** The packages this record refers to (`glassspec`, `control-plane`, `adapters`, and the Reference Console) are not part of this repository. The decision text is kept as history.

## Context

The core value proposition of GlassOps is the **Layered Contract Model**, where governance decisions are made against standard, interoperable schemas (SARIF, CloudEvents).

Analysis of the adapter ecosystem revealed a risk of **schema fragmentation**: allowing adapters to define custom output formats weakens the governance layer's ability to enforce universal policy. To ensure long-term interoperability and protocol integrity, the documentation must strictly enforce the use of canonical, versioned contracts over ad-hoc JSON structures.

### The Strategic Gap

**Custom Schema vs. Protocol Standard:**

Teaching developers to build adapters with custom schemas creates a "Tower of Babel" effect where the Policy Engine cannot reliably interpret findings across different tools. Aligning all adapters to **SARIF 2.1.0** is necessary to unlock the platform's ability to normalize and govern diverse signals.

---

## Decision

**The Manual (glassops/docs/) MUST defer to the Protocol (glassspec/).**

All adapter development documentation will:

1. **Use SARIF 2.1.0** as defined in `glassspec/adapter-interface.md`
2. **Link to the Protocol** for authoritative schema definitions
3. **Reference Anti-Patterns** from `glassspec/protocol.md` for scope boundaries
4. **Emit contracts** to `.glassops/glassops-contract.sarif.json` (not custom filenames)

### Protocol Supremacy Hierarchy

```
glassspec/                       [LAW - Immutable, Versioned]
  ├── protocol.md               [The Protocol definition]
  └── adapter-interface.md      [Canonical SARIF schema]
        │
        ▼
glassops/docs/                   [MANUAL - Defers to Law]
  └── guides/adapter-development.md  [Must comply with Protocol]
```

**When in conflict:** `glassspec/` is always correct.

---

## Rationale

### Why This Matters

1. **Protocol Integrity** - The Layered Contract Model depends on SARIF conformance
2. **Interoperability** - Tools expect standard SARIF, not custom schemas
3. **Governance Enforcement** - Policy Engine parses SARIF results, not custom JSON
4. **Audit Trail** - SARIF contracts are immutable governance records

### Why Custom Schemas Are Dangerous

- **Schema Proliferation** - Every adapter invents its own format
- **Parser Fragmentation** - Governance layer can't normalize findings
- **Protocol Violation** - Breaks the Layered Contract Model
- **Ecosystem Confusion** - Developers don't know which schema to use

---

## Implementation

### Changes Made to adapter-development.md

1. **Added Protocol Supremacy Warning:**

    ```markdown
    > [!IMPORTANT]
    > **Protocol Supremacy:** All adapters MUST emit SARIF 2.1.0 as defined in
    > adapter-interface.md. Do NOT invent custom schemas.
    ```

2. **Replaced Custom Schema with SARIF 2.1.0:**
    - Removed custom JSON structure
    - Added minimal compliant SARIF example
    - Updated all code examples to emit SARIF

3. **Added Mapping Table:**
    - Tool Concept → SARIF Field mappings
    - Deployment metadata → `invocation.properties.glassops`
    - Governance findings → `results[]` array

4. **Added "What Goes in SARIF vs. What Doesn't":**
    - ✅ DO: Policy violations, test failures, coverage findings
    - ❌ DON'T: CPU metrics (use OTel), logs (use OTel Logs), trace IDs (use CloudEvents)
    - Links to Protocol Anti-Patterns section

5. **Updated Integration Logic:**
    - Check `invocation.executionSuccessful` (not custom `status` field)
    - Filter `results` for errors (not custom `quality.coverage.met`)
    - Extract from `invocation.properties.glassops.deploymentId`

---

## Consequences

### Positive

- ✅ Developers build Protocol-compliant adapters
- ✅ Documentation reinforces glassspec/ authority
- ✅ Clear guidance on SARIF scope boundaries
- ✅ Interoperability with SARIF tooling ecosystem

### Negative

- ⚠️ More verbose examples (SARIF is larger than custom JSON)
- ⚠️ Developers must learn SARIF structure
- ⚠️ Existing adapters may be non-compliant (requires audit)

### Neutral

- 📝 Manual must be updated when Protocol changes
- 📝 Reinforces separation between Law (glassspec/) and Guide (docs/)

---

## Validation

### How to Verify Compliance

```bash
# Validate SARIF schema
jq -e '.version == "2.1.0"' .glassops/glassops-contract.sarif.json
jq -e '.$schema | contains("sarif-2.1.0")' .glassops/glassops-contract.sarif.json

# Validate structure
jq -e '.runs[0].tool.driver.name' .glassops/glassops-contract.sarif.json
jq -e '.runs[0].invocations[0].executionSuccessful' .glassops/glassops-contract.sarif.json
```

---

## Future Considerations

### Potential Next Steps

1. **Audit Existing Adapters** - Check glassops-native, glassops-hardis, glassops-scanner for SARIF compliance
2. **Schema Validation** - Add automated SARIF schema validation to CI
3. **Contract Linter** - Create tool to validate SARIF contracts against Protocol
4. **Migration Guide** - If legacy adapters exist, provide migration path

---

## Related ADRs

- [ADR-001: System API Primitive](001-system-api-primitive.md)
- [ADR-004: Schema Contract Config](004-schema-contract-config.md)
- [ADR-006: Documentation as Governed Artifact](006-documentation-as-governed-artifact.md)

---

## References

- [OASIS SARIF 2.1.0 Specification](https://docs.oasis-open.org/sarif/sarif/v2.1.0/sarif-v2.1.0.html)

---

**Author:** Ryan Bumstead  
**Status:** Accepted. The `glassspec` and adapter packages it governs were removed from this repository.

## Alternatives

- None considered.
