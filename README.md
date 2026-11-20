# GlassOps™

[![Docs Governance & Build](https://github.com/glassops-platform/glassops/actions/workflows/docs-governance-deploy.yml/badge.svg)](https://github.com/glassops-platform/glassops/actions/workflows/docs-governance-deploy.yml)
[![Pages Deployment](https://github.com/glassops-platform/glassops/actions/workflows/pages/pages-build-deployment/badge.svg)](https://github.com/glassops-platform/glassops/actions/workflows/pages/pages-build-deployment)

> **Governance-first Salesforce DevOps.**

## Status

GlassOps is a CI-agnostic command line tool that acts as the CI/CD layer for Salesforce delivery, with other platforms possible later. It is written in Go and replaces the older GitHub Action runtime that this repository once held.

The current release is v1.1.0. It adds opt-in Sigstore signing, Rego rules, a pull request comment and OpenTelemetry export. v1.0.0 was the first final release. All releases are on the [releases page](https://github.com/glassops-platform/glassops/releases). GlassOps has deployed a real Salesforce project to production through its full deploy pipeline: plan, validate, gate, deploy, verify and report, in a protected GitHub environment. [What is in which release](docs/reference/releases-and-licenses.md#what-is-in-which-release) lists every command and feature with the release that first has it.

---

## What Exists Today

### The `glassops` CLI

The CLI lives at the repository root (`cmd/`, `internal/`, `schemas/`). Each step is its own command, and `glassops run` runs a whole pipeline of them. Salesforce is the built-in provider, and provider `exec` runs your own commands for any other deploy tool. The CI layers are GitHub Actions, Bitbucket Pipelines and GitLab CI. The last two have not run on a real runner yet.

| Command | What it does |
|---|---|
| `init` | Detects the provider and CI system, then writes `glassops.yaml` and a CI file |
| `doctor` | Checks the config, tools, git checkout and environment variables. With `--auth` it also logs in to targets |
| `describe` | Logs in to a target, runs the org Id guard and records what the target reports |
| `plan` | Works out the change set between two git refs |
| `validate` | Runs a check-only validate of the change set. On a scratch org it first seeds what a destructive change needs |
| `gate` | Evaluates the policy rules for a target into a signed PERMIT |
| `permit verify` | Checks that a PERMIT is signed, unexpired and bound to this target, commit, plan and config |
| `deploy` | Deploys the change set, only under a valid PERMIT, then moves the deployed tag |
| `verify` | After a deploy, checks that the target holds what the plan says |
| `report` | Summarises every result as markdown and SARIF |
| `env create`, `env delete`, `env sweep` | Create, delete and sweep ephemeral scratch orgs that later steps use |
| `test` | Runs Apex tests with coverage on a target and writes JUnit XML |
| `check` | Runs your own project scripts |
| `analyze` | Runs the Code Analyzer and records violation counts. Needs `analyze.enabled: true` and the `glassops-analyzer` image |
| `run` | Runs a named pipeline from `glassops.yaml`, such as the default `pr` and `deploy` |

Commands exit 0 on pass, 1 on a failed check and 2 on a tool error, and every command except `report` writes a result file to `.glassops/results/`. See the [CLI reference](docs/reference/cli.md), the [configuration reference](docs/reference/config.md), [PERMIT and rules](docs/reference/permit-and-rules.md), [Run a pipeline](docs/guides/run-a-pipeline.md) and the [GitHub deploy guide](docs/guides/deploy-with-github.md).

---

## Learn More

- **Architecture overview:** [`overview.md`](docs/architecture/overview.md)
- **CLI reference:** [`cli.md`](docs/reference/cli.md)
- **Configuration reference:** [`config.md`](docs/reference/config.md)
- **PERMIT and rules:** [`permit-and-rules.md`](docs/reference/permit-and-rules.md)
- **Run a pipeline:** [`run-a-pipeline.md`](docs/guides/run-a-pipeline.md)
- **Deploy with GitHub:** [`deploy-with-github.md`](docs/guides/deploy-with-github.md)
- **Bitbucket and GitLab:** [`other-ci.md`](docs/guides/other-ci.md)
- **Scratch org definitions and settings:** [`scratch-org-settings.md`](docs/guides/scratch-org-settings.md)
- **The exec provider:** [`exec-provider.md`](docs/reference/exec-provider.md)
- **Sigstore signing (opt-in):** [`sigstore-signing.md`](docs/reference/sigstore-signing.md)
- **Rego rules (opt-in):** [`rego-rules.md`](docs/reference/rego-rules.md)
- **Pull request comment and OpenTelemetry export (opt-in):** [`pr-comment-and-telemetry.md`](docs/reference/pr-comment-and-telemetry.md)
- **Verify, coverage and YAML settings:** [`verify-and-coverage-settings.md`](docs/reference/verify-and-coverage-settings.md)
- **Releases, images and licenses:** [`releases-and-licenses.md`](docs/reference/releases-and-licenses.md)
- **Design history (direction and earlier designs, not current functionality):** the [vision](https://github.com/glassops-platform/glassops/blob/main/docs/vision.md), the [white paper](https://github.com/glassops-platform/glassops/blob/main/docs/whitepaper.md) and the [ADRs](https://github.com/glassops-platform/glassops/blob/main/docs/adr/README.md) are kept in the repository and are not part of the docs site

---

## Who This Is For

Teams that:

- Already use (or want to use) **GitHub Actions**
- Care about **auditability and compliance**
- Prefer **transparent systems over black boxes**

---

## Quick Start

The image to use is `ghcr.io/glassops-platform/glassops:v1`, which follows the latest 1.x release. For a reproducible run, use `ghcr.io/glassops-platform/glassops:1.0.0` or pin the image by digest. To build from a checkout instead, run `go build -o glassops ./cmd/glassops` for the CLI or `docker build -t glassops .` for the container image.

On GitHub, [`examples/consumer/`](examples/consumer/) is the smallest working layout. It holds a `glassops.yaml` and two short workflows that call the reusable workflows `glassops-pr.yml` and `glassops-deploy.yml` from this repository. The deploy caller passes `GLASSOPS_PERMIT_KEY` by name, and the value lives only as a secret of the protected `glassops-deploy` environment. See the [GitHub deploy guide](docs/guides/deploy-with-github.md). Or run `glassops init` in a Salesforce DX project to generate a similar setup.

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
