---
type: Documentation
domain: agent
origin: packages/tools/agent/src/index.ts
last_modified: 2026-01-31
generated: true
source: packages/tools/agent/src/index.ts
generated_at: 2026-01-31T10:20:14.426234
hash: eab86e8c516702845a7338f7e46e27a9f0d332309fbcd22b01d200c37c38bae5
---

## GlassOps Agent Documentation

**Overview**

This tool is an AI-powered agent designed to automatically generate documentation and metadata for GlassOps projects. It analyzes source code and related files to produce up-to-date and informative documentation.

**Key Features**

*   **Automated Documentation:** Simplifies the process of creating and maintaining documentation.
*   **Multi-Language Support:** Supports TypeScript, Python, JavaScript, Terraform, YAML, JSON, Markdown, and other common file types.
*   **Configurable Patterns:** Allows you to specify which files and directories should be included in the documentation generation process.
*   **Error Handling:** Provides informative error messages to assist with troubleshooting.

**Installation**

This tool is designed to be used as a command-line application. Installation instructions will be provided with the distribution package.

**Usage**

The primary command is `glassops-agent generate`.

**Command: `generate [patterns...]`**

This command initiates the documentation generation process.

*   **Description:** Generates documentation for the repository.
*   **Arguments:**
    *   `patterns` (optional): One or more file patterns to include in the documentation generation. If no patterns are provided, a default set of patterns will be used. Patterns are glob-style.
*   **Example:**

    *   `glassops-agent generate` – Generates documentation using the default file patterns.
    *   `glassops-agent generate 'src/**/*.ts' 'docs/**/*.md'` – Generates documentation only for TypeScript files in the `src` directory and Markdown files in the `docs` directory.

**Default File Patterns**

If you do not specify any patterns, the following file types will be included by default:

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

**Technical Details**

The agent determines the project root directory automatically by traversing up four levels from its own location. This ensures accurate documentation generation even when the agent is executed from within a nested directory structure.

**Error Handling**

If an error occurs during the documentation generation process, the agent will display an error message to the console, including the error stack trace if available. The process will then exit with a non-zero exit code.

**Future Enhancements**

We plan to add support for additional file types and customization options in future releases. We are also exploring integration with other documentation tools and platforms.

**Support**

For questions or issues, please consult the project’s support channels.