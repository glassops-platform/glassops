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
| `deploy-start-pass.json` | `sf project deploy start` that passed. Exit 0, `Succeeded`, `checkOnly` false, 4 of 4 components |
| `deploy-start-compile-fail.json` | the same with an Apex compile error. Exit 1, and unlike `deploy validate` the body is full: `Failed`, 2 of 4, two component failures |
| `deploy-start-async.json` | `sf project deploy start --async`. Exit 0, `Queued`, and only the job Id |
| `deploy-report-pass.json` | `sf project deploy report --job-id <id>` for the async deploy above, after it finished. The full body |
| `deploy-resume-pass.json` | `sf project deploy resume --job-id <id>` for the same deploy. The same full body |
| `deploy-start-succeeded-partial.json` | `sf project deploy start --ignore-errors` with one bad class. Exit 68, `SucceededPartial`, 2 of 4. Without `--ignore-errors` a component error fails the whole deploy |
| `deploy-validate-delta-pass.json` | `sf project deploy validate` of a delta with post-destructive changes. Exit 0, `Succeeded`, `checkOnly` true |
| `deploy-quick-pass.json` | `sf project deploy quick --job-id <id>` of that validation. Exit 0, `Succeeded`, 4 of 4, no tests run, and a new deploy Id |
| `deploy-report-quick.json` | `sf project deploy report` for the quick deploy |
| `deploy-quick-already-used.json` | a second quick deploy of the same validation. Exit 1, `CannotQuickDeployError`. A validated Id can be quick deployed once |
| `deploy-quick-invalid-id.json` | a quick deploy of an Id that was never validated. Exit 1, `sf:UNKNOWN_EXCEPTION`, "invalid parameter value" |
| `deploy-validate-notestrun-rejected.json` | `sf project deploy validate --test-level NoTestRun`. Exit 1. `deploy validate` accepts only RunAllTestsInOrg, RunLocalTests, RunSpecifiedTests and RunRelevantTests |
| `deploy-validate-destructive-not-found.json` | the delta validated again after the field was already deleted. Exit 1, `FailedValidationError`, and the body holds only `data.deployId` |
| `deploy-report-destructive-not-found.json` | `sf project deploy report` for that validation. `Failed`, 3 of 3, `numberComponentErrors` 0, and one component failure on `destructiveChanges.xml`: "No CustomField named: Account.Tier__c found" |
| `deploy-start-seed-field.json` | `sf project deploy start --metadata CustomField:Account.Tier__c` from the base ref, which puts the deleted field back. Exit 0, 1 of 1 |
| `deploy-validate-delta-after-seed.json` | the same delta validated after that seed. Exit 0, `Succeeded`, 4 of 4, check-only |
| `deploy-start-seed-missing-parent.json` | `sf project deploy start --metadata CustomField:Widget__c.Size__c` when the object does not exist. Exit 1, "Entity 'Widget__c' not found." |
| `apex-test-pass.json` | `sf apex run test --code-coverage --result-format json` with `--wait`, all passing. Exit 0, `summary.outcome` "Passed", `summary.orgWideCoverage` "100%" |
| `apex-test-fail.json` | the same with one failing test. Exit 100, `summary.outcome` "Failed", `summary.failing` 1 |
| `apex-test-uncovered-class.json` | the same after deploying a class no test touches. `coverage.coverage[]` does not list that class, while `summary.orgWideCoverage` drops to "66%". Org-wide coverage catches an untested class and the per-class list does not |
| `apex-test-async.json` | the same without `--wait`. Exit 0, and only `testRunId` |
| `apex-get-test-fail.json` | `sf apex get test --test-run-id <id>` for that run, which has a failing test. Exit 100 with the full body |
| `query-active-scratch-org.json` | `sf data query` of `ActiveScratchOrg` on the Dev Hub: `Id`, `ScratchOrg`, `SignupUsername`, `OrgName`, `CreatedDate`, `ExpirationDate`, `ScratchOrgInfoId` |
| `query-scratch-org-info.json` | `sf data query` of `ScratchOrgInfo` on the Dev Hub: `Id`, `Status`, `ErrorCode`, `ScratchOrg`, `OrgName`, `CreatedDate`, `ExpirationDate` |

## `sgd/`

