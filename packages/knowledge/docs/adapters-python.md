---
type: Documentation
domain: knowledge
origin: packages/knowledge/generation/adapters/python.py
last_modified: 2026-01-28
generated: true
source: packages/knowledge/generation/adapters/python.py
generated_at: 2026-01-28T22:43:16.696858
hash: 01f09935871c40b71b24ec76a18a51e62995920a97c119c08c60e789e00bf4cd
---

## Python Adapter Documentation

This document details the functionality of the Python Adapter, a component designed for generating documentation from Python source code files. It serves as an interface within a larger documentation generation system.

**Module Purpose and Responsibilities**

The Python Adapter is responsible for identifying, parsing, and preparing Python files for documentation generation. It handles the process of breaking down large files into smaller, manageable chunks suitable for processing by a language model. The adapter ensures that these chunks respect logical boundaries within the code, such as class and function definitions, to maintain context.

**Key Classes and Their Roles**

*   **`PythonAdapter`**: This is the core class of the adapter. It inherits from `BaseAdapter` and implements the specific logic for handling Python files. It determines if a file can be processed, parses the file content into chunks, formats those chunks, and constructs a prompt for the language model.

**Important Functions and Their Behavior**

*   **`can_handle(file_path: Path) -> bool`**: This function checks if the adapter can process a given file based on its extension. It returns `True` if the file has a `.py` extension, indicating a Python source file, and `False` otherwise. The `file_path` argument is a `Path` object representing the file's location.
*   **`parse(file_path: Path, content: str) -> List[str]`**: This function takes the file path and content of a Python file as input and divides the content into a list of strings (chunks). It employs a line-based chunking strategy, aiming to keep each chunk below `TARGET_CHUNK_SIZE` (approximately 6000 tokens). The parsing attempts to split the content at logical boundaries – class or function definitions – to preserve code context. If the file content is smaller than the target chunk size, it returns a list containing the entire content as a single chunk.
*   **`_format_chunk(file_path: Path, content: str, part: int = None) -> str`**: This is a helper function that formats a single chunk of Python code. It prepends the file path and an optional part number to the chunk, and wraps the code within a markdown code block for clarity. The `content` argument is the string representing the code chunk, and `part` is an optional integer indicating the chunk number.
*   **`get_prompt(file_path: Path, parsed_content: str) -> str`**: This function constructs the prompt that will be sent to the language model. It includes instructions for the model, specifying its role as a principal architect, the desired output format, and a list of constraints. It then appends the parsed content of the Python file to the prompt. The `parsed_content` argument is a string representing the code chunk to be documented.

**Type Hints and Their Significance**

The code extensively uses type hints (e.g., `file_path: Path`, `content: str`, `-> List[str]`). These hints improve code readability and maintainability by explicitly defining the expected data types for function arguments and return values. They also enable static analysis tools to detect potential type errors during development.

**Notable Patterns or Design Decisions**

*   **Adapter Pattern**: The `PythonAdapter` follows the Adapter pattern, allowing the documentation generation system to work with different file types without modifying the core logic.
*   **Chunking Strategy**: The line-based chunking strategy with awareness of class/function boundaries is designed to balance chunk size limitations with the need to preserve code context.
*   **Prompt Engineering**: The `get_prompt` function demonstrates a deliberate approach to prompt engineering, providing clear instructions and constraints to the language model to guide its documentation generation process.
*   **Configuration**: The `TARGET_CHUNK_SIZE` constant allows for easy adjustment of the maximum chunk size without modifying the core parsing logic.