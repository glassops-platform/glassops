---
type: Documentation
domain: runtime
origin: packages/runtime/eslint.config.mjs
last_modified: 2026-01-26
generated: true
source: packages/runtime/eslint.config.mjs
generated_at: 2026-01-26T14:05:03.575Z
hash: d503daf56d501a737fbefc255d43a9e5cbef1a559247f87477997259741d563b
---

## ESLint Configuration: Runtime Package

**Document Version:** 1.0
**Date:** October 26, 2023
**Author:** Principal Architect

**1. Introduction**

This document details the ESLint configuration for the `runtime` package. ESLint is a static code analysis tool used to enforce coding standards, identify potential errors, and improve code quality. This configuration builds upon a shared root configuration and extends it with rules specific to the runtime environment.

**2. Configuration Overview**

The ESLint configuration is defined in `packages/runtime/eslint.config.mjs`. It is an array of configuration objects, allowing for layered and modular rule sets.  This configuration leverages a base configuration defined in `../../config/eslint.config.mjs` and extends it with runtime-specific overrides.

**3. Key Components**

*   **Base Configuration:** The configuration inherits all rules and settings from the root ESLint configuration located at `../../config/eslint.config.mjs`. This ensures consistency across the project.
*   **Global Ignores:** The following directories are excluded from linting:
    *   `dist/`:  Generated distribution files.
    *   `node_modules/`:  Installed dependencies.
    *   `coverage/`:  Test coverage reports.
*   **TypeScript-Specific Rules:**  Rules are applied specifically to TypeScript (`.ts`) files. These rules modify or disable rules from the base configuration to better suit the runtime package’s needs.
    *   `@lwc/lwc/no-async-operation`: Disabled. Allows asynchronous operations where the base configuration might disallow them.
    *   `no-await-in-loop`: Disabled. Permits the use of `await` within loops.
    *   `@typescript-eslint/no-explicit-any`: Set to `warn`.  Discourages the use of the `any` type, but flags it as a warning rather than an error.
    *   `@typescript-eslint/no-unused-vars`: Set to `warn` with `argsIgnorePattern: "^_"`.  Flags unused variables as warnings, but ignores variables starting with an underscore (`_`).
*   **TypeScript Language Options:** Configures the TypeScript parser for accurate and reliable analysis.
    *   `projectService`: Enables type checking using `tsconfig.json` files. `allowDefaultProject: ['*.ts', '*.js']` allows the parser to find project configurations for both TypeScript and JavaScript files.
    *   `tsconfigRootDir`: Specifies the root directory for resolving `tsconfig.json` files, set to the directory of the current configuration file using `import.meta.dirname`.

**4. Purpose and Benefits**

This configuration aims to:

*   Maintain consistent code style across the project.
*   Identify potential runtime errors early in the development process.
*   Improve code readability and maintainability.
*   Provide flexibility for TypeScript-specific requirements within the runtime package.