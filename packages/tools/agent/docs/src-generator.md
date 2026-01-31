---
type: Documentation
domain: agent
origin: packages/tools/agent/src/generator.ts
last_modified: 2026-01-31
generated: true
source: packages/tools/agent/src/generator.ts
generated_at: 2026-01-31T10:20:00.344163
hash: 7c69d26ea41989b22a3650371168dda792888bcdf63bb35cc47921f990776a6d
---

## GlassOps Agent Documentation

This document details the functionality and usage of the GlassOps Agent, a tool designed to automatically generate documentation from source code and other file types.

**Overview**

The Agent analyzes files matching specified patterns, leverages Large Language Models (LLMs) to create documentation, and saves the generated content to a designated output directory. It incorporates caching to avoid redundant processing and validation to ensure documentation quality.

**Key Components**

*   **GeminiClient:**  Handles communication with the LLM for content generation.
*   **Scanner:**  Identifies files within the project that match user-defined patterns.
*   **Adapters:**  Responsible for parsing specific file types (TypeScript, Python, Apex, LWC, Terraform, Dockerfile, YAML, JSON) and preparing them for LLM processing.  Currently supported adapters include:
    *   TSAdapter
    *   PyAdapter
    *   ApexAdapter
    *   LWCAdapter
    *   TerraformAdapter
    *   DockerAdapter
    *   YMLAdapter
    *   JSONAdapter
*   **Validator:** Checks generated documentation for potential issues.
*   **DocCache:** Stores file hashes and metadata to prevent reprocessing unchanged files.
*   **PromptConfig:** Defines prompts used to instruct the LLM during documentation generation.

**Functionality**

1.  **Initialization:**
    *   The Agent is initialized with a root directory.
    *   It loads a cache of previously processed files from a `doc-cache.json` file located in the `config` directory within the root directory. If the cache file does not exist, it starts with an empty cache.
    *   It loads prompts from a `prompts.yml` file located in `packages/tools/agent/src`. If the file is not found, it defaults to internal logic.
    *   It instantiates a set of adapters for supported file types.

2.  **File Processing:**
    *   The Agent scans the project for files matching the provided target patterns.
    *   For each file:
        *   It determines the appropriate adapter based on the file extension.
        *   It calculates a SHA256 hash of the file content.
        *   If the file hash matches a cached entry, the file is skipped.
        *   The adapter parses the file content into chunks.
        *   For each chunk, a prompt is constructed using the loaded prompts or adapter-specific defaults.
        *   The prompt is sent to the LLM via the GeminiClient to generate documentation.
        *   The generated documentation is post-processed by the adapter.
        *   Metadata (type, domain, origin, last modified date) is inferred and added as frontmatter.
        *   The documentation is written to a `.md` file in the appropriate output directory.
        *   The file hash and metadata are added to the cache.

3.  **Caching:**
    *   The Agent maintains a cache of processed files to avoid redundant work.
    *   The cache is stored in a `doc-cache.json` file.
    *   The cache is loaded at startup and saved after each run.

4.  **Validation:**
    *   Generated documentation is validated to identify potential issues.
    *   Validation warnings are logged to the console.

**Usage**

You can run the Agent by calling the `run` method with an array of target patterns. For example:

```typescript
const generator = new Generator('/path/to/your/project');
await generator.run(['**/*.ts', '**/*.py']);
```

This will process all TypeScript and Python files within the project.

**Configuration**

*   **`prompts.yml`:**  Customize the prompts used to generate documentation. This file should be placed in `packages/tools/agent/src`.
*   **`rootDir`:** The root directory of your project, provided during instantiation of the `Generator` class.
*   **Target Patterns:**  Specify the files to process using glob patterns.

**Output**

Generated documentation is saved as Markdown files (`.md`) in the `docs` directory within the project or in package-specific `docs` directories if the source file resides within a package. The output directory structure mirrors the source file structure.