The manifests `sfdx-git-delta` wrote for the fixtures above. `delta/` is a change that edits two
Apex classes and a permission set and deletes a custom field. `empty/` is a delta between a commit
and itself.

## `codeanalyzer/`

Output from `sf code-analyzer run` (plugin 5.16.0, core 0.53.0) in a container, over a project
with three low severity violations.

| File | Command |
|---|---|
| `code-analyzer.json` | `--output-file` with a `.json` name. `violationCounts` holds `total` and `sev1` to `sev5`, and `violations` lists each one. Exit 0 with violations |
| `code-analyzer.sarif` | `--output-file` with a `.sarif` name. SARIF 2.1.0 written by the plugin itself. It has one run, for `pmd`, the only engine with violations |
| `code-analyzer-threshold-stdout.txt` | the terminal output with `--severity-threshold 4`. Exit 4, "Error (ThresholdExceeded)" |

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
- `deploy start` failures carry the full result body. Only `deploy validate` failures are thin.
- A quick deploy is single use. A second attempt is `CannotQuickDeployError`.
- When a destructive change names a component the org no longer has, the validate fails on a
  warning. `numberComponentErrors` stays 0, and the reason is only in the component failures.
- `summary.orgWideCoverage` is a string with a percent sign, such as "100%". A failing test
  run is exit 100.
- `ActiveScratchOrg` has no `Status` field. Its `ScratchOrg` is the 15-character org Id.
- Code Analyzer exits 0 when it finds violations. Only `--severity-threshold` changes the exit
  code.

## `sf/`, verify and timeout captures

Captured with `sf` 2.156.8 against a one-day Developer Edition scratch org, from a project with
`Greeter`, `GreeterTest`, a class no test touches (`Orphan`), a slow test class (`SlowTest`) and an
autolaunched Flow (`Glassops_Probe`). Scrubbed the same way as the files above.

