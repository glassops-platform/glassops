---
type: Documentation
domain: knowledge
origin: packages/knowledge/generation/adapters/python.py
last_modified: 2026-01-31
generated: true
source: packages/knowledge/generation/adapters/python.py
generated_at: 2026-01-31T09:52:16.224364
hash: 01f09935871c40b71b24ec76a18a51e62995920a97c119c08c60e789e00bf4cd
---

## Python Adapter Documentation

This document details the functionality of the Python Adapter, a component designed for generating documentation from Python source code files. It is part of a larger documentation generation system.

**Module Purpose and Responsibilities**

The Python Adapter’s primary responsibility is to read Python files, split their content into manageable chunks, and prepare those chunks for processing by a language model to produce documentation. It handles the specifics of Python syntax to ensure logical splitting of code, respecting class and function boundaries where possible.

**Key Classes and Their Roles**

*   **`PythonAdapter`**: This class inherits from `BaseAdapter` and implements the adapter logic for Python files. It determines if a file can be handled, parses the file content into chunks, and formats those chunks with relevant context.

**Important Functions and Their Behavior**

*   **`can_handle(file_path: Path) -> bool`**: This function checks if the adapter can process a given file based on its extension. It returns `True` if the file has a `.py` extension, and `False` otherwise. The `file_path` argument is a `Path` object representing the file's location.

*   **`parse(file_path: Path, content: str) -> List[str]`**: This function takes the file path and content of a Python file as input and splits the content into a list of strings (chunks). The goal is to create chunks that are suitable for input to a language model, respecting a target size (`TARGET_CHUNK_SIZE`). The function attempts to split the content at logical boundaries, such as class or function definitions, to maintain context within each chunk. If the file content is smaller than the target chunk size, it returns a list containing the entire content as a single chunk.

*   **`_format_chunk(file_path: Path, content: str, part: int = None) -> str`**: This is a helper function that formats a chunk of Python code with file context. It prepends the file path to the chunk and optionally adds a part number if the file was split into multiple chunks. The `content` argument is the code chunk itself, and `part` is an optional integer indicating the chunk number. The function returns a formatted string.

*   **`get_prompt(file_path: Path, parsed_content: str) -> str`**: This function constructs the prompt that will be sent to the language model. It includes instructions for the language model, specifying its role as a principal architect and outlining the desired documentation format and constraints. It then appends the parsed content of the Python file to the prompt. The `parsed_content` argument is a string representing the code chunk to be documented.

**Type Hints and Their Significance**

The code makes extensive use of type hints (e.g., `file_path: Path`, `content: str`, `-> List[str]`). These hints improve code readability and allow for static analysis, helping to catch potential errors during development. They also clarify the expected input and output types for each function.

**Notable Patterns or Design Decisions**

*   **Adapter Pattern**: The `PythonAdapter` follows the Adapter pattern, inheriting from a `BaseAdapter` class. This allows for easy addition of support for other languages by creating new adapter classes that implement the same interface.

*   **Chunking Strategy**: The `parse` function employs a line-based chunking strategy with an attempt to respect code structure. This approach balances the need to stay within the `TARGET_CHUNK_SIZE` with the desire to maintain context within each chunk. The `TARGET_CHUNK_SIZE` is set to 24000 characters, which is estimated to be around 6000 tokens.

*   **Prompt Engineering**: The `get_prompt` function demonstrates careful prompt engineering, providing detailed instructions to the language model to ensure the generated documentation meets specific quality and formatting requirements. The prompt explicitly prohibits certain words and styles to control the output.