# Engine hardis (experimental)

With `salesforce.engine: hardis`, GlassOps deploys through sfdx-hardis instead of calling
`sf project deploy start` itself. Everything else stays GlassOps: the plan, the org Id guard,
`gate`, the PERMIT, `verify` and `report`.

```yaml
salesforce:
  engine: hardis
  hardis:
    min_version: "8.20.1"   # optional, this is the default
```

sfdx-hardis is a separate program under its own license (AGPL-3.0). GlassOps does not ship it,
install it or build an image with it. `Dockerfile` in this folder shows how to build your own
image from the GlassOps image. Read the sfdx-hardis license before you build, run or share an
image that contains it. This page is not legal advice.

## What the engine runs

| Step | Command |
|---|---|
| `plan` | GlassOps' own plan. sfdx-hardis does not choose what is deployed |
| `validate` | `sf project deploy validate`, the same as the native engine |
| `deploy` | `sf hardis:project:deploy:start` with the plan's `package.xml` and destructive manifests |
| `deploy` of a validated Id | `sf project deploy quick`. If the org refuses the Id, a full deploy through sfdx-hardis follows, unless `salesforce.deploy.quick` is `require` |
| `deploy --resume` | `sf project deploy report` |

The deploy result records which path ran, for example `salesforce.hardis.validate_path: native`
and `salesforce.hardis.deploy_path: hardis`.

The engine never runs `hardis:project:deploy:smart`. That command works out its own delta and
can change source files before it deploys, so what it deploys could differ from the plan the
PERMIT was issued for.

## sfdx-hardis 8.20.1: its validate command deploys for real

Tested with sfdx-hardis 8.20.1 against a scratch org:

- `sf hardis:project:deploy:validate` ran a real deploy. The org changed and the deploy result
  says `checkOnly: false`. Do not use it as a check-only validate, by hand or in a pipeline.
- `sf hardis:project:deploy:quick` refused both `--job-id` and `-i`, so a quick deploy through
  it does not work.

That is why the engine validates and quick deploys with `sf` directly. If a later version changes
this, GlassOps still refuses a validate result that is not check-only.

## What the engine refuses

- **Deploy commands in the sfdx-hardis config.** If `config/.sfdx-hardis.yml`, another
  `config/.sfdx-hardis*.yml` or a file in `config/branches/` defines `commandsPreDeploy` or
  `commandsPostDeploy`, `validate` and `deploy` fail (exit 1) before any `sf` call, and `doctor`
  fails `salesforce.hardis.commands`. Those commands would run against the org outside the plan
  the PERMIT binds. Run such steps as their own pipeline steps instead.
- **A missing or old sfdx-hardis.** `doctor`, `validate` and `deploy` check the installed version
  against `min_version` and print the install steps. `doctor` warns when the version is newer
  than the one GlassOps was tested with (8.20.1).

## Pull request comments

sfdx-hardis posts pull request comments and notifications when it finds a CI token. The engine
blanks `GITHUB_TOKEN`, `SYSTEM_ACCESSTOKEN`, `CI_SFDX_HARDIS_GITLAB_TOKEN` and
`CI_SFDX_HARDIS_BITBUCKET_TOKEN` for the sfdx-hardis process, so `glassops report` stays the only
thing that writes to the pull request.
