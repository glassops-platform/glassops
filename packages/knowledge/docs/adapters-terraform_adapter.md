---
type: Documentation
domain: knowledge
origin: packages/knowledge/generation/adapters/terraform_adapter.py
last_modified: 2026-01-28
generated: true
source: packages/knowledge/generation/adapters/terraform_adapter.py
generated_at: 2026-01-28T22:43:40.002956
hash: 34f8eb184d2b7e76a39b0b848459e812a805111c250f893d6e922c6e463b029d
---

## Terraform Adapter Documentation

This document details the Terraform Adapter, a component designed for generating documentation from Terraform configuration files. It is part of a larger system for automated knowledge extraction from infrastructure code.

**Module Purpose and Responsibilities:**

The Terraform Adapter’s primary responsibility is to ingest Terraform files (.tf extension), split them into manageable chunks if they exceed a defined size limit, and prepare them for processing by a language model. It also constructs a prompt that instructs the language model on how to document the Terraform code.

**Key Classes and Their Roles:**

* **`TerraformAdapter`:** This class inherits from `BaseAdapter` and implements the specific logic for handling Terraform files. It determines if a file can be processed, parses the file content into chunks, formats those chunks, and generates a prompt for the language model.

**Important Functions and Their Behavior:**

* **`can_handle(file_path: Path) -> bool`:** This function checks if the adapter can process a given file based on its extension. It returns `True` if the file extension is ".tf", indicating a Terraform file, and `False` otherwise. The `file_path` argument is a `Path` object representing the file's location.
* **`parse(file_path: Path, content: str) -> List[str]`:** This function takes the file path and content as input and splits the content into smaller chunks if the total content length exceeds `TARGET_CHUNK_SIZE` (currently 24000 characters).  It returns a list of strings, where each string represents a chunk of the Terraform configuration. The function ensures that chunks are created at logical breaks (lines) within the file. If the content is smaller than the target size, it returns a list containing the entire content as a single chunk.
* **`_format_chunk(file_path: Path, content: str, part: int = None) -> str`:** This private helper function formats a single chunk of Terraform configuration. It prepends the file path and an optional part number (if the file was split into multiple chunks) to the content, and wraps the content in a Markdown code block using the `hcl` language identifier.
* **`get_prompt(file_path: Path, parsed_content: str) -> str`:** This function constructs a prompt that will be sent to the language model. The prompt instructs the model to act as an Infrastructure as Code expert and document the provided Terraform configuration. It specifies the areas of focus for the documentation (resources, variables, outputs, dependencies, security) and includes strict formatting rules for the output. The `parsed_content` argument represents a chunk of the Terraform configuration.

**Type Hints and Their Significance:**

The code extensively uses type hints (e.g., `file_path: Path`, `content: str`, `-> List[str]`). These hints improve code readability and maintainability by explicitly defining the expected data types for function arguments and return values. They also enable static analysis tools to catch potential type errors during development.

**Notable Patterns and Design Decisions:**

* **Adapter Pattern:** The `TerraformAdapter` follows the Adapter pattern, inheriting from a `BaseAdapter` class. This allows for easy integration of different file types into the documentation generation system without modifying the core logic.
* **Chunking:** The `parse` function implements a chunking mechanism to handle large Terraform files that might exceed the input limits of the language model. This ensures that even large configurations can be processed.
* **Prompt Engineering:** The `get_prompt` function demonstrates careful prompt engineering to guide the language model towards generating high-quality, relevant documentation. The prompt includes specific instructions, formatting requirements, and constraints to ensure consistent and accurate results.
* **Markdown Formatting:** The adapter consistently formats the Terraform content within Markdown code blocks, ensuring that the generated documentation is easily readable and can be rendered correctly in Markdown viewers.