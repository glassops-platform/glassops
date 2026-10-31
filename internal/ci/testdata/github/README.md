# GitHub API fixtures

These files are real responses from the GitHub REST API for this repository. They were
scrubbed before they were committed. The approval comment in `run-approvals.json` was replaced
during scrubbing. Do not edit them by hand to suit a test. When the API
changes its output, capture it again.

| File | Endpoint | What it shows |
|---|---|---|
| `commit-pulls.json` | `GET /repos/{owner}/{repo}/commits/{sha}/pulls` | the pull request a merge commit came from, with its author |
| `commit-pulls-labelled.json` | the same endpoint for a commit whose pull request has labels | `labels` on each pull request |
| `commit-pulls-unknown.json` | the same endpoint for a SHA the repository does not have | HTTP 422, "No commit found for SHA". The body carries `status` as the string "422" |
| `pull.json` | `GET /repos/{owner}/{repo}/pulls/{number}` | a merged pull request. `merged_by` names who merged it |
| `pull-reviews.json` | `GET /repos/{owner}/{repo}/pulls/{number}/reviews` | a pull request with no reviews, an empty array |
| `issue-labels.json` | `GET /repos/{owner}/{repo}/issues/{number}/labels` | a pull request with no labels, an empty array |
| `deployments.json` | `GET /repos/{owner}/{repo}/deployments` | a deployment to the `glassops-deploy` environment |
| `run-approvals.json` | `GET /repos/{owner}/{repo}/actions/runs/{run_id}/approvals` | an approved environment review: state, approver login, comment and the environments it covered. Readable by a job's `GITHUB_TOKEN` with `actions: read` |
| `issue-events-labelled.json` | `GET /repos/{owner}/{repo}/issues/{number}/events` | `labeled` events, each with `label.name` and the `actor` who added it |
