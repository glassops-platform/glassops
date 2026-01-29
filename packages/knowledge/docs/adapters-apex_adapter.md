---
type: Documentation
domain: knowledge
origin: packages/knowledge/generation/adapters/apex_adapter.py
last_modified: 2026-01-28
generated: true
source: packages/knowledge/generation/adapters/apex_adapter.py
generated_at: 2026-01-28T22:41:11.958315
hash: 7fd9d575b3ebb4b889c9e55dcb24940b5cc37ec212bc40d9dbb3868136103744
---

## Apex Adapter Documentation

This document details the functionality of the Apex Adapter, a component designed for generating documentation from Salesforce Apex code. It serves as an interface between the documentation generation system and Apex files (.cls and .trigger).

**Module Purpose and Responsibilities:**

The Apex Adapter is responsible for identifying, parsing, and formatting Apex code files into manageable chunks suitable for processing by a language model. It then constructs a prompt that instructs the language model to generate comprehensive documentation for the provided code. The adapter handles both Apex classes and triggers.

**Key Classes and Their Roles:**

* **ApexAdapter:** This is the primary class within the adapter. It inherits from the `BaseAdapter` class, providing a standardized interface for handling different file types. Its core responsibilities include determining if a file can be handled, parsing the file content into chunks, formatting those chunks, and constructing the prompt for the language model.

**Important Functions and Their Behavior:**

* **`can_handle(file_path: Path) -> bool`:** This function determines whether the adapter can process a given file based on its extension. It returns `True` if the file extension is ".cls" (Apex class) or ".trigger" (Apex trigger), and `False` otherwise. The `file_path` argument is a `Path` object representing the file's location.
* **`parse(file_path: Path, content: str) -> List[str]`:** This function takes the file path and its content as input and splits the content into smaller chunks. This is necessary because language models have input length limitations. The function aims to create chunks that are no larger than `TARGET_CHUNK_SIZE` (24000 characters). It iterates through the lines of the content, building up each chunk until it reaches the size limit. Any remaining content is added as a final chunk. The function returns a list of strings, where each string represents a chunk of Apex code.
* **`_format_chunk(file_path: Path, content: str, part: int = None) -> str`:** This private helper function formats a single chunk of Apex code into a string that includes the file name, file type (class or trigger), and an optional part number if the file was split into multiple chunks. The `content` argument is the Apex code chunk, and `part` is an integer indicating the chunk number. The function returns a formatted string suitable for inclusion in the prompt.
* **`get_prompt(file_path: Path, parsed_content: str) -> str`:** This function constructs the prompt that will be sent to the language model. The prompt instructs the model to act as a Salesforce architect and document the provided Apex code. It specifies the desired documentation elements (purpose, methods, governor limits, integration points, test coverage) and includes strict formatting rules to ensure the output is valid Markdown. The `parsed_content` argument is a string containing the Apex code chunk.

**Type Hints and Their Significance:**

The code extensively uses type hints (e.g., `file_path: Path`, `content: str`, `-> List[str]`). These hints improve code readability and maintainability by explicitly specifying the expected data types for function arguments and return values. They also enable static analysis tools to detect potential type errors.

**Notable Patterns or Design Decisions:**

* **Adapter Pattern:** The `ApexAdapter` follows the Adapter pattern, allowing the documentation generation system to work with different file types without modification. Each adapter is responsible for handling a specific file type and converting it into a standardized format.
* **Chunking:** The `parse` function implements a chunking mechanism to handle large Apex code files that exceed the language model's input length limit. This ensures that the entire file can be processed, even if it requires splitting it into multiple chunks.
* **Prompt Engineering:** The `get_prompt` function demonstrates careful prompt engineering to guide the language model towards generating high-quality, relevant documentation. The prompt includes specific instructions, formatting rules, and constraints to ensure the output meets the desired requirements.