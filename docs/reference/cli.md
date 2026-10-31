# CLI Reference

> **Early work.** v0.3.0 is released as a prerelease. It has `init`, `doctor`, `describe`, `plan`, `validate`, `gate`, `permit verify`, `deploy` and `report`. The commands `env`, `test`, `run`, `check` and `analyze` are in the repository and ship in the next release.

`glassops` is one Go binary that runs each delivery step as its own command. The goal is a governed pipeline that runs the same way in any CI system or on a laptop. Salesforce is the first provider.

---

## Commands

| Command | What it does |
|---|---|
| `glassops init` | Detects the provider and the CI system, then writes `glassops.yaml` and a CI file |
| `glassops doctor` | Checks the config, the tools, the git checkout and the environment variables the config names. With `--auth` it also logs in to targets |
| `glassops describe` | Logs in to one target, runs the guard, and records what the target reports about itself |
| `glassops plan` | Works out what changed between two git refs and writes the manifests and a change set |
| `glassops validate` | Runs a check-only validate of the change set against a target |
| `glassops gate` | Evaluates the policy rules for a target and writes a signed PERMIT |
| `glassops permit verify` | Checks that a PERMIT is signed, unexpired and bound to this target, commit, plan and config |
| `glassops deploy` | Deploys the change set to a target, only under a valid PERMIT |
| `glassops report` | Summarises every result for the CI and writes SARIF |
| `glassops env create` | Creates the ephemeral environment (a scratch org) that later steps use |
| `glassops env delete` | Deletes that environment |
| `glassops env sweep` | Deletes old environments that a cancelled run left behind |
| `glassops test` | Runs Apex tests with coverage on a target and writes JUnit XML |
| `glassops run` | Runs a whole pipeline from `glassops.yaml` |
| `glassops check` | Runs your own project scripts |
| `glassops analyze` | Runs the Code Analyzer and records the violation counts |

### `glassops init`

- Detects a Salesforce project from `sfdx-project.json`.
- Detects the CI system from the project files (`.github`, `bitbucket-pipelines.yml`, `.gitlab-ci.yml`, `azure-pipelines.yml`) or from the environment it runs in.
- Writes `glassops.yaml` with commented stub targets and the default `pr` and `deploy` pipelines.
- For GitHub, also writes `.github/workflows/glassops.yml`, which runs `glassops doctor` on full git history. Other CI systems are detected but get no CI file yet.
- Never overwrites a file unless you pass `--force`.

### `glassops doctor`

Runs these checks and fails if any of them fails:

