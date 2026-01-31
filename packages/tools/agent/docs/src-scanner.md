---
type: Documentation
domain: agent
origin: packages/tools/agent/src/scanner.ts
last_modified: 2026-01-31
generated: true
source: packages/tools/agent/src/scanner.ts
generated_at: 2026-01-31T10:21:02.261432
hash: e72d6c8da430da632f90e9e248c9596fa06da0de3f12d3731f0267728bbf3356
---

## Agent Scanner Documentation

**Introduction**

This document details the functionality of the Agent Scanner, a component designed to locate files within a specified directory based on defined patterns, while respecting exclusion rules. It provides a robust and configurable method for identifying relevant files for processing.

**Functionality**

The Scanner identifies files matching given patterns within a root directory. It incorporates a flexible ignore system, based on `.gitignore` content and a default set of exclusions, to refine the search results. The scanner returns absolute paths to the identified files.

**Key Features**

*   **Pattern-Based Search:** Locates files using glob patterns.
*   **.gitignore Support:** Respects rules defined in a `.gitignore` file within the root directory.
*   **Default Exclusions:** Automatically excludes common build artifacts, dependency folders, and environment files.
*   **Absolute Paths:** Returns file paths as absolute references.

**Classes**

*   **`Scanner`**

    This class encapsulates the file scanning logic.

    *   **Constructor (`rootDir: string`)**

        Initializes the Scanner with the root directory to scan. It loads `.gitignore` rules, if present, and establishes a default set of exclusions.

        *   `rootDir`: The base directory for the file search.

    *   **`findFiles(patterns: string[]): Promise<string[]>`**

        Asynchronously searches for files matching the provided patterns within the root directory.

        *   `patterns`: An array of glob patterns to match against file paths.
        *   Returns: A Promise that resolves to an array of absolute file paths that match the patterns and are not ignored.

**Usage**

1.  **Instantiation:** Create a `Scanner` instance, providing the root directory as an argument.

    ```typescript
    const scanner = new Scanner('/path/to/your/project');
    ```

2.  **File Search:** Call the `findFiles` method with an array of glob patterns.

    ```typescript
    const files = await scanner.findFiles(['*.txt', 'src/**/*.js']);
    console.log(files); // Output: An array of absolute file paths
    ```

**Default Ignored Patterns**

The following patterns are always excluded from search results:

*   `node_modules/**`
*   `dist/**`
*   `package-lock.json`
*   `.env`
*   `docs/generated/**`
*   `venv/**`
*   `__pycache__/**`

**.gitignore Integration**

If a `.gitignore` file exists in the root directory, its contents are automatically loaded and applied to the file search. This allows you to define exclusions specific to your project.

**Error Handling**

The `findFiles` method handles potential errors during file system access and glob matching. Errors are not explicitly thrown but may be logged internally by the underlying `fast-glob` library.

**Dependencies**

*   `fast-glob`: For efficient file system traversal and glob matching.
*   `ignore`: For managing ignore patterns based on `.gitignore` and other rules.
*   `fs`: For file system operations.
*   `path`: For path manipulation.