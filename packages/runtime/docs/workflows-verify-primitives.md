---
type: Documentation
domain: runtime
origin: packages/runtime/.github/workflows/verify-primitives.yml
last_modified: 2026-01-26
generated: true
source: packages/runtime/.github/workflows/verify-primitives.yml
generated_at: 2026-01-26T05:14:41.172Z
hash: 4482a192ff89ad3a853944510034151ce386cb5c22f46fe1ee9bf6d4568aaf87
---

# Primitive Logic Verification Workflow Documentation

## Purpose

This workflow automates the verification of primitive logic within the project. It performs unit tests, code hygiene checks (ESLint and Prettier), dependency audits, and dependency reviews on code changes.  The workflow is triggered by pushes to the `main` branch, pull requests targeting `main`, and manual dispatch.

## Structure

The workflow consists of three main jobs: `unit-tests`, `code-hygiene`, and `dependency-review`. Each job runs on a specific environment and performs a set of steps.

### 1. `unit-tests` Job

*   **Name:** `Run Logic Unit Tests (${{ matrix.os }})`
*   **`runs-on`:**  `${{ matrix.os }}` - Dynamically selects the operating system based on the matrix.
*   **`strategy`:** Defines a matrix of configurations for running the tests across different environments.
    *   **`matrix.os`:**  `[ubuntu-latest, windows-latest, macos-latest]` - Runs tests on Ubuntu, Windows, and macOS.
    *   **`matrix.node-version`:** `["18", "20"]` - Runs tests with Node.js versions 18 and 20.
    *   **`exclude`:**  Excludes specific combinations from the matrix. In this case, it excludes running Node.js 18 on Ubuntu.
*   **Steps:**
    *   **`Checkout`:** Uses `actions/checkout@v6` to checkout the repository code.
    *   **`Setup Node.js ${{ matrix.node-version }}`:** Uses `actions/setup-node@v6` to set up the specified Node.js version.  `cache: "npm"` enables caching of npm dependencies for faster builds.
    *   **`Install Dependencies`:** Runs `npm ci` to install dependencies from `package-lock.json`.
    *   **`Run Jest Suite`:** Executes the Jest test suite using `npm test`.  The output is redirected to `test-output.log`, and a summary of the test results is written to the GitHub step summary.  The workflow fails if tests fail.
    *   **`Upload Coverage Reports`:** Uses `codecov/codecov-action@v5` to upload code coverage reports to Codecov. This step only runs on Ubuntu with Node.js 20.  `file: ./coverage/lcov.info` specifies the coverage report file. `flags: unittests` and `name: codecov-umbrella` are used for organization within Codecov.

### 2. `code-hygiene` Job

*   **Name:** `Verify Code Standards`
*   **`runs-on`:** `ubuntu-latest` - Runs on Ubuntu.
*   **Steps:**
    *   **`Checkout`:** Uses `actions/checkout@v6` to checkout the repository code.
    *   **`Setup Node.js`:** Uses `actions/setup-node@v6` to set up Node.js version 20.
    *   **`Install Dependencies`:** Runs `npm ci` to install dependencies.
    *   **`Run Code Hygiene Checks`:** Executes ESLint (`npm run lint`) and Prettier (`npm run format:check`) to verify code style and formatting.  The results are written to the GitHub step summary in a table format.
    *   **`Verify Build`:** Runs `npm run build` to ensure the project builds successfully.
    *   **`Check Dependencies`:** Runs `npm audit --audit-level moderate` to scan for security vulnerabilities in dependencies, reporting those with a moderate severity or higher.

### 3. `dependency-review` Job

*   **Name:** `Dependency Review`
*   **`runs-on`:** `ubuntu-latest` - Runs on Ubuntu.
*   **`if: github.event_name == 'pull_request'`:** This job only runs when triggered by a pull request event.
*   **Steps:**
    *   **`Checkout`:** Uses `actions/checkout@v6` to checkout the repository code.
    *   **`Dependency Review`:** Uses `actions/dependency-review-action@v4` to perform a review of the pull request's dependencies, identifying potential security vulnerabilities and outdated packages.