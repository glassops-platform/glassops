# Deploy with GitHub

> **Early work.** The `glassops` CLI is being built in phases and nothing is released yet. This guide uses only the commands that exist today.

This guide sets up a deploy to production that runs when a change lands on `main`. One job plans the change, validates it, asks the policy rules for a signed PERMIT, deploys under that PERMIT and reports. It runs the `glassops` container image the same way the [end-to-end workflow](https://github.com/glassops-platform/glassops/blob/main/.github/workflows/cli-e2e.yml) does, with `docker run`. For what a PERMIT is, see [PERMIT and rules](../reference/permit-and-rules.md).

---

## Prerequisites

- A `glassops.yaml` with a target named `prod`, tier `prod`, and a login. See the [CLI reference](../reference/cli.md#authentication).
- Repository secrets, or secrets in the environment below, for the login: `SF_PROD_USERNAME`, `SF_CLIENT_ID` and `SF_JWT_KEY` in this example.
- A GitHub Environment named `glassops-deploy`. Add required reviewers to it if you want an approval before the job runs. The run's approval is read as evidence for the separation of duties rule.
- A signing key stored as the secret `GLASSOPS_PERMIT_KEY` **in that environment**, not as a repository secret. Make one with `openssl rand -base64 48`. It must decode to at least 32 bytes.

**The key must never be available to `pull_request` jobs.** Anyone who can run a job with the key can sign an allow PERMIT. A secret held in the `glassops-deploy` environment reaches only jobs that name that environment, and you control which branches may use it. `glassops doctor` warns when `GLASSOPS_PERMIT_KEY` is set on a `pull_request` or `pull_request_target` run.

---

## Installation

The image holds the CLI, the Salesforce CLI and `sfdx-git-delta`. The release workflow publishes it to `ghcr.io/glassops-platform/glassops` when a version tag is pushed. No release exists yet. Until one does, build the image from a checkout of this repository and use its name:

```bash
docker build -t glassops .
```

Then set `GLASSOPS_IMAGE` in the workflow to `glassops`, or to a published tag once one exists.

---

## Usage

Save this as `.github/workflows/glassops-deploy.yml`. It is also in the repository as [`examples/workflows/glassops-deploy.yml`](https://github.com/glassops-platform/glassops/blob/main/examples/workflows/glassops-deploy.yml).

```yaml
name: GlassOps Deploy

on:
  push:
    branches: [main]

concurrency:
  group: glassops-deploy-prod
  cancel-in-progress: false

env:
  GLASSOPS_IMAGE: ghcr.io/glassops-platform/glassops:latest

jobs:
  deploy:
    runs-on: ubuntu-latest
    environment: glassops-deploy
    permissions:
      contents: write # deploy pushes the glassops/deployed/prod tag
      pull-requests: read # gate reads the pull request author and labels
      actions: read # gate reads the environment approvals of this run
      issues: read # gate reads who added an approval label
    env:
      GLASSOPS_PERMIT_KEY: ${{ secrets.GLASSOPS_PERMIT_KEY }}
      GITHUB_TOKEN: ${{ github.token }}
      SF_PROD_USERNAME: ${{ secrets.SF_PROD_USERNAME }}
      SF_CLIENT_ID: ${{ secrets.SF_CLIENT_ID }}
      SF_JWT_KEY: ${{ secrets.SF_JWT_KEY }}
    steps:
      - uses: actions/checkout@v7
        with:
          fetch-depth: 0 # a shallow clone is an error, never an empty plan

      - name: Add a glassops command that runs the image
        run: |
          mkdir -p "$RUNNER_TEMP/bin"
          cat > "$RUNNER_TEMP/bin/glassops" <<'WRAPPER'
          #!/bin/sh
          exec docker run --rm \
            -e GITHUB_ACTIONS -e GITHUB_REPOSITORY -e GITHUB_REF -e GITHUB_SHA \
            -e GITHUB_EVENT_NAME -e GITHUB_ACTOR -e GITHUB_RUN_ID -e GITHUB_TOKEN \
            -e GITHUB_STEP_SUMMARY -e GITHUB_OUTPUT \
            -e GLASSOPS_PERMIT_KEY -e SF_PROD_USERNAME -e SF_CLIENT_ID -e SF_JWT_KEY \
            -v "$GITHUB_STEP_SUMMARY:$GITHUB_STEP_SUMMARY" -v "$GITHUB_OUTPUT:$GITHUB_OUTPUT" \
            -v "$PWD:/work" -w /work \
            "$GLASSOPS_IMAGE" "$@"
          WRAPPER
          chmod +x "$RUNNER_TEMP/bin/glassops"
          echo "$RUNNER_TEMP/bin" >> "$GITHUB_PATH"

      - name: Plan
        run: glassops plan --target prod

      - name: Validate
        run: glassops validate --target prod

      - name: Gate
        run: glassops gate --target prod

      - name: Deploy
        run: glassops deploy --target prod

      - name: Report
        if: always()
        run: glassops report

      - name: Upload results
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: glassops-results
          path: |
            .glassops/results/
            .glassops/plan.json
            .glassops/permit.json
            .glassops/report.sarif
          if-no-files-found: warn
```

### How the steps fit together

- **Plan.** With `--target prod` and no `--from`, `plan` starts from the `glassops/deployed/prod` tag. If that tag does not exist yet, this is the first deploy and the whole source is planned.
- **Validate.** A check-only validate against `prod`. It records which plan it validated, so `gate` can bind it into the PERMIT.
- **Gate.** Evaluates the rules and writes `.glassops/permit.json`. A deny exits 1, so the job stops here and the deploy never starts. `gate` reads the pull request, its labels and the run's approvals through the token, which is why the job needs the permissions listed above.
- **Deploy.** Verifies the PERMIT, then logs in and deploys. When the PERMIT holds a passed validation of this plan, it quick deploys it. Afterwards it moves the `glassops/deployed/prod` tag and pushes it.
- **Report.** `if: always()` makes it run after a failed step. It writes the run summary and SARIF, and exits with the worst status of the results, so a failure still fails the job.

### Notes

- The wrapper passes the `GITHUB_*` variables into the container because the CLI reads the commit, actor, run Id and token from them. It also mounts the step summary and output files so the container can write to them.
- The container runs as root, so files it writes under `.glassops/` are owned by root on the runner.
- A failed tag push is a warning, not a failure, because the deploy has already happened. The tag push uses the git credentials in the checkout. If the container cannot use them, the next `plan` will not start from this deploy until the tag is pushed some other way.
- Separation of duties needs the rule in `glassops.yaml`. When the person who merges is also the pull request author, the rule fails unless an approval satisfies it. Set `approval.environment: glassops-deploy` so an approval of the environment satisfies the rule. See [PERMIT and rules](../reference/permit-and-rules.md#separation_of_duties).
