---
type: Documentation
domain: knowledge
last_modified: 2026-02-02
generated: true
source: packages/knowledge/adapters/json_adapter.py
generated_at: 2026-02-02T22:25:13.537416
hash: fc8972e787717de8b86255c2cce0b6cb93a5c35df2759316d3e75be142b6c9d3
---

## JSON Adapter Documentation

This document describes the JSON Adapter, a component designed to process JSON files for documentation generation. It inherits from the `BaseAdapter` class and provides specific functionality for handling JSON-formatted content.

**Module Purpose:**

The primary responsibility of this module is to read, parse, validate, and format JSON files into chunks suitable for processing by a language model to generate technical documentation. It filters out common project files like `package.json` and `tsconfig.json` to avoid irrelevant documentation.

**Key Classes:**

*   **`JSONAdapter`**: This class implements the adapter pattern for JSON files. It extends `BaseAdapter` and provides methods for determining if a file can be handled, parsing its content, validating its structure, and formatting it for input to a language model.

**Important Functions:**

*   **`can_handle(file_path: Path) -> bool`**:
    This function determines whether the adapter can process a given file based on its path. It returns `True` if the file has a `.json` extension and is not one of the excluded files (`package.json`, `package-lock.json`, `tsconfig.json`). The `file_path` argument is a `Path` object representing the file's location.
*   **`parse(file_path: Path, content: str) -> List[str]`**:
    This function parses the content of a JSON file and splits it into chunks if the content exceeds `TARGET_CHUNK_SIZE` (24000 characters). It iterates through the lines of the content, building chunks until the `TARGET_CHUNK_SIZE` is reached. The `file_path` argument is a `Path` object, and `content` is a string containing the file's content. The function returns a list of strings, where each string represents a chunk of the file content.
*   **`validate_content(content: str) -> List[str]`**:
    This function validates the JSON syntax of the provided content. It attempts to parse the content using `json.loads()`. If the parsing is successful, it returns an empty list, indicating no errors. If a `json.JSONDecodeError` occurs, it returns a list containing an error message with details about the syntax error, including the error message and line number. The `content` argument is a string containing the JSON content.
*   **`_format_chunk(file_path: Path, content: str, part: int = None) -> str`**:
    This is a helper function that formats a chunk of JSON content into a string suitable for input to a language model. It includes the file path and an optional part number in the formatted string. The `file_path` argument is a `Path` object, `content` is the chunk of JSON content, and `part` is an optional integer representing the chunk number. The function returns a formatted string.
*   **`get_prompt(file_path: Path, parsed_content: str) -> str`**:
    This function constructs a prompt for a language model, instructing it to document the provided JSON schema or data structure. The prompt includes specific instructions regarding the desired output format (Markdown), content focus (data representation, required/optional fields, use cases), and restrictions (no conversational text, no mention of specific names, no emojis, and avoidance of certain words). The `file_path` argument is a `Path` object, and `parsed_content` is the JSON content to be documented. The function returns a string containing the prompt.

**Type Hints:**

The code extensively uses type hints (e.g., `file_path: Path`, `content: str`, `-> List[str]`). These hints improve code readability and allow for static analysis, helping to catch potential errors during development.

**Notable Patterns and Design Decisions:**

*   **Adapter Pattern:** The `JSONAdapter` class implements the adapter pattern, allowing the documentation generation process to work with different file types in a consistent manner.
*   **Chunking:** The `parse` function splits large JSON files into smaller chunks to avoid exceeding the input limits of the language model.
*   **Error Handling:** The `validate_content` function provides basic JSON syntax validation and returns informative error messages.
*   **Prompt Engineering:** The `get_prompt` function carefully crafts a prompt to guide the language model towards generating high-quality documentation.
*   **File Exclusion:** The `can_handle` function excludes common project files that are not relevant for documentation.