# ADR-004: Schema Contract over YAML Parsing

**Date:** 2026-01-19
**Status:** Accepted
**Deciders:** Ryan Bumstead

---

## Context

Parsing workflow_dispatch YAML files in Apex (Level 3) to build dynamic UI forms is brittle (regex, indentation issues) and prone to breakage.

## Decision

We enforce a **Schema Contract** pattern. Inputs must be explicitly defined in Salesforce Custom Metadata (Workflow_Template\_\_mdt). Apex does NOT parse YAML.

## Consequences

**Positive:**

- Robustness. Eliminates YAML parsing bugs.
- Typed inputs in Salesforce.

**Negative:**

- Double maintenance (YAML + CMDT records).

## Mitigation

- Automation tools to sync YAML inputs to CMDT (future roadmap).

## Alternatives

- None considered.
