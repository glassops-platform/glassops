# Rego rules

> **Opt-in, and not in a published release yet.** This is planned for v1.1.0. Without `policy.rego` in `glassops.yaml`, `gate` runs only its built-in rules. See [what is in which release](releases-and-licenses.md#what-is-in-which-release).

The built-in rules in [PERMIT and rules](permit-and-rules.md#the-v1-rules) cover the common cases. When you need a rule of your own, `gate` can run [Rego](https://www.openpolicyagent.org/docs/latest/policy-language/) policies after them. A Rego rule can only add a denial. It can never turn a failed built-in rule into a pass.

GlassOps does not evaluate Rego itself. It runs the separate program `opa`, which is in both GlassOps images (see [Releases, images and licenses](releases-and-licenses.md#tools-in-the-images)). A working example with tests is in [`examples/rego`](https://github.com/glassops-platform/glassops/tree/main/examples/rego).

---

## Turn it on

List the Rego files in `glassops.yaml`. Relative paths are resolved against the project directory.

```yaml
policy:
  rego:
    files: [policy/deploy.rego]
```

`gate` then runs this for the target it checks, after the built-in rules:

```bash
opa eval --format json --input <facts file> --data policy/deploy.rego data.glassops.deny
```

An ephemeral target gets the automatic PERMIT and its rules are not run, so Rego is not run for it either.

---

## What your policy sees

The input is the facts document that `gate` builds and also writes to `.glassops/facts.json`:

```json
{
  "schema": "glassops.facts/v1",
  "facts": { "target.tier": "prod", "salesforce.profiles": ["Admin"] }
}
```

These are the same facts the built-in rules read, such as `target.tier`, `ci.event`, `ci.actor`, `ci.labels`, `change.destructive_count` and the component lists the provider records, such as `salesforce.apex_classes`, `salesforce.permission_sets` and `salesforce.profiles`. Run `gate` once and read `.glassops/facts.json` to see what your project produces.

The facts come from `gate` alone. A result file or a change set file cannot add a `ci.`, `target.` or `change.` fact, so a policy can trust them.

---

## Write a rule

Declare `package glassops` and add entries to the `deny` set. Each entry is an object with a `rule` name and a `detail` text.

```rego
package glassops

import rego.v1

deny contains {"rule": "protected_profile", "detail": "the plan changes the Admin profile"} if {
	input.facts["target.tier"] == "prod"
	"Admin" in input.facts["salesforce.profiles"]
}
```

Each entry becomes a failed check named `rego.<rule>` in the PERMIT, and any entry denies the deploy. `gate` exits 1 for a denial. When the set is empty, the PERMIT gets one passing check named `rego`, whose detail names the files and the first 12 characters of a SHA-256 over their names and contents. That records which policy was in force.

A rule name starts with a letter or digit and holds only letters, digits, `_`, `.` and `-`, up to 100 characters. A detail is cut at 500 bytes, and email addresses in it are masked.

---

## It fails closed

Rego never fails open. When `policy.rego` is set and any of these happens, `gate` still signs and writes a PERMIT, the PERMIT denies, and `gate` exits 2:

- `opa` is not on `PATH`.
- A file is missing or cannot be read.
- `opa` reports an error or runs longer than 2 minutes.
- `data.glassops.deny` is undefined or is not a set. A file with the wrong package name leaves it undefined, so it fails instead of passing quietly.
- An entry has no valid `rule`.

---

## Check it before `gate`

`glassops doctor` checks, when `policy.rego` is set, that `opa` is installed (`tool.opa`) and that the files load and define `data.glassops.deny` as a set (`policy.rego`).

To try your rules over made-up facts, run `opa` yourself. Name the files, because passing a whole folder also loads the sample inputs as data:

```bash
opa test examples/rego/deploy.rego examples/rego/deploy_test.rego
opa eval --format json --input examples/rego/facts-deny.json --data examples/rego/deploy.rego data.glassops.deny
opa eval --format json --input examples/rego/facts-pass.json --data examples/rego/deploy.rego data.glassops.deny
```

---

## Tested here, and not

- The evaluator is covered by unit tests with a fake `opa`. The example policy and its tests also run against a real `opa` 1.22.0, the version in the images, in a test that runs when `opa` is on `PATH`.
- A full `gate` run in a workflow on GitHub has not been proven yet.
