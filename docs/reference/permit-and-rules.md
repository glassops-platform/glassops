# PERMIT and Rules

> **Early work.** The `glassops` CLI is being built in phases and nothing is released yet. This page covers what exists today.

A PERMIT is a signed file that records one decision: whether one plan may be deployed to one target. `glassops gate` makes the decision and writes the PERMIT. `glassops deploy` and `glassops permit verify` refuse to go on unless the PERMIT is genuine, still valid and about the thing being deployed. This page describes the PERMIT, the rules that feed it, and the evidence `gate` reads. The commands are in the [CLI reference](cli.md).

---

## What a PERMIT is

The PERMIT is JSON in `.glassops/permit.json`, with the schema name `glassops.permit/v1`. It holds:

| Field | Meaning |
|---|---|
| `id` | A random identifier, in hex |
| `allow` | `true` or `false`. A deny is recorded as a PERMIT too |
| `issued_at`, `expires_at` | When it was made and when it stops being valid. The lifetime is `policy.permit_ttl` |
| `target` | The target `name`, `tier`, `id` and `kind`. The Id and kind are what the login reported, unless `--no-connect` was used |
| `commit` | The commit the CI reported |
| `plan_sha256` | The hash of the plan files |
| `config_sha256` | The hash of `glassops.yaml` |
| `validation` | Optional. The deploy Id of a passed check-only validation, its test level and `passed` |
| `separation_of_duties` | How the separation of duties rule was met. See below |
| `checks` | Every rule result, each with its rule name, outcome and detail |
| `signature` | The algorithm, the key id and the signature value |

A PERMIT binds to a specific target, commit, plan and config. Change any of them after `gate` ran and the PERMIT no longer matches.

---

## The signing key

The PERMIT is signed with HMAC-SHA256. The key comes from the environment variable `GLASSOPS_PERMIT_KEY`, which holds base64 text that decodes to at least 32 bytes. Standard base64 is accepted with or without padding. A shorter key is refused, and the error never contains the key. One way to make a key:

```bash
openssl rand -base64 48
```

The `key_id` in the signature is the first eight hex characters of the SHA-256 of the key. It names the key without revealing it. `permit verify` and `deploy` refuse a PERMIT whose `key_id` differs from their own key's, with the message that it was signed with a different key.

The signature covers the whole PERMIT, with the signature value left out: keys sorted, no extra whitespace, times in UTC. A PERMIT file with a field the format does not know, or with trailing data, is refused as unreadable, because such a field would not be covered by the signature.

**The key must never be available to pull request jobs.** Anyone who can run a job with the key can sign an allow. Put it in a protected CI environment that only the deploy job uses. `glassops doctor` adds a warning when `GLASSOPS_PERMIT_KEY` is set on a `pull_request` or `pull_request_target` run. See [Deploy with GitHub](../guides/deploy-with-github.md).

---

## What `gate` does

`gate` always writes a signed PERMIT, including when policy denies the deploy. The deny is part of the record. Before it does anything else, it deletes any PERMIT already at the output path, so a stale allow can never stand in for the current run.

Two things stop it before a PERMIT is written, and both exit 2: `GLASSOPS_PERMIT_KEY` is missing or unusable, and the plan file or the target cannot be found. Everything after that still produces a signed PERMIT.

