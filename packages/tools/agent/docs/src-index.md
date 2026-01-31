---
type: Documentation
domain: agent
origin: packages/tools/agent/src/index.ts
last_modified: 2026-01-31
generated: true
source: packages/tools/agent/src/index.ts
generated_at: 2026-01-31T09:23:01.433678
hash: eab86e8c516702845a7338f7e46e27a9f0d332309fbcd22b01d200c37c38bae5
---

## GlassOps Agent Documentation

**Overview**

This tool is an automated agent designed to produce documentation and metadata for GlassOps projects. It analyzes source code and related files within a repository to create up-to-date documentation assets.

**Purpose**

The agent simplifies the process of maintaining accurate and comprehensive documentation, reducing the manual effort required for project onboarding, knowledge sharing, and overall maintainability.

**Installation & Execution**

The agent is designed to be executed from the command line. Ensure Node.js and npm are installed on your system. Installation is typically handled through npm package management.

**Command-Line Interface**

The primary command is `glassops-agent`.

*   `glassops-agent generate [patterns...]`

    This command initiates the documentation generation process.

    *   `[patterns...]` (Optional):  A space-separated list of file patterns to include in the documentation generation. If no patterns are provided, a default set of patterns will be used (see “Default File Patterns” below).  You can specify patterns like `'packages/**/*.ts'` to focus on TypeScript files within the `packages` directory.

**Default File Patterns**

If no patterns are specified during the `generate` command, the agent will process files matching the following patterns:

*   `packages/**/*.ts`
*   `packages/**/*.py`
*   `packages/**/*.mjs`
*   `packages/**/*.cls`
*   `packages/**/*.trigger`
*   `packages/**/*.js`
*   `packages/**/*.tf`
*   `packages/**/Dockerfile`
*   `docs/**/*.md`
*   `packages/**/*.yml`
*   `packages/**/*.yaml`
*   `packages/**/*.json`
*   `scripts/**/*.py`
*   `scripts/**/*.ts`
*   `*.py`

**Operation**

1.  **Root Directory Detection:** The agent automatically determines the root directory of the GlassOps project. It assumes a project structure where the agent’s execution location is four levels deep from the root.
2.  **Pattern Application:** The agent applies the provided (or default) file patterns to identify relevant files within the project.
3.  **Documentation Generation:** The agent analyzes the identified files and generates documentation. The specific output format and content are managed internally by the agent’s core logic.
4.  **Error Handling:**  If an error occurs during the process, the agent will display an error message to the console and exit with a non-zero status code.

**Dependencies**

The agent relies on the following dependencies:

*   Commander: For parsing command-line arguments.
*   Built-in Node.js modules: `url`, `path`.

**Maintainability**

I am designed for ongoing maintenance and improvement. Updates will include new features, bug fixes, and support for additional file types. We are committed to providing a reliable and effective documentation solution for GlassOps projects.

**Error Reporting**

If you encounter issues or have suggestions for improvement, please report them through the project’s issue tracker. Include detailed information about the problem, including the command used, any error messages, and your project’s structure.