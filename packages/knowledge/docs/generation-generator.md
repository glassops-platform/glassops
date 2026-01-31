---
type: Documentation
domain: knowledge
origin: packages/knowledge/generation/generator.py
last_modified: 2026-01-31
generated: true
source: packages/knowledge/generation/generator.py
generated_at: 2026-01-31T09:53:49.868676
hash: 0a0ea08d2da4d3e70cf37f855e65aeb860ba654725e4f99d4fd6d8b26f988e1d
---

## Documentation Generator Overview

This tool automates the generation of documentation for a codebase. It scans source files, identifies their type, and uses Large Language Models (LLMs) to create documentation based on configurable prompts and adapter-specific logic. The generated documentation can be output alongside the source files or to a dedicated output directory.

### Core Components

**1. Generator Class:**

The `Generator` class is the central orchestrator. It manages the entire documentation generation process, from file scanning to output writing.

*   **Initialization (`__init__`)**:
    *   `root_dir`: The base directory of the codebase. This is a required argument.
    *   `output_dir`: An optional directory where generated documentation files will be placed. If not provided, documentation is created alongside the source files.
    *   Initializes an `LLMClient` for interacting with the language model.
    *   Loads configuration data, including prompts and a cache of previously generated documentation.
    *   Instantiates a list of `BaseAdapter` instances, which handle different file types. The order of adapters is significant; the first matching adapter is used.

*   **Scanning (`scan_files`)**:
    *   Takes a list of glob patterns as input.
    *   Recursively searches for files matching the patterns within the `root_dir`.
    *   Excludes files and directories specified in the `IGNORED_DIRS` set.
    *   Returns a sorted list of `Path` objects representing the matched files.

*   **File Processing (`generate_for_file`)**:
    *   Takes a `Path` object representing a source file.
    *   Determines the appropriate `BaseAdapter` for the file type.
    *   Reads the file content.
    *   Checks the cache to see if documentation for this file already exists and is up-to-date.
    *   Parses the file content into chunks using the adapter.
    *   Generates documentation for each chunk using the LLM and prompts.
    *   Combines the generated documentation for all chunks.
    *   Returns the combined documentation string.

*   **Running the Generator (`run`)**:
    *   Takes a list of glob patterns as input.
    *   Loads the documentation cache and prompts.
    *   Scans for files matching the patterns.
    *   Iterates through the found files, calling `generate_for_file` for each one.
    *   Writes the generated documentation to the appropriate output path.
    *   Updates the documentation cache.
    *   Provides summary statistics on the generation process.

**2. Adapters (`knowledge.generation.adapters`)**:

Adapters are responsible for handling specific file types. They provide the following functionality:

*   `can_handle(file_path)`: Determines if the adapter can process a given file based on its extension or other characteristics.
*   `parse(file_path, content)`:  Parses the file content into smaller chunks suitable for LLM processing.
*   `get_prompt(file_path, chunk)`:  Generates a prompt for the LLM based on the file type and content chunk.
*   `post_process(file_path, outputs)`:  Performs any necessary post-processing on the LLM output, such as combining chunks or formatting the documentation.

Available adapters include: `GoAdapter`, `PythonAdapter`, `TypeScriptAdapter`, `YAMLAdapter`, `JSONAdapter`, `DockerAdapter`, `TerraformAdapter`, `ApexAdapter`, and `LWCAdapter`.

**3. LLM Client (`knowledge.llm.client`)**:

The `LLMClient` handles communication with the Large Language Model. It provides a `generate(prompt)` method that sends a prompt to the LLM and returns the generated text.

**4. Validator (`knowledge.generation.validator`)**:

The `Validator` class validates the generated documentation against a set of rules. It reports any warnings or errors found during validation.

### Key Data Structures

*   **`IGNORED_DIRS`**: A set of directory names that are excluded from the file scanning process.
*   **`EXTENSION_TO_PROMPT_KEY`**: A dictionary that maps file extensions to prompt keys used for selecting the appropriate prompt configuration.
*   **`cache`**: A dictionary that stores previously generated documentation to avoid redundant LLM calls. The keys are file paths, and the values are dictionaries containing the generated documentation, a hash of the file content, and a timestamp.
*   **`prompts`**: A dictionary that stores prompt configurations loaded from a YAML file.

### Design Considerations

*   **Adapter Pattern**: The use of adapters allows the tool to easily support new file types without modifying the core `Generator` class.
*   **Caching**: Caching significantly improves performance by avoiding redundant LLM calls for unchanged files.
*   **Prompt Configuration**:  Externalizing prompts in a YAML file allows for easy customization and experimentation.
*   **Modular Design**: The separation of concerns into distinct classes (Generator, Adapters, LLMClient, Validator) promotes maintainability and testability.
*   **Type Hints**: Extensive use of type hints improves code readability and helps prevent errors.
*   **Pathlib**: Using `Pathlib` provides a more object-oriented and platform-independent way to work with file paths.