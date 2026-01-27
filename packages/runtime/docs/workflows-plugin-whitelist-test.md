---
type: Documentation
domain: runtime
origin: packages/runtime/.github/workflows/plugin-whitelist-test.yml
last_modified: 2026-01-26
generated: true
source: packages/runtime/.github/workflows/plugin-whitelist-test.yml
generated_at: 2026-01-26T05:13:25.490Z
hash: f523841210feb49671d5338e1daf7ad18d254158769bb1369f62911ef8438c3a
---

## Plugin Whitelist Governance Test Workflow Documentation

This workflow validates the plugin whitelist governance feature. It tests scenarios involving whitelisted plugins, non-whitelisted plugins, missing whitelist configurations, version constraints, and the blocking of specific plugins (like the scanner).

### Workflow Triggers

The workflow is triggered by:

*   **`push` events:** When code is pushed to the `main` branch, specifically changes in the `src/` directory, the `action.yml` file, or the workflow file itself (`.github/workflows/plugin-whitelist-test.yml`).
*   **`pull_request` events:** When a pull request is opened against the `main` branch.
*   **`workflow_dispatch` events:**  Allows manual triggering of the workflow.

### Jobs

The workflow consists of five jobs:

1.  **`test-whitelisted-plugin`**: Tests the successful installation of a plugin that *is* included in the whitelist.
2.  **`test-non-whitelisted-plugin`**: Tests that the workflow correctly rejects the installation of a plugin that is *not* in the whitelist.
3.  **`test-no-whitelist`**: Tests the behavior when no plugin whitelist is configured.
4.  **`test-version-constraints`**: Tests the enforcement of version constraints specified in the whitelist.
5.  **`test-scanner-blocked`**: Specifically tests that the deprecated `@salesforce/sfdx-scanner` plugin is blocked, even if governance is enabled.

### Job Structure (Common Elements)

Each job shares a common structure:

*   **`name`**: A descriptive name for the job.
*   **`runs-on`**: Specifies the virtual machine environment to use (`ubuntu-latest`).
*   **`steps`**: A sequence of steps to execute.  These steps generally include:
    *   **`Checkout`**: Checks out the repository code using `actions/checkout@v6`.
    *   **`Setup Node.js`**: Sets up a Node.js environment using `actions/setup-node@v6`, specifying version 20 and enabling npm caching.
    *   **`Install Dependencies`**: Installs project dependencies using `npm ci`.
    *   **`Build Action`**: Builds the action using `npm run build`.
    *   **`Create Test Config with Plugin Whitelist`**: Creates a `devops-config.json` file containing the governance configuration, including the `plugin_whitelist`. The content of this file varies between jobs to test different scenarios.
    *   **`Test Whitelisted/Non-Whitelisted Plugin Installation`**: Executes the action (`./`) with specific input parameters to simulate plugin installation.  Key parameters include:
        *   `jwt_key`: A mock JWT key for authentication.
        *   `client_id`: A mock client ID.
        *   `username`: A mock username.
        *   `plugins`: The name of the plugin to install (either whitelisted or non-whitelisted).
        *   `enforce_policy`: Set to `"false"` to allow testing of the rejection logic without actually enforcing the policy.
        *   `skip_auth`: Set to `"true"` to bypass authentication.
    *   **`Verify Plugin Was Rejected` (in `test-non-whitelisted-plugin`):**  Checks the outcome of the previous step (which is expected to fail) and reports success or failure based on whether the non-whitelisted plugin was correctly rejected.  The result is written to the GitHub Step Summary.
    *   **`Validate Scanner Blocked` (in `test-scanner-blocked`):** Checks the outcome of the scanner installation attempt and validates that the scanner was blocked as expected. The result is written to the GitHub Step Summary.

### `devops-config.json` Configuration

The `devops-config.json` file is dynamically created in each job and defines the governance policy.  The relevant section is:

```json
{
  "governance": {
    "enabled": true,
    "plugin_whitelist": ["plugin1", "plugin2@version"]
  },
  "runtime": {
    "cli_version": "latest"
  }
}
```

*   **`governance.enabled`**: A boolean value indicating whether governance is enabled.  Set to `true` in all jobs.
*   **`governance.plugin_whitelist`**: An array of strings representing the whitelisted plugins.  Each string can include a version constraint using semantic versioning (e.g., `"sfdx-hardis@^6.0.0"`).  If this array is empty or missing, no whitelist is enforced.
*   **`runtime.cli_version`**: Specifies the CLI version to use. Set to `"latest"` in all jobs.

### Key Takeaways

This workflow provides comprehensive testing of the plugin whitelist governance feature, ensuring that:

*   Whitelisted plugins are installed successfully.
*   Non-whitelisted plugins are rejected.
*   The workflow behaves correctly when no whitelist is configured.
*   Version constraints are enforced.
*   Specific, deprecated plugins (like the scanner) are blocked.