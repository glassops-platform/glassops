---
type: Documentation
domain: knowledge
origin: packages/knowledge/generation/adapters/terraform_adapter.py
last_modified: 2026-02-01
generated: true
source: packages/knowledge/generation/adapters/terraform_adapter.py
generated_at: 2026-02-01T19:32:02.156233
hash: 34f8eb184d2b7e76a39b0b848459e812a805111c250f893d6e922c6e463b029d
---

## Terraform Adapter Documentation

This document details the Terraform Adapter, a component designed for generating documentation from Terraform configuration files. It is part of a larger system for knowledge generation from infrastructure code.

**Module Purpose and Responsibilities**

The Terraform Adapter’s primary responsibility is to process Terraform files (.tf) and prepare their content for documentation generation by a language model. It handles file identification, content parsing into manageable chunks, and formatting prompts for the language model. This adapter ensures that large Terraform configurations are broken down into pieces suitable for processing, while preserving file context.

**Key Classes and Their Roles**

*   **TerraformAdapter:** This class inherits from the `BaseAdapter` and implements the specific logic for handling Terraform files. It determines if a file is a Terraform file, parses its content, formats the content into chunks, and constructs a prompt for the language model.

**Important Functions and Their Behavior**

*   **`can_handle(file_path: Path) -> bool`**: This function checks if the adapter can process a given file based on its extension. It returns `True` if the file path’s suffix is ".tf", indicating a Terraform file, and `False` otherwise.
*   **`parse(file_path: Path, content: str) -> List[str]`**: This function takes the file path and content of a Terraform file as input. It splits the content into chunks, ensuring each chunk does not exceed `TARGET_CHUNK_SIZE` (24000 characters). The function returns a list of strings, where each string represents a chunk of the Terraform configuration. If the entire file content is smaller than `TARGET_CHUNK_SIZE`, it returns a list containing a single string with the entire content.
*   **`_format_chunk(file_path: Path, content: str, part: int = None) -> str`**: This private helper function formats a single chunk of Terraform configuration. It prepends the file path and an optional part number to the content, wraps the content in a code block using the HCL (HashiCorp Configuration Language) syntax, and returns the formatted string.
*   **`get_prompt(file_path: Path, parsed_content: str) -> str`**: This function constructs a prompt for the language model. The prompt instructs the model to act as an Infrastructure as Code expert and document the provided Terraform configuration. It specifies the aspects to focus on (resources, variables, outputs, dependencies, security) and includes strict formatting rules to ensure the output is valid Markdown without conversational text or prohibited terms.

**Type Hints and Their Significance**

The code extensively uses type hints (e.g., `file_path: Path`, `content: str`, `-> List[str]`). These hints improve code readability and maintainability by clearly defining the expected data types for function arguments and return values. They also enable static analysis tools to catch potential type errors during development.

**Notable Patterns and Design Decisions**

*   **Adapter Pattern:** The `TerraformAdapter` follows the Adapter pattern, inheriting from a `BaseAdapter` class. This allows for easy integration of different infrastructure-as-code formats into the knowledge generation system without modifying the core logic.
*   **Chunking:** The `parse` function implements a chunking mechanism to handle large Terraform files. This is necessary because language models have input length limitations. By splitting the content into smaller chunks, we can process files of any size.
*   **Prompt Engineering:** The `get_prompt` function demonstrates careful prompt engineering. The prompt is designed to elicit specific, well-formatted documentation from the language model, while also enforcing constraints on the output style and content.
*   **HCL Code Blocks:** The use of “\`\`\`hcl” code blocks ensures that the Terraform configuration is displayed correctly in Markdown, with proper syntax highlighting.