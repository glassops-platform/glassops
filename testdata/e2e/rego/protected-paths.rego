# The Rego rule of the OPA end-to-end run (scripts/e2e/opa-e2e.sh). It does what
# examples/rego/deploy.rego does for profiles, for a path in an exec project: it denies a plan
# that changes a file under a protected path. The exec.plan command reports the changed files
# as the fact exec.changed_files. Each entry becomes a failed check named rego.<rule>.
package glassops

import rego.v1

protected_path_prefixes := {"app/protected/"}

deny contains {
	"rule": "protected_path",
	"detail": sprintf("the plan changes %s, which is under a protected path", [file]),
} if {
	some file in input.facts["exec.changed_files"]
	some prefix in protected_path_prefixes
	startswith(file, prefix)
}
