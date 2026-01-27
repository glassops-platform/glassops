---
type: Documentation
domain: runtime
origin: packages/runtime/.github/workflows/release.yml
last_modified: 2026-01-26
generated: true
source: packages/runtime/.github/workflows/release.yml
generated_at: 2026-01-26T05:13:46.042Z
hash: 8981e08ca5d3b34ffcb4a12a3be0ba582fff00ca77ca2db7f14a45676902f3b6
---

# Release Workflow Documentation

This document details the `release.yml` workflow file located in the `.github/workflows` directory. This workflow automates the release process for the project, triggered by pushes to tags (versioned releases) or manual dispatch.

## Workflow Structure

The workflow consists of a single job named `release`. This job is executed on an Ubuntu-latest runner.

## Key Configuration Details

### `name: Release`

*   **Purpose:** Defines the name of the workflow, displayed in the GitHub Actions interface.

### `on:`

*   **Purpose:** Specifies the events that trigger the workflow.
*   **`push:`** Triggers the workflow on push events.
    *   **`tags:`**  Further filters the push event to only trigger when a tag is pushed. The tag name must match the pattern `"v*"`.
*   **`workflow_dispatch:`** Enables manual triggering of the workflow from the GitHub Actions interface.

### `jobs:`

*   **Purpose:** Defines the jobs to be executed in the workflow.
*   **`release:`** Defines the release job.
    *   **`name: Create Release`** Sets the name of the job.
    *   **`runs-on: ubuntu-latest`** Specifies the runner environment for the job (Ubuntu latest).
    *   **`permissions:`** Defines the permissions granted to the workflow.
        *   **`contents: write`** Allows the workflow to write to the repository's contents (e.g., create/update files).
        *   **`packages: write`** Allows the workflow to write to the repository's packages.
        *   **`id-token: write`** Allows the workflow to write to the repository's ID token.
    *   **`steps:`**  A sequence of tasks to be executed within the job.

#### Steps Breakdown

*   **`Checkout`**
    *   **`uses: actions/checkout@v6`** Checks out the repository code to the runner.
*   **`Setup Node.js`**
    *   **`uses: actions/setup-node@v6`** Sets up a Node.js environment.
    *   **`with:`** Configuration options for the Node.js setup.
        *   **`node-version: "20"`** Specifies the Node.js version to use.
        *   **`cache: "npm"`** Enables caching of npm dependencies for faster builds.
        *   **`registry-url: "https://registry.npmjs.org"`** Sets the npm registry URL.
*   **`Install Dependencies`**
    *   **`run: npm ci`** Installs project dependencies using `npm ci` (clean install).
*   **`Run Tests`**
    *   **`run: npm test`** Executes the project's test suite.
*   **`Build`**
    *   **`run: npm run build`** Executes the project's build script.
*   **`Create GitHub Release`**
    *   **`uses: softprops/action-gh-release@v2`** Creates a GitHub release.
    *   **`with:`** Configuration options for the release creation.
        *   **`generate_release_notes: true`** Automatically generates release notes based on commit messages.
        *   **`body:`** The body of the release notes. Includes installation instructions and a link to the CHANGELOG.md file.
        *   **`draft: false`** Creates a published release (not a draft).
        *   **`prerelease: false`** Creates a stable release (not a pre-release).
*   **`Generate Changelog`**
    *   **`run:`** A multi-line script to generate a basic CHANGELOG.md file.  It creates a new file or overwrites an existing one with a header, release version, date, and placeholder changes.
*   **`Commit Changelog`**
    *   **`run:`** A multi-line script to commit the generated CHANGELOG.md file back to the repository. It configures git with a generic user, adds the CHANGELOG.md file, commits the changes, and pushes them to the repository.  Includes error handling to prevent commit failures if no changes are detected.
*   **`Publish to NPM (if applicable)`**
    *   **`run:`** A conditional script to publish the package to npm. It checks if the tag name matches the pattern `v[0-9]+\.[0-9]+\.[0-9]+` (a semantic version). If it does, it publishes the package using `npm publish`. Otherwise, it skips the publishing step.
    *   **`env:`** Environment variables for the step.
        *   **`NODE_AUTH_TOKEN: ${{ secrets.NPM_TOKEN }}`** Provides the npm authentication token from a GitHub secret.