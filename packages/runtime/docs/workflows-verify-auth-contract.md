---
type: Documentation
domain: runtime
origin: packages/runtime/.github/workflows/verify-auth-contract.yml
last_modified: 2026-01-26
generated: true
source: packages/runtime/.github/workflows/verify-auth-contract.yml
generated_at: 2026-01-26T05:14:03.555Z
hash: fa2c79011a5bc105a5f27bd76dd3f33204428a9a1d6c4f1412ad71a8d86496dd
---

# Auth Contract Verification Workflow Documentation

This document details the `verify-auth-contract.yml` workflow file, which automates the verification of the authentication contract for the GlassOps Runtime.

## Purpose

The primary goal of this workflow is to ensure that the GlassOps Runtime is correctly initialized and configured with valid authentication credentials and organizational information. It verifies the output of the runtime initialization process, specifically checking for a valid organization ID and a readiness signal.  This workflow is triggered on pushes and pull requests to the `main` branch, and can also be manually triggered.

## Structure

The workflow consists of a single job named `verify-identity-contract`. This job is executed on an Ubuntu-latest runner.

### `name: Auth Contract Verification`

*   **Purpose:** Defines the name of the workflow as displayed in the GitHub Actions interface.

### `on:`

*   **Purpose:** Specifies the events that trigger the workflow.
    *   `push:`
        *   `branches: [main]` - Triggers the workflow on pushes to the `main` branch.
    *   `pull_request:`
        *   `branches: [main]` - Triggers the workflow on pull requests targeting the `main` branch.
    *   `workflow_dispatch:` - Allows manual triggering of the workflow.

### `jobs:`

*   **Purpose:** Defines the jobs to be executed in the workflow.

    #### `verify-identity-contract:`

    *   **`name: Verify Identity & Output Contract`** -  Defines the name of the job.
    *   **`runs-on: ubuntu-latest`** - Specifies that the job should run on a virtual machine with the latest Ubuntu operating system.
    *   **`if: github.repository_owner == 'glassops-platform' && github.actor != 'dependabot[bot]'`** -  A conditional statement that determines whether the job should run. It only runs if the repository owner is `glassops-platform` and the actor is not the `dependabot[bot]` user.

    ##### `steps:`

    *   **Purpose:** Defines the sequence of steps to be executed within the job.

        *   **`- name: Checkout`**
            *   `uses: actions/checkout@v6` - Checks out the repository code to the runner.
        *   **`- name: Initialize GlassOps Runtime™`**
            *   `id: runtime` - Assigns an ID to this step, allowing its outputs to be referenced in subsequent steps.
            *   `uses: ./` - Uses a custom action defined within the repository (presumably the GlassOps Runtime initialization action).
            *   `with:`
                *   `jwt_key: ${{ secrets.SF_JWT_KEY }}` - Passes the JWT key as a secret to the action.
                *   `client_id: ${{ secrets.SF_CLIENT_ID }}` - Passes the client ID as a secret to the action.
                *   `username: ${{ vars.SF_USERNAME }}` - Passes the username as a variable to the action.
                *   `enforce_policy: "false"` -  Disables policy enforcement during runtime initialization.
        *   **`- name: Validate Contract Primitives`**
            *   `run: |` - Executes a multi-line shell script.
                *   The script verifies the `org_id` and `glassops_ready` outputs from the `runtime` step.
                *   It checks if the `org_id` starts with "00D".
                *   It checks if `glassops_ready` is equal to "true".
                *   If either check fails, the script exits with an error code (1).
                *   If both checks pass, it prints a success message.
                *   It also generates a step summary in Markdown format, displaying the verification results.
        *   **`- name: Audit Environment Setup`**
            *   `run: |` - Executes a multi-line shell script.
                *   The script verifies the Salesforce CLI (sf) environment.
                *   It attempts to display the org ID using `sf org display --json` and pipes the output to `jq` to extract the ID. The output is redirected to `/dev/null` to suppress it.
                *   If the command succeeds, it prints a success message indicating that the CLI is authenticated and functional.