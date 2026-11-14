# Tests for deploy.rego. Run them with: opa test examples/rego
package glassops_test

import rego.v1

import data.glassops

prod_facts(profiles, permission_sets) := {"facts": {
	"target.tier": "prod",
	"salesforce.profiles": profiles,
	"salesforce.permission_sets": permission_sets,
}}

test_protected_profile_denies_in_prod if {
	denials := glassops.deny with input as prod_facts(["Admin"], [])
	{d.rule | some d in denials} == {"protected_profile"}
}

test_protected_permission_set_denies_in_prod if {
	denials := glassops.deny with input as prod_facts([], ["Platform_Admin"])
	{d.rule | some d in denials} == {"protected_permission_set"}
}

test_other_components_pass if {
	count(glassops.deny) == 0 with input as prod_facts(["Sales_User"], ["Greeter_User"])
}

test_protected_profile_passes_outside_prod if {
	count(glassops.deny) == 0 with input as {"facts": {
		"target.tier": "uat",
		"salesforce.profiles": ["Admin"],
	}}
}

test_no_input_gives_an_empty_set if {
	count(glassops.deny) == 0
}
