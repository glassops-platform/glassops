---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/config/jest.integration.config.js
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/config/jest.integration.config.js
generated_at: 2026-01-31T11:09:14.883870
hash: e015b22c69711d008eba43205f6f91622c2a2a9133f6434e4762ff636395b2a1
---

## Integration Test Configuration

This document details the configuration for integration tests within the project. These tests verify the interaction between different parts of the system.

### Overview

The configuration defines how tests are discovered, executed, and reported. It specifies settings for the Jest testing framework, including the test environment, timeout values, code coverage, and setup procedures.

### Key Settings

* **Preset:** `ts-jest` – Configures Jest to work with TypeScript projects, handling compilation and transformation.
* **Test Environment:** `node` – Specifies that tests should be run in a Node.js environment.
* **Root Directory:** `../src` – Sets the base directory for resolving test paths to the source code directory.
* **Test Match:** `**/*.integration.test.ts` – Defines the pattern used to identify integration test files. Only files ending in `.integration.test.ts` will be included in the test run.
* **Test Timeout:** `30000` milliseconds – Increases the default timeout for individual tests to 30 seconds, accommodating potentially longer-running integration tests.
* **Transform Ignore Patterns:** `[/node_modules/] ` – Prevents Jest from attempting to transform files within the `node_modules` directory, improving performance.
* **Coverage Collection:** `true` – Enables code coverage analysis during test execution.
* **Coverage Directory:** `../coverage/integration` – Specifies the directory where coverage reports will be stored.
* **Coverage Reporters:** `["text", "lcov", "html", "json-summary"]` – Configures Jest to generate coverage reports in multiple formats:
    * `text`:  A human-readable text summary in the console.
    * `lcov`:  A format suitable for integration with code coverage tools.
    * `html`:  An HTML report for easy visualization in a web browser.
    * `json-summary`: A JSON file containing a summary of the coverage results.
* **Coverage Source Files:**  The following patterns determine which source files are included in the coverage analysis:
    * `**/*.ts`: All TypeScript files.
    * `!**/*.test.ts`: Excludes unit test files.
    * `!**/*.integration.test.ts`: Excludes integration test files.
    * `!**/node_modules/**`: Excludes files within the `node_modules` directory.
    * `!**/index.ts`: Excludes the project’s entry point, as it is tested through end-to-end tests.
    * `!**/integration/test-helpers.ts`: Excludes test helper files, which are not part of the production code.
* **Setup Files:** `<rootDir>/../config/jest.integration.setup.js` – Specifies a file to be executed before running the tests. This file is used to configure the test environment, such as setting up database connections or mocking external dependencies.

### Usage

You do not typically interact with this configuration directly. It is used by the test runner when you execute integration tests. To run the integration tests, use the appropriate command defined in the project’s `package.json` file (e.g., `npm run test:integration`).

### Maintainer Notes

I have designed this configuration to provide comprehensive coverage reporting for integration tests while excluding unnecessary files and optimizing performance. The increased timeout value is intended to support tests that interact with external systems or perform complex operations. We will continue to monitor and adjust these settings as the project evolves.