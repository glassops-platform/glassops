---
type: Documentation
domain: agent
origin: packages/tools/agent/src/scanner.ts
last_modified: 2026-01-26
generated: true
source: packages/tools/agent/src/scanner.ts
generated_at: 2026-01-26T14:13:21.162Z
hash: b2a4b10c7db5fa3f95ba057e8cb0b95314437d68bdc05cd2a3ea84962d74b070
---

## Agent File Scanner Documentation

**Overview**

The File Scanner component is responsible for discovering files within a specified directory based on provided patterns, while respecting exclusion rules defined in `.gitignore` files and a default set of ignored patterns. It provides a simple interface for locating files relevant to the agent’s operation.

**Key Features**

*   **Pattern-Based Search:** Locates files matching glob patterns.
*   **.gitignore Support:** Automatically incorporates rules from `.gitignore` files within the target directory.
*   **Default Exclusions:**  Includes a pre-defined list of commonly ignored files and directories (e.g., `node_modules`, `dist`, lock files, environment files).
*   **Absolute Paths:** Returns file paths as absolute references.

**Architecture**

The `Scanner` class encapsulates the file discovery logic. It initializes with a root directory and builds an ignore instance to manage exclusion rules. The core functionality resides in the `findFiles` method, which leverages the `fast-glob` library for efficient file system traversal.

**Usage**

1.  **Initialization:**
    *   Instantiate the `Scanner` class, providing the root directory to scan as a parameter.

    ```typescript
    const scanner = new Scanner('/path/to/your/project');
    ```

2.  **File Discovery:**
    *   Call the `findFiles` method, passing an array of glob patterns to search for.

    ```typescript
    const files = await scanner.findFiles(['*.txt', 'src/**/*.js']);
    ```

    *   The `findFiles` method returns a Promise that resolves to an array of strings, where each string is the absolute path to a matching file.

**Configuration**

*   **Root Directory:** The `rootDir` is set during instantiation and defines the base directory for all file searches.
*   **.gitignore:** The scanner automatically reads and applies rules from a `.gitignore` file located in the `rootDir`.
*   **Default Ignored Patterns:** The following patterns are always ignored:
    *   `node_modules/**`
    *   `dist/**`
    *   `package-lock.json`
    *   `.env`
    *   `docs/generated/**`
    *   `venv/**`
    *   `__pycache__/**`
*   **Custom Ignore Patterns:**  I do not currently support adding custom ignore patterns beyond those loaded from `.gitignore` and the defaults.

**Dependencies**

*   `fast-glob`:  A high-performance glob library for file system traversal.
*   `ignore`: A library for matching files against `.gitignore` patterns.
*   `fs`: Node.js file system module.
*   `path`: Node.js path module.

**Error Handling**

The `fast-glob` library handles most file system errors.  I do not currently implement specific error handling beyond allowing those errors to propagate.

**Future Considerations**

*   Allow users to specify custom ignore patterns.
*   Implement more robust error handling.
*   Add options for controlling glob behavior (e.g., only files, only directories).