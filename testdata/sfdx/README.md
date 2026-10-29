# Test Salesforce project

A small SFDX project the Phase 2 tests build a two-commit git repo from. It sits under
`testdata/` so `init` and project detection never see it as the repo's own project.

`sh testdata/sfdx/build-repo.sh DEST` creates `DEST` as a git repo on `main` with two commits
and no `glassops.yaml`.

| Commit | Content | What it exercises |
|---|---|---|
| 1 (`commit1/`) | `sfdx-project.json`, `.forceignore`, a scratch definition, Apex class `Greeter` with test `GreeterTest`, custom field `Account.Tier__c`, permission set `Greeter_User` with class access and edit on the field | A full plan, and the first delta from an empty tree |
| 2 (`commit2/`, `commit2.remove`) | `Greeter` and `GreeterTest` change, `Greeter_User` loses the field permission, the field file is deleted | A delta with changed Apex classes, a touched permission set and one destructive member, `CustomField Account.Tier__c` |

The delta from commit 1 to commit 2 is `ApexClass` Greeter and GreeterTest, `PermissionSet`
Greeter_User, and a destructive `CustomField` `Account.Tier__c`.

Commit 2 holds only the changed files. `commit2.remove` lists the paths to delete, one per line.
