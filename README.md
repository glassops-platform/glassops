# GlassOps™

[![Runtime Build & Test](https://github.com/glassops-platform/glassops/actions/workflows/runtime-docker.yml/badge.svg)](https://github.com/glassops-platform/glassops/actions/workflows/runtime-docker.yml)
[![Docs Governance & Build](https://github.com/glassops-platform/glassops/actions/workflows/docs-governance-deploy.yml/badge.svg)](https://github.com/glassops-platform/glassops/actions/workflows/docs-governance-deploy.yml)
[![Pages Deployment](https://github.com/glassops-platform/glassops/actions/workflows/pages/pages-build-deployment/badge.svg)](https://github.com/glassops-platform/glassops/actions/workflows/pages/pages-build-deployment)

> **Governance-first Salesforce DevOps.**

## Status

GlassOps is being rebuilt as a CI-agnostic command line tool that acts as the CI/CD layer for Salesforce delivery, with other platforms possible later. This repository is mid-rebuild and nothing here is released. What exists today is the first part of the new `glassops` CLI and the older GitHub Action runtime in `packages/runtime`, both written in Go. It is early work, not a finished product.

---

## What Exists Today

### The `glassops` CLI

The new CLI lives at the repository root (`cmd/`, `internal/`, `schemas/`). So far it has these commands:

- `glassops init` detects a Salesforce project and the CI system, then writes `glassops.yaml` and a CI file.
- `glassops doctor` checks the config against its schema, the tools (`git`, `sf`, `sfdx-git-delta`), the git checkout and the environment variables the config names.
- `glassops describe`, `glassops plan` and `glassops validate` log in to a target with an org Id and tier guard, work out a change set between two git refs, and run a check-only validate. `doctor --auth` checks the logins.

Every command writes a result file to `.glassops/results/` and exits 0 on pass, 1 on a failed check and 2 on a tool error. See the [CLI reference](docs/reference/cli.md).

### The runtime

The runtime in [`packages/runtime`](packages/runtime/README.md) runs as a Docker container action. In order, it:

1. Loads `devops-config.json` and checks configured freeze windows.
2. Optionally runs static analysis through the Salesforce CLI.
3. Authenticates to a Salesforce org with the JWT flow.
4. Writes a permit file and a deployment contract file under `.glassops/`.

The permit and the contract are plain JSON files. They are not signed and they do not yet carry a trustworthy record of what ran. Treat them as a design sketch of the final output, not as an audit record.

The runtime does not deploy anything. Your workflow runs the deploy after the runtime step finishes.

---

## Learn More

- **Architecture overview:** [`overview.md`](docs/architecture/overview.md)
- **Vision:** [`vision.md`](docs/vision.md)
- **CLI reference:** [`cli.md`](docs/reference/cli.md)
- **Runtime reference:** [`platform-reference.md`](docs/reference/platform-reference.md)
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

**Not yet released.** No tag or release exists, so there is no pinned version to reference. To try the runtime today, build it from a checkout of this repository. See [`packages/runtime/README.md`](packages/runtime/README.md) for local development and for the inputs and outputs. The example workflow is in [`examples/workflows/glassops-deploy.yml`](examples/workflows/glassops-deploy.yml).

---

## Contributing

- Report bugs or suggest features
- Improve the Go runtime in `packages/runtime`
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
