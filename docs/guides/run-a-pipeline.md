# Run a pipeline

`glassops run <pipeline>` runs a whole list of steps from `glassops.yaml` in one command. It runs the same way on a laptop and in any CI system, so a pipeline you run locally is the one your CI runs. The commands it calls are described in the [CLI reference](../reference/cli.md#glassops-run).

---

## Run it locally

From the project root, with `git`, the Salesforce CLI, `sfdx-git-delta` and the environment variables your targets name:

```bash
glassops run pr --dry-run   # print the steps, run nothing
glassops run pr
```

`--dry-run` shows each step, its `when` and the global flags it receives. It writes nothing and removes nothing. A real run writes one result file per step to `.glassops/results/`, plus `run.json` for the whole run. The exit code is the worst of all steps: 0 pass, 1 a check or policy failed, 2 a tool error.

## Run it in CI

In any CI, run the same command in the `glassops` container image, or with a `glassops` binary on the runner:

```bash
docker run --rm -v "$PWD:/work" -w /work \
  -e SF_JWT_KEY -e SF_CLIENT_ID -e SF_USERNAME -e SF_DEVHUB_ORG_ID \
  ghcr.io/glassops-platform/glassops:<version> run pr
```

`run` is not in v0.3.0. Use the image of the next release, or one you build from this repository with `docker build -t glassops .`.

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
    - step: report
      when: always
```

**pr.** Plans the change, runs your project scripts (`check`), runs the Code Analyzer (`analyze`), creates a short-lived scratch org named `ephemeral` from the Dev Hub (`env create`), and runs a check-only validate there. `--fallback-target prod` makes `validate` use the `prod` target when no environment exists, for example when `environments.ephemeral` is not configured. `report` and `env delete` run even after a failure, so the run always leaves a summary and never leaves an org behind. `validate` runs the tests at the configured test level. To run the tests with coverage minimums as a separate step, add `test --target ephemeral` after `validate`.

**deploy.** Plans from what `prod` last received (`--target prod` starts from the `glassops/deployed/prod` tag), validates, asks the policy rules for a signed PERMIT, and deploys under it. A denied gate exits 1 and the deploy step is skipped. `report` runs either way. The deploy needs `GLASSOPS_PERMIT_KEY`, which should be available only to the protected deploy job. See [Deploy with GitHub](deploy-with-github.md).

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
