## glassops report: error

| Step | Status | Target | Detail |
|---|---|---|---|
| doctor | pass |  | 2 checks |
| plan | pass |  | 1 check |
| validate | fail | qa | problem.1: ApexClass AccountService: Variable does not exist: acct (line 14) |
| check | pass |  | 1 check |
| test | skipped | qa |  |
| gate | fail | prod | window.prod: outside the deploy window |
| deploy | error | prod | deploy: the target stopped answering for u***@example.com |

### validate: fail

- problem.1 (fail): ApexClass AccountService: Variable does not exist: acct (line 14)
- problem.2 (fail): force-app/main/default/classes/Util.cls:27 Unexpected token \| here
- problem.3 (fail): force-app\main\default\triggers\T.trigger: missing semicolon

### gate: fail

- window.prod (fail): outside the deploy window

### deploy: error

- deploy (error): the target stopped answering for u***@example.com
- error: poll timed out

### PERMIT: deny

Target: prod (production). Issued 2026-10-30 09:05Z, expires 2026-10-30 11:05Z.

| Rule | Outcome | Detail |
|---|---|---|
| plan_not_empty | pass | 3 components |
| window.prod | fail | outside the deploy window |
| separation_of_duties | satisfied_by_approval | approved |

Separation of duties: satisfied_by_approval by environment, approver reviewer1.

### Deploy

Deploy Id: `0AfXX0000000001AAA`

