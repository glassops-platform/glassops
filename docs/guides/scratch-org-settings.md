# Scratch org definitions and settings

`glassops env create` makes a Salesforce scratch org for a pull request from a definition file. It reads the same file `sf org create scratch` reads, including the `settings` block, and applies the settings before it reports the org ready. This page says what is applied, what is refused, and what happens when the org refuses a setting.

## Prerequisites

- A Dev Hub target in `glassops.yaml` and an `environments` entry that names it.
- A scratch org definition file in the project.

## Installation

Nothing extra. `env create` uses the Salesforce CLI in the GlassOps image.

## Usage

Point the environment at the file with `definition`:

```yaml
environments:
  ephemeral:
    devhub: devhub
    definition: config/project-scratch-def.json
    ttl_days: 1
```

## What the definition may hold

| Key | Handled as |
|---|---|
| `edition`, `features`, `hasSampleData`, `language`, `country`, `description`, `release`, `sourceOrg` | Fields of the scratch org request. A list becomes a semicolon list |
| `orgName` | Ignored on purpose. The org is named after the environment so that `env delete` and `env sweep` can find it |
| `settings` | Converted to Settings metadata and deployed to the new org. See below |
| Any other key | Refused. `env create` names the keys it cannot apply and creates no org |

`env create` reads and converts the whole definition before it asks the Dev Hub for an org. A definition it cannot apply never creates an org, so you do not pay an org to find out.

## How settings are applied

Each key under `settings` that ends in `Settings` becomes one file. For example `lightningExperienceSettings` becomes `LightningExperience.settings-meta.xml`. The files go into a throwaway source project and deploy to the new org with `sf project deploy start --test-level NoTestRun`, after the login to the new org works. Fields keep the order the definition gives them. A list becomes one element per item, and a nested object becomes a nested element.

The conversion refuses, with a message that names the field:

- `objectSettings`, because it describes object metadata and not a Settings type.
- A key that does not name a Settings type, such as one that does not end in `Settings`.
- A settings block that is not an object.
- A field name that is not a valid element name.
- A `null` value, or a list inside a list.
- The same key twice in one object.

## When the org refuses a setting

Some settings need features the edition or the Dev Hub does not have. The org then fails the settings deploy. What `env create` does:

1. It deletes the new scratch org, even when the run itself was stopped.
2. It records a failed `create` check that lists each failure as `<type> <name>: <problem>`, followed by "The scratch org was deleted", and exits 1. `env.json` records the attempt.
3. If the delete also fails, the check is an error, the command exits 2 and the message says "Deleting the scratch org failed too" with the reason. `env.json` is written, so `env delete` or `env sweep` can remove the org later.

Exit 1 means the definition is at fault and no org is left. Exit 2 means the tooling failed: the deploy never got an answer, the result could not be read, or an org may still exist. The run never goes on in an org that is missing part of what its definition asked for.

The result records `salesforce.settings` (the member names), `salesforce.settings_deployed`, `salesforce.settings_deploy_id` when the org reported one, and `salesforce.settings_org_deleted` after a failure.

## What was tested

The conversion and the refusal path are covered by unit tests with a fake `sf`. How a real org answers depends on its edition and features, and some settings are refused there. A setting can also come back different from what the definition set. Run your own definition once and retrieve the settings you depend on before you trust them in a pull request pipeline.
