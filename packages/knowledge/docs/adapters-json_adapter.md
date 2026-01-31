---
type: Documentation
domain: knowledge
origin: packages/knowledge/generation/adapters/json_adapter.py
last_modified: 2026-01-31
generated: true
source: packages/knowledge/generation/adapters/json_adapter.py
generated_at: 2026-01-31T09:51:33.970976
hash: 51047741270ca1d311ab2b2ad45b14dbe6f1ad33eb711851b4a87c690b2cea1d
---

## JSON Adapter Documentation

This document describes the JSON Adapter, a component responsible for processing JSON files as input for documentation generation. It inherits from the `BaseAdapter` class and provides specific logic for handling JSON-formatted content.

**Module Purpose:**

The primary purpose of this module is to adapt JSON files into a format suitable for large language models (LLMs) to generate technical documentation. It handles file identification, content parsing into manageable chunks, and prompt creation for the LLM.

**Key Classes:**

*   **`JSONAdapter`**: This class implements the adapter pattern for JSON files. It extends `BaseAdapter` and provides the specific logic for handling JSON content.

    *   **`TARGET_CHUNK_SIZE`**: A class-level constant set to 24000. This defines the maximum size, in characters, of each chunk the adapter will create when splitting large JSON files.

    *   **`can_handle(file_path: Path) -> bool`**: This method determines if the adapter can process a given file based on its path. It returns `True` if the file has a `.json` extension and is not a `package.json`, `package-lock.json`, or `tsconfig.json` file. These specific files are excluded as they typically contain build or dependency information, not architectural data. The `Path` type hint indicates that the input is a file path object.

    *   **`parse(file_path: Path, content: str) -> List[str]`**: This method takes a file path and its content as input and splits the content into a list of strings (chunks). Each chunk is designed to be within the `TARGET_CHUNK_SIZE` limit. The `List[str]` type hint indicates that the method returns a list of strings. If the content is smaller than the target size, it returns a list containing a single chunk. Otherwise, it iterates through the lines of the content, building chunks until the `TARGET_CHUNK_SIZE` is reached.

    *   **`_format_chunk(file_path: Path, content: str, part: int = None) -> str`**: This is a private helper method that formats a single chunk of JSON content. It adds a header indicating the file path and, if applicable, a part number for chunked files. The `part` argument is optional and defaults to `None`. The method returns a formatted string containing the file path, part number (if any), and the JSON content enclosed in a code block.

    *   **`get_prompt(file_path: Path, parsed_content: str) -> str`**: This method constructs a prompt for the LLM, providing instructions and the parsed JSON content. The prompt instructs the LLM to act as a technical documentation expert and to explain the JSON schema or data structure, focusing on its architectural significance, field requirements, and common use cases. It also includes strict rules for the LLM’s output, prohibiting conversational text, specific words, and the mention of certain names. The `parsed_content` is directly embedded into the prompt.



**Design Decisions and Patterns:**

*   **Adapter Pattern:** The `JSONAdapter` implements the adapter pattern, allowing the documentation generation system to work with different file types without modifying the core logic.
*   **Chunking:** The `parse` method implements a chunking mechanism to handle large JSON files that might exceed the LLM’s input token limit. This ensures that the entire file content can be processed.
*   **Type Hints:** The code extensively uses type hints (e.g., `file_path: Path`, `content: str`, `-> List[str]`) to improve code readability and maintainability. These hints also enable static analysis tools to catch potential errors.
*   **Prompt Engineering:** The `get_prompt` method demonstrates careful prompt engineering to guide the LLM towards generating high-quality technical documentation. The prompt includes specific instructions, constraints, and a clear task definition.