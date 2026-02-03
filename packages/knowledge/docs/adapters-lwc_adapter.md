---
type: Documentation
domain: knowledge
last_modified: 2026-02-02
generated: true
source: packages/knowledge/adapters/lwc_adapter.py
generated_at: 2026-02-02T22:25:32.809546
hash: 7bf7cab97960f829b21a0a9e22be0b77440bf39e55ce6645db8c6cfe671017bc
---

## Lightning Web Component (LWC) Adapter Documentation

This document details the functionality of the LWC Adapter, a component within a larger documentation generation system. It is responsible for processing Salesforce Lightning Web Component files (.js, .html, .css) and preparing them for documentation creation using large language models.

**Module Purpose:**

The LWC Adapter serves as a bridge between the core documentation generation pipeline and Salesforce’s LWC technology. It handles file identification, content parsing into manageable chunks, basic content validation, and prompt engineering for optimal documentation results.

**Key Classes:**

*   **`LWCAdapter`**: This class inherits from `BaseAdapter` and implements the specific logic for handling LWC files. It encapsulates the parsing, validation, and formatting processes tailored to LWC structures.

**Important Functions:**

*   **`can_handle(file_path: Path) -> bool`**:  This function determines if the adapter is capable of processing a given file. It checks if the file path contains "lwc" in its parts (indicating it resides within an LWC directory) and if the file extension is one of the supported types: ".js", ".html", or ".css". The `file_path` argument is a `Path` object representing the file's location.
*   **`parse(file_path: Path, content: str) -> List[str]`**: This function takes the file path and content of an LWC file as input and splits the content into smaller chunks. This is necessary because large language models have input length limitations. The `TARGET_CHUNK_SIZE` constant (set to 24000 characters) defines the maximum size of each chunk. The function returns a list of strings, where each string represents a chunk of the original content.
*   **`validate_content(content: str) -> List[str]`**: This function performs basic validation of the LWC content. Currently, it checks for valid XML syntax if the content appears to be an HTML template (starts with `<template`).  If the content is not valid XML, it returns a list containing an error message.  It returns an empty list if validation passes.
*   **`_format_chunk(file_path: Path, content: str, part: int = None) -> str`**: This is a helper function that formats a chunk of LWC content into a string suitable for inclusion in a prompt to a language model. It includes the file path, an optional part number (if the content was chunked), and wraps the content in a code block with the appropriate language identifier (javascript, html, or css).
*   **`get_prompt(file_path: Path, parsed_content: str) -> str`**: This function constructs a prompt that will be sent to a language model (like Gemma 12b IT) to generate documentation for the LWC file. The prompt instructs the model to act as a Salesforce Lightning expert and to document the component's purpose, properties, wire adapters, event handling, lifecycle hooks, and CSS styling. It includes strict rules for the model's output, prohibiting conversational text, emojis, and specific words. The `parsed_content` argument is a string representing the content of the LWC file.

**Type Hints:**

The code extensively uses type hints (e.g., `file_path: Path`, `content: str`, `-> List[str]`). These hints improve code readability and allow for static analysis, helping to catch potential errors during development.

**Notable Patterns and Design Decisions:**

*   **Adapter Pattern:** The `LWCAdapter` follows the Adapter pattern, inheriting from a `BaseAdapter` class. This allows for a consistent interface for handling different file types and promotes code reusability.
*   **Chunking:** The `parse` function implements a chunking mechanism to handle large files that exceed the input limits of large language models.
*   **Prompt Engineering:** The `get_prompt` function demonstrates careful prompt engineering to guide the language model towards generating high-quality documentation. The prompt includes specific instructions, constraints, and a clear definition of the desired output format.
*   **Content Validation:** The `validate_content` function provides a basic level of content validation to prevent errors during documentation generation.