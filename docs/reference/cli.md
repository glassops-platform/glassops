# CLI Reference

> **Early work.** The `glassops` CLI is being built in phases and nothing is released yet. This page covers what exists today.

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

Further commands are planned and not built yet.

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

`glassops.yaml` is optional. Its schema is [`schemas/config.json`](https://github.com/glassops-platform/glassops/blob/main/schemas/config.json). Beyond the schema, these rules are checked:

- Org Ids look like Salesforce org Ids, unless written as a `$VAR`.
- No two targets share an org Id.
- At most one target has tier `prod`.
- Every policy window names a time zone that exists.
- Every blackout ends on or after its start date.

Secrets are always named with `$VAR`, so values stay in your CI's secret store.

```yaml
version: 1
provider: salesforce
targets:
  prod:
    id: 00D000000000000
    tier: prod
    branch: main
    auth: { method: jwt, username: $SF_PROD_USERNAME, client_id: $SF_CLIENT_ID, key: $SF_JWT_KEY }
policy:
  windows:
    prod: [{ days: [tue, thu], from: "18:00", to: "22:00", tz: America/Chicago }]
```

---

## End-to-end check

The workflow [`cli-e2e.yml`](https://github.com/glassops-platform/glassops/blob/main/.github/workflows/cli-e2e.yml) creates a scratch org in a Dev Hub, runs `doctor --auth`, `plan` and `validate` against it, and deletes the org. It runs on demand and on pushes to `main`, and it skips with a notice when the Salesforce secrets are not set. The scripts are in [`scripts/e2e`](https://github.com/glassops-platform/glassops/tree/main/scripts/e2e).

---

## Install

There is no release yet. To try it, build from a checkout with Go 1.27.3 or later:

```bash
go build -o glassops ./cmd/glassops
./glassops doctor
```

The container image adds Node 22, the Salesforce CLI and `sfdx-git-delta`. Build it with `docker build -t glassops .` from the repository root.
