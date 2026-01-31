---
type: Documentation
domain: knowledge
origin: packages/knowledge/generation/adapters/apex_adapter.py
last_modified: 2026-01-31
generated: true
source: packages/knowledge/generation/adapters/apex_adapter.py
generated_at: 2026-01-31T08:54:25.703140
hash: 7fd9d575b3ebb4b889c9e55dcb24940b5cc37ec212bc40d9dbb3868136103744
---

## Apex Adapter Documentation

This document details the functionality of the Apex Adapter, a component designed for generating documentation from Salesforce Apex code. It serves as an interface between the documentation generation system and Apex files (.cls and .trigger).

**Module Purpose and Responsibilities:**

The Apex Adapter is responsible for identifying, parsing, and formatting Apex code files into manageable chunks suitable for processing by a language model. It then constructs a prompt that instructs the language model to generate comprehensive documentation for the provided code. The adapter handles both Apex classes and triggers.

**Key Classes and Their Roles:**

*   **ApexAdapter:** This is the primary class within the adapter. It inherits from the `BaseAdapter` class, providing a standardized interface for handling different file types. The `ApexAdapter` specifically implements the logic for Apex code, including file type recognition, content parsing, and prompt generation.

**Important Functions and Their Behavior:**

*   **`can_handle(file_path: Path) -> bool`:** This function determines if the adapter can process a given file based on its extension. It returns `True` if the file extension is ".cls" (Apex class) or ".trigger" (Apex trigger), and `False` otherwise. The `file_path` argument is a `Path` object representing the file's location.
*   **`parse(file_path: Path, content: str) -> List[str]`:** This function takes the file path and content of an Apex file as input and splits the content into smaller chunks. This is necessary because language models have input length limitations. The function aims to split the code at logical points to avoid breaking code blocks. It returns a list of strings, where each string represents a chunk of Apex code. The `TARGET_CHUNK_SIZE` constant defines the maximum size of each chunk (24000 characters).
*   **`_format_chunk(file_path: Path, content: str, part: int = None) -> str`:** This private helper function formats a single chunk of Apex code into a string that includes the file name, file type (class or trigger), and an optional part number if the file was split into multiple chunks. The `content` is enclosed in a Markdown code block with the `apex` language identifier.
*   **`get_prompt(file_path: Path, parsed_content: str) -> str`:** This function constructs the prompt that will be sent to the language model. The prompt includes instructions to act as a Salesforce architect and document the provided Apex code, focusing on purpose, methods, governor limits, integration points, and test coverage. It also includes strict rules for the language model’s output, prohibiting conversational text, specific words, and the mention of certain names. The `parsed_content` argument is a string containing the Apex code chunk.

**Type Hints and Their Significance:**

The code extensively uses type hints (e.g., `file_path: Path`, `content: str`, `-> List[str]`). These hints improve code readability and maintainability. They also enable static analysis tools to catch potential type errors during development.

**Notable Patterns or Design Decisions:**

*   **Adapter Pattern:** The `ApexAdapter` follows the Adapter pattern, allowing the documentation generation system to work with different file types without modifying its core logic. Each adapter is responsible for handling a specific file type.
*   **Chunking:** The `parse` function implements a chunking mechanism to handle large Apex files that exceed the language model's input length limit. This ensures that the entire file can be processed, even if it requires splitting it into multiple parts.
*   **Prompt Engineering:** The `get_prompt` function demonstrates careful prompt engineering to guide the language model towards generating high-quality, relevant documentation. The prompt includes specific instructions, constraints, and a clear definition of the desired output format.
*   **Markdown Formatting:** The adapter consistently formats the code chunks and the overall prompt using Markdown, ensuring that the generated documentation is well-structured and easy to read.