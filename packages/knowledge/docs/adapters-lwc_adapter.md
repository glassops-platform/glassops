---
type: Documentation
domain: knowledge
last_modified: 2026-03-08
generated: true
source: packages/knowledge/adapters/lwc_adapter.py
generated_at: 2026-03-08T22:43:53.867538
hash: 04247cda5ca4104b2431873f70c612870b867f22ea078bf07ef7d7d4ec96f329
---

## Lightning Web Component (LWC) Adapter Documentation

This adapter is responsible for processing Salesforce Lightning Web Component files (.js, .html, .css) during documentation generation. It handles parsing the component’s code, splitting large files into manageable chunks, validating content, and preparing prompts for a language model to generate documentation.

**Key Classes:**

*   **`LWCAdapter`**: This class inherits from `BaseAdapter` and implements the specific logic for handling LWC files. It defines how to identify LWC files, parse their content, validate it, and format it for documentation.

**Important Functions:**

*   **`can_handle(file_path: Path) -> bool`**:  Determines if the adapter can process a given file based on its path. It checks if the file is located within an "lwc" directory and has a ".js", ".html", or ".css" extension. The `file_path` argument is a `Path` object representing the file's location.
*   **`parse(file_path: Path, content: str) -> List[str]`**: Parses the content of an LWC file and splits it into chunks if the content exceeds `TARGET_CHUNK_SIZE` (24000 characters).  For JavaScript files, it attempts to read a corresponding metadata file (filename + "-meta.xml") and includes its content in the first chunk. The function returns a list of strings, where each string represents a chunk of the file content formatted for documentation. The `file_path` argument is a `Path` object, and `content` is the file's content as a string.
*   **`validate_content(content: str) -> List[str]`**: Validates the content of an LWC file. Currently, it performs a basic check: if the content appears to be HTML (starts with `<template`), it attempts to parse it as XML using `xml.etree.ElementTree`. If parsing fails, it returns a list containing an error message. The `content` argument is the file's content as a string.
*   **`_format_chunk(file_path: Path, content: str, part: int = None, metadata: str = None) -> str`**: Formats a chunk of LWC content into a string suitable for inclusion in the documentation. It includes the file path, a part number (if applicable), the content itself enclosed in a code block with the appropriate language identifier, and the metadata (if provided). The `file_path` argument is a `Path` object, `content` is the chunk's content, `part` is the chunk number, and `metadata` is the metadata content.
*   **`get_prompt(file_path: Path, parsed_content: str) -> str`**: Constructs a prompt for a language model to generate documentation for the LWC file. The prompt instructs the model to act as a Salesforce Lightning expert and explain the component’s purpose, properties, wire adapters, event handling, lifecycle hooks, and CSS styling. The `file_path` argument is a `Path` object, and `parsed_content` is the content to be documented.

**Type Hints:**

The code extensively uses type hints (e.g., `file_path: Path`, `content: str`, `-> List[str]`). These hints improve code readability and allow for static analysis, helping to catch potential errors during development.

**Notable Patterns and Design Decisions:**

*   **Adapter Pattern:** The `LWCAdapter` follows the adapter pattern, inheriting from a base class (`BaseAdapter`) to provide a consistent interface for handling different file types.
*   **Chunking:** Large files are split into smaller chunks to avoid exceeding the input limits of the language model.
*   **Metadata Handling:** The adapter attempts to include metadata associated with JavaScript files, providing additional context for documentation.
*   **Content Validation:** Basic content validation is performed to catch syntax errors in HTML files.
*   **Prompt Engineering:** A carefully crafted prompt is used to guide the language model in generating accurate and informative documentation.