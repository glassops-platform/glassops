# Configuration reference

`glassops.yaml` is optional. Without it, the commands use their defaults and the default [pipelines](../guides/run-a-pipeline.md#the-default-pipelines). The schema is [`schemas/config.json`](https://github.com/glassops-platform/glassops/blob/main/schemas/config.json). It rejects keys it does not know, so a typo is an error. `glassops doctor` validates the file.

Secrets are never written in the file. A secret field names an environment variable with a `$VAR` value, and the value stays in your CI's secret store.

Beyond the schema, these rules are checked:

- Org Ids look like Salesforce org Ids, unless written as a `$VAR`.
- No two targets share an org Id.
- At most one target has tier `prod`.
- Every policy window names a time zone that exists, and its `from` and `to` differ.
- Every blackout has valid dates and ends on or after its start date.
- `policy.permit_ttl` is a duration from 5 minutes to 24 hours.
- A separation of duties approval sets `environment` or `label`, and a `label` needs `approvers`.
- Every `environments.<name>.devhub` names a configured target.
- No pipeline step is empty.

Names (targets, tiers, environments, pipelines) are lowercase letters, digits, `_` and `-`, starting with a letter.

## Top-level keys

| Key | Required | What it holds |
|---|---|---|
| `version` | yes | The schema version. Always `1` |
| `provider` | no | `salesforce`. Detected from the project files when absent |
| `targets` | no | Orgs that changes go to, with their login |
| `policy` | no | Deploy windows, blackouts and the rules `gate` evaluates |
| `salesforce` | no | Salesforce provider settings |
| `environments` | no | Ephemeral environments that `env create` makes |
| `test` | no | Test levels and coverage minimums for `test` |
| `check` | no | Project scripts that `check` runs |
| `analyze` | no | Code Analyzer settings |
| `pipelines` | no | Named lists of steps for `run` |

### `version` and `provider`

```yaml
version: 1
provider: salesforce
```

### `targets`

One entry per org. The key is the target name that `--target` takes.

| Field | Meaning |
|---|---|
| `id` | The org Id the login must land on (the [guard](cli.md#the-guard)), or a `$VAR` holding it |
| `tier` | A name that policy windows, rules and test levels key on. `prod`, `devhub` and `ephemeral` have fixed meaning in the guard |
| `branch` | The git branch that deploys to this target. `plan --from auto` uses its merge base when the target has no deployed tag |
| `auth` | The login. See [Authentication](cli.md#authentication) |

```yaml
targets:
  devhub:
    id: $SF_DEVHUB_ORG_ID
    tier: devhub
    auth: { method: jwt, username: $SF_USERNAME, client_id: $SF_CLIENT_ID, key: $SF_JWT_KEY }
  prod:
    id: $SF_PROD_ORG_ID
    tier: prod
    branch: main
    auth: { method: jwt, username: $SF_PROD_USERNAME, client_id: $SF_CLIENT_ID, key: $SF_JWT_KEY }
```

### `policy`

What `gate` evaluates. The rules are described in [PERMIT and rules](permit-and-rules.md#the-v1-rules).

| Field | Meaning |
|---|---|
| `windows` | Per tier, a list of allowed deploy windows: `days`, `from`, `to` (24 hour `HH:MM`) and `tz` (an IANA zone) |
| `blackouts` | Date ranges (`from`, `to`, optional `tiers` and `reason`) when no deploy is allowed |
| `rules` | `coverage_min`, `analyzer_max_severity`, `separation_of_duties`, `destructive_needs_label` |
| `permit_ttl` | How long a PERMIT stays valid, from `5m` to `24h`. The default is `2h` |
| `actors` | Accepted by the schema. No command reads it yet |

```yaml
policy:
  permit_ttl: 2h
  windows:
    prod: [{ days: [tue, thu], from: "18:00", to: "22:00", tz: America/Chicago }]
  blackouts:
    - { from: "2026-12-20", to: "2027-01-03", tiers: [prod], reason: holiday freeze }
  rules:
    coverage_min: 75
    analyzer_max_severity: 2
    destructive_needs_label: { label: destructive-ok, tiers: [prod] }
    separation_of_duties:
      tiers: [prod]
      approval: { environment: glassops-deploy }
```

### `salesforce`

| Field | Meaning |
|---|---|
| `test_level` | Map of tier to the test level `validate` and `deploy` use: `NoTestRun`, `RunSpecifiedTests`, `RunLocalTests` or `RunAllTestsInOrg`. The `default` key applies to any tier not listed |
| `delta.enabled` | `false` makes `plan` list the whole source instead of a delta |
| `deploy.quick` | `prefer` (the default), `require` or `off`. See [validation binding](permit-and-rules.md#validation-binding-and-quick-deploy) |
| `ephemeral.seed` | `minimal` (the default), `full` or `off`. See [the smart seed](cli.md#the-smart-seed) |
| `engine`, `api_version` | Accepted by the schema (`native`, and `auto` or a version such as `62.0`). No command reads them yet |

```yaml
salesforce:
  test_level: { default: RunLocalTests, prod: RunLocalTests }
  deploy: { quick: prefer }
  ephemeral: { seed: minimal }
```

### `environments`

One entry per ephemeral environment, keyed by the target name the environment registers as. The default pr pipeline uses the name `ephemeral`. See [`env create`](cli.md#glassops-env-create).

| Field | Meaning |
|---|---|
| `devhub` | The target that creates the environment (required) |
| `definition` | Scratch definition file, relative to the project root |
| `ttl_days` | Days the environment lives. The default is `1` |
| `shape` | Source org Id for an org shape, or a `$VAR` holding one |
| `prefix` | Starts every environment name. The default is `glassops` |

```yaml
environments:
  ephemeral:
    devhub: devhub
    definition: config/project-scratch-def.json
    ttl_days: 1
```

### `test`

Settings for [`glassops test`](cli.md#glassops-test).

| Field | Meaning |
|---|---|
| `level` | Map of tier to test level, with a `default` key. The `--level` flag overrides it |
| `coverage.org_min` | Minimum org-wide coverage, 0 to 100 |
| `coverage.class_min` | Minimum coverage for each class the test run reports, 0 to 100 |

```yaml
test:
  level: { default: RunLocalTests }
  coverage: { org_min: 75, class_min: 75 }
```

### `check`

Scripts that [`glassops check`](cli.md#glassops-check) runs with `sh -c`.

```yaml
check:
  scripts:
    - { name: lint, run: npm run lint }
    - { name: unit, run: npm test }
```

### `analyze`

Settings for [`glassops analyze`](cli.md#glassops-analyze). Nothing runs until `enabled` is `true`.

| Field | Meaning |
|---|---|
| `enabled` | `true` turns `analyze` on. It also needs the `glassops-analyzer` image |
| `rule_selector` | Code Analyzer rules to run. The default is `Recommended` |
| `max_severity` | The default for `policy.rules.analyzer_max_severity`, 1 (critical) to 5. The policy value wins when both are set |

```yaml
analyze:
  enabled: true
  rule_selector: Recommended
  max_severity: 2
```

### `pipelines`

Named lists of steps for [`glassops run`](cli.md#glassops-run). Each step is a string (a glassops command line without the `glassops` prefix) or a map with `step`, optional `when` (`success`, `failure` or `always`) and, for the step `exec`, a `run` shell command. A pipeline needs at least one step.

```yaml
pipelines:
  pr:
    - plan
    - check
    - step: exec
      run: npm run build
    - step: report
      when: always
```

The default `pr` and `deploy` pipelines apply only when the project has no `glassops.yaml`. A file that has a `pipelines` key replaces them, and one without the key has no pipelines, so `glassops run pr` fails with an unknown pipeline. `glassops init` writes the defaults into the file. See [Run a pipeline](../guides/run-a-pipeline.md).
