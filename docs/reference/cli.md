# CLI Reference

> **Early work.** The `glassops` CLI is being built in phases and nothing is released yet. This page covers what exists today.

`glassops` is one Go binary that runs each delivery step as its own command. The goal is a governed pipeline that runs the same way in any CI system or on a laptop. Salesforce is the first provider.

---

## Commands

| Command | What it does |
|---|---|
| `glassops init` | Detects the provider and the CI system, then writes `glassops.yaml` and a CI file |
| `glassops doctor` | Checks the config, the tools, the git checkout and the environment variables the config names |

Other commands, such as `plan`, `gate`, `deploy` and `run`, are planned and not built yet.

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

## Install

There is no release yet. To try it, build from a checkout with Go 1.27.3 or later:

```bash
go build -o glassops ./cmd/glassops
./glassops doctor
```

The container image adds Node 22, the Salesforce CLI and `sfdx-git-delta`. Build it with `docker build -t glassops .` from the repository root.
