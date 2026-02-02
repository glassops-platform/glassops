---
type: Documentation
domain: knowledge
origin: packages/knowledge/generation/adapters/json_adapter.py
last_modified: 2026-02-01
generated: true
source: packages/knowledge/generation/adapters/json_adapter.py
generated_at: 2026-02-01T19:31:09.136226
hash: 51047741270ca1d311ab2b2ad45b14dbe6f1ad33eb711851b4a87c690b2cea1d
---

## JSON Adapter Documentation

This document describes the JSON Adapter, a component responsible for processing JSON files as input for documentation generation. It inherits from the `BaseAdapter` class and provides specific logic for handling JSON-formatted content.

**Module Purpose:**

The primary responsibility of this module is to read JSON files, split them into manageable chunks if they exceed a defined size limit, and format these chunks for inclusion in prompts sent to a language model for documentation generation. It also defines a prompt template instructing the language model on how to document the JSON content.

**Key Classes:**

*   **`JSONAdapter`**: This class implements the adapter pattern for JSON files. It determines if a file should be handled by this adapter, parses the file content into chunks, formats those chunks, and constructs a prompt for the language model.

    *   `TARGET_CHUNK_SIZE`: A constant integer defining the maximum size (in characters) of each chunk. Currently set to 24000.

**Important Functions:**

*   **`can_handle(file_path: Path) -> bool`**: This function determines whether the adapter can process a given file based on its path. It returns `True` if the file has a `.json` extension and is not one of the excluded files (`package.json`, `package-lock.json`, `tsconfig.json`). The `file_path` argument is a `Path` object representing the file's location.

*   **`parse(file_path: Path, content: str) -> List[str]`**: This function takes the file path and its content as input and splits the content into a list of strings (chunks). If the content is smaller than `TARGET_CHUNK_SIZE`, it returns a list containing the entire content formatted as a single chunk. Otherwise, it splits the content into multiple chunks, ensuring no chunk exceeds the size limit. The `file_path` argument is a `Path` object, and `content` is a string. The return value is a `List` of strings, where each string represents a chunk of the original content.

*   **`_format_chunk(file_path: Path, content: str, part: int = None) -> str`**: This is a private helper function that formats a single chunk of content. It adds a header indicating the file path and, if applicable, the chunk number. The `file_path` argument is a `Path` object, `content` is the string representing the chunk, and `part` is an optional integer indicating the chunk number. The function returns a formatted string.

*   **`get_prompt(file_path: Path, parsed_content: str) -> str`**: This function constructs the prompt that will be sent to the language model. It includes instructions for the model to act as a technical documentation expert, focusing on explaining the JSON schema or data structure's purpose, fields, and use cases. It also includes strict rules for the model's output, prohibiting conversational text, specific words, and the mention of certain names. The `file_path` argument is a `Path` object, and `parsed_content` is the string representing the JSON content. The function returns a string containing the prompt.

**Type Hints:**

The code extensively uses type hints (e.g., `file_path: Path`, `content: str`, `-> List[str]`). These hints improve code readability and allow for static analysis, helping to catch potential errors during development. They clearly define the expected data types for function arguments and return values.

**Design Decisions and Patterns:**

*   **Adapter Pattern:** The `JSONAdapter` class implements the adapter pattern, allowing the documentation generation process to work with different file types without modifying the core logic.
*   **Chunking:** The `parse` function implements a chunking mechanism to handle large JSON files that might exceed the language model's input limits. This ensures that the entire file content can be processed.
*   **Prompt Engineering:** The `get_prompt` function demonstrates careful prompt engineering, providing clear instructions and constraints to the language model to ensure high-quality documentation output.
*   **String Formatting:** The use of f-strings for formatting strings improves readability and maintainability.