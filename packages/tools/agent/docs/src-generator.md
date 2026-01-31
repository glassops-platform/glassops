---
type: Documentation
domain: agent
origin: packages/tools/agent/src/generator.ts
last_modified: 2026-01-31
generated: true
source: packages/tools/agent/src/generator.ts
generated_at: 2026-01-31T09:22:39.545996
hash: 7c69d26ea41989b22a3650371168dda792888bcdf63bb35cc47921f990776a6d
---

## GlassOps Agent Documentation

This document details the functionality and usage of the GlassOps Agent, a tool designed to automatically generate documentation from source code. It supports multiple languages and file types, leveraging a large language model (LLM) to create comprehensive and consistent documentation.

**Overview**

The Agent scans specified files, parses their content, and uses an LLM to generate documentation. It supports caching to avoid redundant processing and includes validation checks to ensure documentation quality. The generated documentation is written to a designated output directory, organized based on the source code structure.

**Key Features**

*   **Multi-Language Support:** Handles TypeScript, Python, Apex, LWC, Terraform, Dockerfile, YAML, JSON, and Markdown files.
*   **LLM Integration:** Uses a GeminiClient to interact with a large language model for documentation generation.
*   **File Scanning:**  Identifies files to process based on provided target patterns.
*   **Caching:** Stores file hashes and generated file information to skip unchanged files, improving performance.
*   **Prompting:** Employs configurable prompts to guide the LLM’s documentation generation process.
*   **Validation:** Checks generated documentation for potential issues.
*   **Automated Organization:** Structures generated documentation based on the source code’s directory structure.
*   **Frontmatter Injection:** Adds metadata to generated Markdown files, including type, domain, origin, and modification dates.

**Architecture**

The Agent consists of the following core components:

*   **Generator:** The main class responsible for orchestrating the documentation generation process.
*   **Scanner:** Locates files matching specified patterns within the project’s root directory.
*   **Adapters:** Interface with different file types, parsing their content and preparing it for the LLM. Supported adapters include:
    *   TSAdapter (TypeScript)
    *   PyAdapter (Python)
    *   ApexAdapter (Apex)
    *   LWCAdapter (LWC)
    *   TerraformAdapter (Terraform)
    *   DockerAdapter (Dockerfile)
    *   YMLAdapter (YAML)
    *   JSONAdapter (JSON)
*   **GeminiClient:**  Handles communication with the LLM.
*   **Validator:**  Performs validation checks on the generated documentation.

**Usage**

You interact with the Agent by providing an array of target file patterns. The Agent will:

1.  Load cached file information to avoid reprocessing unchanged files.
2.  Load prompts from a `prompts.yml` file (if available).
3.  Scan the project for files matching the provided patterns.
4.  For each file:
    *   Determine the appropriate adapter based on the file extension.
    *   Read the file content and calculate its hash.
    *   If the file has not changed (based on hash comparison), skip it.
    *   Parse the file content using the selected adapter.
    *   Generate a prompt for the LLM, incorporating the file content and any configured prompts.
    *   Send the prompt to the LLM and receive the generated documentation.
    *   Post-process the generated documentation.
    *   Write the documentation to a Markdown file in the appropriate output directory.
    *   Update the cache with the new file hash and generated file information.
5.  Save the updated cache.

**Configuration**

*   **`prompts.yml`:**  A YAML file containing prompts for different file types. This file allows you to customize the instructions given to the LLM. If not found, default prompts are used.
*   **`rootDir`:** The root directory of the project. This is specified during the Generator’s instantiation.
*   **`cachePath`:** The path to the cache file, located within the `config` directory of the `rootDir`.

**Output**

The Agent generates Markdown files (`.md`) containing the documentation. These files are organized within the `docs` directory, mirroring the source code’s directory structure.  Generated files include frontmatter containing metadata about the documentation.

**Error Handling**

The Agent logs errors encountered during processing.  Validation warnings are also logged, but do not prevent file creation.