---
type: Documentation
domain: agent
origin: packages/tools/agent/eslint.config.mjs
last_modified: 2026-02-01
generated: true
source: packages/tools/agent/eslint.config.mjs
generated_at: 2026-02-01T19:46:50.928587
hash: 642d832bd1e30cccfe62dd437a93aa9a78a6f926e506cf45c4ea171fe501209c
---

## Agent ESLint Configuration

This document details the configuration for the ESLint tool used within the Agent package. It outlines how code quality and style are maintained, and how potential issues are identified. This configuration builds upon a shared root configuration and adds specific rules relevant to this package.

**Purpose**

The purpose of this configuration is to enforce consistent coding standards, identify potential errors, and improve the overall maintainability of the Agent codebase. It helps ensure code quality and reduces the risk of bugs.

**Configuration Structure**

The ESLint configuration is structured as an array of objects. Each object represents a set of rules or configuration options. This configuration extends a base configuration and adds package-specific overrides.

**Base Configuration**

This configuration extends the base ESLint configuration located at `../../../config/eslint.config.mjs`. This base configuration provides a foundation of rules applicable across multiple packages.

**Ignores**

The following directories are excluded from ESLint analysis:

*   `dist/**`:  The distribution directory, containing compiled code.
*   `node_modules/**`: The directory containing installed dependencies.

**TypeScript Specific Rules**

The following rules are specifically applied to TypeScript files (`**/*.ts`):

*   `@lwc/lwc/no-async-operation`: This rule is disabled. It prevents asynchronous operations within certain contexts, which are permitted in this package.
*   `no-await-in-loop`: This rule is disabled. It prevents the use of `await` inside loops, which is sometimes necessary for the Agent’s operation.
*   `@typescript-eslint/no-explicit-any`:  This rule is set to `warn`. It flags the use of the `any` type, encouraging more specific type definitions.  Using `any` is not prohibited, but a warning is issued to prompt review.
*   `@typescript-eslint/no-unused-vars`: This rule is set to `warn`. It flags unused variables. Arguments starting with an `_` are excluded from this check, allowing for intentional omission of unused parameters. The configuration is: `['warn', { "argsIgnorePattern": "^_" }]`.

**How to Use**

You do not typically interact with this configuration directly. It is used automatically by the ESLint tool during development and continuous integration processes. To run ESLint, use the command provided in the package’s documentation or your development environment.

**Maintenance**

I will maintain this configuration to ensure it remains aligned with best practices and the evolving needs of the Agent package. Updates will be documented in the project’s release notes.