The checks in the PERMIT come in this order: the `guard` check, the `validation.bound` check, the rule results, then any `evidence.*` failures. The guard is the org Id and tier check described in the [CLI reference](cli.md#the-guard). With `--no-connect` it does not run, the PERMIT records the configured Id and the kind the tier requires, and the `guard` check is `skipped`.

`gate` also writes the facts it evaluated to `.glassops/facts.json`.

### Fail closed

Missing evidence is never a pass.

- A rule that needs a fact that is not there fails, or says so in its detail.
- A tool or infrastructure problem while gathering evidence adds a failing `evidence.<name>` check. The PERMIT denies, is still signed and written, and `gate` exits 2 instead of 1. Examples: the guard refusing the target, a GitHub API call that fails, a result file that exists but cannot be read, a missing commit or plan hash, and plan files that no longer match the plan hash.
- A rule set with no rules is a defect, never a pass.

### The ephemeral automatic PERMIT

A target whose login the guard accepted as a scratch org gets an automatic PERMIT. Only the org Id guard applies. The rules are not run, no GitHub API call is made, and the PERMIT carries an `ephemeral` check and a `not_applicable` separation of duties record. It is issued only when the guard really ran and saw a scratch org, so `--no-connect` on an ephemeral target runs the full rule set instead. The result fact `gate.ephemeral` says which happened.

---

## What `permit verify` and `deploy` check

Both use the same verification, in this order. The first failure is the one reported.

`permit verify`, before the PERMIT checks:

1. The key. A missing or unusable key is an error (exit 2) before anything else is read.
2. The config, the target, the change set and the commit. A problem reading any of them is an error (exit 2).
3. The PERMIT file. A missing file is an error (exit 2). A file that is not a readable PERMIT is a failure (exit 1).
4. The plan files. If the files under `.glassops/plan/` no longer hash to the change set's `plan_sha256`, the plan was edited after `plan` and the check fails (exit 1).

`deploy`, before the PERMIT checks, reads the target, the change set, the config hash and the commit, then checks the plan files the same way. In `deploy` every refusal is exit 2, because nothing may touch the target. The login comes after verification.

The PERMIT checks, shared by both:

1. The signature: it exists, the algorithm is `HMAC-SHA256`, the key id matches, and the value verifies.
2. The schema is `glassops.permit/v1`.
3. `allow` is true.
4. It has not expired.
5. The target name matches.
6. The target Id matches. Ids compare on their canonical form, so a 15 and an 18 character form of one Id are equal.
7. The commit matches.
8. The plan hash matches.
9. The config hash matches.

Checking the signature first means a forged or damaged PERMIT is rejected before any of its contents are trusted. `deploy` additionally compares the PERMIT against the deployed commit, which is the CI's commit, or `git rev-parse HEAD` when the CI reports none.

---

## Validation binding and quick deploy

`gate` can bind a check-only validation into the PERMIT, so `deploy` can quick deploy it instead of deploying again. It binds the validation only when all of these hold for `.glassops/results/validate.json`:

- The result passed.
- It is for the same target name.
- It was a check-only run with a deploy Id.
- It records the plan hash it validated (`validate.plan_sha256`) and that hash equals the current plan.

The PERMIT then carries `validation` with the deploy Id and the recorded test level. The `validation.bound` check says whether it was bound and why not. It never denies. Without a bound validation, `deploy` runs a full deploy.

How `deploy` uses it depends on `salesforce.deploy.quick` in `glassops.yaml`:

| Value | Behaviour |
|---|---|
| `prefer` (default) | Quick deploy the bound validation. If the target refuses it, fall back to a full deploy and record `salesforce.quick_refused` |
| `require` | Quick deploy only. With no passed validation in the PERMIT, `deploy` fails with exit 1 before any login. If the target refuses the quick deploy, no full deploy runs |
| `off` | Always run a full deploy |

The target refuses a quick deploy when the validation was never run, was already deployed, or is more than ten days old.

---

## The v1 rules

`gate` runs these seven rules, in this order. Every rule runs, so the PERMIT lists every reason for a deny. Only a `fail` outcome denies. The outcomes are `pass`, `fail`, `skipped`, `waived_by_config` and `satisfied_by_approval`.

| Rule | Outcomes |
|---|---|
| `plan_not_empty` | `fail` if the plan has no changes, or the facts do not say. Otherwise `pass` |
| `window.<tier>` | See below |
| `blackout` | See below |
| `coverage_min` | See below |
| `destructive_needs_label` | See below |
| `separation_of_duties` | See below |
| `analyzer_max_severity` | See below |

### `window.<tier>`

Named after the target's tier, such as `window.prod`. Configured as `policy.windows.<tier>`, a list of weekly windows with `days`, `from`, `to` and `tz`.

- `skipped` when no window is configured for the tier.
- `pass` when now is inside any window for the tier.
- `fail` when now is outside every window, or a window cannot be used (an unknown time zone or a bad clock).
- `fail`, under the plain name `window`, when the tier is unknown and any window is configured anywhere.

Times are wall clock times in the window's zone, so the window moves with daylight saving time. A window is `from` up to but not including `to`. A window whose `to` is earlier than its `from` runs past midnight and belongs to the day it starts on. A window where `from` equals `to` never matches.

### `blackout`

Configured as `policy.blackouts`, a list with `from` and `to` dates (`YYYY-MM-DD`), optional `tiers` and an optional `reason`. Dates are compared in UTC, because a blackout has no time zone, and both ends are included.

- `skipped` when no blackouts are configured.
- `fail` when today (UTC) falls in a blackout that covers the target's tier. An empty `tiers` list covers every tier. When the tier is not known, every blackout in range counts.
- `pass` otherwise.

### `coverage_min`

Configured as `policy.rules.coverage_min`, a whole number from 0 to 100.

- `skipped` when it is not set.
- `fail` when the plan facts do not list the changed Apex classes.
- Otherwise it checks each changed Apex class against the per class coverage facts from the validate result. A class below the minimum fails. A class with no coverage fact fails too, unless its name ends in `Test` or `_Test`. Those are treated as test classes, and the detail says how many.
- `pass` when every class is at or above the minimum.

### `destructive_needs_label`

Configured as `policy.rules.destructive_needs_label` with a `label` and optional `tiers`. An empty `tiers` list covers every tier.

- `skipped` when it is not set, or the target's tier is not covered.
- `fail` when the tier is unknown, or the facts do not say how many destructive changes the plan has.
- `pass` when the plan has no destructive changes.
- With destructive changes: `fail` when no pull request was found, or the pull request lacks the label. `pass` when it has the label.

### `separation_of_duties`

Configured as `policy.rules.separation_of_duties` with `tiers` and an optional `approval`. It makes sure the person who deploys is not the author of the change.

- `skipped` when the rule is not set.
- `waived_by_config` when the target's tier is not in `tiers`. **An empty `tiers` list turns the rule off everywhere**, because every tier is then outside the list.
- Enforced when the tier is in `tiers`. The author is the pull request author and the deployer is the CI actor, compared without regard to case.
    - With no `approval` configured, it `pass`es when both are known and differ, and `fail`s when either is unknown or they are the same.
    - With an `approval` configured, it `pass`es when they differ. When they are the same or unknown, it looks for a configured approval and gives `satisfied_by_approval`, or `fail` when none is found.

Two kinds of approval can satisfy the rule:

| Mechanism | Configured as | Counts when |
|---|---|---|
| `environment` | `approval.environment` | The run has an approved review for that environment |
| `label` | `approval.label` and `approval.approvers` | The pull request has the label, and the person who added it is one of the approvers |

A label approver may be the pull request author. The explicit approval by a listed approver is the control.

The PERMIT's `separation_of_duties` block records the result. Its `mode` is `enforced`, `waived_by_config`, `satisfied_by_approval` or `not_applicable` (the rule is off, or the PERMIT is an ephemeral automatic one). It also holds the author, the deployer and, for an approval, the mechanism and approver.

### `analyzer_max_severity`

Configured as `policy.rules.analyzer_max_severity`, a whole number from 1 to 5, where 1 is the worst.

- `skipped` when it is not set.
- `skipped` when there are no analyze facts, because the analyzer did not run. This is the one place a missing fact does not fail: a limit with no analysis behind it is not enforced until analysis exists.
- `fail` when the facts show any violations at the configured severity or worse.
- `pass` otherwise.

---

## Configuration example

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
  permit_ttl: 2h
  windows:
    prod:
      - { days: [tue, thu], from: "18:00", to: "22:00", tz: America/Chicago }
  blackouts:
    - { from: "2026-12-20", to: "2027-01-03", tiers: [prod], reason: holiday freeze }
  rules:
    coverage_min: 85
    destructive_needs_label:
      label: destructive-ok
      tiers: [prod]
    separation_of_duties:
      tiers: [prod]
      approval:
        environment: glassops-deploy
    analyzer_max_severity: 2
salesforce:
  test_level:
    prod: RunLocalTests
  deploy:
    quick: prefer
```

`permit_ttl` is a duration from 5 minutes to 24 hours. The default is 2 hours.

---

## Evidence `gate` reads from GitHub

Some rules need facts that only GitHub knows. When the CI is GitHub and a token is available in `GITHUB_TOKEN` or `GH_TOKEN`, `gate` reads them through the GitHub API. Without a token it reads nothing, the facts stay absent, and the rules that need them fail. The ephemeral automatic PERMIT makes no API calls. `GITHUB_API_URL` overrides the API address.

| Read | Used for | Token permission |
|---|---|---|
| `GET /repos/{repo}/pulls/{number}` on a `pull_request` run | Pull request author and labels | `pull-requests: read` |
| `GET /repos/{repo}/commits/{sha}/pulls` on a push | The merged pull request for the commit, its author and labels | `pull-requests: read` |
| `GET /repos/{repo}/actions/runs/{run_id}/approvals` | Environment approvals for this run | `actions: read` |
| `GET /repos/{repo}/issues/{number}/events` | Who added the approval label, only when `approval.label` is configured and the pull request has it | `issues: read` |

On a push, the actor is `GITHUB_ACTOR`, the person who pushed or merged. A push that was not merged from a pull request has no author, and a rule that needs one fails. A call that fails adds an `evidence.ci.*` failure, so `gate` denies and exits 2.

The job also needs `contents: write` for `deploy`, because it pushes the `glassops/deployed/<target>` tag. Without it the deploy still succeeds and only the tag push warns.
