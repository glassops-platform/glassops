---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/eslint.config.mjs
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/eslint.config.mjs
generated_at: 2026-01-31T11:10:12.361137
hash: 6c18960259bdddac150641f8016881667a04e07d58087c38a3b44b6229458846
---

## ESLint Configuration for Runtime Package

This document details the ESLint configuration for the runtime package. It defines the rules and settings applied to TypeScript code within this project, ensuring code quality and consistency.

**Overview**

This configuration extends a base ESLint configuration defined in `../../config/eslint.config.mjs`, adding specific rules tailored to the runtime package’s needs. It focuses on TypeScript-specific linting and manages ignored files.

**Configuration Structure**

The configuration is an array of ESLint configurations. This allows for layering of rules and settings.

**1. Base Configuration Extension**

The first element in the array extends the root ESLint configuration. This ensures that all base rules and settings are inherited.

```typescript
...rootConfig,
```

**2. Ignore Patterns**

This section defines files and directories that should be excluded from linting.

```typescript
{
    ignores: ['dist/**', 'node_modules/**', 'coverage/**']
}
```

*   `dist/**`:  Ignores all files and directories within the `dist` directory (typically containing compiled code).
*   `node_modules/**`: Ignores all files and directories within the `node_modules` directory (containing project dependencies).
*   `coverage/**`: Ignores all files and directories within the `coverage` directory (containing test coverage reports).

**3. TypeScript-Specific Rules**

This section applies specific rules to TypeScript files (`*.ts`).

```typescript
{
    files: ['**/*.ts'],
    rules: {
        '@lwc/lwc/no-async-operation': 'off',
        'no-await-in-loop': 'off',
        '@typescript-eslint/no-explicit-any': 'warn',
        '@typescript-eslint/no-unused-vars': ['warn', { "argsIgnorePattern": "^_" }]
    },
    languageOptions: {
        parserOptions: {
            projectService: {
                allowDefaultProject: ['*.ts', '*.js']
            },
            tsconfigRootDir: import.meta.dirname
        }
    }
}
```

*   `files: ['**/*.ts']`: Specifies that the following rules apply only to TypeScript files.
*   `@lwc/lwc/no-async-operation: 'off'`: Disables the rule that prevents asynchronous operations.
*   `no-await-in-loop: 'off'`: Disables the rule that prevents the use of `await` inside loops.
*   `@typescript-eslint/no-explicit-any: 'warn'`:  Issues a warning when the `any` type is explicitly used.  This encourages more specific type definitions.
*   `@typescript-eslint/no-unused-vars: ['warn', { "argsIgnorePattern": "^_" }]`: Issues a warning for unused variables. Variables starting with an underscore (`_`) are excluded from this check.
*   `languageOptions`: Configures the TypeScript parser.
    *   `parserOptions`: Specifies parser options.
        *   `projectService`: Enables type checking using the project’s `tsconfig.json` file. `allowDefaultProject: ['*.ts', '*.js']` allows the parser to find the configuration file.
        *   `tsconfigRootDir: import.meta.dirname`: Sets the root directory for resolving the `tsconfig.json` file.

**Usage**

You do not directly interact with this configuration file. It is used internally by ESLint when linting the project’s TypeScript code. To run ESLint, use the command specified in the project’s `package.json` file (e.g., `npm run lint`).

**Maintainer Notes**

I maintain this configuration to ensure consistent code style and quality. Changes to this file should be carefully considered and tested to avoid introducing unintended consequences. We aim to balance strictness with practicality, allowing for flexibility while maintaining a high standard of code quality.