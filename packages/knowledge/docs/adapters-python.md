---
type: Documentation
domain: knowledge
last_modified: 2026-02-02
generated: true
source: packages/knowledge/adapters/python.py
generated_at: 2026-02-02T22:25:50.974175
hash: ff0834066863d556c68cf4606cee7ffb1ea087d352c29c7cbbe82225e8b238bc
---

## Python Adapter Documentation

This module provides a language adapter for generating documentation from Python source files. It is designed to be part of a larger documentation generation system.

**Responsibilities:**

The `PythonAdapter` handles the parsing, chunking, validation, and prompt formatting specific to Python code. It prepares Python source code for processing by a language model to produce documentation.

**Key Classes:**

*   **`PythonAdapter`**: This class inherits from `BaseAdapter` and implements the logic for handling Python files. It defines methods for determining if a file can be handled, parsing the file content into chunks, validating the content, and formatting prompts for a language model.

**Important Functions:**

*   **`can_handle(file_path: Path) -> bool`**:
    This function checks if the adapter can handle a given file based on its extension. It returns `True` if the file has a `.py` extension, and `False` otherwise.

*   **`parse(file_path: Path, content: str) -> List[str]`**:
    This function parses the content of a Python file into smaller chunks. It aims to split the content intelligently, respecting class and function boundaries to maintain context. The `TARGET_CHUNK_SIZE` constant (set to 24000 characters, approximately 6k tokens) controls the maximum size of each chunk. If the file content is smaller than this size, it returns a single chunk containing the entire content. Otherwise, it splits the content into multiple chunks, attempting to break at logical points (e.g., the end of a function or class definition) to avoid splitting code in the middle of a statement. The function returns a list of strings, where each string represents a chunk of the original content.

*   **`validate_content(content: str) -> List[str]`**:
    This function validates the Python content using the `ast` (Abstract Syntax Trees) module. It checks for syntax errors and, potentially, undefined variables. If any errors are found, they are returned as a list of error messages. Currently, the undefined variable check is disabled to reduce false positives with code snippets.

*   **`_format_chunk(file_path: Path, content: str, part: int = None) -> str`**:
    This is a helper function that formats a chunk of Python code with file context. It adds a header indicating the file path and, if applicable, the chunk number. The code content is enclosed in a code block using backticks.

*   **`get_prompt(file_path: Path, parsed_content: str) -> str`**:
    This function constructs a prompt for a language model. The prompt includes instructions for the model, specifying its role as a principal architect and outlining the desired documentation style and content. It also includes the parsed content of the Python file.

**Type Hints:**

The code makes extensive use of type hints (e.g., `file_path: Path`, `content: str`, `-> List[str]`). These hints improve code readability and allow for static analysis, helping to catch potential errors during development.

**Notable Patterns and Design Decisions:**

*   **Adapter Pattern:** The `PythonAdapter` follows the adapter pattern, inheriting from a `BaseAdapter` class. This allows for easy extension to support other languages by creating new adapter classes.
*   **Chunking Strategy:** The `parse` function employs a line-based chunking strategy with awareness of code structure. This approach balances the need to keep chunks within a manageable size for the language model with the desire to preserve context.
*   **Error Handling:** The `validate_content` function includes basic error handling to catch syntax errors and potential issues with the Python code.
*   **Prompt Engineering:** The `get_prompt` function carefully crafts a prompt to guide the language model in generating high-quality documentation. The prompt includes specific instructions, constraints, and a clear definition of the desired output format.