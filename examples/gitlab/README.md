# GlassOps on GitLab CI

`.gitlab-ci.yml` runs the pr pipeline on merge requests and the deploy pipeline on a push to the default branch. `glassops.yaml` is the same Salesforce config as `examples/consumer`. Copy both into a Salesforce DX project, or run `glassops init` there.

This layer is written from GitLab's documentation of the predefined variables. It has not been run on a real GitLab runner yet, so treat the first run as a proof.

## Set up

1. Add the variables `glassops.yaml` names with `$VAR` (for example `SF_JWT_KEY`) as masked CI/CD variables.
2. Protect the default branch. Add `GLASSOPS_PERMIT_KEY` as a CI/CD variable that is protected, masked and limited to the `production` environment. Generate the value with `openssl rand -base64 32`. Merge request pipelines never see it.

## How it behaves

- `GIT_DEPTH: "0"` fetches the full history, so change sets between commits are complete.
- The image sets `entrypoint: [""]`, because GitLab needs a shell to run script lines.
- Merge request pipelines report `Event` as `pull_request`, with the merge request IID and the target branch. Other pipeline sources (schedule, web, api) are passed through as they are.
- Outputs are appended to `.glassops/outputs.env`, which the job publishes with `artifacts:reports:dotenv`. The run summary goes to `.glassops/summary.md`. Set `GLASSOPS_OUTPUT_FILE` or `GLASSOPS_SUMMARY_FILE` to move them.
- Approval evidence (labels, environment approvals) is not read from GitLab. On this CI, separation of duties is enforce or off per tier.
- A job token usually cannot push, so moving the `glassops/deployed/*` tag only warns until you provide a push credential.
