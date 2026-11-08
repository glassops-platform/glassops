# Run GlassOps on Bitbucket or GitLab

GlassOps runs the same pipelines on GitHub Actions, Bitbucket Pipelines, GitLab CI and a laptop. The CLI picks the CI from the environment it runs in (GitHub first, then Bitbucket, then GitLab), or from the global flag `--ci github|bitbucket|gitlab|local`.

> **Not proven on a real runner.** The Bitbucket and GitLab layers are written from Atlassian's and GitLab's documentation of their predefined variables, and they are covered by unit tests that set those variables. No pipeline has run on a real Bitbucket or GitLab runner yet. Every statement on this page about a real runner is untested. Treat your first run as the proof.

Working files are in the repository: [`examples/bitbucket`](https://github.com/glassops-platform/glassops/tree/main/examples/bitbucket) and [`examples/gitlab`](https://github.com/glassops-platform/glassops/tree/main/examples/gitlab). Or run `glassops init` in the project. It writes `bitbucket-pipelines.yml` or `.gitlab-ci.yml` when it detects that CI.

---

## Prerequisites

- A Salesforce DX project with a `glassops.yaml`, or run `glassops init` in it.
- A signing key for the PERMIT, stored only where the deploy job can read it. Make one with `openssl rand -base64 48`.
- The login secrets that `glassops.yaml` names with `$VAR`.

## Installation

Both examples run the container image `ghcr.io/glassops-platform/glassops:v1`, which holds the CLI, the Salesforce CLI and `sfdx-git-delta`. See [Releases, images and licenses](../reference/releases-and-licenses.md) for the tags and for how to build the image yourself.

---

## What the CI layer does

A CI layer tells GlassOps who is running what, and gives it somewhere to put outputs and a summary.

| | Bitbucket Pipelines | GitLab CI |
|---|---|---|
| Detected from | `BITBUCKET_BUILD_NUMBER` is set | `GITLAB_CI` is `true` |
| Repository | `BITBUCKET_REPO_FULL_NAME` | `CI_PROJECT_PATH` |
| Commit | `BITBUCKET_COMMIT` | `CI_COMMIT_SHA` |
| Actor | UUID of whoever started the build (`BITBUCKET_STEP_TRIGGERER_UUID`) | `GITLAB_USER_LOGIN` |
| Run Id | `BITBUCKET_BUILD_NUMBER` | `CI_PIPELINE_ID` |
| Pull request | `BITBUCKET_PR_ID` is set | `CI_PIPELINE_SOURCE` is `merge_request_event` |
| Event name | `pull_request` for a pull request, `push` for everything else | `pull_request` for a merge request, `push` when the source is empty, any other source passed through as it is |
| Base branch | `BITBUCKET_PR_DESTINATION_BRANCH` | `CI_MERGE_REQUEST_TARGET_BRANCH_NAME` |

Bitbucket exposes no user name, so the actor is a UUID. GitLab's user email variable is never read.

### Outputs and the summary

Neither CI has a command channel like GitHub's `GITHUB_OUTPUT`, so GlassOps writes files and you publish them as artifacts:

- Step outputs (`status` and `result`) are appended as `NAME=value` lines to `.glassops/outputs.env`. A name must be letters, digits and underscores, and a value must be one line.
- The run summary is appended as markdown to `.glassops/summary.md`.
- Set `GLASSOPS_OUTPUT_FILE` or `GLASSOPS_SUMMARY_FILE` to move either file. A relative path is relative to the project directory.

On GitLab, `.glassops/outputs.env` is a dotenv file, so the job can publish it with `artifacts:reports:dotenv` and later jobs read the values as variables. Annotations are plain log lines, such as `error: ...`.

### What is not read

Approval evidence (the PR author, labels and environment approvals) is not read from Bitbucket or GitLab. The separation of duties rule needs that evidence, so on these CIs it can be `enforce` or `off` per tier but it cannot decide on an approval. See [PERMIT and rules](../reference/permit-and-rules.md).

---

## Usage on Bitbucket

1. Enable Pipelines for the repository.
2. Add the variables `glassops.yaml` names with `$VAR` (for example `SF_JWT_KEY`) as secured repository variables.
3. Create a `production` deployment environment, restrict it to `main`, and add `GLASSOPS_PERMIT_KEY` as a secured variable of that environment only. Make the value with `openssl rand -base64 48`. The pull request pipeline never sees it.
4. Copy `bitbucket-pipelines.yml` to the repository root.

The file clones with `depth: full` so change sets between commits are complete, runs `glassops run pr` on pull requests, and runs `glassops run deploy` on a push to `main` in the `production` deployment. Both save `.glassops/**` as artifacts.

## Usage on GitLab

1. Add the variables `glassops.yaml` names with `$VAR` as masked CI/CD variables.
2. Protect the default branch. Add `GLASSOPS_PERMIT_KEY` as a CI/CD variable that is protected, masked and limited to the `production` environment. The merge request pipeline never sees it.
3. Copy `.gitlab-ci.yml` to the repository root.

The file sets `GIT_DEPTH: "0"` for the full history and `entrypoint: [""]` on the image, because GitLab needs a shell to run script lines.

---

## Moving the deployed tag

After a successful deploy, GlassOps moves the git tag `glassops/deployed/<target>` in the checkout, and tries to push it. The next `plan --target` starts from that tag.

- The Bitbucket example pushes `glassops/deployed/*` itself after `glassops run deploy`.
- On GitLab, a job token usually cannot push, so moving the tag only warns until you give the job a credential that can push. Without the push, the next plan falls back to the merge base with the target's branch.

A failed push is a warning, never a failed deploy, because the deploy has already happened.

---

## What you can check without a runner

You can run the CI layer's mapping on your own machine by setting the variables by hand, for example `CI_PIPELINE_SOURCE=merge_request_event GITLAB_CI=true glassops doctor`, and reading the result file. That checks the mapping, not the runner. Whether a real runner sets the variables as documented, whether the artifacts are collected, and whether the dotenv report reaches later jobs are the parts that stay untested here.
