---
type: Documentation
domain: knowledge
last_modified: 2026-03-08
generated: true
source: packages/knowledge/adapters/aura_adapter.py
generated_at: 2026-03-08T22:43:35.029900
hash: f011e08e06a7d2c2089e634069420ef0a9dd2c3782facf02ebc9883499314f95
---

## Aura Adapter Documentation

This document describes the Aura Adapter, a component responsible for processing Salesforce Aura components within the larger knowledge system. It acts as an interface between the core system and files specific to the Aura framework.

**Module Purpose:**

The Aura Adapter’s primary responsibility is to identify, parse, and prepare Aura component files for documentation. It determines if a given file is an Aura component based on its path and extension, extracts its content, and optionally includes associated metadata. It then formats this information into a prompt suitable for a language model.

**Key Classes:**

*   **`AuraAdapter`**: This class inherits from `BaseAdapter` and implements the specific logic for handling Aura components. It encapsulates the file identification, parsing, prompt generation, and content validation processes.

**Important Functions:**

*   **`can_handle(file_path: Path) -> bool`**: This function determines whether the adapter can process a given file. It checks if the file path contains “aura” and if the file extension is one of the recognized Aura component extensions (.cmp, .app, .evt, .intf, .tokens, .js, .css, .auradoc). The `Path` type hint indicates that the function expects a file path object. The function returns `True` if the file is an Aura component, and `False` otherwise.
*   **`parse(file_path: Path, content: str) -> List[str]`**: This function parses the content of an Aura component file. It reads the file content and, for specific file types (.cmp, .app, .evt, .intf, .tokens), attempts to read associated metadata files (e.g., `file.cmp-meta.xml`). If the metadata file exists, its content is appended to the main file content. The function returns a list containing the combined content as a single string element. The `List[str]` type hint indicates that the function returns a list of strings, representing content chunks.
*   **`get_prompt(file_path: Path, parsed_content: str) -> str`**: This function generates a prompt for a language model, using the parsed content of the Aura component. It constructs a default prompt instructing the model to document the provided content. The `str` type hint indicates that the function returns a string representing the prompt.
*   **`validate_content(content: str) -> List[str]`**: This function currently provides a placeholder for content validation. It can be extended to perform checks on the Aura component content, such as syntax validation or adherence to coding standards. It currently returns an empty list. The `List[str]` type hint indicates that the function returns a list of strings, potentially containing validation errors.

**Type Hints:**

The code extensively uses type hints (e.g., `file_path: Path`, `content: str`, `-> bool`). These hints improve code readability and allow for static analysis, helping to catch potential errors during development. They clearly define the expected data types for function arguments and return values.

**Design Decisions and Patterns:**

*   **Adapter Pattern:** The `AuraAdapter` follows the Adapter pattern, allowing the system to work with Aura components without needing to know the specifics of the Aura framework. This promotes loose coupling and makes it easier to add support for other component types in the future.
*   **Metadata Handling:** The adapter attempts to include metadata associated with Aura definition files. This provides additional context for documentation. Error handling is included to prevent failures when reading metadata files from interrupting the process.
*   **Chunking:** The `parse` function returns a list containing a single string. This design anticipates potential future requirements for breaking down larger files into smaller chunks for processing.
*   **Configuration Fallback:** The `get_prompt` function provides a default prompt if a specific prompt is not found in a configuration. This ensures that documentation can always be generated, even without custom prompts.

You can extend the `validate_content` function to add more robust validation rules for Aura components.