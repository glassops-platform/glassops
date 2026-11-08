# sfdx-hardis captures

Output from `sf` 2.155.7 with sfdx-hardis 8.20.1, run in a local image built from
`examples/hardis/Dockerfile` (never published), against a one-day Developer Edition scratch org.
No CI token was set in the container. Scrubbed like the files in `../sf/`. These files are output
only. No hardis code, schema or doc text is in this repo.

**Shape.** With `--json`, hardis prints its own envelope `{status, result: {outputstring}, warnings}`.
When the deploy succeeds, `outputstring` is an object holding the whole `sf` envelope. When it fails,
`outputstring` is a string: a short summary ("Deployment summary ... Status: Failed ...", or "No
deployment result available to summarize") followed by the `sf` error JSON as text.

| File | Command and result |
|---|---|
| `sf-plugins-hardis.json` | the `sf`, `sfdx-git-delta` and `sfdx-hardis` entries of `sf plugins --json` |
| `hardis-agent-flag-passed-through.json` | `hardis:project:deploy:validate --agent ...`. Exit 2. The wrapper passes `--agent` through to `sf`, which refuses it as a nonexistent flag |
| `hardis-validate-pass.json`, `hardis-validate-pass-report.json` | `hardis:project:deploy:validate --manifest ... --ignore-conflicts --test-level RunSpecifiedTests`. Exit 0. Then `sf project deploy report` on its Id. **It is not check-only**: `checkOnly` false, `Succeeded`, 2 components deployed. The 8.20.1 validate wrapper runs a real deploy |
| `hardis-validate-fail.json`, `hardis-validate-fail-report.json` | the same with a compile error in `Greeter`. Exit 1, string `outputstring` |
| `hardis-start-testfail.json` | `hardis:project:deploy:start` with a failing test. Exit 1, summary "Apex tests: 0 completed, 1 failed" |
| `hardis-start-timeout.json` | `hardis:project:deploy:start --wait 1` with slow tests. Exit 1, `ClientTimeoutError` inside the string. The follow-up report was not captured |
| `hardis-quick-jobid-flag.json`, `hardis-quick-pass.json`, `hardis-quick-refused.json` | `hardis:project:deploy:quick` with `--job-id` and with `-i`. Exit 2 every time, "Nonexistent flag" under context `DeployMetadata`. Quick deploy through the 8.20.1 wrapper does not work |
| `hardis-start-destructive.json` | `hardis:project:deploy:start --manifest ... --post-destructive-changes ...`. Exit 0, the class in the destructive manifest deleted |
| `hardis-start-pass.json`, `hardis-start-pass-report.json` | `hardis:project:deploy:start --manifest ... --test-level NoTestRun`. Exit 0, object `outputstring` |
| `hardis-start-with-commands.json`, `hardis-validate-with-commands.json` | the same two commands with a `config/.sfdx-hardis.yml` that defines `commandsPreDeploy` and `commandsPostDeploy` writing marker files. Exit 0, and neither marker file was written |

Every wrapper above (validate, quick, start) ran the `sf` deploy start command, as the error
contexts and stacks show. The job Id is only inside `outputstring`.
