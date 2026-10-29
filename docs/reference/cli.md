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

Other commands, such as `gate`, `deploy` and `run`, are planned and not built yet.

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

### `glassops doctor --auth`

Adds a login to each target in the config, or only the one named by `--target`, and runs the [guard](#the-guard) on it. Without `--auth`, `doctor` never logs in. The environment check covers every `$VAR` the config names, so set them all before you run `doctor`.

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

Works out the change set between two git refs. It writes the provider's manifests to `.glassops/plan/` and the change set to `.glassops/plan.json`. A shallow clone is an error, never an empty plan.

| Flag | Default | Meaning |
|---|---|---|
| `--from` | `auto` | Base git ref, or `auto` |
| `--to` | `auto` | Head git ref, or `auto` (`HEAD`) |
| `--target` | | Target the plan is for, used by `--from auto` |

With `--from auto`, a pull request uses the merge base with its base branch. With `--target`, `auto` uses the git tag `glassops/deployed/<target>`, then the merge base with the target's branch.

### `glassops validate`

Logs in to the target, runs the guard, and runs a check-only validate of the change set in `.glassops/plan.json`. It never deploys. See [Validate is check-only](#validate-is-check-only).

| Flag | Default | Meaning |
|---|---|---|
| `--target` | | Target name from `glassops.yaml` (required) |
| `--fallback-target` | | Target to use when `--target` is not configured or not reachable |
| `--plan` | `<dir>/.glassops/plan.json` | Change set file |
| `--check-only` | `true` | Always true. `false` is refused |

---

## Authentication

Each target names its login under `auth`. There are three methods, and every secret value is a `$VAR` that names an environment variable. Values stay in your CI's secret store and never go in the file.

| Method | Fields | Notes |
|---|---|---|
| `jwt` | `username`, `client_id`, `key`, optional `instance_url` | `key` is a `$VAR` that holds the private key as PEM text, not a file path |
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

Every command that talks to an org runs the Salesforce CLI with its own temporary home directory (`HOME` and `USERPROFILE`). It never reads or writes the Salesforce CLI logins you have on your machine. The temporary directory, and any key file written into it, is removed when the command ends.

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

`validate` is check-only by construction. It always passes the check-only option to the deploy, and `--check-only=false` is refused. After the run it reads the full result with `sf project deploy report` and exits 2 if the target reports anything other than a check-only run. Check-only validates never target a Dev Hub. Use a scratch org or a sandbox.

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
