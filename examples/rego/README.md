# Rego rules example

GlassOps can run your own Rego rules after its built-in rules. This folder holds a small policy
that refuses production deploys of protected profiles and permission sets. Rego rules are
opt-in and need the `opa` program on `PATH`.

## Turn it on

List the Rego files in `glassops.yaml`. Paths are relative to the project directory.

```yaml
policy:
  rego:
    files: [policy/deploy.rego]
```

`gate` then runs this for the target it checks:

```
opa eval --format json --input <facts> --data policy/deploy.rego data.glassops.deny
```

The input is the facts document that `gate` also writes to `.glassops/facts.json`:

```json
{
  "schema": "glassops.facts/v1",
  "facts": { "target.tier": "prod", "salesforce.profiles": ["Admin"] }
}
```

## Write a rule

Declare `package glassops` and add entries to the `deny` set. Each entry is an object with a
`rule` name and a `detail` text.

```rego
deny contains {"rule": "protected_profile", "detail": "the plan changes the Admin profile"} if {
	input.facts["target.tier"] == "prod"
	"Admin" in input.facts["salesforce.profiles"]
}
```

Each entry becomes a failed check named `rego.<rule>` in the PERMIT, and any entry denies the
deploy. When the set is empty, the PERMIT gets one passing check named `rego` that names the
files and a hash of their contents.

A rule name starts with a letter or digit and holds only letters, digits, `_`, `.` and `-`, up
to 100 characters. A detail is cut at 500 bytes, and email addresses in it are masked.

## What Rego cannot do

- The built-in rules always run first and their results never change. Rego can only add
  denials. It cannot turn a failed built-in rule into a pass.
- `gate` fails closed. When `policy.rego` is set and `opa` is missing, a file is missing, `opa`
  reports an error, `data.glassops.deny` is undefined or not a set, or an entry has no valid
  `rule`, `gate` writes a PERMIT that denies and exits 2.
- A file with the wrong package name leaves `data.glassops.deny` undefined, so it fails instead
  of passing quietly.

## Check it before gate

`glassops doctor` checks that `opa` is installed and that the files load and define
`data.glassops.deny` as a set.

Run the tests and try the rules over the two sample inputs, which hold made-up facts:

```
opa test examples/rego/deploy.rego examples/rego/deploy_test.rego
opa eval --format json --input examples/rego/facts-deny.json --data examples/rego/deploy.rego data.glassops.deny
opa eval --format json --input examples/rego/facts-pass.json --data examples/rego/deploy.rego data.glassops.deny
```

Name the files for `opa test`. Passing the whole folder also loads the sample inputs as data,
and `opa` refuses the two because their values conflict.

The facts come from your plan, so a rule can look at the component lists the provider records,
such as `salesforce.apex_classes`, `salesforce.permission_sets` and `salesforce.profiles`, and
at the target and CI facts such as `target.tier`, `ci.event` and `ci.labels`.
