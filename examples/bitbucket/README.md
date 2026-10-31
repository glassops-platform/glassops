# GlassOps on Bitbucket Pipelines

`bitbucket-pipelines.yml` runs the pr pipeline on pull requests and the deploy pipeline on a push to `main`. `glassops.yaml` is the same Salesforce config as `examples/consumer`. Copy both into a Salesforce DX project, or run `glassops init` there.

This layer is written from Atlassian's documentation of the predefined variables. It has not been run on a real Bitbucket runner yet, so treat the first run as a proof.

## Set up

1. Enable Pipelines for the repository.
2. Add the variables `glassops.yaml` names with `$VAR` (for example `SF_JWT_KEY`) as secured repository variables.
3. Create a `production` deployment environment, restrict it to `main`, and add `GLASSOPS_PERMIT_KEY` as a secured variable of that environment only. Generate the value with `openssl rand -base64 32`. The pr pipeline never sees it.

## How it behaves

- The clone is full depth, so change sets between commits are complete.
- Pull request pipelines report `Event` as `pull_request`, with the pull request number and the destination branch.
- Bitbucket exposes no user name, so the actor is the UUID of whoever started the build.
- Outputs are appended to `.glassops/outputs.env` and the run summary to `.glassops/summary.md`. Both are collected as artifacts. Set `GLASSOPS_OUTPUT_FILE` or `GLASSOPS_SUMMARY_FILE` to move them.
- Approval evidence (labels, environment approvals) is not read from Bitbucket. On this CI, separation of duties is enforce or off per tier.
