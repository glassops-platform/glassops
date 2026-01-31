---
type: Documentation
domain: knowledge
origin: packages/knowledge/generation/adapters/json_adapter.py
last_modified: 2026-01-31
generated: true
source: packages/knowledge/generation/adapters/json_adapter.py
generated_at: 2026-01-31T08:56:06.682029
hash: 51047741270ca1d311ab2b2ad45b14dbe6f1ad33eb711851b4a87c690b2cea1d
---

## JSON Adapter Documentation

This document details the functionality of the JSON Adapter, a component designed for processing JSON files during documentation generation. It is responsible for identifying, parsing, and preparing JSON content for input into a language model to produce technical documentation.

**Module Purpose:**

The primary purpose of this module is to adapt JSON files into a format suitable for documentation generation. This involves determining if a file is a supported JSON file, splitting large files into smaller chunks, and formatting the content with relevant metadata.

**Key Classes:**

*   **`JSONAdapter`**: This class inherits from `BaseAdapter` and implements the specific logic for handling JSON files. It defines how JSON files are identified, parsed into chunks, and formatted for prompt creation.

**Important Functions:**

*   **`can_handle(file_path: Path) -> bool`**: This function determines whether the adapter can process a given file based on its path. It returns `True` if the file has a `.json` extension and is not a common package or configuration file (e.g., `package.json`, `package-lock.json`, `tsconfig.json`). The type hint `Path` indicates that the function expects a file path object as input.
*   **`parse(file_path: Path, content: str) -> List[str]`**: This function takes a file path and its content as input and splits the content into chunks if the content exceeds `TARGET_CHUNK_SIZE` (24000 characters). It returns a list of strings, where each string represents a chunk of the original content, formatted for documentation. The type hint `List[str]` specifies that the function returns a list of strings.
*   **`_format_chunk(file_path: Path, content: str, part: int = None) -> str`**: This is a helper function that formats a single chunk of JSON content. It adds metadata such as the file path and chunk number (if applicable) and wraps the content in a Markdown code block. The `part` parameter, an optional integer, indicates the chunk number. The type hint `str` indicates that the function returns a formatted string.
*   **`get_prompt(file_path: Path, parsed_content: str) -> str`**: This function constructs a prompt for the language model. It includes instructions for the model to act as a technical documentation expert, focusing on explaining the JSON schema or data structure, its architectural role, field requirements, and common use cases. It also includes strict formatting rules and constraints. The type hint `str` indicates that the function returns a string representing the prompt.

**Notable Patterns and Design Decisions:**

*   **Adapter Pattern:** The `JSONAdapter` follows the Adapter pattern, inheriting from `BaseAdapter`. This allows for easy integration of different file types into the documentation generation process.
*   **Chunking:** Large JSON files are split into smaller chunks to avoid exceeding the input limits of the language model. The `TARGET_CHUNK_SIZE` constant defines the maximum size of each chunk.
*   **Type Hints:** The code extensively uses type hints (e.g., `file_path: Path`, `content: str`, `-> List[str]`) to improve code readability and maintainability. Type hints help to clarify the expected data types for function arguments and return values.
*   **Markdown Formatting:** The adapter formats the JSON content within Markdown code blocks to ensure proper rendering in the generated documentation.
*   **Prompt Engineering:** The `get_prompt` function carefully crafts a prompt that guides the language model to generate high-quality technical documentation. It includes specific instructions, constraints, and formatting requirements.
*   **File Exclusion:** The `can_handle` function explicitly excludes common package and configuration files to prevent irrelevant content from being processed.