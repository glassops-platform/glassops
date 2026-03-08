---
type: Documentation
domain: knowledge
last_modified: 2026-03-08
generated: true
source: packages/knowledge/adapters/visualforce_adapter.py
generated_at: 2026-03-08T22:44:07.408601
hash: 7df3c8c5de52169bec47f05b03c391549baa9f5c88ec931a0c9dfd9f9b36fe4b
---

## Visualforce Adapter Documentation

This document describes the Visualforce Adapter, a component designed to process Salesforce Visualforce page and component files. It serves as an interface between the core knowledge processing system and Visualforce content, enabling documentation and analysis of these files.

**Module Purpose:**

The primary responsibility of this adapter is to identify, parse, and prepare Visualforce files for processing by a larger knowledge system. It handles the specific file types associated with Visualforce and provides a standardized way to access and present the file content.

**Key Classes:**

* **VisualforceAdapter:** This class inherits from the `BaseAdapter` class and implements the specific logic for handling Visualforce files. It defines how to recognize these files, extract their content, and formulate a prompt for documentation.

**Important Functions:**

* **`can_handle(file_path: Path) -> bool`:** This function determines if the adapter can process a given file based on its extension. It returns `True` if the file extension is either ".page" or ".component" (case-insensitive), indicating a Visualforce page or component, and `False` otherwise. The `Path` type hint ensures the input is a file path object.
* **`parse(file_path: Path, content: str) -> List[str]`:** This function takes the file path and the file content as input and prepares the content for further processing. Currently, it treats the entire file content as a single chunk and returns it as a list containing that single string. The `List[str]` type hint indicates the function returns a list of strings.
* **`get_prompt(file_path: Path, parsed_content: str) -> str`:** This function generates a prompt string that will be used to instruct a language model to document the Visualforce content. It constructs a prompt that includes the file type and the file content itself. The `str` type hint indicates the function returns a string.
* **`validate_content(content: str) -> List[str]`:** This function is intended to perform validation checks on the file content. Currently, it is a placeholder and returns an empty list, indicating no validation errors. Future development could include XML or HTML validation to ensure the Visualforce content is well-formed. The `List[str]` type hint indicates the function returns a list of strings, where each string represents a validation error.

**Type Hints:**

The code extensively uses type hints (e.g., `file_path: Path`, `content: str`, `-> bool`). These hints improve code readability and allow for static analysis, helping to catch potential errors during development. They clearly define the expected data types for function arguments and return values.

**Design Decisions & Patterns:**

* **Adapter Pattern:** The `VisualforceAdapter` follows the Adapter pattern, allowing the core knowledge processing system to work with Visualforce files without needing to know the specifics of their format. This promotes loose coupling and makes it easy to add support for other file types in the future.
* **Simple Parsing:** The `parse` function currently performs minimal parsing, treating the entire file as a single unit. This approach is suitable for initial documentation efforts. More sophisticated parsing could be implemented later to extract specific elements or metadata from the Visualforce files.
* **Fallback Prompt:** The `get_prompt` function provides a basic prompt for documentation. You can customize this prompt to provide more specific instructions to the language model, improving the quality of the generated documentation.