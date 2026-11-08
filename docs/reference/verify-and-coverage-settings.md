# Verify, coverage and YAML settings

Settings in `glassops.yaml` for the post-deploy `verify` step, for the class coverage check, and the limits the file must stay inside. The schema is [`schemas/config.json`](https://github.com/glassops-platform/glassops/blob/main/schemas/config.json). The exec provider's own settings are on [the exec provider page](exec-provider.md).

## `verify`

All keys are optional. See [`glassops verify`](cli.md#glassops-verify) for what each check does.

```yaml
verify:
  smoke: ./ci/smoke.sh
  foreign_deploys: warn
  flows: fail
```

| Key | Values | Default | Meaning |
|---|---|---|---|
| `verify.smoke` | a shell command | none | Runs after verify's own checks, with `sh -c` in the project directory. Exit 0 passes and any other exit fails |
| `verify.foreign_deploys` | `fail`, `warn`, `off` | `warn` | What another deploy to the target during your deploy window does |
| `verify.flows` | `fail`, `warn`, `off` | `fail` | What an active Flow that the target does not run as its latest version does |

The smoke command gets the usual environment plus `GLASSOPS_RESULTS_DIR`, `GLASSOPS_TARGET` and `GLASSOPS_TARGET_ID`. A command that cannot start is an error.

## `test.coverage.class_scope`

`test.coverage.class_min` fails the test step when any class is below the minimum. `sf apex run test` lists only the classes a test touched, so a class that no test touched is missing from its answer, and without more help it would never be checked. After a finished test run, GlassOps asks the target for its coverage aggregate and counts the untouched classes at 0%. `class_scope` says which untouched classes count.

```yaml
test:
  coverage:
    class_min: 75
    class_scope: changed
```

| Value | Classes counted at 0% when no test touched them |
|---|---|
| `changed` (default) | The Apex classes in `.glassops/plan.json`. With no plan file, every untouched class the target reports, because nothing says which classes changed |
| `all` | The classes in the plan and every untouched class the target reports |

Rules that apply to both:

- A class whose name ends in `Test` is taken for a test class and skipped.
- The classes added this way are recorded in the fact `test.coverage_unreported`. The provider records the untouched classes it found as `test.coverage_untouched`.
- If the coverage aggregate cannot be read, the step fails instead of passing on a missing list. The fact `test.coverage_unavailable` holds the reason.
- A plan file that exists and cannot be read fails the step.
- With `class_min` or `org_min` set and no Apex tests found, the step fails, because coverage cannot be checked.

## YAML limits

`glassops.yaml` is read as YAML with anchors and aliases allowed, within limits that stop a hostile file from crashing the tool:

- An alias that refers to itself, directly or through other anchors, is refused. `&a [*a]` is an error with a line number.
- The file may not nest deeper than 64 levels.
- After aliases are expanded, the file may not hold more than 100000 values. An anchor used many times over many levels can reach that, and it is refused instead of expanded.
- An alias that names no anchor is refused.

A real config is a few hundred values, so none of these limits is reachable by accident. Each refusal is a config error with a message, and the command that loaded the config does not go on.
