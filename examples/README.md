# GlassOps examples

`consumer/` is the smallest repository layout that runs GlassOps on GitHub. Copy it into a
Salesforce DX project, or run `glassops init` to generate a similar `glassops.yaml` and workflow.

- [`glassops.yaml`](consumer/glassops.yaml): two targets, a policy for production, and the default pipelines.
- [`.github/workflows/pr.yml`](consumer/.github/workflows/pr.yml): runs the pr pipeline on pull requests by calling the reusable workflow.
- [`.github/workflows/deploy.yml`](consumer/.github/workflows/deploy.yml): runs the deploy pipeline on a push to main by calling the reusable workflow. The job runs in the `glassops-deploy` environment.

## Set up the environment

1. Create the `glassops-deploy` environment in the repository settings.
2. Add the required reviewers, and a deployment branch rule that allows `main` only.
3. Add `GLASSOPS_PERMIT_KEY` as a secret of that environment, not of the repository. Generate the value with `openssl rand -base64 32`.
4. Add the Salesforce secrets `SF_JWT_KEY`, `SF_CLIENT_ID`, `SF_USERNAME`, `SF_DEVHUB_ORG_ID`, `SF_PROD_USERNAME` and `SF_PROD_ORG_ID` as repository secrets.

The pr workflow never receives the permit key. Only a job in the protected environment can read it.
The [GitHub deploy guide](../docs/guides/deploy-with-github.md) explains each part of the deploy job.

## Other commands in a workflow

The `glassops-platform/glassops` action runs one command in the container, for steps that do not fit a pipeline:

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
