---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/config/jest.config.js
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/config/jest.config.js
generated_at: 2026-01-31T11:08:57.028250
hash: b18205ff56a3b280190ff342fc2274acbbac97f5209fb9cad0deaf2e74b1808b
---

## Runtime Package: Testing Configuration

This document details the configuration for the testing framework used within this runtime package. It outlines how tests are discovered, executed, and how code coverage is measured.

### Overview

The testing setup employs Jest, a popular JavaScript testing framework, configured specifically for TypeScript projects. This configuration ensures consistent and reliable testing practices throughout the development lifecycle.

### Configuration Details

The following settings govern the testing process:

*   **`preset: "ts-jest"`**:  Specifies the use of the `ts-jest` preset, which provides built-in support for compiling and running TypeScript tests.
*   **`testEnvironment: "node"`**:  Defines the test environment as Node.js, suitable for server-side JavaScript and TypeScript code.
*   **`rootDir: "../"`**: Sets the root directory for the project to the parent directory of the `config` folder. This ensures tests are located relative to the project’s source code.
*   **`testMatch: ["<rootDir>/src/**/*.test.ts"]`**:  Defines the pattern used to locate test files.  All files ending in `.test.ts` within the `src` directory and its subdirectories will be recognized as tests.
*   **`testPathIgnorePatterns: [...]`**:  Specifies patterns for files or directories to exclude from testing.
    *   `"<rootDir>/src/integration/"`: Excludes the entire `src/integration` directory.
    *   `".*\\.integration\\.test\\.ts$"`: Excludes any file with “integration” in its name and ending with `.test.ts`. This allows for focused unit testing by excluding larger integration tests during specific runs.
*   **`moduleFileExtensions: ["ts", "js", "json", "node"]`**:  Lists the file extensions that Jest should recognize as modules.
*   **`collectCoverage: true`**: Enables code coverage collection during test execution.
*   **`coverageDirectory: "coverage"`**:  Specifies the directory where code coverage reports will be stored.
*   **`coverageReporters: ["text", "lcov", "html", "json-summary"]`**:  Defines the formats for the generated code coverage reports.
    *   `text`:  A human-readable text summary in the console.
    *   `lcov`:  A format commonly used for integration with code coverage services.
    *   `html`:  An interactive HTML report for detailed coverage analysis.
    *   `json-summary`: A JSON file containing a summary of the coverage results.
*   **`coverageThreshold: { global: { ... } }`**:  Sets minimum acceptable code coverage thresholds.  Tests will report a failure if coverage falls below these levels.
    *   `branches: 80`: Requires at least 80% branch coverage.
    *   `functions: 80`: Requires at least 80% function coverage.
    *   `lines: 80`: Requires at least 80% line coverage.
    *   `statements: 80`: Requires at least 80% statement coverage.

### Running Tests

To execute the tests, you should use the standard Jest command: `jest`.  This command will discover and run all tests matching the `testMatch` pattern, collect code coverage information, and report any failures.

### Customization

You can modify this configuration file to adjust test discovery patterns, coverage thresholds, or reporting options to suit your specific needs.  Changes to this file will affect all subsequent test runs.