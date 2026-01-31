---
type: Documentation
domain: knowledge
origin: packages/knowledge/generation/adapters/terraform_adapter.py
last_modified: 2026-01-31
generated: true
source: packages/knowledge/generation/adapters/terraform_adapter.py
generated_at: 2026-01-31T08:57:01.018714
hash: 34f8eb184d2b7e76a39b0b848459e812a805111c250f893d6e922c6e463b029d
---

## Terraform Adapter Documentation

This document details the Terraform Adapter, a component designed for generating documentation from Terraform configuration files. It is part of a larger system for knowledge generation from infrastructure code.

**Module Purpose:**

The Terraform Adapter’s primary responsibility is to read Terraform files (.tf), split them into manageable chunks if they exceed a defined size limit, and prepare them for processing by a language model. It also constructs a prompt that instructs the language model on how to document the Terraform code.

**Key Classes:**

* **`TerraformAdapter`:** This class inherits from `BaseAdapter` and implements the specific logic for handling Terraform files. It determines if a file is a Terraform file, parses its content into chunks, formats those chunks, and generates a prompt for documentation.

**Important Functions:**

* **`can_handle(file_path: Path) -> bool`:** This function checks if the adapter can process a given file based on its extension. It returns `True` if the file extension is ".tf", indicating a Terraform file, and `False` otherwise. The `file_path` argument is a `Path` object representing the file's location.
* **`parse(file_path: Path, content: str) -> List[str]`:** This function takes the file path and content of a Terraform file as input. If the content is within the `TARGET_CHUNK_SIZE` limit, it returns a list containing a single formatted chunk. Otherwise, it splits the content into multiple chunks, ensuring no chunk exceeds the size limit. Each chunk is then formatted using the `_format_chunk` method. The function returns a list of strings, where each string represents a chunk of the Terraform configuration.
* **`_format_chunk(file_path: Path, content: str, part: int = None) -> str`:** This private helper function formats a single chunk of Terraform configuration. It prepends the file path and an optional part number (if the content was chunked) to the content, wraps the content in a code block using HCL syntax highlighting, and returns the formatted string.
* **`get_prompt(file_path: Path, parsed_content: str) -> str`:** This function constructs a prompt that will be sent to a language model. The prompt instructs the model to act as an Infrastructure as Code expert and document the provided Terraform configuration. It specifies the areas of focus for the documentation (resources, variables, outputs, dependencies, security) and includes strict formatting rules for the output. The `parsed_content` argument is a string containing the Terraform configuration chunk.

**Type Hints:**

The code extensively uses type hints (e.g., `file_path: Path`, `content: str`, `-> List[str]`). These hints improve code readability and allow for static analysis, helping to catch potential errors during development. They clearly define the expected data types for function arguments and return values.

**Notable Patterns and Design Decisions:**

* **Adapter Pattern:** The `TerraformAdapter` follows the Adapter pattern, inheriting from a `BaseAdapter` class. This allows for easy addition of support for other infrastructure-as-code languages or file formats without modifying the core knowledge generation system.
* **Chunking:** The `parse` function implements a chunking mechanism to handle large Terraform files that might exceed the input size limits of the language model. This ensures that even large configurations can be processed.
* **Prompt Engineering:** The `get_prompt` function demonstrates careful prompt engineering to guide the language model towards generating high-quality, focused documentation. The prompt includes specific instructions, formatting requirements, and constraints.
* **String Formatting:** The code uses f-strings for clear and concise string formatting, improving readability.