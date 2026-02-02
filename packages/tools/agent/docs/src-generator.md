---
type: Documentation
domain: agent
origin: packages/tools/agent/src/generator.ts
last_modified: 2026-02-01
generated: true
source: packages/tools/agent/src/generator.ts
generated_at: 2026-02-01T19:50:14.555604
hash: 7c69d26ea41989b22a3650371168dda792888bcdf63bb35cc47921f990776a6d
---

## GlassOps Agent Documentation

This document details the functionality and usage of the GlassOps Agent, a tool designed to automatically generate documentation from source code. It supports multiple languages and file types, leveraging a large language model (LLM) to create comprehensive and up-to-date documentation.

**Overview**

The Agent scans specified files, parses their content, and uses an LLM to generate documentation. It supports caching to avoid redundant processing and includes validation checks to ensure documentation quality. The generated documentation is written to a designated output directory, organized based on the source code structure.

**Key Components**

*   **GeminiClient:**  Handles communication with the LLM for content generation.
*   **Scanner:**  Locates files matching specified patterns within the project’s root directory.
*   **Adapters:**  Responsible for parsing specific file types (TypeScript, Python, Apex, LWC, Terraform, Dockerfile, YAML, JSON) and preparing them for the LLM.  Currently supported adapters include:
    *   `TSAdapter`: TypeScript files.
    *   `PyAdapter`: Python files.
    *   `ApexAdapter`: Apex files.
    *   `LWCAdapter`: Lightning Web Component files.
    *   `TerraformAdapter`: Terraform files.
    *   `DockerAdapter`: Dockerfile.
    *   `YMLAdapter`: YAML files.
    *   `JSONAdapter`: JSON files.
*   **Validator:**  Performs validation checks on the generated documentation.
*   **DocCache:**  Stores file hashes and metadata to prevent reprocessing unchanged files.
*   **PromptConfig:** Defines the system and user prompts sent to the LLM, allowing customization of the generated documentation style and content.

**Installation & Configuration**

The Agent requires a root directory to operate. A `doc-cache.json` file is created within a `config` directory inside the root directory to store cached file information.  A `prompts.yml` file, containing customizable prompts, is expected to be located in `packages/tools/agent/src/`. If not found, default prompts are used.

**Usage**

You initiate the documentation generation process by calling the `run` method, providing an array of target file patterns.

```typescript
const generator = new Generator('/path/to/your/project');
await generator.run(['**/*.ts', '**/*.py']);
```

This example instructs the Agent to process all TypeScript and Python files within the project.

**Workflow**

1.  **Initialization:** The Agent loads its cache and prompts.
2.  **Scanning:** The Scanner identifies files matching the provided target patterns.
3.  **Processing:** For each file:
    *   The Agent determines the appropriate adapter based on the file extension.
    *   It calculates a hash of the file content.
    *   If the file hasn’t changed (hash matches cache), it’s skipped.
    *   The adapter parses the file content into chunks.
    *   Prompts are constructed using the loaded `PromptConfig` or adapter defaults.
    *   The LLM generates documentation based on the prompts and file chunks.
    *   The adapter post-processes the generated documentation.
    *   Metadata (type, domain, origin, last modified date) is inferred and added as frontmatter.
    *   The documentation is written to the output directory.
    *   The cache is updated with the new file hash and metadata.
4.  **Caching:** The updated cache is saved to disk.

**Output**

Generated documentation is written as Markdown (`.md`) files to a `docs` directory, mirroring the source code structure. Frontmatter is added to each file containing metadata about the documentation. The output directory can be customized based on the project structure.

**Customization**

*   **Prompts:**  Modify the `prompts.yml` file to customize the LLM prompts.
*   **Adapters:** Extend the `AgentAdapter` interface to support additional file types.
*   **Output Directory:** The output directory is determined by the project structure and can be adjusted within the adapter implementations.

**Validation**

The Agent includes a validation step to identify potential issues in the generated documentation. Validation warnings are logged to the console.