---
type: Documentation
domain: agent
origin: packages/tools/agent/src/generator.ts
last_modified: 2026-01-26
generated: true
source: packages/tools/agent/src/generator.ts
generated_at: 2026-01-26T14:12:32.453Z
hash: 6e3e11b59b4e72025de8b269db9b3d13ad0344f106f135c664f03d15fe00dcf1
---

## GlassOps Agent Documentation

This document details the functionality of the GlassOps Agent, a tool designed to automatically generate documentation from source code. It supports multiple languages and file types, leveraging a Large Language Model (LLM) to create comprehensive and up-to-date documentation.

**Overview**

The Agent scans specified files, parses their content, and uses an LLM to generate documentation. It incorporates a caching mechanism to avoid redundant processing and a validation step to ensure documentation quality. The generated documentation is then written to a designated output directory, organized based on the source code’s structure.

**Key Components**

*   **GeminiClient:**  Handles communication with the LLM for content generation.
*   **Scanner:**  Locates files matching specified patterns within the project’s root directory.
*   **Adapters:**  Interface with different file types (TypeScript, Python, Apex, LWC, Terraform, Dockerfile, YAML, JSON) to parse their content and prepare it for the LLM. Currently supported adapters include:
    *   TSAdapter (TypeScript)
    *   PyAdapter (Python)
    *   ApexAdapter (Apex)
    *   LWCAdapter (LWC)
    *   TerraformAdapter (Terraform)
    *   DockerAdapter (Dockerfile)
    *   YMLAdapter (YAML)
    *   JSONAdapter (JSON)
*   **Validator:** Checks the generated documentation for potential issues.
*   **DocCache:** Stores file hashes and metadata to prevent reprocessing unchanged files.
*   **PromptConfig:** Defines the prompts sent to the LLM, allowing customization of the generated documentation’s style and content.

**Workflow**

1.  **Initialization:** The Agent initializes its components, including the LLM client, scanner, adapters, and cache. It attempts to load prompts from a `prompts.yml` file.
2.  **Scanning:** The scanner identifies files matching the provided target patterns.
3.  **Processing:** For each file:
    *   The Agent determines the appropriate adapter based on the file extension.
    *   It calculates a hash of the file’s content.
    *   If the file hasn’t changed (based on the cache), it’s skipped.
    *   The adapter parses the file’s content into chunks.
    *   Prompts are constructed using configured prompts or adapter defaults.
    *   The LLM generates documentation for each chunk.
    *   The adapter post-processes the generated documentation.
    *   Metadata (type, domain, origin, last modified date) is inferred.
    *   The documentation is written to a file in the designated output directory, with a filename derived from the source file’s name and location.
    *   The cache is updated with the file’s hash and metadata.
4.  **Validation:** The generated documentation is validated. Warnings are logged for any issues.
5.  **Caching:** The updated cache is saved to disk.

**Configuration**

*   **rootDir:** Specifies the root directory of the project.
*   **targetPatterns:** An array of file patterns to scan (e.g., `['**/*.ts', '**/*.py']`).
*   **prompts.yml:** (Optional) A YAML file containing custom prompts for different file types.

**Usage**

You can run the Agent by providing an array of target patterns:

```
agent.run(['**/*.ts', '**/*.py']);
```

**Cache Management**

The Agent uses a cache to store file hashes and metadata. This prevents reprocessing unchanged files, improving performance. The cache is stored in a `doc-cache.json` file within the `config` directory of the project’s root.

**Output**

Generated documentation is written as Markdown (`.md`) files to a designated output directory. The directory structure mirrors the source code’s structure, allowing for easy navigation and organization. Frontmatter is added to each file containing metadata about the documentation.

**Error Handling**

The Agent logs errors encountered during processing. Validation warnings are also logged, providing feedback on potential documentation issues.