- `glassops.yaml`, when present, against the schema and its rules. Without the file, the defaults are used.
- The provider: the config and the project files must agree.
- The tools: `git` always, plus `sf` and the `sfdx-git-delta` plugin for Salesforce projects. Versions are recorded.
- The analyzer, only when `analyze.enabled` is `true`: Java 11 or newer and the Code Analyzer plugin (`tool.java` and `tool.code-analyzer`). Both are in the `glassops-analyzer` image.
- The checkout: a shallow clone fails, because it would make every change set empty. On GitHub, use `actions/checkout` with `fetch-depth: 0`.
- The environment: every `$VAR` the config names must be set. Only presence is checked. Values are never printed or stored.
- The PERMIT key: on a `pull_request` or `pull_request_target` run, `GLASSOPS_PERMIT_KEY` should not be set. If it is, `doctor` adds a `permit.key` check whose detail starts with `warning:` and a CI warning annotation. This never fails the command. Only presence is checked. See [PERMIT and rules](permit-and-rules.md#the-signing-key).

### `glassops doctor --auth`

Adds a login to each target in the config, or only the one named by `--target`, and runs the [guard](#the-guard) on it. Without `--auth`, `doctor` never logs in. The environment check covers every `$VAR` the config names. With `--auth --target`, it covers only that target's variables, so a target that does not exist yet, such as a scratch org that a later step creates, does not fail the check.

| Flag | Meaning |
|---|---|
| `--auth` | Also log in to targets and run the guard |
| `--target` | With `--auth`, check only this target |

### `glassops describe`

Logs in to one target, runs the guard, and writes what the target reports about itself to `.glassops/results/describe.json`. A guard mismatch exits 2 before anything else runs.

| Flag | Meaning |
|---|---|
| `--target` | Target name from `glassops.yaml` (required) |

### `glassops plan`

Works out the change set between two git refs. It writes the provider's manifests to `.glassops/plan/` and the change set to `.glassops/plan.json`. For Salesforce, `sfdx-git-delta` works out the delta and the project's `.forceignore` applies. A shallow clone is an error, never an empty plan. With `salesforce.delta.enabled: false`, `plan` lists the whole source instead, and it reads no git history.

| Flag | Default | Meaning |
|---|---|---|
| `--from` | `auto` | Base git ref, or `auto` |
| `--to` | `auto` | Head git ref, or `auto` (`HEAD`) |
| `--target` | | Target the plan is for, used by `--from auto` |

With `--from auto`, a pull request uses the merge base with its base branch. With `--target`, `auto` uses the git tag `glassops/deployed/<target>`, then the merge base with the target's branch. Without either, pass `--from`.

First deploy rule: with `--target` and `--from auto`, a target that has no `glassops/deployed/<target>` tag has never received a deploy from glassops. `plan` then plans the whole source instead of a delta, prints a notice, and records the facts `plan.first_deploy` and `plan.from_source: first-deploy`. A shallow clone is still an error.

### `glassops validate`

Logs in to the target, runs the guard, and runs a check-only validate of the change set in `.glassops/plan.json`. It never deploys. See [Validate is check-only](#validate-is-check-only).

| Flag | Default | Meaning |
|---|---|---|
| `--target` | | Target name from `glassops.yaml` (required) |
| `--fallback-target` | | Target to use when `--target` is not configured or not reachable |
| `--plan` | `<dir>/.glassops/plan.json` | Change set file |
| `--check-only` | `true` | Always true. `false` is refused |

A target counts as not configured when it is missing from `glassops.yaml` or when its `id` variable is not set, for example because the scratch org was never created. A guard mismatch never falls back. The test level comes from `salesforce.test_level` for the tier of the target used, then its `default` entry, then the Salesforce CLI's own default. An empty change set ends with status `empty` and no login. `validate` records the facts `validate.plan_sha256` and `validate.test_level`, which `gate` uses to decide whether a deploy may reuse the validation. See [validation binding](permit-and-rules.md#validation-binding-and-quick-deploy).

### `glassops gate`

Reads the change set, the earlier results in the results directory and the CI context. It builds the facts for one target, runs the policy rules over them, and writes a signed PERMIT to `.glassops/permit.json`. It writes the PERMIT whether the answer is allow or deny, and it deletes any PERMIT already at that path first, so an old allow never survives a failed run. Unless `--no-connect` is set, it logs in to the target and runs the [guard](#the-guard). The rules, the PERMIT and the evidence `gate` reads are described in [PERMIT and rules](permit-and-rules.md).

| Flag | Default | Meaning |
|---|---|---|
| `--target` | | Target name from `glassops.yaml` (required) |
| `--plan` | `<dir>/.glassops/plan.json` | Change set file |
| `--permit` | `<dir>/.glassops/permit.json` | Where the PERMIT is written |
| `--no-connect` | `false` | Do not log in. The PERMIT records the configured Id and the kind the guard requires for the tier, and the `guard` check is `skipped` |

Needs `GLASSOPS_PERMIT_KEY`. Reads `.glassops/plan.json`, `.glassops/results/plan.json` and `.glassops/results/validate.json` when they exist, and the CI context. Writes `.glassops/permit.json`, `.glassops/facts.json` and `.glassops/results/gate.json`.

| Exit | When |
|---|---|
| `0` | The PERMIT allows the deploy |
| `1` | Policy denies the deploy. The signed PERMIT is still written |
| `2` | `GLASSOPS_PERMIT_KEY` is missing or unusable, the plan file is missing or unreadable, the target is not in the config, or evidence could not be gathered. A missing key, plan or target writes no PERMIT. Evidence that could not be gathered writes a signed PERMIT that denies |

Result facts: `gate.allow`, `gate.ephemeral`, `gate.permit_id` and `gate.expires_at`.

### `glassops permit verify`

Checks a PERMIT without logging in to the target, so any deploy tool can honour it. It checks the key, the signature, that the PERMIT allows, that it has not expired, and that it matches the target, the commit, the plan and the config. The order is in [PERMIT and rules](permit-and-rules.md#what-permit-verify-and-deploy-check).

| Flag | Default | Meaning |
|---|---|---|
| `--target` | | Target name from `glassops.yaml` (required) |
| `--permit` | `<dir>/.glassops/permit.json` | PERMIT file |
| `--plan` | `<dir>/.glassops/plan.json` | Change set file |

Needs `GLASSOPS_PERMIT_KEY`. Reads the PERMIT, the change set, the plan files under `.glassops/plan/`, `glassops.yaml` and the CI commit. Writes `.glassops/results/permit-verify.json`.

| Exit | When |
|---|---|
| `0` | The PERMIT holds |
| `1` | The PERMIT does not hold: bad signature, denied, expired, a different target, commit, plan or config, a file that is not a readable PERMIT, or plan files changed after `plan` |
| `2` | The key is missing or unusable, or the PERMIT file, the change set, the config or the commit cannot be read or found |

Result facts: `permit.key_id`, `permit.id` and `permit.expires_at`.

### `glassops deploy`

Verifies the PERMIT before it logs in. Then it runs the [guard](#the-guard) and deploys the change set in `.glassops/plan.json`. When the PERMIT holds a validated deploy Id, it quick deploys that validation. See [validation binding](permit-and-rules.md#validation-binding-and-quick-deploy). Success is read from the target's result body. After a successful deploy it moves the git tag `glassops/deployed/<target>` to the deployed commit and pushes it. The tag is lightweight. A failed move or push is a warning, because the deploy has already happened: the `tag` check is `skipped`, `deploy.tag_pushed` is `false`, and the CI gets a warning annotation.

| Flag | Default | Meaning |
|---|---|---|
| `--target` | | Target name from `glassops.yaml` (required) |
| `--plan` | `<dir>/.glassops/plan.json` | Change set file |
| `--permit` | `<dir>/.glassops/permit.json` | PERMIT file |
| `--wait` | the provider's | How long to wait for the deploy, such as `30m`. The Salesforce provider waits 60 minutes |
| `--resume` | | Job Id of an earlier deploy to wait on again. Skips quick deploy |

Needs `GLASSOPS_PERMIT_KEY`. The deployed commit is the CI's commit, or `git rev-parse HEAD` when the CI reports none. It must be the commit the PERMIT names. An empty change set under a valid PERMIT ends with status `empty` and no login.

| Exit | When |
|---|---|
| `0` | The deploy succeeded, or the change set was empty |
| `1` | The deploy ran and the target reported failures, or `salesforce.deploy.quick` is `require` and the PERMIT holds no passed validation to quick deploy |
| `2` | The PERMIT is missing, unreadable, denied, expired or does not match, the plan files changed, login or the guard failed, or the deploy did not finish within `--wait`. A timeout keeps the job Id and prints the `--resume` command, and does not move the tag |

Result facts: `deploy.permit_id`, `deploy.quick_mode`, `deploy.id`, `deploy.quick`, `deploy.timed_out`, `deploy.commit`, `deploy.tag` and `deploy.tag_pushed`, plus the provider's own facts.

### `glassops report`

Reads every result in the results directory. It writes a markdown summary to the CI's run summary, and with `--summary` also to a file. It writes SARIF to `--sarif` for policy denials found in the PERMIT and for failed checks in the `validate` and `deploy` results. SARIF files already in the results directory, such as Code Analyzer's, are merged into the output. The SARIF file it writes is skipped when it reads the directory again. `report` writes no result file of its own.

| Flag | Default | Meaning |
|---|---|---|
| `--sarif` | `<dir>/.glassops/report.sarif` | SARIF file to write |
| `--summary` | | Also write the markdown summary to this file |

`report` exits with the worst status among the results: `0` when all pass, `1` when any failed, `2` when any errored. That lets a CI run it after a failed step with `if: always()` and still fail the job. It also exits `2` when there are no results or one cannot be read.

### `glassops env create`

Creates the ephemeral environment named under `environments` in `glassops.yaml`, from its Dev Hub target, for later steps to use. It logs in to the hub and runs the [guard](#the-guard) on it. The default pr pipeline runs it before `validate`.

For Salesforce it does not run `sf org create scratch`. It creates a `ScratchOrgInfo` record in the Dev Hub (a signup), polls until the record is `Active` for up to 15 minutes, and exchanges the record's auth code for a login. A signup refused for capacity or a transient error is retried up to three times. A signup that is still not active after 15 minutes is not retried, because it could still become active. The scratch definition file (`environments.<target>.definition`) supports the keys `edition`, `features`, `hasSampleData`, `language`, `country`, `description`, `release` and `sourceOrg`. Any other key, including `settings`, is an error, so an org is never created without something the file asks for. `orgName` is replaced by the environment name.

| Flag | Default | Meaning |
|---|---|---|
| `--target` | `ephemeral` | Environment name from the `environments` section |
| `--name` | `<prefix>-<repo>-<pr or branch>-<run>` | Environment name. The prefix is `environments.<target>.prefix`, default `glassops`. Outside CI the repo is the project directory and the run is a UTC timestamp |
| `--ttl-days` | `environments.<target>.ttl_days`, else `1` | Days the environment lives |
| `--shape` | `environments.<target>.shape` | Source org Id for an org shape. A `$VAR` is read from the environment |
| `--env-file` | `<dir>/.glassops/env/<target>/env.json` | Where to write `env.json` |

**How later steps use the org.** `env create` writes `env.json` and a login file. `env.json` records the environment name, Id, expiry and the path of the login file, and holds no secret. The login file holds an `sfdx-url` login, is readable only by the current user, and lives in the operating system's temporary directory. When each command runs in its own container, set `TMPDIR` to a mount they share. A later command that resolves a target named under `environments`, such as `validate --target ephemeral`, finds `env.json` and logs in with that file instead of the `$VAR` settings. With no `env.json`, the target falls back to its `targets` entry, and `--fallback-target` applies if that is not configured. An `env.json` that records a failed create (`ready: false`) counts as not configured. On GitHub, `env create` also sets the step outputs `env_id`, `env_file` and `auth_file`.

`env create` refuses to run when `env.json` already exists, so run `env delete` first. If the create fails after the signup was requested, `env.json` is still written so `env delete` can remove the org.

Writes `.glassops/results/env-create.json` with the facts `env.file`, `env.id`, `env.record_id`, `env.name`, `env.auth_file` and `env.expires_at`, plus the provider's own.

| Exit | When |
|---|---|
| `0` | The environment is active |
| `2` | `environments.<target>` is not configured, `env.json` already exists, the hub login or guard failed, or the signup failed or timed out |

### `glassops env delete`

Deletes the environment from its hub. It reads `env.json`, or finds the environment by `--name`. With neither, nothing is recorded to delete: the `delete` check is `skipped` and the command exits 0. Deleting an environment that is already gone passes. After a delete from `env.json`, it removes the login file and `env.json`, so later steps fall back to the config. The default pr pipeline runs it with `when: always`.

| Flag | Default | Meaning |
|---|---|---|
| `--target` | `ephemeral` | Environment name from the `environments` section |
| `--name` | the one `env create` made | Delete by name instead of from `env.json` |
| `--env-file` | `<dir>/.glassops/env/<target>/env.json` | `env.json` to read |

Writes `.glassops/results/env-delete.json`. Exit `0` when the environment is deleted, already gone, or nothing is recorded. Exit `2` when the config, the hub login or `env.json` (for another target or hub) is wrong, or the delete fails.

### `glassops env sweep`

Deletes every environment on the hub whose name starts with `<prefix>-` and that is older than `--max-age`, and lists what it deleted. It never matches the bare prefix. Run it on a schedule to remove environments that a cancelled run left behind.

| Flag | Default | Meaning |
|---|---|---|
| `--target` | `ephemeral` | Environment name from the `environments` section, which supplies the hub and the prefix |
| `--max-age` | `environments.<target>.ttl_days` as a duration | Delete environments older than this, such as `24h`. Must be above zero |

Writes `.glassops/results/env-sweep.json` with the facts `env.prefix`, `env.max_age`, `env.deleted` and `env.deleted_count`. Exit `0` when the sweep finished, `2` when the config, the hub login or a delete failed.

### `glassops test`

Logs in to the target, runs the guard, and runs the target's Apex tests with code coverage. It never deploys. It writes JUnit XML and checks the result against `test.coverage`.

| Flag | Default | Meaning |
|---|---|---|
| `--target` | | Target name from `glassops.yaml` (required) |
| `--level` | `test.level` for the target's tier, then `test.level.default` | Test level, such as `RunLocalTests` |
| `--tests` | | Comma separated tests to run, for a level that names them |
| `--wait` | the provider's | How long to wait for the run, such as `30m` |
| `--resume` | | Run Id of an earlier run to wait on again, instead of starting a new one |
| `--junit` | `<results-dir>/test.junit.xml` | JUnit XML file to write |

**Coverage.** Percentages are rounded down, so a rounded-up value never passes a minimum. Both settings are optional:

- `test.coverage.org_min` compares against the org-wide coverage that `sf` reports. That figure counts every class in the org, so it catches classes that no test touches.
- `test.coverage.class_min` compares each class that the test run reports. `test` sees per-class coverage only for the classes in that report, so a class that no test ran against can be missing from it. Set `org_min` as well to catch those.

If a minimum is set and its figure is not reported, that is a failure, never a pass. A run that does not finish within `--wait` fails and names the run Id, and `--resume` waits on it again.

**No tests found.** When `sf` reports that no Apex tests were found, `test` records the fact `test.none_found`. It fails when `org_min` or `class_min` is set, because coverage cannot be checked. When neither is set, the status is `empty`, the `tests` check is `skipped`, and it exits 0.

Writes `.glassops/results/test.json` and the JUnit file (listed as an artifact). Facts: `test.ran`, `test.failed`, `test.finished`, `test.run_id`, `test.org_coverage` (a whole percent, absent when unknown), `test.coverage.<Class>.percent` and `target.source`.

| Exit | When |
|---|---|
| `0` | The tests passed and every set minimum is met, or no tests were found and no minimum is set |
| `1` | A test failed, the run did not finish, coverage is below a minimum or was not reported, or no tests ran while a minimum is set |
| `2` | The target could not be resolved, the login or guard failed, or the tests could not be run |

### `glassops run`

Runs the steps of one pipeline from `glassops.yaml`, in order, in one process. See [Run a pipeline](../guides/run-a-pipeline.md) for the default pipelines.

```bash
glassops run pr
glassops run deploy --dry-run
```

| Flag | Default | Meaning |
|---|---|---|
| `--dry-run` | `false` | Print the steps and the global flags each one gets. Runs nothing and removes nothing |

A step is a glassops command line such as `validate --target prod`, or a map with `step`, `when` and, for an `exec` step, `run`.

**When a step runs.** `when` decides what happens after an earlier step failed.

| `when` | The step runs |
|---|---|
| `success` (the default) | Only when no earlier step failed |
| `failure` | Only when an earlier step failed |
| `always` | Every time, including after a failure or a cancel |

A step has failed when its exit code is not 0. After a `plan` step that exits 0, `run` reads `.glassops/plan.json`. An empty plan skips every later step that is not `when: always`, with the reason "the plan is empty". A plan that cannot be read is an error (exit 2), and later `success` steps are skipped, so it is never read as an empty plan. A cancel (Ctrl+C or SIGTERM) stops the current step and skips the rest except `when: always` steps, which still run with up to 2 minutes each. A cancelled run exits at least 2.

**Exec steps.** A step named `exec` runs its `run` command with `sh -c` in the project directory. The command inherits the environment and gets `GLASSOPS_RESULTS_DIR`, the absolute path of the results directory. Its output goes straight to the run's output. A zero exit passes and any other exit is a fail (exit 1). A command that cannot start, is stopped by a signal or is cancelled is an error (exit 2). Each exec step writes `exec.<n>.json`, where `n` counts exec steps from 1, with one `command` check and the facts `exec.run`, `exec.index` and `exec.exit_code`.

**Global flags.** Every glassops step receives this run's `--dir`, `--config`, `--results-dir` and `--ci`, so all steps read one project and write to one results directory.

**Stale results.** Before the first step, a real run removes `run.json` and every top-level `*.json` file in the results directory whose `schema` is `glassops.result/v1`. It does not recurse and keeps every other file, such as SARIF and JUnit files. This stops an earlier run from reading as evidence for this one. The count is the fact `run.stale_removed`. A removal failure exits 2 before any step runs.

**Exit code.** The worst exit code of all steps. Any code other than 0 and 1 ranks as 2.

Writes `.glassops/results/run.json` with the facts `run.pipeline`, `run.steps` (index, step, `when`, argv, result, exit code, duration, skip reason), `run.exit_code`, `run.empty_plan`, `run.cancelled` and `run.stale_removed`, and one check per step. A pipeline that is not in the config, or a step that does not parse, exits 2 before anything runs.

### `glassops check`

Runs your own project scripts, such as lint or unit test commands, and records the outcome. It does not replace those tools.

Scripts come from `check.scripts` in `glassops.yaml`, a list of `name` and `run` entries. Each runs with `sh -c` in the project directory, in order, and every script runs even after an earlier one failed. A script's output goes to the log and never into the result.

| Flag | Default | Meaning |
|---|---|---|
| `--script` | | Run only the script with this name. An unknown name is an error |

Writes `.glassops/results/check.json` with one check per script named `script.<name>` and the facts `check.scripts`, `check.passed` and `check.failed`. With no scripts configured the status is `empty` and the exit code is 0.

| Exit | When |
|---|---|
| `0` | Every script exited 0, or none are configured |
| `1` | A script exited non-zero |
| `2` | `sh` could not start, or `--script` names no entry |

### `glassops analyze`

Runs the Salesforce Code Analyzer over the project and records how many violations it found. It runs only when `analyze.enabled: true` is set in `glassops.yaml`. Without that, it records a `skipped` check, the status `empty`, and exits 0.

The analyzer needs Java 11 or newer and the Code Analyzer plugin of the Salesforce CLI. The standard `glassops` image has neither, so run `analyze` in the separate `glassops-analyzer` image, which adds them. `glassops doctor` checks for Java and the plugin when `analyze.enabled` is true. Only the Salesforce provider is supported.

| Flag | Default | Meaning |
|---|---|---|
| `--rule-selector` | `analyze.rule_selector`, else `Recommended` | Rules to run |
| `--sarif` | `<results-dir>/analyze.sarif` | SARIF file to write |

`analyze` does not decide what is too severe. It exits 0 whatever it found, so long as the analyzer ran. The `gate` rule `analyzer_max_severity` reads the counts and decides. It writes `.glassops/results/analyze.json` with the facts `analyze.rule_selector`, `analyze.total` and `analyze.sev1` to `analyze.sev5`. The SARIF file and the analyzer's own JSON (`.glassops/analyze/code-analyzer.json`, beside the results directory) are listed as artifacts, and `report` merges the SARIF file into its output.

| Exit | When |
|---|---|
| `0` | The analyzer ran, or `analyze` is off |
| `2` | The analyzer could not run, its output could not be read, or the provider is not Salesforce |

### The smart seed

A destructive change, such as a deleted field, cannot be validated on a fresh scratch org, because the component it deletes was never there. When `validate` runs against a scratch org and the change set has destructive members, it first deploys the base state they need. The base ref is the plan's `from` ref. `salesforce.ephemeral.seed` chooses how:

| Mode | What it deploys |
|---|---|
| `minimal` (the default) | Only the destructive members, from the base ref, plus the parent object of any child component (field, record type, validation rule and similar) whose parent is a custom object (`__c`, `__mdt` or `__e`) |
| `full` | The whole source at the base ref |
| `off` | Nothing. The destructive change is validated against the org as it is |

If a minimal seed fails because something else is missing (an error such as `Entity 'Widget__c' not found`), the seed falls back to the full seed, records `validate.seed.fallback: true` and raises a CI warning. A failed seed stops `validate` before it runs. The seed applies to scratch targets only. `validate` records `validate.seed.mode`, `validate.seed.fallback` and `validate.seed.members`.

---

## Authentication

Each target names its login under `auth`. There are three methods, and every secret value is a `$VAR` that names an environment variable. Values stay in your CI's secret store and never go in the file.

| Method | Fields | Notes |
|---|---|---|
| `jwt` | `username`, `client_id`, `key`, optional `instance_url` | `key` is a `$VAR` that holds the private key as PEM text. A value that is not PEM text is read as the path of a key file |
| `sfdx-url` | `url` | `url` is a `$VAR` that holds a `force://` auth URL |
| `access-token` | `token`, `instance_url` | `token` is a `$VAR` that holds an access token |

```yaml
targets:
  devhub:
    id: $SF_DEVHUB_ORG_ID
    tier: devhub
    auth: { method: jwt, username: $SF_USERNAME, client_id: $SF_CLIENT_ID, key: $SF_JWT_KEY }
  ephemeral:
    id: $SF_EPHEMERAL_ORG_ID
    tier: ephemeral
    auth: { method: sfdx-url, url: $SF_EPHEMERAL_AUTH_URL }
```

### Isolated Salesforce CLI state

Every command that talks to an org runs the Salesforce CLI with its own temporary home directory (`HOME` and `USERPROFILE`). It never reads or writes the Salesforce CLI logins you have on your machine. The temporary directory, and any key file written into it, is removed when the command ends, including when it is stopped with Ctrl+C or SIGTERM. Logins are retried only on network and server errors, never on a bad key, an unknown user or an expired token.

### The guard

Before any step does anything to a target, the guard compares the org the login reached with the target in the config. A mismatch exits 2.

- The Id is compared on its first 15 characters, case-sensitively, so a 15 and an 18 character form of one Id match.
- A target with no Id configured fails, because it cannot be checked.
- Tier `prod` and tier `devhub` need a production org.
- Tier `ephemeral` needs a scratch org.
- Every other tier checks the Id only.

### What is never logged

Results, step outputs, CI annotations and error messages never contain the org's company name, tokens, private keys, auth URLs or full usernames. Usernames are masked to their first character and domain, such as `u***@example.com`. Email-shaped text in a tool's error message is masked the same way.

### Validate is check-only

`validate` is check-only by construction. It only ever runs `sf project deploy validate`, never a deploy, and `--check-only=false` is refused. After the run it reads the full result with `sf project deploy report` and exits 2 if the target reports anything other than a check-only run. Keep a Dev Hub for managing scratch orgs, and validate against a scratch org or a sandbox.

---

## Global flags

| Flag | Default | Meaning |
|---|---|---|
| `--dir` | `.` | Project root |
| `--config` | `<dir>/glassops.yaml` | Config file |
| `--results-dir` | `<dir>/.glassops/results` | Where results are written |
| `--ci` | `auto` | CI system: `auto`, `github` or `local` |

---

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Pass |
| `1` | A policy or check failed. The tool itself worked |
| `2` | A tool or infrastructure error stopped the command, or the command line was invalid |

---

## Results

Every command writes `.glassops/results/<step>.json` in one envelope, described by [`schemas/result.json`](https://github.com/glassops-platform/glassops/blob/main/schemas/result.json):

```json
{
  "schema": "glassops.result/v1",
  "step": "doctor",
  "status": "pass",
  "started": "2026-10-26T12:00:00Z",
  "ended": "2026-10-26T12:00:02Z",
  "provider": "salesforce",
  "checks": [{ "name": "tool.sf", "status": "pass", "detail": "2.155.7" }],
  "facts": { "tools.sf": "2.155.7" },
  "artifacts": [],
  "errors": []
}
```

`status` is one of `pass`, `fail`, `error`, `skipped` or `empty`.

On GitHub, each command also sets the step outputs `status` and `result` (the result file path), adds an error annotation for each failed check, and writes a table to the job summary.

---

## Configuration

`glassops.yaml` is optional. Every key, with an example, is in the [configuration reference](config.md). The schema is [`schemas/config.json`](https://github.com/glassops-platform/glassops/blob/main/schemas/config.json).

---

## End-to-end check

The workflow [`cli-e2e.yml`](https://github.com/glassops-platform/glassops/blob/main/.github/workflows/cli-e2e.yml) creates a scratch org in a Dev Hub, runs `doctor --auth`, `plan` and `validate` against it, and deletes the org. It runs on demand and on pushes to `main`, and it skips with a notice when the Salesforce secrets are not set. The scripts are in [`scripts/e2e`](https://github.com/glassops-platform/glassops/tree/main/scripts/e2e).

---

## Install

v0.3.0 is released as a prerelease, with a container image at `ghcr.io/glassops-platform/glassops:0.3.0`. The image adds Node 22, the Salesforce CLI and `sfdx-git-delta`. The commands `env`, `test`, `run`, `check` and `analyze` are not in v0.3.0. To use them now, build from a checkout with Go 1.27.3 or later:

```bash
go build -o glassops ./cmd/glassops
./glassops doctor
```

Or build the container image with `docker build -t glassops .` from the repository root. The `analyze` command needs the separate `glassops-analyzer` image, which adds Java and the Code Analyzer plugin. Build it with `docker build -f Dockerfile.analyzer -t glassops-analyzer .`.
