# Run a pipeline

`glassops run <pipeline>` runs a whole list of steps from `glassops.yaml` in one command. It runs the same way on a laptop and in any CI system, so a pipeline you run locally is the one your CI runs. The commands it calls are described in the [CLI reference](../reference/cli.md#glassops-run).

---

## Run it locally

From the project root, with `git`, the Salesforce CLI, `sfdx-git-delta` (`sf plugins install sfdx-git-delta`) and the environment variables your targets name:

```bash
glassops run pr --dry-run   # print the steps, run nothing
```

`--dry-run` shows each step, its `when` and the global flags it receives. It writes nothing and removes nothing. A real run writes one result file to `.glassops/results/` per step, plus `run.json` for the whole run. The exit code is the worst of all steps: 0 pass, 1 a check or policy failed, 2 a tool error. Add `.glassops/` to your `.gitignore`: `init` does not, and the directory holds generated plans, PERMITs and results.

**A fresh `init` project does not finish `glassops run pr` on a laptop.** The default `pr` pipeline is written for a pull request in CI, and two things are missing outside one:

- `plan` has no base. Outside a pull request, `--from auto` needs a target or a base branch, so the first step stops with "--from auto needs a target or a pull request base branch. Pass --from." Every later step is skipped except those marked `when: always`.
- `environments.ephemeral` is not configured. `env create` and `env delete` need that entry and a Dev Hub target, so `env delete` fails with "environments.ephemeral is not configured". See [`environments`](../reference/config.md#environments).

To run something that finishes, add a pipeline of your own to `glassops.yaml` that names the base and leaves out what you have not set up. This one plans from the previous commit, runs your project scripts and reports:

```yaml
pipelines:
  local:
    - plan --from HEAD~1
    - check
    - step: report
      when: always
```

```bash
glassops run local
```

Add `validate --target prod` before `report` to run a check-only validate against a real target. That logs in to the target with the credentials in your environment. The CI pipeline is the same steps with a base the CI provides, so a local `local` pipeline is a rehearsal and not the CI run.

## Run it in CI

In any CI, run the same command in the `glassops` container image, or with a `glassops` binary on the runner:

```bash
docker run --rm -v "$PWD:/work" -w /work \
  -e SF_JWT_KEY -e SF_CLIENT_ID -e SF_USERNAME -e SF_DEVHUB_ORG_ID \
  ghcr.io/glassops-platform/glassops:<version> run pr
```

`run` is in v0.4.0 and later. For anything newer, build the image from this repository. See [Install](../reference/cli.md#install).

On GitHub, the reusable workflows do this for you and add the checkout, the SARIF upload and the evidence upload. See [Deploy with GitHub](deploy-with-github.md). For a single command outside a pipeline, the `glassops-platform/glassops` action takes `command: run pr` or any other command line.

---

## The default pipelines

A project with no `glassops.yaml` gets these two pipelines, exactly as `config.DefaultPipelines()` returns them. `glassops init` writes them into the file it creates. A `glassops.yaml` that has no `pipelines` key has none, and one that has the key uses only what it lists.

```yaml
pipelines:
  pr:
    - plan
    - check
    - analyze
    - env create
    - validate --target ephemeral --fallback-target prod
    - step: report
      when: always
    - step: env delete
      when: always
  deploy:
    - plan --target prod
    - validate --target prod
    - gate --target prod
    - deploy --target prod
    - verify --target prod
    - step: report
      when: always
```

**pr.** Plans the change, runs your project scripts (`check`), runs the Code Analyzer (`analyze`), creates a short-lived scratch org named `ephemeral` from the Dev Hub (`env create`), and runs a check-only validate there. `--fallback-target prod` makes `validate` use the `prod` target when no environment exists, for example when `environments.ephemeral` is not configured. `report` and `env delete` run even after a failure, so the run always leaves a summary and never leaves an org behind. `validate` runs the tests at the configured test level. To run the tests with coverage minimums as a separate step, add `test --target ephemeral` after `validate`.

**deploy.** Plans from what `prod` last received (`--target prod` starts from the `glassops/deployed/prod` tag), validates, asks the policy rules for a signed PERMIT, deploys under it, and then runs [`verify`](../reference/cli.md#glassops-verify) to check that the target holds what the plan says. A denied gate exits 1 and the deploy step is skipped. `verify` is newer than the published prereleases. See [what is in which release](../reference/releases-and-licenses.md#what-is-in-which-release). `report` runs either way. The deploy needs `GLASSOPS_PERMIT_KEY`, which should be available only to the protected deploy job. See [Deploy with GitHub](deploy-with-github.md).

Two steps skip themselves when they have nothing to do. `check` does nothing without `check.scripts`, and `analyze` does nothing without `analyze.enabled: true`. Both end with the status `empty` and exit 0.

---

## Write your own

A step is a glassops command line without the `glassops` prefix, or a map. Add `when` to decide what runs after a failure, and an `exec` step to run any shell command.

```yaml
pipelines:
  pr:
    - plan
    - step: exec
      run: npm run build
    - env create
    - validate --target ephemeral
    - test --target ephemeral
    - step: report
      when: always
    - step: env delete
      when: always
```

| `when` | The step runs |
|---|---|
| `success` (the default) | Only when no earlier step failed |
| `failure` | Only when an earlier step failed |
| `always` | Every time, including after a failure or a cancel |

How `run` treats the steps:

- **Empty plan.** After `plan` writes an empty plan, every later step that is not `when: always` is skipped. A pull request that changes no deployable source passes without logging in anywhere. A plan that cannot be read is an error, never an empty plan.
- **Failure.** A step fails when its exit code is not 0. Later `success` steps are skipped. `failure` and `always` steps still run. The run exits with the worst code.
- **Exec steps.** An `exec` step runs `sh -c` in the project directory. It gets `GLASSOPS_RESULTS_DIR`, the absolute path of the results directory, so a script can read or add to the results. Any non-zero exit is a fail. A command that cannot start is an error. The result is `exec.<n>.json`, with `n` counting exec steps from 1.
- **Same flags everywhere.** Every glassops step receives the run's `--dir`, `--config`, `--results-dir` and `--ci`.
- **Stale results.** A real run first removes `run.json` and the glassops result files left in the results directory by an earlier run, and keeps everything else. A `report` step cannot read an old run as evidence for this one.
- **Cancel.** Ctrl+C or SIGTERM stops the current step, skips the rest, and still runs the `when: always` steps. The run exits at least 2.

A created environment is used automatically by later steps. `env create` writes `.glassops/env/ephemeral/env.json` and a login file, and `validate --target ephemeral` finds them. See [`env create`](../reference/cli.md#glassops-env-create).
