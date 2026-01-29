---
type: Documentation
domain: knowledge
origin: packages/knowledge/generation/generator.py
last_modified: 2026-01-28
generated: true
source: packages/knowledge/generation/generator.py
generated_at: 2026-01-28T22:44:57.734424
hash: 99c77a8a9d1216321beb54743f8b4fbb14ad2ed6401abbd069ac524aa048fd9e
---

## Documentation Generator: Overview

This tool automates the generation of documentation for a codebase. It scans source files, identifies their type, and uses Large Language Models (LLMs) to create documentation based on configurable prompts and adapter-specific logic. The generated documentation can be output alongside the source files or to a dedicated output directory.

### Core Components

**1. `Generator` Class:**

This is the central orchestrator of the documentation generation process. It manages file scanning, adapter selection, LLM interaction, caching, and output writing.

   * **Initialization (`__init__`)**:  The `Generator` is initialized with the root directory of the codebase (`root_dir`) and an optional output directory (`output_dir`). It also initializes an LLM client (`LLMClient`), sets up paths for the documentation cache and prompts file, and loads a list of adapters.
   * **Configuration**: The tool uses a `doc-cache.json` file for caching generated documentation to avoid redundant LLM calls. It also uses a `prompts.yml` file to configure prompts for different file types.
   * **Adapters**: A list of `BaseAdapter` instances are used to handle different file types (e.g., Go, Python, TypeScript). The order of adapters in the list is significant; the first adapter that can handle a file type will be used.
   * **`scan_files(patterns)`**:  This function scans the codebase for files matching the provided glob patterns, excluding files and directories listed in `IGNORED_DIRS`. It supports exclusion patterns using a "!" prefix.
   * **`generate_for_file(file_path)`**: This method generates documentation for a single file. It finds the appropriate adapter, reads the file content, generates documentation using the LLM, and writes the output to a file.
   * **`run(patterns)`**: This method orchestrates the entire documentation generation process. It scans files, iterates through them, generates documentation for each file, and saves the results.

**2. `BaseAdapter` (Abstract Class):**

This is an abstract base class for adapters that handle different file types. Adapters are responsible for:

   * **`can_handle(file_path)`**: Determining if the adapter can handle a given file based on its extension or other criteria.
   * **`parse(file_path, content)`**: Parsing the file content into smaller chunks suitable for LLM processing.
   * **`get_prompt(file_path, chunk)`**:  Generating a prompt for the LLM based on the file path and content chunk.
   * **`post_process(file_path, outputs)`**: Combining the LLM outputs into a final documentation string.

**3. Concrete Adapters (e.g., `GoAdapter`, `PythonAdapter`, `TypeScriptAdapter`):**

These classes implement the `BaseAdapter` interface for specific file types. Each adapter provides its own logic for parsing, prompting, and post-processing.  Specific adapters include: `LWCAdapter`, `GoAdapter`, `PythonAdapter`, `TypeScriptAdapter`, `YAMLAdapter`, `JSONAdapter`, `DockerAdapter`, `TerraformAdapter`, and `ApexAdapter`.

### Key Functions and Processes

* **File Scanning**: The `scan_files` function uses glob patterns to identify files for documentation generation, respecting ignore rules.
* **Adapter Selection**: The `_find_adapter` function iterates through the list of adapters to find the first one that can handle a given file type.
* **Prompt Generation**: The `_get_prompt_for_file` function prioritizes prompts defined in the `prompts.yml` configuration file. If no specific prompt is found, it falls back to the adapter's `get_prompt` method.
* **LLM Interaction**: The `LLMClient` is used to interact with the LLM, sending prompts and receiving generated documentation.
* **Caching**: The tool caches generated documentation in `doc-cache.json` to avoid redundant LLM calls. The cache key is based on the file's relative path and content hash.
* **Output Generation**: The `_get_output_path` function determines the output path for the generated documentation, either mirroring the source structure in a dedicated output directory or placing the documentation alongside the source files.
* **Frontmatter Generation**: The `_generate_frontmatter` function creates YAML frontmatter for each documentation file, including metadata such as file type, origin, last modification date, and content hash.
* **Validation**: The `Validator.validate` function checks the generated documentation for potential issues.

### Type Hints

The code extensively uses type hints (e.g., `root_dir: str`, `output_dir: Optional[str]`) to improve code readability and maintainability. Type hints help to clarify the expected data types for function arguments and return values, enabling static analysis and reducing the risk of runtime errors.

### Design Patterns and Decisions

* **Adapter Pattern**: The use of the `BaseAdapter` interface and concrete adapter classes promotes loose coupling and allows for easy extension to support new file types.
* **Configuration-Driven**: The use of `prompts.yml` allows for flexible configuration of prompts without modifying the code.
* **Caching**: The caching mechanism improves performance by avoiding redundant LLM calls.
* **Clear Separation of Concerns**: The `Generator` class orchestrates the process, while adapters handle file-specific logic, and the `LLMClient` manages LLM interaction.
* **Error Handling**: The code includes basic error handling to gracefully handle file reading errors, cache loading/saving errors, and LLM failures.