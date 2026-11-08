# Deploy with GitHub

> **Early work.** v0.3.0 is released as a prerelease. The reusable workflows, the `run` command and the other new commands this guide uses ship in the next release.

This guide sets up pull request checks and a deploy to production on GitHub. Two reusable workflows in this repository run the `glassops` container image for you. Your repository calls them with a few lines each. [`examples/consumer/`](https://github.com/glassops-platform/glassops/tree/main/examples/consumer) is the smallest working layout, and `glassops init` writes a similar `glassops.yaml` and workflow into your project. For what a PERMIT is, see [PERMIT and rules](../reference/permit-and-rules.md).

| Reusable workflow | Runs | Triggered by |
|---|---|---|
| `glassops-pr.yml` | `glassops run pr` | `pull_request` |
| `glassops-deploy.yml` | `glassops run deploy`, in a protected environment | a push to `main` |

The pipelines are the ones in your `glassops.yaml`. See [Run a pipeline](run-a-pipeline.md) for the defaults.

---

## Prerequisites

- A `glassops.yaml` with a Dev Hub target and a target named `prod` with tier `prod`, each with a login. See the [configuration reference](../reference/config.md) and [Authentication](../reference/cli.md#authentication).
- The Salesforce secrets your targets name, such as `SF_JWT_KEY`, `SF_CLIENT_ID`, `SF_USERNAME`, `SF_DEVHUB_ORG_ID`, `SF_PROD_USERNAME` and `SF_PROD_ORG_ID`, as repository secrets.
- A GitHub environment named `glassops-deploy`. Give it required reviewers if you want an approval before the deploy job runs, and a deployment branch rule that allows `main` only. The run's approval is read as evidence for the separation of duties rule.
- The signing key `GLASSOPS_PERMIT_KEY` as a secret **of that environment**, not of the repository. Make one with `openssl rand -base64 48`. It must decode to at least 32 bytes.

**The key must never be available to pull request jobs.** Anyone who can run a job with the key can sign an allow PERMIT. A secret held only in the `glassops-deploy` environment reaches only jobs that name that environment, and you control which branches may use it. `glassops doctor` warns when `GLASSOPS_PERMIT_KEY` is set on a `pull_request` or `pull_request_target` run.

---

## Pull requests

Save this as `.github/workflows/pr.yml`:

```yaml
name: glassops pr

on:
  pull_request:

permissions:
  contents: read

jobs:
  glassops:
    uses: glassops-platform/glassops/.github/workflows/glassops-pr.yml@v1
    permissions:
      contents: read
      pull-requests: read
      security-events: write
    with:
      targets: devhub
    secrets:
      SF_JWT_KEY: ${{ secrets.SF_JWT_KEY }}
      SF_CLIENT_ID: ${{ secrets.SF_CLIENT_ID }}
      SF_USERNAME: ${{ secrets.SF_USERNAME }}
      SF_DEVHUB_ORG_ID: ${{ secrets.SF_DEVHUB_ORG_ID }}
      SF_PROD_USERNAME: ${{ secrets.SF_PROD_USERNAME }}
      SF_PROD_ORG_ID: ${{ secrets.SF_PROD_ORG_ID }}
```

The workflow checks out full history, runs `glassops run pr`, uploads the SARIF file to code scanning, and keeps the results as an artifact. It never receives the permit key.

| Input | Default | Meaning |
|---|---|---|
| `image` | `ghcr.io/glassops-platform/glassops:v1` | The container image. Pass the `glassops-analyzer` image when `analyze.enabled` is true |
| `targets` | | Space separated target names to log in to before the pipeline, so a bad login fails early |
| `upload-sarif` | `true` | Upload the SARIF file. Needs code scanning on the repository |

---

## Deploy

Save this as `.github/workflows/deploy.yml`:

```yaml
name: glassops deploy

on:
  push:
    branches: [main]

permissions:
  contents: read

jobs:
  glassops:
    uses: glassops-platform/glassops/.github/workflows/glassops-deploy.yml@v1
    permissions:
      contents: write
      actions: read
      issues: read
      pull-requests: read
    with:
      environment: glassops-deploy
    secrets:
      GLASSOPS_PERMIT_KEY: ${{ secrets.GLASSOPS_PERMIT_KEY }}
      SF_JWT_KEY: ${{ secrets.SF_JWT_KEY }}
      SF_CLIENT_ID: ${{ secrets.SF_CLIENT_ID }}
      SF_USERNAME: ${{ secrets.SF_USERNAME }}
      SF_DEVHUB_ORG_ID: ${{ secrets.SF_DEVHUB_ORG_ID }}
      SF_PROD_USERNAME: ${{ secrets.SF_PROD_USERNAME }}
      SF_PROD_ORG_ID: ${{ secrets.SF_PROD_ORG_ID }}
```

`glassops-deploy.yml` takes the inputs `image`, `environment` (default `glassops-deploy`) and `targets`, like the pull request workflow.

**The permit key is passed by name.** The caller must list `GLASSOPS_PERMIT_KEY` under `secrets`, because a called workflow does not see environment secrets unless the caller passes them. The workflow declares the secret `required: true`, so a caller that omits it fails. The value lives only as a secret of the protected `glassops-deploy` environment. Do not add it as a repository secret. The deploy job runs in that environment, so the reviewers and branch rules you set there apply to it.

**Permissions.** The calling job needs these, because a called workflow cannot hold more than its caller grants:

| Permission | Why |
|---|---|
| `contents: write` | The runner pushes the `glassops/deployed/prod` tag |
| `actions: read` | `gate` reads this run's environment approvals |
| `pull-requests: read` | `gate` reads the pull request behind the commit, its author and labels |
| `issues: read` | `gate` reads who added an approval label |

### What the deploy job does

1. Checks out full history. A shallow clone is an error, never an empty plan.
2. Installs a `glassops` command on the runner that runs the image with `docker run --rm`, one container per command, with the workspace mounted at `/work`. The checkout, the tag push and the uploads run on the runner, not in the container.
3. Logs in to each target in `targets`, when set, with `glassops doctor --auth --target <name>`.
4. Runs `glassops run deploy`. With the default pipeline that is `plan --target prod`, `validate --target prod`, `gate --target prod`, `deploy --target prod` and `report`.
5. **Pushes the deployed tag.** `deploy` moves the git tag `glassops/deployed/prod` inside the container, but the container has no git credentials, so its own push is only a warning. A runner step after the container exits pushes the `glassops/deployed/*` tags with the checkout's credentials. Without that tag, the next `plan` has no record of this deploy and starts from the merge base instead.
6. Runs `glassops report` (also after a failure), and uploads `.glassops/results/`, `plan.json` and `permit.json` as the `glassops-deploy-evidence` artifact.

One deploy runs at a time per environment. A queued deploy waits and is never cancelled.

### How the steps fit together

- **Plan.** With `--target prod` and no `--from`, `plan` starts from the `glassops/deployed/prod` tag. If that tag does not exist yet, this is the first deploy and the whole source is planned.
- **Validate.** A check-only validate against `prod`. It records which plan it validated, so `gate` can bind it into the PERMIT.
- **Gate.** Evaluates the rules and writes `.glassops/permit.json`. A deny exits 1, so the run stops here and the deploy never starts.
- **Deploy.** Verifies the PERMIT, then logs in and deploys. When the PERMIT holds a passed validation of this plan, it quick deploys it.
- **Report.** Runs after a failed step, writes the run summary and SARIF, and exits with the worst status of the results.

### Notes

- The references say `@v1`. The `v1` workflow tag and the `:v1` image tag move only on a final 1.x release. Until one exists, replace `@v1` with a release tag or a commit SHA, and set the `image` input to a version tag.
- Separation of duties needs the rule in `glassops.yaml`. When the person who merges is also the pull request author, the rule fails unless an approval satisfies it. Set `approval.environment: glassops-deploy` so an approval of the environment satisfies the rule. See [PERMIT and rules](../reference/permit-and-rules.md#separation_of_duties).
- The container runs as root, so files it writes under `.glassops/` are owned by root on the runner.
- The warning about the tag push inside the container appears on every deploy. That is expected, because the runner step pushes the tag.

---

## Other CI systems and single commands

`glassops` is one binary, so any CI can run `glassops run pr` and `glassops run deploy` with the same `glassops.yaml`. On GitHub, the `glassops-platform/glassops` action runs one command in the container, for steps that do not fit a pipeline:

```yaml
- uses: glassops-platform/glassops@v1
  with:
    command: doctor --auth --target prod
    env: |
      SF_JWT_KEY
      SF_CLIENT_ID
  env:
    SF_JWT_KEY: ${{ secrets.SF_JWT_KEY }}
    SF_CLIENT_ID: ${{ secrets.SF_CLIENT_ID }}
```
