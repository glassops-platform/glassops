# Pull request comment and OpenTelemetry export

> **Opt-in, and not in a published release yet.** Both are planned for v1.1.0. Neither does anything until you turn it on. See [what is in which release](releases-and-licenses.md#what-is-in-which-release).

---

## Pull request comment

`glassops report` can post its summary as a comment on the pull request. It creates the comment the first time and updates the same comment on every later run, so a pull request has one GlassOps comment, not one per push.

Turn it on with the flag:

```bash
glassops report --pr-comment
```

or in `glassops.yaml`, which also covers the `report` step of a pipeline:

```yaml
report:
  pr_comment: true
```

### What it posts

The comment body is the same redacted markdown that `report` writes to the CI's run summary, with a hidden line `<!-- glassops-report -->` first. That line is how a later run finds its comment: `report` reads the comments on the pull request and updates the first one that holds it. A body longer than 60000 characters is cut at a line end with a note that the run summary has the rest, because GitHub rejects longer comments.

### When it applies

- GitHub only. On Bitbucket Pipelines and GitLab CI it posts nothing.
- On a `pull_request` run only. On a push or any other event it posts nothing and adds a notice.
- It needs the repository, the pull request number and a token. The token is `GITHUB_TOKEN`, or `GH_TOKEN` when that is not set. `GITHUB_API_URL` changes the API address. When something is missing, `report` adds a warning and posts nothing.

### Permission

The job needs `pull-requests: write`. Give it to the job that runs `report` and to no other. A pull request from a fork gets a read-only token, so the comment cannot be posted there. It warns instead.

The reusable workflow `glassops-pr.yml` grants `pull-requests: write` to its one job, the one that runs `report`. The caller must grant `pull-requests: write` too, because a called workflow cannot hold more than its caller grants. A caller that grants less is refused when the run starts. The workflow's other permissions stay `contents: read` and `security-events: write`.

### It never changes the result

A comment that cannot be posted adds a warning annotation, for example "the job needs pull-requests: write", and `report` goes on to write the SARIF file and exits with the worst status among the results, as it always does. The comment is a convenience, so the job's result never depends on it.

### What is tested for the comment

The calls are covered by unit tests against a local fake of the GitHub issue comments API, written from GitHub's REST documentation. The create, then update, flow has not been proven on a live pull request yet.

---

## OpenTelemetry export

`glassops run` can send a finished run to an OpenTelemetry collector as trace spans, so a pipeline shows up in the tracing tool you already use. It is off until you set an endpoint.

```yaml
telemetry:
  otlp:
    endpoint: https://otel.example.com
    headers:
      authorization: $OTEL_AUTH_HEADER
    service_name: glassops
```

| Field | Meaning |
|---|---|
| `endpoint` | Where to send spans, over OTLP with the JSON encoding on HTTP. Setting it turns export on. A URL with no path gets `/v1/traces` added |
| `headers` | Headers to send, such as a token. Each value must be a whole `$VAR` reference, never the value itself. A header whose variable is not set is an error |
| `service_name` | The `service.name` in the spans. The default is `glassops` |

No SDK is involved. GlassOps posts the JSON itself.

### What is sent

One run is one trace. There is a root span named `glassops run`, and one child span per step named `glassops <step>`, such as `glassops gate`. A span whose command exited 1 or 2 has the error status.

Only these attributes are ever sent:

| Span | Attributes |
|---|---|
| Root | `glassops.duration_ms`, `glassops.pipeline`, `glassops.provider`, `glassops.exit_code`, `glassops.empty_plan`, `glassops.cancelled` |
| Step | `glassops.step`, `glassops.step.index`, `glassops.step.when`, `glassops.duration_ms`, `glassops.step.skipped`, `glassops.step.outcome` (`pass`, `fail`, `error` or `skipped`), `glassops.exit_code` when the step ran, and `glassops.target.name` and `glassops.target.tier` when the step names a target that is in your config |

The resource carries `service.name` and nothing else. Fact values, logins, org Ids, error text and command output are never copied into a span. Target names and tiers are sent, so do not put anything private in a target name.

### Safety rules

- Headers are sent only over `https`. Plain `http` is accepted for a loopback host such as `localhost`, so you can test against a local collector. Headers over plain `http` to any other host are refused.
- The endpoint must not hold credentials. Use `headers`.
- An export waits up to 10 seconds.
- A failed export adds a warning annotation, for example `run: telemetry export failed`. It never changes the exit code, and it runs after the run record is written.
- Every `$VAR` in `headers` is covered by the `glassops doctor` environment check, like the variables of a target.

### What is tested for the export

The output was sent to `otel/opentelemetry-collector` 0.140.0 with an OTLP HTTP receiver and a file exporter, after a three step pipeline whose second step failed. The collector accepted it without a warning and wrote one trace with the spans described above. No export has been sent from a CI job to a hosted tracing service.
