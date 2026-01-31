---
type: Documentation
domain: knowledge
origin: packages/knowledge/generation/adapters/lwc_adapter.py
last_modified: 2026-01-31
generated: true
source: packages/knowledge/generation/adapters/lwc_adapter.py
generated_at: 2026-01-31T08:56:26.283943
hash: 46fc5ec245213835235f1563576f82635e9c7c77480f8813a58f700dc8574dba
---

## Lightning Web Component (LWC) Adapter Documentation

This document details the functionality of the LWC Adapter, a component within a documentation generation system. It is designed to process Salesforce Lightning Web Component files (.js, .html, .css) and prepare them for documentation creation using a large language model.

**Module Purpose:**

The LWC Adapter’s primary responsibility is to identify, parse, and format LWC files into manageable chunks suitable for input to a language model. It handles the specific file types and directory structure associated with LWCs within a Salesforce project.

**Key Classes:**

*   **`LWCAdapter`**: This class inherits from `BaseAdapter` and implements the logic for handling LWC files. It defines how to determine if a file is an LWC, how to split the file content into smaller parts, and how to format those parts for the language model.

**Important Functions:**

*   **`can_handle(file_path: Path) -> bool`**: This function determines whether the adapter can process a given file. It checks if the file path contains "lwc" in its parts (indicating it’s within an LWC directory) and if the file extension is one of the supported types (.js, .html, or .css). The `file_path` argument is a `Path` object representing the file's location.
*   **`parse(file_path: Path, content: str) -> List[str]`**: This function takes the file path and content as input and splits the content into a list of strings (chunks). Each chunk is designed to be within a defined size limit (`TARGET_CHUNK_SIZE`) to accommodate language model input constraints. It iterates through the file line by line, creating chunks that do not exceed the size limit. The `file_path` argument is a `Path` object, and `content` is the file's content as a string. The function returns a list of strings, where each string represents a chunk of the original file content.
*   **`_format_chunk(file_path: Path, content: str, part: int = None) -> str`**: This private helper function formats a single chunk of content into a string suitable for the language model. It includes the file path, an optional part number (if the file was split into multiple chunks), and wraps the content in a code block with the appropriate language identifier (javascript, html, or css). The `file_path` argument is a `Path` object, `content` is the chunk's content, and `part` is an optional integer indicating the chunk number. It returns a formatted string.
*   **`get_prompt(file_path: Path, parsed_content: str) -> str`**: This function constructs the prompt that will be sent to the language model. The prompt instructs the model to act as a Salesforce Lightning expert and document the provided LWC file. It specifies the areas of focus for the documentation, including component purpose, properties, wire adapters, event handling, lifecycle hooks, and CSS styling. It also includes strict rules for the model’s output, prohibiting conversational text, emojis, and specific words. The `file_path` argument is a `Path` object, and `parsed_content` is the content of the LWC file. It returns a string containing the prompt.

**Type Hints:**

The code extensively uses type hints (e.g., `file_path: Path`, `content: str`, `-> List[str]`). These hints improve code readability and allow for static analysis, helping to catch potential errors during development. They clearly define the expected data types for function arguments and return values.

**Notable Patterns and Design Decisions:**

*   **Adapter Pattern:** The `LWCAdapter` follows the Adapter pattern, inheriting from a `BaseAdapter` class. This allows for easy extension to support other file types or documentation sources in the future.
*   **Chunking:** The `parse` function implements a chunking mechanism to handle large files that exceed the language model’s input size limit. This ensures that the entire file content can be processed, even if it needs to be split into multiple parts.
*   **Prompt Engineering:** The `get_prompt` function demonstrates careful prompt engineering to guide the language model towards generating high-quality, focused documentation. The prompt includes specific instructions, constraints, and a clear definition of the desired output format.
*   **File Type Detection:** The `can_handle` function uses the file extension and directory structure to accurately identify LWC files.