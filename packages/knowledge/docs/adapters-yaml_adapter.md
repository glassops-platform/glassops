---
type: Documentation
domain: knowledge
last_modified: 2026-02-02
generated: true
source: packages/knowledge/adapters/yaml_adapter.py
generated_at: 2026-02-02T22:26:43.857070
hash: 9005d910fabe6895fe8632f553933a415eb60b5aecc0259ab4dcd86f82609b1b
---

## YAML Adapter Documentation

This document details the functionality of the YAML Adapter, a component designed for processing YAML configuration files within a documentation generation pipeline. It handles parsing, validation, and formatting of YAML content to prepare it for documentation by a language model.

**Module Responsibilities:**

The primary responsibility of this module is to read YAML files, split them into manageable chunks if they exceed a defined size limit, validate their syntax, and format them into a prompt suitable for a language model. This adapter ensures that large YAML files can be processed without exceeding the context window limitations of the language model.

**Key Classes:**

*   **`YAMLAdapter`**: This class inherits from `BaseAdapter` and implements the specific logic for handling YAML files. It encapsulates the parsing, validation, and formatting processes.

    *   **`TARGET_CHUNK_SIZE`**: A class-level constant set to 24000, defining the maximum size (in characters) of a single chunk of YAML content. This value is used to split large files into smaller, more manageable pieces.

**Important Functions:**

*   **`can_handle(file_path: Path) -> bool`**: This function determines whether the adapter can process a given file based on its extension. It returns `True` if the file path has a `.yml` or `.yaml` extension, and `False` otherwise. The `file_path` argument is a `Path` object representing the file's location.

*   **`parse(file_path: Path, content: str) -> List[str]`**: This function parses the YAML content of a file and splits it into chunks if necessary. It takes the `file_path` (a `Path` object) and the file `content` (a string) as input. If the content's length is within the `TARGET_CHUNK_SIZE`, it returns a list containing the entire content formatted as a single chunk. Otherwise, it splits the content into multiple chunks, ensuring no chunk exceeds the size limit. Each chunk is then formatted using the `_format_chunk` function. The function returns a list of strings, where each string represents a chunk of YAML content.

*   **`validate_content(content: str) -> List[str]`**: This function validates the YAML syntax of the provided content. It attempts to parse the `content` (a string) using `yaml.safe_load()`. If the parsing is successful, it returns an empty list, indicating no errors. If a `yaml.YAMLError` occurs, it catches the exception and returns a list containing an error message describing the syntax error.

*   **`_format_chunk(file_path: Path, content: str, part: int = None) -> str`**: This is a helper function that formats a single chunk of YAML content into a string suitable for inclusion in a prompt. It takes the `file_path` (a `Path` object), the `content` (a string), and an optional `part` number (an integer) as input. It constructs a string that includes the file path, an optional part number (e.g., " (Part 1)"), and the YAML content enclosed in a code block.

*   **`get_prompt(file_path: Path, parsed_content: str) -> str`**: This function generates a prompt for the language model, incorporating the parsed YAML content. It takes the `file_path` (a `Path` object) and the `parsed_content` (a string) as input. The prompt instructs the language model to act as a DevOps engineer and technical writer, documenting the provided YAML configuration. It emphasizes the need for valid Markdown output and explicitly prohibits certain phrasing and content.

**Type Hints:**

The code extensively uses type hints (e.g., `file_path: Path`, `content: str`, `-> List[str]`). These hints improve code readability and allow for static analysis, helping to catch potential errors during development. They clearly define the expected data types for function arguments and return values.

**Design Decisions and Patterns:**

*   **Adapter Pattern:** The `YAMLAdapter` class follows the Adapter pattern, allowing the system to work with YAML files without needing to know the specifics of the YAML format. This promotes loose coupling and makes it easy to add support for other configuration file types in the future.
*   **Chunking:** The implementation of chunking addresses the limitation of language model context windows. By splitting large YAML files into smaller chunks, the adapter ensures that the entire file can be processed without exceeding the model's capacity.
*   **Validation:** The inclusion of YAML syntax validation helps to prevent errors and ensures that the language model receives valid input.
*   **Prompt Engineering:** The `get_prompt` function demonstrates careful prompt engineering, providing clear instructions to the language model and specifying desired output characteristics.