# Engine hardis (experimental)

> **Experimental, opt-in, and not in a published release yet.** This is planned for v1.1.0. The default engine is `native` and needs nothing from this page. See [what is in which release](releases-and-licenses.md#what-is-in-which-release).

With `salesforce.engine: hardis`, GlassOps deploys through the `sfdx-hardis` plugin of the Salesforce CLI instead of calling `sf project deploy start` itself. Everything else stays GlassOps: the plan, the org Id guard, `gate`, the PERMIT, the clean tree check, `verify` and `report`. The engine only changes which program makes the deploy call.

```yaml
salesforce:
  engine: hardis
  hardis:
    min_version: "8.20.1"   # optional, this is the default
```

---

## A separate program under its own license

`sfdx-hardis` is a separate program under the AGPL-3.0 license. GlassOps does not ship it, install it or build an image that contains it, and the published GlassOps images do not hold it. GlassOps only runs it as a command. You install it yourself. [`examples/hardis/Dockerfile`](https://github.com/glassops-platform/glassops/blob/main/examples/hardis/Dockerfile) shows how to build your own image from a GlassOps image, and this repository never builds or publishes that image.

Read the `sfdx-hardis` license before you build, run or share an image that contains it. This page is not legal advice.

Install the plugin at the version you want to run, for example:

```bash
echo y | sf plugins install sfdx-hardis@8.20.1
```

---

## What the engine runs

| Step | What runs |
|---|---|
| `plan` | GlassOps' own plan. `sfdx-hardis` does not choose what is deployed |
| `validate` | `sf project deploy validate`, exactly as the native engine does |
| `deploy` | `sf hardis:project:deploy:start` with the plan's `package.xml` and destructive manifests, the test level and the target |
| `deploy` of a validated Id | `sf project deploy quick`. If the org refuses the Id, a full deploy through `sfdx-hardis` follows, unless `salesforce.deploy.quick` is `require` |
| `deploy --resume` | `sf project deploy report` |

The deploy result records which path ran, for example `salesforce.hardis.validate_path: native` and `salesforce.hardis.deploy_path: hardis`, and the plugin version as `salesforce.hardis.version`.

The engine never runs `hardis:project:deploy:smart`. That command works out its own delta and can change source files before it deploys, so what it deploys could differ from the plan the PERMIT was issued for.

`deploy` checks the PERMIT and the working tree before it reaches the engine, so the clean tree rule in [PERMIT and rules](permit-and-rules.md#what-permit-verify-and-deploy-check) applies to this engine too.

---

## Why validate and quick deploy bypass the plugin

Tested with `sfdx-hardis` 8.20.1 against a scratch org:

- `sf hardis:project:deploy:validate` made a real deploy. The org changed, and the deploy result says `checkOnly: false`. **Do not use it as a check-only validate, by hand or in a pipeline.** GlassOps never does.
- `sf hardis:project:deploy:quick` refused both `--job-id` and `-i`, so a quick deploy through it does not work.

That is why the engine validates and quick deploys with `sf` directly. If a later version changes this, GlassOps still refuses a validate result that is not check-only.

---

## What the engine refuses

- **Deploy commands in the `sfdx-hardis` config.** If `config/.sfdx-hardis.yml`, another `config/.sfdx-hardis*.yml` or a file in `config/branches/` defines `commandsPreDeploy` or `commandsPostDeploy`, `validate` and `deploy` fail (exit 1) before any `sf` call, and `glassops doctor` fails the check `salesforce.hardis.commands`. Those commands would run against the org outside the plan the PERMIT binds. A config file that cannot be read counts as a problem too. Run such steps as their own pipeline steps instead.
- **A missing or old `sfdx-hardis`.** `doctor` (check `tool.sfdx-hardis`), `validate` and `deploy` compare the installed version with `min_version` and print the install steps. `doctor` passes with a `warning:` when the version is newer than 8.20.1, the version GlassOps was tested with.

## Pull request comments

`sfdx-hardis` posts pull request comments and notifications when it finds a CI token. The engine removes `GITHUB_TOKEN`, `SYSTEM_ACCESSTOKEN`, `CI_SFDX_HARDIS_GITLAB_TOKEN` and `CI_SFDX_HARDIS_BITBUCKET_TOKEN` from the `sfdx-hardis` process environment, so they are absent there and not set empty, and `glassops report` stays the only thing that writes to the pull request. See [the PR comment](pr-comment-and-telemetry.md#pull-request-comment).

---

## Tested here, and not

- The engine is covered by unit tests over output captured from `sfdx-hardis` 8.20.1 and `sf` 2.155.7 in a locally built image against a scratch org: a passing deploy, a failing validate, a test failure, a timeout, a destructive manifest and a config with deploy commands.
- The follow-up `deploy report` after a hardis timeout was not captured, so that step is untested against a real capture.
- No pipeline has run this engine end to end on a CI runner.
