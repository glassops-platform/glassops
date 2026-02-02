---
type: Documentation
domain: knowledge
origin: packages/knowledge/generation/adapters/apex_adapter.py
last_modified: 2026-02-01
generated: true
source: packages/knowledge/generation/adapters/apex_adapter.py
generated_at: 2026-02-01T19:30:02.923182
hash: 7fd9d575b3ebb4b889c9e55dcb24940b5cc37ec212bc40d9dbb3868136103744
---

## Apex Adapter Documentation

This document details the functionality of the Apex Adapter, a component designed for generating documentation from Salesforce Apex code. It is part of a larger system for automated documentation generation across various knowledge sources.

**Module Purpose and Responsibilities**

The Apex Adapter is responsible for identifying, parsing, and preparing Apex code files (.cls and .trigger) for documentation generation. It handles splitting large files into smaller chunks to accommodate limitations of large language models and formats the code for inclusion in a prompt.

**Key Classes and Their Roles**

*   **ApexAdapter:** This class inherits from the `BaseAdapter` class and implements the specific logic for handling Apex files. It determines if a file can be processed, parses the file content into chunks, formats those chunks, and constructs a prompt for a language model to generate documentation.

**Important Functions and Their Behavior**

*   **`can_handle(file_path: Path) -> bool`**: This function checks if the adapter can process a given file based on its extension. It returns `True` if the file extension is ".cls" (Apex class) or ".trigger" (Apex trigger), and `False` otherwise. The `file_path` argument is a `Path` object representing the file's location.
*   **`parse(file_path: Path, content: str) -> List[str]`**: This function takes the file path and content as input and splits the content into smaller chunks if the content exceeds `TARGET_CHUNK_SIZE` (24000 characters). It returns a list of strings, where each string represents a chunk of Apex code. The function ensures that chunks are created at logical breaks (lines) and avoids splitting code in the middle of a line. If the content is smaller than the target size, it returns a list containing the entire content as a single chunk.
*   **`_format_chunk(file_path: Path, content: str, part: int = None) -> str`**: This private helper function formats a single chunk of Apex code. It adds metadata such as the file path, file type (Apex Class or Apex Trigger), and a part number (if the content was split into multiple chunks). The formatted chunk is returned as a string, ready to be included in a prompt.
*   **`get_prompt(file_path: Path, parsed_content: str) -> str`**: This function constructs a prompt for a language model. The prompt instructs the model to act as a Salesforce architect and document the provided Apex code. It specifies the desired documentation elements (purpose, methods, governor limits, integration points, test coverage) and includes strict formatting rules to ensure the output is valid Markdown and adheres to project guidelines. The `parsed_content` argument is a string containing the Apex code chunk.

**Type Hints and Their Significance**

The code extensively uses type hints (e.g., `file_path: Path`, `content: str`, `-> List[str]`). These hints improve code readability, maintainability, and allow for static analysis to catch potential errors early in the development process. They clearly define the expected data types for function arguments and return values.

**Notable Patterns or Design Decisions**

*   **Adapter Pattern:** The `ApexAdapter` follows the Adapter pattern, inheriting from a `BaseAdapter` class. This allows for easy addition of support for other code types by creating new adapter classes without modifying the core documentation generation logic.
*   **Chunking:** The `parse` function implements a chunking mechanism to handle large Apex files. This is necessary because large language models have input length limitations. The `TARGET_CHUNK_SIZE` constant defines the maximum size of each chunk.
*   **Prompt Engineering:** The `get_prompt` function demonstrates careful prompt engineering to guide the language model towards generating high-quality, relevant documentation. The prompt includes specific instructions, formatting requirements, and constraints.
*   **String Formatting:** The code uses f-strings for clear and concise string formatting, improving readability and maintainability.