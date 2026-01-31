---
type: Documentation
domain: knowledge
origin: packages/knowledge/generation/adapters/typescript.py
last_modified: 2026-01-31
generated: true
source: packages/knowledge/generation/adapters/typescript.py
generated_at: 2026-01-31T08:57:29.350150
hash: 38edc897ad829164a4fa8e31f1d033c31f53ae2e1924bbe7f308e5d525489af4
---

## TypeScript Adapter Documentation

This document details the TypeScript Adapter, a component designed for processing TypeScript and JavaScript source code during documentation generation. It is a port of functionality originally present in a related TypeScript project.

**Module Purpose:**

The primary responsibility of this adapter is to take TypeScript or JavaScript files as input, split them into manageable chunks, and format those chunks for use with a language model. This prepares the code for documentation creation by ensuring the input size is appropriate and providing necessary context.

**Key Classes:**

*   **`TypeScriptAdapter`**: This class inherits from `BaseAdapter` and implements the specific logic for handling TypeScript and JavaScript files. It determines if a file can be processed, parses the file content into chunks, and formats those chunks with relevant metadata.

**Important Functions:**

*   **`can_handle(file_path: Path) -> bool`**: This function checks if the adapter can process a given file based on its extension. It returns `True` if the file has a `.ts`, `.js`, `.mjs`, `.tsx`, or `.jsx` extension; otherwise, it returns `False`. The `file_path` argument is a `Path` object representing the file's location.
*   **`parse(file_path: Path, content: str) -> List[str]`**: This function takes the file path and content as input and splits the content into a list of strings (chunks). The size of each chunk is limited by `TARGET_CHUNK_SIZE` (currently 24000 characters, approximately 6000 tokens). It intelligently splits the content at line breaks to avoid breaking code mid-line. If the content is smaller than the target size, it returns a list containing the entire content as a single chunk. The function returns a `List[str]`, where each string is a formatted chunk of the original file content.
*   **`_format_chunk(file_path: Path, content: str, part: int = None) -> str`**: This is a helper function that formats a single chunk of code. It adds a header indicating the file path and, if applicable, a part number for chunked files. The `content` argument is the code chunk itself, and `part` is an optional integer indicating the chunk number. The function returns a formatted string containing the file path, part number (if any), and the code content enclosed in a ```typescript``` block.
*   **`get_prompt(file_path: Path, parsed_content: str) -> str`**: This function constructs the prompt that will be sent to the language model. It includes instructions for the model, specifying its role as a principal architect and outlining the desired characteristics of the generated documentation (concise, inclusive, professional). It also includes strict rules regarding the output format and prohibited words. The `parsed_content` argument is the formatted chunk of code that will be included in the prompt.

**Type Hints:**

The code extensively uses type hints (e.g., `file_path: Path`, `content: str`, `-> List[str]`). These hints improve code readability and allow for static analysis, helping to catch potential errors during development. They clearly define the expected data types for function arguments and return values.

**Notable Patterns and Design Decisions:**

*   **Adapter Pattern:** The `TypeScriptAdapter` follows the Adapter pattern, inheriting from a `BaseAdapter` class. This allows for easy addition of support for other languages by creating new adapter classes that implement the same interface.
*   **Chunking Strategy:** The `parse` function implements a chunking strategy to handle large files. It splits the content into smaller chunks based on a target size, ensuring that the input to the language model remains within acceptable limits. The splitting occurs at line breaks to maintain code integrity.
*   **Contextual Formatting:** The `_format_chunk` function adds contextual information (file path and part number) to each chunk, providing the language model with valuable information about the source code.
*   **Prompt Engineering:** The `get_prompt` function carefully crafts a prompt that guides the language model to generate high-quality documentation. The prompt includes specific instructions, constraints, and a defined role for the model.