| File | Command |
|---|---|
| `deploy-start-coverage-fail.json`, `deploy-report-coverage-fail.json` | `sf project deploy start --test-level RunLocalTests` with `Orphan` in the deploy. Exit 1, `Failed`, `numberComponentErrors` 0, and every component still listed in `componentSuccesses`. The reason is only in `runTestResult.codeCoverageWarnings`: org-wide coverage 50%, 75% required. Nothing was deployed |
| `tooling-deployrequest-coverage-fail.json` | Tooling `DeployRequest` by Id for that deploy: `Status` Failed, `NumberComponentsDeployed` 4 of 4 |
| `deploy-start-flow-pass.json`, `deploy-report-flow-pass.json` | the same without `Orphan`. Exit 0, `Succeeded`, 3 of 3, one test |
| `tooling-deployrequest-pass.json` | Tooling `DeployRequest` by Id: `Status`, `CheckOnly`, `NumberComponentsDeployed`, `NumberComponentsTotal`, `NumberComponentErrors`, `NumberTestsCompleted`, `NumberTestErrors`, `NumberTestsTotal`, `RunTestsEnabled`, `StartDate`, `CompletedDate`, `CreatedById` |
| `tooling-flowdefinition-active.json`, `tooling-flow-active.json` | Tooling `FlowDefinition` (`ActiveVersionId`, `LatestVersionId`) and `Flow` (`VersionNumber`, `Status`, `LastModifiedDate`, `DefinitionId`) after the Flow deployed as Active. Both version Ids are the same |
| `deploy-start-flow-draft.json` | the same Flow deployed again with `<status>Draft</status>`. Exit 0, `Succeeded` |
| `tooling-flowdefinition-draft.json`, `tooling-flow-draft.json` | after that deploy. `ActiveVersionId` is still version 1, `LatestVersionId` is the new version 2 with `Status` Draft. A deploy can succeed and leave the deployed Flow inactive |
| `deploy-start-foreign.json`, `tooling-deployrequest-window.json` | a second deploy (`Orphan`, NoTestRun), then Tooling `DeployRequest` filtered on `CreatedDate` from the start of the window. Four deploys, ordered by `CreatedDate` |
| `apex-test-with-orphan.json` | `sf apex run test --code-coverage` with `Orphan` deployed. The per-class list leaves `Orphan` out |
| `tooling-coverage-aggregate.json` | Tooling `ApexCodeCoverageAggregate` after that run. It lists `Orphan` with `NumLinesCovered` 0 and `NumLinesUncovered` 2. The aggregate catches the class the test output leaves out |
| `tooling-classes-without-aggregate.json` | Tooling `ApexClass` with no aggregate row. Only the test class |
| `deploy-start-timeout.json` | `sf project deploy start --wait 1` with tests that take longer. Exit 1, `ClientTimeoutError`, and the job Id only in `data.id` |
| `deploy-report-inprogress.json` | `sf project deploy report --job-id <id>` right after. Exit 0, `InProgress`, `done` false, 9 of 12 tests |
| `deploy-report-after-timeout-done.json` | the same report after the deploy finished. `Succeeded`, 12 tests |
| `deploy-validate-for-checkonly.json`, `tooling-deployrequest-checkonly.json` | `sf project deploy validate`, then Tooling `DeployRequest` by Id. `CheckOnly` true, `RunTestsEnabled` true |
| `deploy-start-notestrun.json`, `tooling-deployrequest-notestrun.json` | `sf project deploy start --test-level NoTestRun`, then Tooling `DeployRequest`. `RunTestsEnabled` false, `NumberTestsTotal` 0 |
| `deploy-start-lock-released.json`, `deploy-report-lock-released.json` | a transient platform error. `sf project deploy start` exited 1 with `MetadataTransferError` "Lock is already released" and the job Id in `data.id`, and `sf project deploy report` on that Id said `Succeeded`. An error from `deploy start` does not mean nothing was deployed |
| `deploy-settings-two.json` | the T8 conversion of the portfolio AI definition's `lightningExperienceSettings` and `mobileSettings`, deployed to a plain Developer scratch org. Exit 0, 2 of 2. Retrieved afterwards, both files equal the ones `sf org create scratch` applied from the same definition, and `enableS1EncryptedStoragePref2` went from true to false |
| `deploy-settings-full-refused.json` | all five converted Settings to the same org, which lacks the definition's features. Exit 1, `Failed`, component failures on `AgentPlatform` and `EinsteinGpt`: "Not available for deploy for this organization" |
| `deploy-report-folders-settings.json`, `package-folders-settings.xml` | trimmed from a real full-source deploy of a larger project (1167 components) into a scratch org, keeping the 19 successes and manifest members that show folder and Settings naming. The manifest names a dashboard or report folder as `Dashboard` or `Report` with a trailing slash (`Portfolio_Owner_Dashboards/`) and an email folder as `EmailTemplate` without one (`Portfolio_Contact`). The deploy reports them as `DashboardFolder`, `ReportFolder` and `EmailFolder` with the bare name. The manifest names `Settings` `EmailAdministration`, and the deploy reports `EmailAdministrationSettings` `EmailAdministration`. The `files` list and the test results were removed |

Also seen, not saved: `sf project deploy resume` refuses `--target-org` ("Nonexistent flag"). The portfolio AI definition, with `Einstein1AIPlatform` and `Communities`, signs up on a Developer Edition Dev Hub, and `Settings:AgentPlatform` does not come back from a retrieve there.

## `sf/`, RunRelevantTests and quick deploy captures

Captured with `sf` 2.156.8 against a Developer Edition production org on 2026-11-09, then
scrubbed the same way as the files above. Each is a Tooling `DeployRequest` by Id with the fields
of `tooling-deployrequest-pass.json`.

| File | What ran |
|---|---|
| `tooling-deployrequest-relevant-noapex.json` | `sf project deploy validate --test-level RunRelevantTests` of an LWC and a custom field, no Apex. `CheckOnly` true, `RunTestsEnabled` true, `NumberTestsTotal` 0. `sf project deploy quick` of it, with no other deploy in between, failed with `CannotQuickDeployError` |
| `tooling-deployrequest-relevant-apex.json` | the same command for 14 Apex classes. `RunTestsEnabled` true, 272 of 272 tests from 19 test classes, picked by Salesforce |
| `tooling-deployrequest-quick.json` | `sf project deploy quick` of that validation. `CheckOnly` false, 14 of 14 components, `RunTestsEnabled` false, `NumberTestsTotal` 0 |

Also seen, not saved: a quick deploy of a validation that ran tests was refused with the same
`CannotQuickDeployError` after another deploy reached the org between the validation and the
quick deploy.
