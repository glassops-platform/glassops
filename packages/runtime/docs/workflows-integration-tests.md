---
type: Documentation
domain: runtime
origin: packages/runtime/.github/workflows/integration-tests.yml
last_modified: 2026-01-26
generated: true
source: packages/runtime/.github/workflows/integration-tests.yml
generated_at: 2026-01-26T05:13:04.759Z
hash: 2b0d418994e7b62476bca886521cd6306fa41b81b95b393c3d73a25cb3345f4a
---

# Integration Tests Workflow Documentation

This document details the `integration-tests.yml` workflow file, which automates the integration testing process for the project.

## Purpose

The primary goal of this workflow is to ensure the quality and stability of the project's code by running integration tests on various environments. It covers unit and integration tests, generates coverage reports, and performs end-to-end testing of an action.

## Workflow Triggers

The workflow is triggered by the following events:

*   **`push` to `main` branch:** When code is pushed to the `main` branch, the workflow runs if the changes are in the `src`, `config`, `package.json`, `tsconfig.json`, or `.github/workflows/integration-tests.yml` directories.
*   **`pull_request` to `main` branch:** When a pull request is created or updated targeting the `main` branch.
*   **`workflow_dispatch`:** Allows manual triggering of the workflow with an optional `debug` input.

## Jobs

The workflow consists of the following jobs:

### 1. `integration-tests`

*   **Name:** Integration Test Suite
*   **Runs on:** `ubuntu-latest`
*   **Purpose:** Executes the core integration tests.
*   **Steps:**
    *   **Checkout Repository:** Checks out the project's code.
    *   **Setup Node.js:** Sets up the Node.js environment with version 20 and enables npm caching.
    *   **Install Dependencies:** Installs project dependencies using `npm ci`.
    *   **Build Project:** Builds the project using `npm run build`.
    *   **Run Integration Tests:** Executes the integration tests using `npm run test:integration`.  Sets environment variables for access within the tests.
    *   **Upload Integration Test Coverage:** Uploads the integration test coverage report from the `coverage/integration/` directory as a workflow artifact.  The artifact is retained for 14 days.

### 2. `combined-coverage`

*   **Name:** Combined Coverage Report
*   **Runs on:** `ubuntu-latest`
*   **Needs:** `integration-tests` (This job depends on the successful completion of the `integration-tests` job.)
*   **Purpose:** Generates a combined coverage summary report from both unit and integration tests.
*   **Steps:**
    *   **Checkout Repository:** Checks out the project's code.
    *   **Setup Node.js:** Sets up the Node.js environment with version 20 and enables npm caching.
    *   **Install Dependencies:** Installs project dependencies using `npm ci`.
    *   **Run All Tests with Coverage:** Runs both unit tests (`npm test`) and integration tests (`npm run test:integration`) to generate coverage data.
    *   **Generate Combined Coverage Summary:** Creates a summary report in the GitHub step output using `jq` to parse the `coverage-summary.json` files for unit and integration tests. The report is formatted as Markdown and includes percentage coverage for statements, branches, functions, and lines.

### 3. `e2e-action-test`

*   **Name:** E2E Action Test
*   **Runs on:** `ubuntu-latest`
*   **Conditional Execution:** Runs only if one of the following conditions is met:
    *   The workflow is triggered manually (`workflow_dispatch`).
    *   The workflow is triggered by a `push` event to the `main` branch.
    *   The workflow is triggered by a `pull_request` event.
*   **Purpose:** Performs an end-to-end test of the project's action.
*   **Steps:**
    *   **Checkout Repository:** Checks out the project's code.
    *   **Setup Node.js:** Sets up the Node.js environment with version 20 and enables npm caching.
    *   **Install Dependencies:** Installs project dependencies using `npm ci`.
    *   **Build Action:** Builds the project's action using `npm run build`.
    *   **Create Test Configuration:** Creates a `devops-config.json` file with a predefined configuration for the action test, including governance settings and runtime parameters.
    *   **Test Action Execution (Dry Run):** Executes the action in dry-run mode using the `./` path (assuming the action is in the current directory).  It provides test credentials and sets the `DEBUG` environment variable based on the `inputs.debug` value. `continue-on-error: true` allows the workflow to continue even if the action fails (as it's expected to fail in dry-run mode without valid credentials).
    *   **Verify Action Outputs:**  Analyzes the outcome of the action test and logs the results to the GitHub step output.  It checks if the action executed successfully and displays the runtime ID if available.  It also includes the `devops-config.json` file used for the test.

### 4. `matrix-integration`

*   **Name:** Matrix Integration (${{ matrix.os }} / Node ${{ matrix.node }})
*   **Runs on:**  Dynamically determined based on the matrix strategy.
*   **Strategy:** Uses a matrix strategy to run the integration tests on multiple operating systems and Node.js versions.
    *   **`os`:** `ubuntu-latest`, `windows-latest`, `macos-latest`
    *   **`node`:** `"20"`, `"22"`
*   **Purpose:**  Ensures compatibility across different environments.
*   **Steps:**
    *   **Checkout Repository:** Checks out the project's code.
    *   **Setup Node.js:** Sets up the Node.js environment with the version specified in the matrix and enables npm caching.
    *   **Install Dependencies:** Installs project dependencies using `npm ci`.
    *   **Build Project:** Builds the project using `npm run build`.
    *   **Run Integration Tests:** Executes the integration tests using `npm run test:integration`. Sets environment variables for access within the tests.

## Inputs

The `workflow_dispatch` event accepts the following input:

*   **`debug`:** (boolean, optional, default: `false`) Enables debug logging if set to `true`.

## Artifacts

The workflow produces the following artifacts:

*   **`integration-coverage-report`:** Contains the integration test coverage report from the `coverage/integration/` directory.  Retained for 14 days.