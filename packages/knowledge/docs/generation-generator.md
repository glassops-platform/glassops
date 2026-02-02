---
type: Documentation
domain: knowledge
origin: packages/knowledge/generation/generator.py
last_modified: 2026-02-01
generated: true
source: packages/knowledge/generation/generator.py
generated_at: 2026-02-01T19:33:09.545595
hash: 0a0ea08d2da4d3e70cf37f855e65aeb860ba654725e4f99d4fd6d8b26f988e1d
---

## Documentation Generator Documentation

This tool automates the generation of documentation for a codebase. It scans source files, identifies their type, and uses Large Language Models (LLMs) to create documentation based on configurable prompts and adapter-specific logic.

### Core Components

**Generator Class:**

The `Generator` class is the central orchestrator of the documentation generation process. It manages file scanning, adapter selection, LLM interaction, caching, and output writing.

*   **`__init__(root_dir: str, output_dir: Optional[str] = None)`:** Initializes the generator with the root directory of the codebase and an optional output directory.  It also initializes the LLM client, cache path, and prompt path. The `root_dir` argument specifies the base directory for scanning source code. If `output_dir` is provided, generated documentation will be placed there, mirroring the source directory structure. Otherwise, documentation will be placed alongside the source files.
*   **`scan_files(patterns: List[str]) -> List[Path]`:** Scans the codebase for files matching the provided glob patterns. It supports exclusion patterns (prefixed with "!"). The function returns a sorted list of `Path` objects representing the matched files.
*   **`generate_for_file(file_path: Path) -> Optional[str]`:** Generates documentation for a single file. It selects the appropriate adapter based on the file extension, extracts content, interacts with the LLM, and returns the generated documentation string. Returns `None` if documentation generation fails.
*   **`run(patterns: List[str]) -> None`:** Executes the entire documentation generation process. It scans files, iterates through them, generates documentation for each, and writes the output to disk.
*   **`_load_cache() -> None`:** Loads the documentation cache from a JSON file, if it exists. This avoids regenerating documentation for unchanged files.
*   **`_save_cache() -> None`:** Saves the documentation cache to a JSON file.
*   **`_load_prompts() -> None`:** Loads prompts from a YAML configuration file. These prompts guide the LLM in generating documentation.
*   **`_get_prompt_for_file(file_path: Path, parsed_content: str) -> Optional[str]`:** Retrieves the appropriate prompt for a given file based on its extension and configured prompts.
*   **`_should_ignore(path: Path) -> bool`:** Checks if a given path should be ignored during scanning (e.g., `node_modules`, `.git`).
*   **`_find_adapter(file_path: Path) -> Optional[BaseAdapter]`:** Selects the appropriate adapter for a given file based on its extension.
*   **`_get_output_path(source_path: Path) -> Path`:** Determines the output path for the generated documentation file.
*   **`_generate_frontmatter(source_path: Path, content: str) -> str`:** Generates YAML front matter for the documentation file, including metadata like file type, origin, and last modification date.
*   **`_clean_llm_output(output: str) -> str`:** Removes markdown code block wrappers from the LLM output.

**BaseAdapter Class:**

The `BaseAdapter` class is an abstract base class for file-specific adapters. Adapters are responsible for parsing file content, generating prompts, and post-processing the LLM output.

*   **`can_handle(file_path: Path) -> bool`:**  Determines if the adapter can handle a given file based on its extension.
*   **`parse(file_path: Path, content: str) -> List[str]`:** Parses the file content into smaller chunks suitable for LLM processing.
*   **`get_prompt(file_path: Path, chunk: str) -> str`:** Generates a prompt for the LLM based on the file and content chunk.
*   **`post_process(file_path: Path, outputs: List[str]) -> str`:** Post-processes the LLM outputs and combines them into a single documentation string.

**Specific Adapters:**

The tool includes adapters for various file types:

*   `GoAdapter`: Handles Go source files (`.go`).
*   `PythonAdapter`: Handles Python source files (`.py`).
*   `TypeScriptAdapter`: Handles TypeScript, JavaScript, and JSX files (`.ts`, `.js`, `.mjs`, `.tsx`, `.jsx`).
*   `YAMLAdapter`: Handles YAML files (`.yml`, `.yaml`).
*   `JSONAdapter`: Handles JSON files (`.json`).
*   `DockerAdapter`: Handles Dockerfiles.
*   `TerraformAdapter`: Handles Terraform files (`.tf`).
*   `ApexAdapter`: Handles Apex class and trigger files (`.cls`, `.trigger`).
*   `LWCAdapter`: Handles Lightning Web Component files.

**LLMClient Class:**

The `LLMClient` class encapsulates the interaction with the Large Language Model.

*   **`generate(prompt: str) -> Optional[str]`:** Sends a prompt to the LLM and returns the generated response.

**Validator Class:**

The `Validator` class provides validation for the generated documentation.

*   **`validate(content: str, file_path: str) -> List[str]`:** Validates the content against a set of rules and returns a list of validation errors.

### Design Considerations

*   **Adapter Pattern:** The use of adapters allows for easy extension to support new file types without modifying the core generator logic.
*   **Caching:** Caching significantly improves performance by avoiding redundant LLM calls for unchanged files.
*   **Prompt Configuration:** Prompts are loaded from a YAML file, allowing for easy customization and experimentation.
*   **Modular Design:** The code is organized into well-defined classes and functions, promoting maintainability and testability.
*   **Type Hints:** Type hints are used extensively to improve code readability and help catch errors early.
*   **Error Handling:** The code includes error handling to gracefully handle file reading errors, cache loading errors, and LLM failures.
*   **File Exclusion:** The ability to exclude directories and files ensures that irrelevant content is not processed.
*   **Frontmatter Generation:** The inclusion of YAML frontmatter provides metadata for the generated documentation.