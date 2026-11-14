# Example GlassOps Rego rules. gate evaluates data.glassops.deny over the facts document
# (.glassops/facts.json), after the built-in rules. Each entry denies the deploy and becomes a
# failed check named rego.<rule> in the PERMIT.
package glassops

import rego.v1

# Components no pipeline may change in production. Edit these lists for your org.
protected_profiles := {"Admin"}

protected_permission_sets := {"Platform_Admin"}

facts := input.facts

production if facts["target.tier"] == "prod"

deny contains {
	"rule": "protected_profile",
	"detail": sprintf("the plan changes the protected profile %s, which production does not accept from a pipeline", [name]),
} if {
	production
	some name in facts["salesforce.profiles"]
	name in protected_profiles
}

deny contains {
	"rule": "protected_permission_set",
	"detail": sprintf("the plan changes the protected permission set %s, which production does not accept from a pipeline", [name]),
} if {
	production
	some name in facts["salesforce.permission_sets"]
	name in protected_permission_sets
}
