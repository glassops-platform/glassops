---
type: Documentation
domain: agent
origin: packages/tools/agent/tsconfig.json
last_modified: 2026-01-26
generated: true
source: packages/tools/agent/tsconfig.json
generated_at: 2026-01-26T05:12:39.219Z
hash: 7eec88c22c4852b835b49fe48be4ebd703fd8c88aeb3c88f95123b3c8b2fc19e
---

# `tsconfig.json` Documentation - Agent Package

This document details the configuration of the TypeScript compiler for the agent package. This file dictates how TypeScript code within the `src` directory is compiled into JavaScript.

## Overview

The `tsconfig.json` file configures the TypeScript compiler (`tsc`) to transpile TypeScript code into JavaScript. It specifies compiler options, files to include in the compilation process, and files to exclude. This configuration is crucial for building and maintaining the agent package.

## Key Sections

### `compilerOptions`

This section defines the options that control how the TypeScript compiler behaves.

*   **`target`**: `"ES2022"` - Specifies the ECMAScript target version for the compiled JavaScript.  `ES2022` means the code will be compiled to be compatible with environments supporting ECMAScript 2022 features. *Required*.
*   **`module`**: `"NodeNext"` - Determines the module code generation style. `"NodeNext"` is designed for Node.js environments and utilizes the latest module features. *Required*.
*   **`moduleResolution`**: `"NodeNext"` - Specifies how TypeScript resolves module imports. `"NodeNext"` aligns with the latest Node.js module resolution strategy. *Required*.
*   **`outDir`**: `"dist"` -  Specifies the output directory where the compiled JavaScript files will be placed. *Required*.
*   **`rootDir`**: `"src"` - Specifies the root directory of the source files.  The compiler will structure the output in `outDir` relative to this directory. *Required*.
*   **`strict`**: `true` - Enables all strict type-checking options. This helps catch potential errors during development. *Required*.
*   **`esModuleInterop`**: `true` - Enables interoperability between CommonJS and ES modules. This allows importing CommonJS modules in ES module code and vice versa. *Required*.
*   **`skipLibCheck`**: `true` -  Skips type checking of declaration files (`.d.ts`). This can improve compilation speed, especially when working with large projects and many dependencies. *Optional, but common for performance*.
*   **`forceConsistentCasingInFileNames`**: `true` -  Ensures that file names are consistently cased across different operating systems. This prevents issues related to case sensitivity. *Optional, but recommended for cross-platform compatibility*.
*   **`declaration`**: `true` - Generates `.d.ts` declaration files alongside the JavaScript files. These files provide type information for the compiled code, enabling better tooling support and type safety for consumers of the package. *Optional, but useful for library development*.

### `include`

This section specifies the files or patterns of files to include in the compilation process.

*   **`["src/**/*"]`**: Includes all TypeScript files (`.ts` and `.tsx`) within the `src` directory and its subdirectories. *Required*.

### `exclude`

This section specifies the files or patterns of files to exclude from the compilation process.

*   **`["node_modules", "dist"]`**: Excludes the `node_modules` directory (containing installed dependencies) and the `dist` directory (containing the compiled output) from the compilation process. This prevents unnecessary compilation of third-party code and the output itself. *Required*.

## Common Use Cases

*   **Development:**  The `tsconfig.json` file is used during development to compile TypeScript code as it is being written, providing real-time feedback and error checking.
*   **Building:**  The `tsconfig.json` file is used as part of the build process to generate production-ready JavaScript code.
*   **Type Checking:** The `strict` option and declaration file generation (`declaration: true`) enable robust type checking and provide type information for other projects that consume this package.
*   **Package Publishing:** When publishing the agent package, the compiled JavaScript and declaration files are included in the published distribution.