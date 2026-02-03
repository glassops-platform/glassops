---
type: Documentation
domain: knowledge
last_modified: 2026-02-02
generated: true
source: packages/knowledge/adapters/apex_adapter.py
generated_at: 2026-02-02T22:24:01.492007
hash: 1e4f12673775d88fb271fa6ccba1b0fc773a901a080790eb1d40c2e1e8f9e4e5
---

## Apex Adapter Documentation

This document details the functionality of the Apex Adapter, a component designed for generating documentation from Salesforce Apex code. It serves as an interface for processing Apex class and trigger files, preparing them for documentation generation by a language model.

**Module Responsibilities:**

The primary responsibility of this adapter is to read Apex code, split it into manageable chunks if necessary, and format it into a structure suitable for input to a documentation generation process. It also provides a prompt template tailored for instructing a language model to document Apex code effectively.

**Key Classes:**

*   **ApexAdapter:** This class inherits from the `BaseAdapter` and implements the specific logic for handling Apex files. It determines if a file can be processed, parses the file content into chunks, validates the content (currently a no-op), formats the chunks, and constructs a prompt for the language model.

**Important Functions:**

*   **`can_handle(file_path: Path) -> bool`**: This function checks if the adapter can process a given file based on its extension. It returns `True` if the file extension is `.cls` (Apex class) or `.trigger` (Apex trigger), and `False` otherwise. The `file_path` argument is a `Path` object representing the file's location.
*   **`parse(file_path: Path, content: str) -> List[str]`**: This function takes the file path and content as input and splits the content into chunks if the content exceeds `TARGET_CHUNK_SIZE` (24000 characters). It returns a list of strings, where each string represents a chunk of the original content. The function ensures that chunks are created at logical breaks (newlines) to avoid splitting code mid-line. If the content is smaller than the target size, it returns a list containing the entire content as a single chunk.
*   **`validate_content(content: str) -> List[str]`**: This function currently performs no validation and returns an empty list. It is reserved for future implementation of content validation checks.
*   **`_format_chunk(file_path: Path, content: str, part: int = None) -> str`**: This private function formats a single chunk of Apex code into a string that includes the file path, file type (Apex Class or Apex Trigger), and an optional part number if the content was split into multiple chunks. The code is enclosed in a Markdown code block with the `apex` language identifier.
*   **`get_prompt(file_path: Path, parsed_content: str) -> str`**: This function constructs a prompt that will be sent to the language model. The prompt instructs the model to act as a Salesforce architect and document the provided Apex code, focusing on purpose, key methods, governor limits, integration points, and test coverage. It includes strict rules for the model’s output, prohibiting conversational text, specific words, and the mention of certain names. The `parsed_content` argument is the Apex code chunk that will be documented.

**Type Hints:**

The code extensively uses type hints (e.g., `file_path: Path`, `content: str`, `-> List[str]`) to improve code readability and maintainability. These hints specify the expected data types for function arguments and return values, enabling static analysis and helping to prevent type-related errors.

**Notable Patterns and Design Decisions:**

*   **Adapter Pattern:** The `ApexAdapter` follows the Adapter pattern, inheriting from a `BaseAdapter` class. This allows for a consistent interface for handling different file types and promotes code reusability.
*   **Chunking:** The `parse` function implements a chunking mechanism to handle large Apex code files that might exceed the input limits of the language model. This ensures that the entire file can be processed, even if it requires splitting it into multiple parts.
*   **Markdown Formatting:** The `_format_chunk` function formats the Apex code within Markdown code blocks, making it easy to render and display the code in documentation.
*   **Prompt Engineering:** The `get_prompt` function is carefully crafted to provide clear instructions to the language model, guiding it to generate high-quality documentation that is specific to Apex code and Salesforce best practices.