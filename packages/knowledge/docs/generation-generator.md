---
type: Documentation
domain: knowledge
origin: packages/knowledge/generation/generator.py
last_modified: 2026-01-31
generated: true
source: packages/knowledge/generation/generator.py
generated_at: 2026-01-31T08:58:15.131675
hash: 4dbb1763dd2b45db267f1a023ef7aacaa7c88994cdec47293e76aa9f6470efa1
---

## Documentation Generator: Overview

This tool automates the generation of documentation for a codebase. It scans source files, identifies their type, and uses Large Language Models (LLMs) to create documentation based on configurable prompts and adapter-specific logic. The generated documentation can be output alongside the source files or to a dedicated output directory.

### Core Components

**1. `Generator` Class:**

This is the central orchestrator of the documentation generation process. It manages file scanning, adapter selection, LLM interaction, caching, and output writing.

   * **Initialization (`__init__`)**:
      * `root_dir`: Specifies the root directory of the codebase. This is where the scanning process begins.
      * `output_dir` (optional):  Defines the directory where generated documentation files will be placed. If not provided, documentation is created alongside the source files.
      * `llm`: An instance of the `LLMClient` class, responsible for interacting with the LLM.
      * `cache_path`:  The file path for storing a cache of generated documentation to avoid redundant LLM calls.
      * `prompts_path`: The file path for loading prompt configurations from a YAML file.
      * `adapters`: A list of `BaseAdapter` instances, used to handle different file types. The order of adapters is significant; the first adapter that can handle a file type will be used.

   * **Key Methods:**
      * `scan_files(patterns)`:  Scans the `root_dir` for files matching the provided glob patterns, excluding files and directories listed in `IGNORED_DIRS`.
      * `generate_for_file(file_path)`:  Generates documentation for a single file. This involves selecting an appropriate adapter, reading the file content, parsing it into chunks, generating documentation for each chunk using the LLM, and combining the results.
      * `run(patterns)`:  Executes the entire documentation generation process, scanning files, generating documentation, and writing the output.
      * `_load_cache()`: Loads the documentation cache from disk to avoid unnecessary LLM calls.
      * `_save_cache()`: Saves the documentation cache to disk.
      * `_load_prompts()`: Loads prompt configurations from a YAML file.
      * `_get_prompt_for_file(file_path, parsed_content)`: Retrieves the appropriate prompt for a given file based on its extension and configured prompts.
      * `_find_adapter(file_path)`:  Determines the appropriate `BaseAdapter` to use for a given file based on its extension.
      * `_get_output_path(source_path)`: Calculates the output path for the generated documentation file, based on the `output_dir` setting.
      * `_generate_frontmatter(source_path, content)`: Creates YAML frontmatter for the documentation file, including metadata like file origin, last modification date, and a hash of the source content.

**2. `BaseAdapter` Class (and subclasses):**

These classes handle the specifics of parsing and generating documentation for different file types.

   * **Responsibilities:**
      * `can_handle(file_path)`:  Determines if the adapter can handle a given file based on its extension.
      * `parse(file_path, content)`:  Parses the file content into smaller chunks suitable for LLM processing.
      * `get_prompt(file_path, chunk)`:  Generates a prompt for the LLM based on the file type and content chunk.
      * `post_process(file_path, outputs)`:  Combines the LLM outputs into a final documentation string and performs any necessary post-processing.

   * **Available Adapters:**
      * `GoAdapter`: Handles Go files (`.go`).
      * `PythonAdapter`: Handles Python files (`.py`).
      * `TypeScriptAdapter`: Handles TypeScript, JavaScript, and JSX files (`.ts`, `.js`, `.mjs`, `.tsx`, `.jsx`).
      * `YAMLAdapter`: Handles YAML files (`.yml`, `.yaml`).
      * `JSONAdapter`: Handles JSON files (`.json`).
      * `DockerAdapter`: Handles Dockerfiles (`Dockerfile`).
      * `TerraformAdapter`: Handles Terraform files (`.tf`).
      * `ApexAdapter`: Handles Apex class and trigger files (`.cls`, `.trigger`).
      * `LWCAdapter`: Handles Lightning Web Component files.

**3. `LLMClient` Class:**

This class encapsulates the interaction with the Large Language Model. It provides a `generate` method to send prompts to the LLM and receive responses.

**4. `Validator` Class:**

This class provides validation capabilities for the generated documentation, checking for potential issues and providing warnings.

### Key Design Decisions and Patterns

* **Adapter Pattern:** The use of `BaseAdapter` and its subclasses allows for easy extension to support new file types without modifying the core `Generator` class.
* **Caching:** The caching mechanism significantly improves performance by avoiding redundant LLM calls for unchanged files.
* **Configuration via YAML:** Prompts and other configuration settings are loaded from a YAML file, making it easy to customize the documentation generation process.
* **Type Hints:**  Extensive use of type hints (`typing` module) improves code readability and maintainability, and enables static analysis.
* **Pathlib:** The use of `pathlib` provides an object-oriented way to interact with files and directories, making the code more robust and platform-independent.
* **Error Handling:** The code includes `try...except` blocks to handle potential errors during file reading, cache loading, and LLM interaction.
* **Frontmatter:** The inclusion of YAML frontmatter provides structured metadata for each documentation file.