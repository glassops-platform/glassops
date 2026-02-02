---
type: Documentation
domain: agent
origin: packages/tools/agent/src/index.ts
last_modified: 2026-02-01
generated: true
source: packages/tools/agent/src/index.ts
generated_at: 2026-02-01T19:50:32.392981
hash: eab86e8c516702845a7338f7e46e27a9f0d332309fbcd22b01d200c37c38bae5
---

## GlassOps Agent Documentation

**Overview**

This tool is an AI-powered agent designed to automatically generate documentation and metadata for GlassOps repositories. It analyzes source code and related files to produce up-to-date and informative documentation.

**Purpose**

The agent simplifies the process of maintaining accurate documentation, reducing manual effort and improving overall project understanding. It supports a variety of file types commonly found in software projects.

**Installation & Execution**

The agent is designed to be executed from the command line. Ensure Node.js is installed on your system.

**Command-Line Interface**

The primary command is `glassops-agent`.

*   `glassops-agent --version`: Displays the agent’s version number (currently 1.0.0).
*   `glassops-agent generate [patterns...]`:  Initiates the documentation generation process.

    *   `[patterns...]`:  Optional. Specifies file patterns to include in the documentation generation. If no patterns are provided, a default set of patterns will be used.

**Default File Patterns**

If you do not specify any patterns, the agent will analyze files matching the following patterns:

*   `packages/**/*.ts`
*   `packages/**/*.py`
*   `packages/**/*.mjs`
*   `packages/**/*.cls`
*   `packages/**/*.trigger`
*   `packages/**/*.js` (for LWC)
*   `packages/**/*.tf`
*   `packages/**/Dockerfile`
*   `docs/**/*.md`
*   `packages/**/*.yml`
*   `packages/**/*.yaml`
*   `packages/**/*.json`
*   `scripts/**/*.py`
*   `scripts/**/*.ts`
*   `*.py`

**How it Works**

1.  **Root Directory Detection:** The agent automatically determines the root directory of the GlassOps repository. It assumes a specific directory structure relative to its location.
2.  **Generator Initialization:** An instance of the `Generator` class is created, initialized with the repository root directory.
3.  **Pattern Application:** The agent applies the provided (or default) file patterns to identify relevant files within the repository.
4.  **Documentation Generation:** The `Generator` processes the identified files, extracting information and generating documentation.
5.  **Error Handling:**  The agent includes error handling to gracefully manage potential issues during the process. Errors are logged to the console, and the process exits with a non-zero code.

**Example Usage**

*   To generate documentation for all supported file types in the repository:

    ```bash
    glassops-agent generate
    ```

*   To generate documentation only for TypeScript files in the `packages` directory:

    ```bash
    glassops-agent generate packages/**/*.ts
    ```

**Error Reporting**

If an error occurs during execution, the agent will print an error message and stack trace to the console.  A non-zero exit code will also be returned.

**Dependencies**

The agent relies on the `commander` package for command-line argument parsing.

**Internal Components**

*   `Generator` Class: This class encapsulates the core logic for traversing the file system, analyzing files, and generating documentation.
*   Command Definition: The `commander` library is used to define the command-line interface and handle user input.