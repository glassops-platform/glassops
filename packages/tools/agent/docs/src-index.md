---
type: Documentation
domain: agent
origin: packages/tools/agent/src/index.ts
last_modified: 2026-01-26
generated: true
source: packages/tools/agent/src/index.ts
generated_at: 2026-01-26T14:12:49.970Z
hash: f4a190c15881960f1608fc56abe70baa6d65bcbc06ec184b10b72e6bdab05757
---

## GlassOps Agent Documentation

**Overview**

The GlassOps Agent is a command-line tool designed to automatically generate documentation and metadata for GlassOps repositories. It analyzes source code and related files to produce up-to-date documentation, improving code understanding and maintainability.

**Key Features**

*   **Automated Documentation:**  The agent simplifies the documentation process by automatically extracting information from your codebase.
*   **Multi-Language Support:** It supports various programming languages and file types commonly used in GlassOps projects, including TypeScript, Python, JavaScript, Terraform, YAML, JSON, and more.
*   **Configurable Patterns:**  You can specify which files and directories to include in the documentation generation process.
*   **Error Handling:**  The agent provides informative error messages to help diagnose and resolve issues.

**Installation**

This tool is intended to be used as part of a larger GlassOps environment. Installation details are provided within that context.

**Usage**

The agent is invoked from the command line. The primary command is `generate`.

```bash
glassops-agent generate [patterns...]
```

*   `glassops-agent`: The name of the tool.
*   `generate`:  The command to initiate documentation generation.
*   `[patterns...]`: (Optional) One or more file patterns specifying the files to document. If no patterns are provided, a default set of patterns will be used.

**Default File Patterns**

If you do not specify any patterns, the agent will analyze the following file types by default:

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

**Specifying Custom Patterns**

You can provide your own file patterns to control which files are processed. For example:

```bash
glassops-agent generate 'src/**/*.js' 'test/**/*.ts'
```

This command will only analyze JavaScript files in the `src` directory and TypeScript files in the `test` directory.

**Root Directory Detection**

The agent automatically detects the root directory of your GlassOps repository. It assumes the agent’s location is four levels deep from the repository root. This ensures correct file path resolution during documentation generation.

**Error Handling**

If an error occurs during the documentation generation process, the agent will display an error message and stack trace to the console. The process will then exit with a non-zero exit code.

**Example**

To generate documentation for all supported file types in your repository, simply run:

```bash
glassops-agent generate
```

**Version**

The current version of the GlassOps Agent is 1.0.0.