---
type: Documentation
domain: knowledge
origin: packages/knowledge/generation/adapters/python.py
last_modified: 2026-02-01
generated: true
source: packages/knowledge/generation/adapters/python.py
generated_at: 2026-02-01T19:31:45.081907
hash: 01f09935871c40b71b24ec76a18a51e62995920a97c119c08c60e789e00bf4cd
---

## Python Adapter Documentation

This document details the functionality of the Python Adapter, a component designed for generating documentation from Python source code. It is part of a larger documentation generation system.

**Module Purpose and Responsibilities**

The Python Adapter’s primary responsibility is to process Python files (`.py`) and prepare their content for documentation generation. This involves reading the file, splitting it into manageable chunks, and formatting those chunks with relevant context for input to a language model. The adapter ensures that the input to the documentation generator respects size limitations while attempting to maintain logical code boundaries.

**Key Classes and Their Roles**

*   **`PythonAdapter`**: This class inherits from `BaseAdapter` and implements the specific logic for handling Python files. It encapsulates the parsing and formatting steps necessary to prepare Python code for documentation.

**Important Functions and Their Behavior**

*   **`can_handle(file_path: Path) -> bool`**: This function determines if the adapter can process a given file based on its extension. It returns `True` if the file path ends with `.py`, and `False` otherwise. The `file_path` argument is a `Path` object representing the file's location.

*   **`parse(file_path: Path, content: str) -> List[str]`**: This function takes the file path and its content as input and splits the content into a list of strings (chunks). The goal is to create chunks that are within the `TARGET_CHUNK_SIZE` limit (approximately 6k tokens). The parsing logic prioritizes breaking chunks at logical boundaries, such as the end of a class or function definition, to maintain code context. If the entire file content is smaller than the target size, it returns a list containing a single chunk with the entire content. The function returns a `List[str]`, where each string represents a chunk of the original file content.

*   **`_format_chunk(file_path: Path, content: str, part: int = None) -> str`**: This is a helper function that formats a single chunk of code with contextual information. It prepends the file path and an optional part number to the chunk, and wraps the code content in a markdown code block. The `file_path` argument is a `Path` object, `content` is the code chunk as a string, and `part` is an optional integer indicating the chunk number. It returns a formatted string.

*   **`get_prompt(file_path: Path, parsed_content: str) -> str`**: This function constructs the prompt that will be sent to the language model. It includes instructions for the language model, specifying its role as a principal architect and outlining the desired documentation style and content. It also includes the parsed content of the Python file. The `file_path` argument is a `Path` object, and `parsed_content` is the string representing the code chunk. It returns a string containing the complete prompt.

**Type Hints and Their Significance**

The code extensively uses type hints (e.g., `file_path: Path`, `content: str`, `-> List[str]`). These hints improve code readability and maintainability by explicitly specifying the expected data types for function arguments and return values. They also enable static analysis tools to detect potential type errors during development.

**Notable Patterns or Design Decisions**

*   **Adapter Pattern**: The `PythonAdapter` follows the Adapter pattern, inheriting from a base class (`BaseAdapter`) to provide a consistent interface for handling different file types. This allows for easy extension to support other languages or file formats.

*   **Chunking Strategy**: The `parse` function employs a line-based chunking strategy with an attempt to respect code structure. This approach balances the need to stay within the token limit with the desire to provide meaningful context to the documentation generator.

*   **Markdown Formatting**: The `_format_chunk` function uses Markdown formatting to clearly delineate the file context and code content, making it easier for the language model to process the information.

*   **Prompt Engineering**: The `get_prompt` function carefully crafts a prompt that guides the language model to generate high-quality documentation, specifying the desired role, style, and content focus.