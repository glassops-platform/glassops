# GlassOps™

[![Docs Governance & Build](https://github.com/glassops-platform/glassops/actions/workflows/docs-governance-deploy.yml/badge.svg)](https://github.com/glassops-platform/glassops/actions/workflows/docs-governance-deploy.yml)
[![Pages Deployment](https://github.com/glassops-platform/glassops/actions/workflows/pages/pages-build-deployment/badge.svg)](https://github.com/glassops-platform/glassops/actions/workflows/pages/pages-build-deployment)

> **Governance-first Salesforce DevOps.**

## Status

GlassOps is being rebuilt as a CI-agnostic command line tool that acts as the CI/CD layer for Salesforce delivery, with other platforms possible later. This repository is mid-rebuild and nothing here is released. What exists today is the first part of the `glassops` CLI, written in Go. It replaces the older GitHub Action runtime that this repository once held. It is early work, not a finished product.

---

## What Exists Today

### The `glassops` CLI

The CLI lives at the repository root (`cmd/`, `internal/`, `schemas/`). Each step is its own command. Today these exist: `init`, `doctor`, `describe`, `plan`, `validate`, `gate`, `permit verify`, `deploy` and `report`. In order, they set up a project, check the environment, log in to a target and run the org Id guard, work out a change set, run a check-only validate, evaluate policy into a signed PERMIT, check that PERMIT, deploy only under it, and summarise the results as markdown and SARIF. Commands exit 0 on pass, 1 on a failed check and 2 on a tool error, and every command except `report` writes a result file to `.glassops/results/`. See the [CLI reference](docs/reference/cli.md), [PERMIT and rules](docs/reference/permit-and-rules.md) and the [GitHub deploy guide](docs/guides/deploy-with-github.md).

---

## Learn More

- **Architecture overview:** [`overview.md`](docs/architecture/overview.md)
- **Vision:** [`vision.md`](docs/vision.md)
- **CLI reference:** [`cli.md`](docs/reference/cli.md)
- **PERMIT and rules:** [`permit-and-rules.md`](docs/reference/permit-and-rules.md)
- **Deploy with GitHub:** [`deploy-with-github.md`](docs/guides/deploy-with-github.md)
- **White paper (describes direction, not current functionality):** [`whitepaper.md`](docs/whitepaper.md)
- **ADRs:** [earlier design decisions](docs/adr/README.md), each with a status line

---

## Who This Is For

Teams that:

- Already use (or want to use) **GitHub Actions**
- Care about **auditability and compliance**
- Prefer **transparent systems over black boxes**

---

## Quick Start

**Not yet released.** No tag or release exists, so there is no pinned version to reference. To try the CLI today, build it from a checkout of this repository with `go build -o glassops ./cmd/glassops`, or build the container image with `docker build -t glassops .`. The example workflow is in [`examples/workflows/glassops-deploy.yml`](examples/workflows/glassops-deploy.yml).

---

## Contributing

- Report bugs or suggest features
- Improve the Go CLI
- See [`CONTRIBUTING.md`](CONTRIBUTING.md) for details.

---

## License

- **Code:** Apache 2.0. See [LICENSE](LICENSE)
- **Documentation:** Creative Commons Attribution 4.0 International. See [docs/LICENSE](docs/LICENSE)

---

## Contact

- **Issues:** [GitHub Issues](https://github.com/glassops-platform/glassops/issues)
- **Author:** Ryan Bumstead ([@rdbumstead](https://github.com/rdbumstead))
- **Maintainer:** Ryan Bumstead ([@rdbumstead](https://github.com/rdbumstead))
