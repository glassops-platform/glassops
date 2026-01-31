---
type: Documentation
domain: knowledge
origin: packages/knowledge/generation/adapters/terraform_adapter.py
last_modified: 2026-01-31
generated: true
source: packages/knowledge/generation/adapters/terraform_adapter.py
generated_at: 2026-01-31T09:52:43.926560
hash: 34f8eb184d2b7e76a39b0b848459e812a805111c250f893d6e922c6e463b029d
---

## Terraform Adapter Documentation

This document describes the Terraform Adapter, a component designed for generating documentation from Terraform configuration files. It is part of a larger knowledge generation system.

**Module Purpose and Responsibilities**

The Terraform Adapter’s primary responsibility is to ingest Terraform files (`.tf` extension), split them into manageable chunks if they exceed a defined size limit, and prepare them for processing by a language model. It also constructs a prompt that instructs the language model to produce detailed documentation for the Terraform code.

**Key Classes and Their Roles**

*   **TerraformAdapter:** This class inherits from the `BaseAdapter` class and implements the specific logic for handling Terraform files. It determines if a file can be processed, parses the file content into chunks, formats those chunks, and generates a prompt for the language model.

**Important Functions and Their Behavior**

*   **`can_handle(file_path: Path) -> bool`**: This function checks if the adapter can process a given file based on its extension. It returns `True` if the file has a `.tf` extension, and `False` otherwise. The `file_path` argument is a `Path` object representing the file's location.
*   **`parse(file_path: Path, content: str) -> List[str]`**: This function takes the file path and content as input and splits the content into smaller chunks if the content length exceeds `TARGET_CHUNK_SIZE` (currently 24000 characters). It returns a list of strings, where each string represents a chunk of the Terraform configuration. The function ensures that chunks are created at logical breaks (newlines) to avoid splitting code mid-line. If the content is smaller than the target size, it returns a list containing the entire content as a single chunk.
*   **`_format_chunk(file_path: Path, content: str, part: int = None) -> str`**: This is a helper function that formats a single chunk of Terraform configuration. It adds a header indicating the file name and, if applicable, the chunk number (e.g., " (Part 2)"). The content is enclosed in a code block using the `hcl` language identifier for syntax highlighting.
*   **`get_prompt(file_path: Path, parsed_content: str) -> str`**: This function constructs the prompt that will be sent to the language model. The prompt instructs the model to act as an Infrastructure as Code expert and document the provided Terraform configuration. It specifies the areas the documentation should cover: resources, variables, outputs, dependencies, and security considerations. It also includes strict instructions regarding the output format (valid Markdown only, no conversational text, and specific word restrictions) and prohibits mentioning certain names.

**Type Hints and Their Significance**

The code makes extensive use of type hints (e.g., `file_path: Path`, `content: str`, `-> List[str]`). These hints improve code readability and allow for static analysis, helping to catch potential errors during development. They also clarify the expected input and output types for each function.

**Notable Patterns and Design Decisions**

*   **Adapter Pattern:** The `TerraformAdapter` follows the Adapter pattern, inheriting from a `BaseAdapter` class. This allows for easy integration of other file types into the knowledge generation system by creating new adapter classes.
*   **Chunking:** The `parse` function implements a chunking mechanism to handle large Terraform files that might exceed the input limits of the language model. This ensures that even extensive configurations can be processed.
*   **Prompt Engineering:** The `get_prompt` function demonstrates careful prompt engineering to guide the language model towards generating high-quality, relevant documentation. The prompt includes specific instructions and constraints to control the output format and content.
*   **String Formatting:** The code uses f-strings for clear and concise string formatting, particularly in the `_format_chunk` and `get_prompt` functions.