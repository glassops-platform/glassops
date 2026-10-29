# Salesforce provider fixtures

These files are real output from `sf` 2.154 and 2.155 and from `sfdx-git-delta` 6.45, captured
against a Developer Edition Dev Hub and a one-day scratch org. They were scrubbed before they
were committed. Org Ids, record Ids, usernames, names, hosts and tokens were replaced with
example values. The shapes are unchanged, and tests depend on the shapes.

Do not edit these files by hand to suit a test. When `sf` changes its output, capture it again
and scrub it the same way.

## `sf/`

| File | Command |
|---|---|
| `login-jwt-devhub.json` | `sf org login jwt` into the Dev Hub |
| `login-access-token-scratch.json` | `sf org login access-token` into the scratch org |
| `login-sfdx-url-scratch.json` | `sf org login sfdx-url` into the scratch org |
| `login-jwt-unknown-user.json` | `sf org login jwt` for a username that does not exist. sf reports "External client app is not installed in this org" |
| `login-jwt-wrong-key.json` | `sf org login jwt` signed by a key the app does not know: "invalid client credentials" |
| `login-jwt-bad-client-id.json` | `sf org login jwt` with a client id that is not an app in the org: "client identifier invalid" |
| `login-access-token-invalid.json` | `sf org login access-token` with a well-formed token that is not valid |
| `login-sfdx-url-invalid.json` | `sf org login sfdx-url` with a refresh token that is not valid. `cause` holds `invalid_grant` |
| `org-display-devhub.json`, `org-display-scratch.json` | `sf org display`. `accessToken` is redacted by `sf` itself, even with `--verbose` |
| `query-organization-devhub.json`, `query-organization-scratch.json` | `sf data query` of `SELECT Id, IsSandbox, OrganizationType, InstanceName, TrialExpirationDate FROM Organization` |
| `deploy-validate-pass.json` | `sf project deploy validate` that passed |
| `deploy-validate-compile-fail.json` | the same with an Apex compile error. Exit 1, and the body holds only `data.deployId` |
| `deploy-validate-test-fail.json` | the same with a failing Apex test. Exit 1, same thin body |
| `deploy-report-compile-fail.json`, `deploy-report-test-fail.json` | `sf project deploy report --job-id <id>` for the two failures. Exit 0 with the full body |
| `create-scratch-eca-login-fail.json` | `sf org create scratch` from a Dev Hub logged in by JWT through an external client app. Signup worked, then the login to the new org failed |
| `sgd-delta.json`, `sgd-delta-empty.json`, `sgd-delta-bad-ref.json` | `sf sgd source delta --json`. The JSON only names the output dir. The manifests are files |

## `sgd/`

The manifests `sfdx-git-delta` wrote for the fixtures above. `delta/` is a change that edits two
Apex classes and a permission set and deletes a custom field. `empty/` is a delta between a commit
and itself.

## Facts the fixtures show

- A scratch org reports `OrganizationType` "Developer Edition", `IsSandbox` true and a
  `TrialExpirationDate`. A sandbox has no `TrialExpirationDate`. A Developer Edition Dev Hub
  reports `IsSandbox` false.
- The login output carries the org's company name in `name`. It must never reach a result or a
  log.
- The login output spells one key `trailExpirationDate`.
- sf retries a JWT login itself. The failure message repeats the same error three to five
  times, one line per attempt.
- Every JWT failure has the name `JwtGrantError` and the same first line. The cause is only in
  the later lines of `message` and in `cause`.
