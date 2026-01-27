---
type: Documentation
domain: agent
origin: packages/tools/agent/eslint.config.mjs
last_modified: 2026-01-26
generated: true
source: packages/tools/agent/eslint.config.mjs
generated_at: 2026-01-26T14:06:24.533Z
hash: 12e70763304b3a704a4daa8e14a3209c232b3e7f39789c88465467dee98defd8
---

## ESLint Configuration: Agent Tools Package

**Document Version:** 1.0
**Date:** October 26, 2023
**Author:** Principal Architect

**1. Introduction**

This document details the ESLint configuration for the `agent` tools package. ESLint is a static code analysis tool used to enforce coding standards and identify potential errors. This configuration builds upon a shared root configuration and extends it with rules specific to the `agent` package.

**2. Configuration Overview**

The ESLint configuration is defined in `packages/tools/agent/eslint.config.mjs`. It is a JavaScript module exporting an array of ESLint configuration objects. This structure allows for modularity and inheritance of rules.

**3. Core Components**

*   **Root Configuration Inheritance:** The configuration begins by inheriting all rules and settings from the root ESLint configuration located at `../../../config/eslint.config.mjs`. This ensures consistency across the project.
*   **Global Ignores:**  The `dist/` and `node_modules/` directories are explicitly excluded from linting. These directories contain generated or third-party code that should not be analyzed.
*   **TypeScript-Specific Rules:** A dedicated configuration block applies specific rules to all TypeScript (`.ts`) files within the package.

**4. TypeScript Rule Customizations**

The following rules are customized for TypeScript files:

*   `@lwc/lwc/no-async-operation`:  Disabled. This rule likely pertains to Lightning Web Components (LWC) and is not relevant or desired for this package.
*   `no-await-in-loop`: Disabled. Allows the use of `await` within loops, which may be necessary for certain asynchronous operations.
*   `@typescript-eslint/no-explicit-any`: Set to `warn`.  Discourages the use of the `any` type, promoting type safety, but issues a warning rather than an error to allow for pragmatic use cases.
*   `@typescript-eslint/no-unused-vars`: Set to `warn` with `argsIgnorePattern: "^_"`.  Flags unused variables as warnings. Variables starting with an underscore (`_`) are intentionally ignored, as they often represent unused function arguments.

**5. Purpose and Benefits**

This configuration aims to:

*   Maintain consistent code quality across the project.
*   Enforce best practices for TypeScript development.
*   Reduce the risk of runtime errors.
*   Improve code readability and maintainability.