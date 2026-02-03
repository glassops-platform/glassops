---
type: Documentation
domain: knowledge
last_modified: 2026-02-02
generated: true
source: packages/knowledge/adapters/terraform_adapter.py
generated_at: 2026-02-02T22:26:07.124279
hash: 6087c9054ce7b910210b33038fdf2cdbd5492fc22b54c1ea124f9d66d10ccaef
---

## Terraform Adapter Documentation

This document details the Terraform Adapter, a component designed for generating documentation from Terraform configuration files. It is part of a larger system for knowledge management and documentation across various infrastructure-as-code formats.

**Module Purpose and Responsibilities**

The Terraform Adapter’s primary responsibility is to ingest Terraform files (.tf extension), split them into manageable chunks if necessary, and prepare them for processing by a language model. It also formats the input for the language model with a specific prompt designed to elicit detailed documentation. The adapter handles file-specific parsing and formatting, abstracting away the details of the Terraform language from the core documentation generation process.

**Key Classes and Their Roles**

*   **TerraformAdapter:** This class inherits from the `BaseAdapter` class and implements the adapter-specific logic for Terraform files. It defines how Terraform files are identified, parsed, and formatted for documentation generation.

**Important Functions and Their Behavior**

*   **`can_handle(file_path: Path) -> bool`**: This function determines if the adapter can process a given file based on its extension. It returns `True` if the file path’s suffix is ".tf", indicating a Terraform file, and `False` otherwise.
*   **`parse(file_path: Path, content: str) -> List[str]`**: This function takes the file path and content of a Terraform file as input. It splits the content into chunks if the file exceeds `TARGET_CHUNK_SIZE` (currently 24000 characters). Each chunk is then formatted using the `_format_chunk` method. The function returns a list of strings, where each string represents a chunk of the Terraform configuration. If the content is smaller than the target size, it returns a list containing a single formatted chunk.
*   **`validate_content(content: str) -> List[str]`**: This function currently returns an empty list. It is intended for future implementation of content validation checks specific to Terraform files.
*   **`_format_chunk(file_path: Path, content: str, part: int = None) -> str`**: This private helper function formats a chunk of Terraform configuration into a string suitable for input to the language model. It includes the file path and an optional part number if the file was split into multiple chunks. The content is wrapped in a code block using the HCL (HashiCorp Configuration Language) syntax highlighter.
*   **`get_prompt(file_path: Path, parsed_content: str) -> str`**: This function constructs the prompt that is sent to the language model. The prompt instructs the model to act as an Infrastructure as Code expert and document the provided Terraform configuration. It specifies the areas of focus for the documentation (resources, variables, outputs, dependencies, security) and includes strict rules for the model’s output, prohibiting conversational text, emojis, and specific words. It also explicitly forbids mentioning the project name.

**Type Hints and Their Significance**

The code extensively uses type hints (e.g., `file_path: Path`, `content: str`, `-> List[str]`). These hints improve code readability and maintainability by clearly specifying the expected data types for function arguments and return values. They also enable static analysis tools to catch type-related errors during development.

**Notable Patterns or Design Decisions**

*   **Adapter Pattern:** The `TerraformAdapter` follows the adapter pattern, allowing the system to work with different infrastructure-as-code formats without modifying the core documentation generation logic.
*   **Chunking:** The `parse` function implements a chunking mechanism to handle large Terraform files that might exceed the input limits of the language model. This ensures that even large configurations can be processed.
*   **Prompt Engineering:** The `get_prompt` function demonstrates careful prompt engineering to guide the language model towards generating high-quality, focused documentation. The prompt includes specific instructions and constraints to ensure the desired output format and content.
*   **HCL Highlighting:** The `_format_chunk` function uses HCL syntax highlighting to improve the readability of the Terraform code within the documentation.