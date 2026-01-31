---
type: Documentation
domain: agent
origin: packages/tools/agent/eslint.config.mjs
last_modified: 2026-01-31
generated: true
source: packages/tools/agent/eslint.config.mjs
generated_at: 2026-01-31T11:10:24.842906
hash: 642d832bd1e30cccfe62dd437a93aa9a78a6f926e506cf45c4ea171fe501209c
---

## Agent ESLint Configuration

This document details the configuration for the ESLint tool used within the Agent package. It outlines how code quality and style are maintained, and how potential issues are identified. This configuration builds upon a shared root configuration and adds specific rules relevant to this package.

**Purpose**

The purpose of this configuration is to enforce consistent coding standards, identify potential errors, and improve the overall maintainability of the Agent codebase. It ensures that all TypeScript code adheres to a defined set of rules.

**Configuration Structure**

The ESLint configuration is structured as an array of objects. Each object represents a configuration set that applies to specific files or the entire project.

**1. Base Configuration**

The configuration begins by extending a base configuration located at `../../../config/eslint.config.mjs`. This ensures consistency across multiple packages by inheriting a common set of rules.

**2. Ignore Patterns**

The following patterns are excluded from linting:

*   `dist/**`:  Files within the `dist` directory (typically compiled output).
*   `node_modules/**`: Files within the `node_modules` directory (dependencies).

**3. TypeScript Specific Rules**

This section defines rules specifically for TypeScript files (`**/*.ts`).

*   `@lwc/lwc/no-async-operation`: This rule is disabled. It prevents asynchronous operations within certain contexts, which are permitted in this package.
*   `no-await-in-loop`: This rule is disabled. It prevents the use of `await` inside loops, which is sometimes necessary for the Agent’s operation.
*   `@typescript-eslint/no-explicit-any`:  This rule is set to `warn`. It flags the use of the `any` type, encouraging more specific type definitions.  Using `any` will generate a warning, but will not cause a build failure.
*   `@typescript-eslint/no-unused-vars`: This rule is set to `warn` with an exception for arguments starting with an `_`. It flags unused variables, promoting cleaner code. Arguments prefixed with an underscore (`_`) are intentionally ignored and will not trigger a warning. The configuration is: `['warn', { "argsIgnorePattern": "^_" }]`.

**Usage**

You do not typically interact with this configuration directly. The ESLint tool is integrated into the development workflow (e.g., through pre-commit hooks, CI/CD pipelines, or IDE integrations).  When you run ESLint, it automatically applies these rules to your code.

**Maintenance**

I will maintain and update this configuration as needed to reflect evolving best practices and project requirements. Contributions and suggestions are welcome.