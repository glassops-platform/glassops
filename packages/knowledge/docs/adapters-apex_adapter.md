---
type: Documentation
domain: knowledge
origin: packages/knowledge/generation/adapters/apex_adapter.py
last_modified: 2026-01-31
generated: true
source: packages/knowledge/generation/adapters/apex_adapter.py
generated_at: 2026-01-31T09:50:08.724000
hash: 7fd9d575b3ebb4b889c9e55dcb24940b5cc37ec212bc40d9dbb3868136103744
---

## Apex Adapter Documentation

This document details the functionality of the Apex Adapter, a component designed for generating documentation from Salesforce Apex code. It is part of a larger system for automated documentation generation across various knowledge sources.

**Module Purpose and Responsibilities**

The Apex Adapter is responsible for identifying, parsing, and preparing Apex code files (.cls and .trigger) for documentation generation. It handles splitting large files into smaller chunks to accommodate limitations of large language models and formats the code for inclusion in a prompt. The adapter then constructs a prompt tailored for a Salesforce architect to document the provided code.

**Key Classes and Their Roles**

*   **ApexAdapter:** This class inherits from `BaseAdapter` and implements the specific logic for handling Apex files. It determines if a file can be processed, parses the file content into chunks, formats those chunks, and generates a prompt for documentation.

**Important Functions and Their Behavior**

*   **`can_handle(file_path: Path) -> bool`**: This function checks if the adapter can process a given file based on its extension. It returns `True` if the file extension is ".cls" (Apex class) or ".trigger" (Apex trigger), and `False` otherwise. The `file_path` argument is a `Path` object representing the file's location.
*   **`parse(file_path: Path, content: str) -> List[str]`**: This function takes the file path and content as input and splits the content into smaller chunks if the content exceeds `TARGET_CHUNK_SIZE` (24000 characters). It returns a list of strings, where each string represents a chunk of the original content. The function ensures that chunks are created at logical breaks (lines) and avoids splitting code in the middle of a line. If the content is smaller than the target size, it returns a list containing the entire content as a single chunk.
*   **`_format_chunk(file_path: Path, content: str, part: int = None) -> str`**: This private helper function formats a single chunk of Apex code. It adds metadata such as the file path, file type (Apex Class or Apex Trigger), and a part number if the content was chunked. The formatted chunk is returned as a string, ready to be included in the prompt. The `part` argument is optional and indicates the chunk number.
*   **`get_prompt(file_path: Path, parsed_content: str) -> str`**: This function constructs the prompt that will be sent to the language model. It includes instructions for the model to act as a Salesforce architect and document the provided Apex code, focusing on purpose, methods, governor limits, integration points, and test coverage. It also includes strict rules for the model's output, prohibiting conversational text, specific words, and the mention of certain names. The `parsed_content` argument is a string containing the formatted Apex code chunk.

**Type Hints and Their Significance**

The code extensively uses type hints (e.g., `file_path: Path`, `content: str`, `-> List[str]`). These hints improve code readability and maintainability by explicitly specifying the expected data types for function arguments and return values. They also enable static analysis tools to catch type-related errors early in the development process.

**Notable Patterns and Design Decisions**

*   **Adapter Pattern:** The `ApexAdapter` follows the Adapter pattern, inheriting from a base class (`BaseAdapter`) and implementing the specific logic for handling Apex files. This allows for easy addition of support for other file types in the future.
*   **Chunking:** The `parse` function implements a chunking mechanism to handle large files that exceed the input limits of the language model. This ensures that the entire file content can be processed, even if it needs to be split into multiple parts.
*   **Prompt Engineering:** The `get_prompt` function demonstrates careful prompt engineering to guide the language model towards generating high-quality documentation. The prompt includes specific instructions, constraints, and a clear request for a particular output format (Markdown).
*   **File Type Detection:** The adapter accurately identifies Apex classes and triggers based on their file extensions, ensuring correct handling and formatting.