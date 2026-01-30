# GlassOps™

[![Docs](https://github.com/glassops-platform/glassops/actions/workflows/docs-governance-deploy.yml/badge.svg)](https://github.com/glassops-platform/glassops/actions/workflows/docs-governance-deploy.yml)

> **Governance-first Salesforce DevOps.**
>
> GlassOps governs outcomes, not implementations.

GlassOps is an **open governance protocol** for Salesforce CI/CD. It enforces architectural and quality standards _before_ deployments execute — without locking teams into a specific tool or vendor.

---

## What Problem Does GlassOps Solve?

Most Salesforce DevOps tools force a tradeoff:

- **Fast but opaque** (black-box platforms)
- **Transparent but chaotic** (raw CLI scripts)

GlassOps removes the tradeoff.

It separates:

- **Governance** (Policy, Standards, Audit)  
  from
- **Execution** (CLI tools, workflows, automation)

You choose _how_ you deploy.  
GlassOps decides _whether you should_.

---

## The Mental Model

```
Intent → Policy → Simulation → Contract → Enforcement → Execution
```

- **Intent:** A deployment is requested
- **Policy:** Quality & architectural standards are resolved
- **Simulation:** A dry run validates the change
- **Contract:** Results are normalized into a governance record
- **Enforcement:** Policy decides pass/fail
- **Execution:** Deployment only proceeds if governance passes

GlassOps never parses raw CLI output. It only reads a **Deployment Contract**.

---

## What GlassOps Is (and Is Not)

### GlassOps _is_:

- A **governance control plane**
- A **portable policy layer (OPA)**
- A **protocol** that execution tools must follow
- A **Container-First architecture** (Executed via GitHub Actions or K8s)
- An **audit and compliance amplifier**

### GlassOps is _not_:

- A CI/CD pipeline engine
- A replacement for sfdx-hardis, Gearset, or Copado
- A proprietary black box
- A commercial SaaS product

---

## Execution Engines (Adapters)

GlassOps works with multiple execution tools via adapters:

- **GlassOps Native** — raw `sf` CLI (maximum transparency)
- **GlassOps Hardis** — wraps `sfdx-hardis` (maximum velocity)
- **Bring Your Own** — implement the adapter interface

Switch engines without rewriting governance.

---

## Adoption Levels

- **Level 1:** Safe execution (adapters only)
- **Level 2:** Governance protocol (policy + contracts)
- **Level 3:** Optional Salesforce UI (Reference Console)

You can stop at any level and still get value.

---

## Project Status

- **In active development**
- **Seeking testers and collaborators**

### Implementation Status

| Component              | Status                 | Description                                                                   |
| :--------------------- | :--------------------- | :---------------------------------------------------------------------------- |
| **Salesforce Adapter** | Active Development     | Containerized metadata translation (`docker://glassops/adapter-salesforce`).  |
| **Scanner Adapter**    | In Development (Draft) | Containerized architectural invariants (`docker://glassops/adapter-scanner`). |
| **Policy Engine**      | Active Development     | OPA-based policy resolution API.                                              |

This project is an **architectural reference implementation**, not a commercial product.

---

## Learn More

- **One-page overview:** [`overview.md`](docs/architecture/overview.md)
- **Vision & philosophy:** [`vision.md`](docs/vision.md)
- **Full architecture spec:** [`architecture.md`](docs/architecture/overview.md)

---

## Who This Is For

GlassOps is designed for teams that:

- Already use (or want to use) **GitHub Actions**
- Care about **auditability and compliance**
- Prefer **transparent systems over black boxes**
- Think in **systems and contracts**

If that's not you — that's okay. GlassOps is opinionated by design.

---

## Quick Start

```yaml
# .github/workflows/deploy.yml
jobs:
    deploy:
        runs-on: ubuntu-latest
        steps:
            - uses: actions/checkout@v4

            - name: Resolve Policy
              uses: docker://glassops/policy-resolver:v1
              with:
                  args: resolve --org ${{ github.repository_owner }} --output .glassops/policy.json

            - name: Deploy with Governance
              uses: docker://glassops/adapter-salesforce:v1
              with:
                  args: transform --input force-app --policy-ref .glassops/policy.json
```

See [doc-map.md](docs/doc-map.md) for available guides.

---

## Contributing

GlassOps is intentionally modular. You don't need to adopt the whole platform to contribute.

- Build an adapter for your favorite tool
- Improve policy resolution logic
- Add audit visualizations
- Report bugs or suggest features
- See [`CONTRIBUTING.md`](CONTRIBUTING.md) for details.

---

## License

- **Code:** Apache 2.0 — See [LICENSE](LICENSE)
- **Documentation:** Creative Commons Attribution 4.0 International — See [LICENSE-DOCS](docs/LICENSE)

---

## Contact

- **Discussions:** [GitHub Discussions](https://github.com/glassops-platform/glassops/discussions)
- **Issues:** [GitHub Issues](https://github.com/glassops-platform/glassops/issues)
- **Author:** Ryan Bumstead ([@rdbumstead](https://github.com/rdbumstead))
- **Maintainer:** Ryan Bumstead ([@rdbumstead](https://github.com/rdbumstead))
