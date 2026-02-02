---
type: Documentation
domain: knowledge
origin: packages/knowledge/generation/adapters/lwc_adapter.py
last_modified: 2026-02-01
generated: true
source: packages/knowledge/generation/adapters/lwc_adapter.py
generated_at: 2026-02-01T19:31:27.644991
hash: 46fc5ec245213835235f1563576f82635e9c7c77480f8813a58f700dc8574dba
---

## Lightning Web Component (LWC) Adapter Documentation

This document details the functionality of the LWC Adapter, a component within a documentation generation system. It is designed to process Salesforce Lightning Web Component files (.js, .html, .css) and prepare them for documentation creation using a large language model.

**Module Purpose:**

The LWC Adapter’s primary responsibility is to identify, parse, and format LWC files into manageable chunks suitable for input to a language model. It handles files found within directories containing "lwc" in their path. The adapter prepares the content and constructs a prompt instructing the language model to generate documentation specific to LWC best practices.

**Key Classes:**

*   **`LWCAdapter`**: This class inherits from `BaseAdapter` and implements the logic for handling LWC files. It defines how to determine if a file is a supported LWC file, how to split the file content into chunks, and how to format those chunks for the language model.

**Important Functions:**

*   **`can_handle(file_path: Path) -> bool`**: This function determines whether the adapter can process a given file based on its path. It returns `True` if the file path contains "lwc" and has a ".js", ".html", or ".css" extension; otherwise, it returns `False`. The `file_path` argument is a `Path` object representing the file's location.
*   **`parse(file_path: Path, content: str) -> List[str]`**: This function takes the file path and content as input and splits the content into smaller chunks if the content exceeds `TARGET_CHUNK_SIZE` (24000 characters). It returns a list of strings, where each string represents a chunk of the file content. The function ensures that chunks are created at logical breaks (newlines) to avoid splitting code or markup in the middle. If the content is smaller than the target size, it returns a list containing the entire content as a single chunk.
*   **`_format_chunk(file_path: Path, content: str, part: int = None) -> str`**: This private helper function formats a single chunk of content into a string suitable for input to the language model. It includes the file path, an optional part number (if the file was split into multiple chunks), and wraps the content in a code block with the appropriate language identifier (javascript, html, or css).
*   **`get_prompt(file_path: Path, parsed_content: str) -> str`**: This function constructs a prompt for the language model. The prompt instructs the model to act as a Salesforce Lightning expert and document the provided LWC file. It specifies the areas of focus for the documentation, including component purpose, properties, wire adapters, event handling, lifecycle hooks, and CSS styling. It also includes strict rules for the model’s output, prohibiting conversational text, specific words, and the inclusion of certain names. The `parsed_content` argument is a string representing the content of the LWC file.

**Type Hints:**

The code extensively uses type hints (e.g., `file_path: Path`, `content: str`, `-> List[str]`). These hints improve code readability and allow for static analysis, helping to catch potential errors during development. They clearly define the expected data types for function arguments and return values.

**Notable Patterns and Design Decisions:**

*   **Adapter Pattern:** The `LWCAdapter` follows the Adapter pattern, inheriting from a `BaseAdapter` class. This allows for easy extension to support other file types or documentation generation systems.
*   **Chunking:** The `parse` function implements a chunking mechanism to handle large files that might exceed the language model’s input limits. This ensures that all content can be processed, even if it requires splitting the file into multiple parts.
*   **Prompt Engineering:** The `get_prompt` function demonstrates careful prompt engineering to guide the language model towards generating high-quality, specific documentation for LWC files. The prompt includes clear instructions, constraints, and a defined role for the model.
*   **File Type Detection:** The `can_handle` function uses the file extension and directory structure to accurately identify LWC files.