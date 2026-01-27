---
type: Documentation
domain: runtime
origin: packages/runtime/.github/workflows/verify-governance.yml
last_modified: 2026-01-26
generated: true
source: packages/runtime/.github/workflows/verify-governance.yml
generated_at: 2026-01-26T05:14:23.849Z
hash: 00cf212c64c3a332abb4f72980bb4e921d397498a26cd4be7137b333752099f5
---

# Governance Validation Workflow Documentation

This workflow validates the governance and freeze window enforcement logic of the runtime component. It consists of two jobs: `test-freeze-logic` and `test-bypass-logic`.

## Workflow Triggers

The workflow is triggered on:

*   **Push to `main` branch:**  Executes when code is pushed to the `main` branch.
*   **Pull Request to `main` branch:** Executes when a pull request is opened or updated targeting the `main` branch.
*   **Workflow Dispatch:** Allows manual triggering of the workflow.

## Jobs

### 1. `test-freeze-logic` - Verify Freeze Window Enforcement

This job verifies that the runtime component correctly blocks deployments during an active freeze window.

*   **`runs-on`:** `ubuntu-latest` - Specifies that the job runs on a virtual machine with the latest Ubuntu operating system.
*   **Steps:**
    *   **`Checkout`:** Uses the `actions/checkout@v6` action to check out the repository code.
    *   **`Generate Active Freeze Policy`:** Creates a `devops-config.json` file containing a governance policy that enables governance and defines a freeze window for the current day, spanning the entire day (00:00 - 23:59).  The `DAY` variable is dynamically set to the current day of the week.
    *   **`Attempt Governed Run (Expected to Fail)`:** Executes the runtime component using the `./` action (assumes the runtime component is in the same repository).  It passes mock values for `jwt_key`, `client_id`, and `username`.  Crucially, `enforce_policy` is set to `"true"`, meaning the runtime should enforce the governance policy. `continue-on-error: true` allows the workflow to continue even if this step fails (as it's *expected* to fail during the freeze window).
    *   **`Validate Blocked Outcome`:**  Checks the outcome of the `governed-run` step. If the step failed (as expected), it logs a success message indicating that the runtime correctly blocked the deployment. If the step succeeded, it logs a failure message and exits with an error code (1).  The results are written to the GitHub Step Summary.

### 2. `test-bypass-logic` - Verify Manual Bypass

This job verifies that the runtime component can be bypassed when `enforce_policy` is set to `"false"`.

*   **`runs-on`:** `ubuntu-latest` - Specifies that the job runs on a virtual machine with the latest Ubuntu operating system.
*   **Steps:**
    *   **`Checkout`:** Uses the `actions/checkout@v6` action to check out the repository code.
    *   **`Generate Active Freeze Policy`:** Creates a `devops-config.json` file, identical to the one in `test-freeze-logic`, enabling governance and defining a freeze window for the current day.
    *   **`Attempt Bypassed Run`:** Executes the runtime component with `enforce_policy` set to `"false"` and `skip_auth` set to `"true"`. This simulates a manual bypass of the governance policy. `continue-on-error: true` allows the workflow to continue even if this step fails.
    *   **`Verify Bypass`:** Checks the outcome of the `test-bypass` step. If the step succeeded, it logs a success message indicating that the policy was successfully bypassed. If the step failed, it logs a failure message, includes the outcome, and exits with an error code (1). The results are written to the GitHub Step Summary. `if: always()` ensures this step runs regardless of the previous step's outcome.

## `devops-config.json` Structure

The `devops-config.json` file used in both jobs has the following structure:

```json
{
  "governance": {
    "enabled": true,
    "freeze_windows": [
      { "day": "DAY_OF_WEEK", "start": "HH:MM", "end": "HH:MM" }
    ]
  },
  "runtime": { "cli_version": "latest" }
}
```

*   **`governance`:**  An object containing governance-related settings.
    *   **`enabled`:** A boolean value indicating whether governance is enabled.
    *   **`freeze_windows`:** An array of objects, each defining a freeze window.
        *   **`day`:** The day of the week the freeze window applies to (e.g., "Monday", "Tuesday").
        *   **`start`:** The start time of the freeze window in HH:MM format (e.g., "00:00").
        *   **`end`:** The end time of the freeze window in HH:MM format (e.g., "23:59").
*   **`runtime`:** An object containing runtime-related settings.
    *   **`cli_version`:** The desired CLI version for the runtime.