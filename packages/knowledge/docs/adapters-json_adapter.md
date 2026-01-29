---
type: Documentation
domain: knowledge
origin: packages/knowledge/generation/adapters/json_adapter.py
last_modified: 2026-01-28
generated: true
source: packages/knowledge/generation/adapters/json_adapter.py
generated_at: 2026-01-28T22:42:35.624454
hash: 51047741270ca1d311ab2b2ad45b14dbe6f1ad33eb711851b4a87c690b2cea1d
---

## JSON Adapter Documentation

This document describes the JSON Adapter, a component responsible for processing JSON files as input for documentation generation. It handles parsing, chunking, and prompt creation for large JSON files to facilitate effective documentation using language models.

**Module Purpose:**

The primary responsibility of this module is to adapt JSON files into a format suitable for documentation generation. This involves determining if a file is a supported JSON file, splitting large files into smaller chunks, and constructing a prompt that instructs a language model to document the JSON content.

**Key Classes:**

*   **`JSONAdapter`**: This class inherits from `BaseAdapter` and implements the specific logic for handling JSON files. It defines how to identify JSON files, parse their content, and prepare prompts for documentation.

**Important Functions:**

*   **`can_handle(file_path: Path) -> bool`**: This function determines whether the adapter can process a given file based on its path. It returns `True` if the file has a `.json` extension and is not a common package or configuration file (like `package.json`, `package-lock.json`, or `tsconfig.json`). The type hint `Path` indicates that the function expects a file path object as input and returns a boolean value.
*   **`parse(file_path: Path, content: str) -> List[str]`**: This function takes the file path and content of a JSON file as input and splits the content into chunks if the file is larger than `TARGET_CHUNK_SIZE` (24000 characters). It returns a list of strings, where each string represents a chunk of the JSON content. The type hint `List[str]` indicates that the function returns a list of strings. The function ensures that chunks are created at logical breaks (newline characters) to avoid breaking JSON structures.
*   **`_format_chunk(file_path: Path, content: str, part: int = None) -> str`**: This is a helper function that formats a single chunk of JSON content. It adds metadata such as the file path and chunk number (if applicable) to the beginning of the chunk, and wraps the content in a Markdown code block. The `part` parameter is optional and allows specifying a chunk number for larger files. The type hint `str` indicates that the function returns a formatted string.
*   **`get_prompt(file_path: Path, parsed_content: str) -> str`**: This function constructs a prompt that will be sent to a language model to generate documentation for the given JSON content. The prompt includes instructions to act as a technical documentation expert, focus on explaining the data's purpose, fields, and use cases, and strict formatting rules (Markdown only, no conversational text, and exclusion of specific terms). The type hint `str` indicates that the function returns a string representing the prompt.

**Type Hints:**

The code extensively uses type hints (e.g., `file_path: Path`, `content: str`, `-> List[str]`). These hints improve code readability and maintainability by explicitly specifying the expected data types for function arguments and return values. They also enable static analysis tools to catch type-related errors early in the development process.

**Notable Patterns and Design Decisions:**

*   **Adapter Pattern:** The `JSONAdapter` class follows the Adapter pattern, allowing it to interface with a generic documentation generation pipeline while providing specific handling for JSON files. This promotes flexibility and extensibility.
*   **Chunking:** The `parse` function implements a chunking mechanism to handle large JSON files that might exceed the input limits of a language model. This ensures that even very large files can be processed.
*   **Prompt Engineering:** The `get_prompt` function demonstrates careful prompt engineering to guide the language model towards generating high-quality, focused documentation. The prompt includes specific instructions, constraints, and formatting requirements.
*   **File Exclusion:** The `can_handle` function explicitly excludes common package and configuration files to avoid processing irrelevant content.