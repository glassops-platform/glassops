# The exec provider

The exec provider lets any deploy tool run under GlassOps. GlassOps keeps the gate, the signed PERMIT, the org Id guard and the evidence. Your own commands do the validate, the deploy and the verify. Use it when the target is not Salesforce, or when you deploy with a tool that GlassOps has no built-in provider for.

The provider is never detected from the project files. You choose it in `glassops.yaml`:

```yaml
version: 1
provider: exec
exec:
  paths: [terraform, config]
  describe: ./ci/describe.sh
  validate: ./ci/validate.sh
  deploy: ./ci/deploy.sh
  verify: ./ci/verify.sh
targets:
  prod:
    id: prod-eu-1
    tier: prod
    branch: main
```

`glassops doctor` checks that `sh` and `git` exist. It does not check your commands.

## The exec section

Every command is a shell command that runs with `sh -c` in the project directory. Only `paths` is required.

| Key | What it is |
|---|---|
| `exec.paths` | Project paths whose changed files make up the plan. Required, at least one. Each is relative to the project root. An absolute path, a path that climbs out with `..` and a path that starts with `:` are refused |
| `exec.plan` | Optional. Runs after GlassOps writes its own plan files and may add more files to `$GLASSOPS_PLAN_DIR`. A non-zero exit fails `plan` |
| `exec.describe` | Prints `{"id": "...", "kind": "..."}` for the target. See [The guard](#the-guard) |
| `exec.validate` | The check-only run. Required to run `validate` |
| `exec.deploy` | The deploy. Required to run `deploy`, unless you pass a command after `--` |
| `exec.verify` | The post-deploy check. Required to run `verify` |

A step whose command is not set is an error, never a pass. A step that ran nothing must not look like one that succeeded.

## What `plan` does

`plan` lists the files under `exec.paths` that changed between the two refs, using `git diff` for a delta and `git ls-tree` for a full plan. It writes two files into `.glassops/plan/`:

- `changes.txt` lists the files added or modified. It is the manifest.
- `deleted.txt` lists the files deleted. It exists only when something was deleted, and it is the destructive manifest.

Each file has one path per line, relative to the project root, with forward slashes, sorted. A file name that holds a newline cannot be listed and is refused. The plan digest covers everything in `.glassops/plan/`, including files `exec.plan` adds. The PERMIT binds that digest and the commit, so it binds both the file list and the file content. A plan with no changed files and no extra files is `empty`.

## What every command gets

Each command runs with your process environment plus these variables:

| Variable | Value |
|---|---|
| `GLASSOPS_RESULTS_DIR` | Absolute path of the results directory |
| `GLASSOPS_TARGET` | Target name, when the step has a target |
| `GLASSOPS_TARGET_ID` | Target Id, when the step has a target |
| `GLASSOPS_PLAN_DIR` | Absolute path of the plan directory |
| `GLASSOPS_PLAN` | Absolute path of `.glassops/plan.json`, when it exists |
| `GLASSOPS_PERMIT` | Absolute path of `.glassops/permit.json`, when it exists |
| `GLASSOPS_FROM`, `GLASSOPS_TO` | The plan's git refs. `exec.plan` only |
| `GLASSOPS_DEPLOY_ID` | The deploy Id from the deploy result. `exec.verify` only |

`GLASSOPS_PERMIT_KEY` is removed from every command's environment. A deploy tool that could read the signing key could sign its own PERMIT. GlassOps verifies the PERMIT before it runs your deploy command, so the command does not need the key.

Your own credentials stay in the environment, because the exec provider logs in to nothing itself. Put them in the CI's secret store and pass them to the job.

## The result file

A command may write `<results-dir>/exec-<step>.json` to say more than its exit code can. The step is `plan`, `validate`, `deploy` or `verify`. The shape is [`schemas/engine-result.json`](https://github.com/glassops-platform/glassops/blob/main/schemas/engine-result.json):

```json
{
  "schema": "glassops.engine-result/v1",
  "step": "deploy",
  "passed": true,
  "id": "release-2026-11-08",
  "components_deployed": 12,
  "problems": [],
  "facts": { "exec.region": "eu-1" }
}
```

| Field | Rule |
|---|---|
| `schema` | Must be `glassops.engine-result/v1` |
| `step` | Must match the step that ran |
| `passed` | Required |
| `id` | Optional. The run's Id, such as a deploy Id. `deploy` records it as `deploy.id` |
| `components_deployed` | Optional. A count of zero or more |
| `problems` | Optional. One line per failure |
| `facts` | Optional. Every key must start with `exec.` |

The file may hold no other field and no second JSON value. GlassOps deletes an earlier result file before each command, so an old file can never stand in for the current run.

How the outcome is decided:

- With no result file, the exit code decides. Zero passes.
- A pass needs exit code zero and `passed: true`.
- Exit zero with `passed: false` fails.
- A non-zero exit with `passed: true` fails, and the problem says so.
- A result file that does not match the schema is an error, not a failure.

## The guard

The guard compares the org Id the target reports with the Id in the config. The exec provider has no login, so `exec.describe` is how a target reports itself. It prints one JSON object on stdout:

```json
{"id": "prod-eu-1", "kind": "production"}
```

`kind` is one of `production`, `sandbox`, `scratch` or `unknown`. The Id is compared exactly. Tier `prod` and `devhub` need `kind: production`, and tier `ephemeral` needs `scratch`.

- A target of tier `prod` must have `exec.describe`. Without it every command that reaches the guard stops with an error, because nothing may touch production without the check.
- Any other tier may leave `exec.describe` out. The guard is then recorded as `skipped` with the reason, and the configured Id is used as it is.

## Deploying

`glassops deploy --target prod` verifies the PERMIT, runs the guard, then runs `exec.deploy`. The exec provider has no quick deploy and cannot resume. `--resume` is refused. Run the deploy again instead.

You can name the command on the command line and leave `exec.deploy` out, but the PERMIT must bind it. Give `gate` the same command as a JSON array of strings, one element per argument:

```bash
glassops gate --target prod --deploy-command '["./ci/deploy.sh","--region","eu-1"]'
glassops deploy --target prod -- ./ci/deploy.sh --region eu-1
```

`gate` stores only the SHA-256 of the array in the PERMIT, as `deploy_command_sha256`, and never the text. Everything after `--` replaces `exec.deploy` for that run, and each argument is passed as one word. `deploy` refuses any other command, and refuses a PERMIT with no command hash, with exit 2 and before it runs anything. `["a b"]` and `["a","b"]` hash differently. The result records `exec.deploy_override: true`. A command after `--` with any other provider is an error, and so is `--deploy-command`. Without a command after `--`, `exec.deploy` runs and the PERMIT binds the config file, which holds that command.

## Verifying

`glassops verify --target prod` runs `exec.verify` with `GLASSOPS_DEPLOY_ID` set. The check `verify.exec` passes or fails on the result above, and each problem line becomes a failing check named `verify.exec.problem.N`. The default `deploy` pipeline includes `verify`, so a project with provider exec must set `exec.verify` or remove that step from its pipeline. See [`glassops verify`](cli.md#glassops-verify).

## Untested here

The provider is covered by unit tests with a fake shell and by the script `scripts/e2e/exec-e2e.sh`, which runs the binary in a container with no network. The script takes the binary's name or path as its one argument and defaults to `glassops`. It runs `plan`, `gate --deploy-command`, `permit verify`, a tampered PERMIT, a wrong describe Id, a command the PERMIT does not bind, a modified tracked file, and the deploy. No exec project has deployed to a production system through it yet.
