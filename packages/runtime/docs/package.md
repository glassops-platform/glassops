---
type: Documentation
domain: runtime
origin: packages/runtime/package.json
last_modified: 2026-01-26
generated: true
source: packages/runtime/package.json
generated_at: 2026-01-26T05:02:41.360Z
hash: 1bb38d70ee58f47a345e9d064d864168bc0bc6be49334b3e6b320226d8dd1a38
---

# `@glassops/runtime` Package Documentation

This document details the structure and purpose of the `@glassops/runtime` package, a core component for Salesforce DevOps execution.

## Overview

The `@glassops/runtime` package serves as the execution primitive for a governance-first Salesforce DevOps pipeline. It provides the core logic and tooling for orchestrating and governing deployments and other DevOps processes within a Salesforce environment.  It is designed to be used within CI/CD systems, particularly GitHub Actions.

## Schema Breakdown

The `package.json` file defines the metadata and dependencies for this Node.js package.  Here's a breakdown of the key fields:

### Core Metadata

*   **`name`**: `@glassops/runtime` -  The unique identifier for the package within the npm registry and the project.
*   **`version`**: `1.0.0` - The current version of the package, following semantic versioning.
*   **`description`**: `The governance-first execution primitive for Salesforce DevOps.` - A brief description of the package's purpose.
*   **`main`**: `dist/index.js` - Specifies the entry point for the package when it's required in a Node.js environment.  This points to the compiled JavaScript file.
*   **`repository`**:  Defines the location of the source code repository.
    *   **`type`**: `git` - Indicates the repository type is Git.
    *   **`url`**: `git+https://github.com/glassops-platform/glassops-runtime.git` - The URL of the Git repository.
*   **`keywords`**: `["salesforce", "devops", "governance", "github-actions"]` -  Keywords used for searching and categorizing the package.
*   **`author`**: `Ryan Bumstead` - The author of the package.
*   **`license`**: `Apache-2.0` - The license under which the package is distributed.

### Development & Build Configuration

*   **`engines`**:  Specifies the Node.js version requirement.
    *   **`node`**: `>=20` -  Requires Node.js version 20 or higher.
*   **`scripts`**: Defines a set of scripts for automating common development tasks.
    *   **`build`**: `ncc build src/index.ts -o dist --source-map --license LICENSE` - Compiles the TypeScript source code (`src/index.ts`) into JavaScript and outputs it to the `dist` directory.  Uses `ncc` (Next.js Compiler) for bundling and includes source maps and the license file.
    *   **`format`**: `prettier --write \"src/**/*.ts\" \"config/**/*.js\"` - Formats TypeScript and JavaScript files in the `src` and `config` directories using Prettier.
    *   **`format:check`**: `prettier --check \"src/**/*.ts\" \"config/**/*.js\"` - Checks if TypeScript and JavaScript files are formatted correctly using Prettier, without making changes.
    *   **`lint`**: `eslint src/**/*.ts` -  Lints TypeScript files in the `src` directory using ESLint.
    *   **`test`**: `jest --config config/jest.config.js` - Runs unit tests using Jest with the configuration file `config/jest.config.js`.
    *   **`test:integration`**: `jest --config config/jest.integration.config.js` - Runs integration tests using Jest with the configuration file `config/jest.integration.config.js`.
    *   **`test:all`**: `npm run test && npm run test:integration` - Runs both unit and integration tests.
    *   **`all`**: `npm run format && npm run lint && npm run test && npm run build` - Executes all development tasks: formatting, linting, testing, and building.

### Dependencies

*   **`dependencies`**: Lists the packages required for the runtime execution of the package.
    *   `@actions/cache`: Used for caching dependencies and other artifacts in GitHub Actions.
    *   `@actions/core`: Provides core functionality for interacting with GitHub Actions.
    *   `@actions/exec`: Allows executing shell commands within GitHub Actions.
    *   `@actions/io`: Provides utilities for file system operations within GitHub Actions.
    *   `zod`: A TypeScript-first schema declaration and validation library.

*   **`devDependencies`**: Lists the packages required for development and testing, but not for runtime execution.  These include:
    *   `@eslint/js`: Core ESLint JavaScript parser.
    *   `@types/jest`: TypeScript definitions for Jest.
    *   `@types/node`: TypeScript definitions for Node.js.
    *   `@typescript-eslint/eslint-plugin`: ESLint plugin for TypeScript.
    *   `@typescript-eslint/parser`: ESLint parser for TypeScript.
    *   `@vercel/ncc`:  A compiler used for bundling Node.js projects.
    *   `eslint`:  A linting tool for JavaScript and TypeScript.
    *   `globals`: Provides global definitions for testing environments.
    *   `jest`: A JavaScript testing framework.
    *   `prettier`: A code formatter.
    *   `ts-jest`: A Jest transformer for TypeScript.
    *   `typescript`: The TypeScript compiler.
    *   `typescript-eslint`: ESLint plugin for TypeScript.

## Common Use Cases

*   **Salesforce Deployment Automation:** Automating the deployment of Salesforce metadata changes.
*   **Governance Enforcement:** Implementing and enforcing governance policies during deployments.
*   **CI/CD Pipelines:** Integrating with CI/CD systems (like GitHub Actions) to automate Salesforce DevOps processes.
*   **Data Management:** Automating data loading and manipulation tasks in Salesforce.
*   **Testing Automation:** Running automated tests against Salesforce environments.