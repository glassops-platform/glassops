---
type: Documentation
domain: agent
origin: packages/tools/agent/src/scanner.ts
last_modified: 2026-01-31
generated: true
source: packages/tools/agent/src/scanner.ts
generated_at: 2026-01-31T09:23:50.473982
hash: e72d6c8da430da632f90e9e248c9596fa06da0de3f12d3731f0267728bbf3356
---

## Agent Scanner Documentation

**Introduction**

This document details the functionality of the Agent Scanner, a component designed to locate files within a specified directory based on defined patterns, while respecting exclusion rules. It provides a robust and configurable method for identifying relevant files for processing.

**Functionality**

The Scanner identifies files matching provided glob patterns within a target directory. It incorporates a flexible ignore system, based on both a `.gitignore` file (if present) and a set of predefined exclusions. The scanner returns an array of absolute file paths.

**Key Features**

*   **Pattern-Based Search:** Locates files using standard glob patterns.
*   **.gitignore Support:** Automatically incorporates rules from a `.gitignore` file located in the target directory.
*   **Predefined Exclusions:** Includes common build artifacts, dependency folders, and environment files in the default exclusion list.
*   **Absolute Paths:** Returns file paths as absolute references, ensuring clarity and consistency.
*   **Hidden File Inclusion:** Includes hidden files and directories in the search results.

**Implementation Details**

The Scanner class is initialized with a root directory. During construction, it loads any existing `.gitignore` file and adds a set of default exclusion rules. These rules are then applied during the file search process. The core file searching is performed using the `fast-glob` library for performance.

**Usage**

1.  **Initialization:** Instantiate the `Scanner` class, providing the root directory to scan as an argument.

    ```typescript
    const scanner = new Scanner('/path/to/your/project');
    ```

2.  **File Search:** Call the `findFiles` method, passing an array of glob patterns.

    ```typescript
    const files = await scanner.findFiles(['*.txt', 'src/**/*.js']);
    ```

    You should await this function as it is asynchronous. The `files` variable will contain an array of strings, each representing the absolute path to a matching file.

**Configuration**

*   **.gitignore:** Place a `.gitignore` file in the root directory to define custom exclusion rules. The scanner automatically loads and applies these rules.
*   **Default Exclusions:** The following patterns are always excluded:
    *   `node_modules/**`
    *   `dist/**`
    *   `package-lock.json`
    *   `.env`
    *   `docs/generated/**`
    *   `venv/**`
    *   `__pycache__/**`

**Dependencies**

*   `fast-glob`: For efficient file system traversal.
*   `ignore`: For managing exclusion rules.
*   `fs`: For file system operations.
*   `path`: For path manipulation.

**Return Value**

The `findFiles` method returns a `Promise` that resolves to a string array. Each string in the array represents the absolute path to a file that matches the provided patterns and is not excluded by the ignore rules.

**Error Handling**

The `fast-glob` library handles most file system errors. Any errors encountered during the process will be propagated as rejections of the `Promise` returned by `findFiles`. You should implement appropriate error handling in your calling code.