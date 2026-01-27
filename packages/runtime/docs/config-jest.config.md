---
type: Documentation
domain: runtime
origin: packages/runtime/config/jest.config.js
last_modified: 2026-01-26
generated: true
source: packages/runtime/config/jest.config.js
generated_at: 2026-01-26T14:06:46.096Z
hash: 13473cbdd251fa5d93b7d60252aff0d2cd2609fb671fadf5067a39b5c60f4df8
---

## Jest Configuration for Runtime Package

**Document Version:** 1.0
**Date:** October 26, 2023
**Author:** Principal Architect

**1. Introduction**

This document details the Jest configuration for the `runtime` package. This configuration defines how automated tests are executed and reported, ensuring code quality and reliability. It is designed for both developers and stakeholders needing to understand the testing framework setup.

**2. Configuration Overview**

The `jest.config.js` file specifies the settings for the Jest testing framework. These settings control test discovery, execution environment, code coverage analysis, and reporting.

**3. Key Configuration Elements**

*   **`preset: "ts-jest"`:**  Utilizes the `ts-jest` preset, enabling Jest to directly execute TypeScript code without pre-transpilation. This simplifies the testing process and improves developer experience.
*   **`testEnvironment: "node"`:**  Specifies that tests should be run in a Node.js environment. This is appropriate for server-side runtime code.
*   **`rootDir: "../"`:** Sets the root directory for the project to the parent directory of the `packages/runtime/config` folder. This ensures that test paths are correctly resolved relative to the project's source code.
*   **`testMatch: ["<rootDir>/src/**/*.test.ts"]`:** Defines the glob pattern used to identify test files.  Tests are located within the `src` directory and have the `.test.ts` extension.
*   **`testPathIgnorePatterns`:**  Specifies patterns for files or directories that should be excluded from test execution.
    *   `"<rootDir>/src/integration/"`: Excludes the entire `integration` directory.
    *   `".*\\.integration\\.test\\.ts$"`: Excludes any file ending in `.integration.test.ts`. This allows for separate, potentially slower, integration tests to be run independently.
*   **`moduleFileExtensions: ["ts", "js", "json", "node"]`:**  Defines the file extensions that Jest should recognize when resolving modules.
*   **`collectCoverage: true`:** Enables code coverage collection during test execution.
*   **`coverageDirectory: "coverage"`:** Specifies the directory where code coverage reports will be stored.
*   **`coverageReporters`:**  Configures the types of code coverage reports generated.
    *   `"text"`:  Displays a text-based summary of coverage in the console.
    *   `"lcov"`: Generates an LCOV report, suitable for integration with code coverage tools.
    *   `"html"`: Creates an HTML report for visual inspection of code coverage.
    *   `"json-summary"`: Produces a JSON summary of coverage data.
*   **`coverageThreshold`:**  Sets minimum acceptable code coverage thresholds.  Tests will fail if coverage falls below these levels.
    *   `global`: Applies to the entire codebase.
        *   `branches: 80`: Requires at least 80% branch coverage.
        *   `functions: 80`: Requires at least 80% function coverage.
        *   `lines: 80`: Requires at least 80% line coverage.
        *   `statements: 80`: Requires at least 80% statement coverage.

**4. Purpose and Benefits**

This configuration ensures:

*   **Comprehensive Testing:**  All relevant TypeScript test files are automatically discovered and executed.
*   **Code Quality:**  Coverage thresholds enforce a minimum level of code coverage, promoting thorough testing.
*   **Detailed Reporting:**  Multiple report formats provide insights into test results and code coverage.
*   **Simplified Development:**  `ts-jest` simplifies the testing process for TypeScript code.
*   **Maintainability:** Clear and consistent configuration promotes maintainability and reduces the risk of errors.