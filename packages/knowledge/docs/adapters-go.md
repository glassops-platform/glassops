---
type: Documentation
domain: knowledge
last_modified: 2026-02-02
generated: true
source: packages/knowledge/adapters/go.py
generated_at: 2026-02-02T22:24:55.803883
hash: 620b071e516120f8ebbfd0bb801601cb11dd0cd5df9ce23957661d69db49c7d2
---

## Go Adapter Documentation

This module provides an adapter for processing Go source files for documentation generation. It inherits from the `BaseAdapter` class and implements specific logic for parsing and validating Go code.

**Responsibilities:**

- Identifying Go files based on their `.go` extension.
- Parsing Go file content into smaller chunks suitable for large language models.
- Validating Go code syntax using the `go fmt` tool.
- Formatting chunks with file context for inclusion in prompts.
- Constructing prompts for documentation generation.

**Key Classes:**

- **`GoAdapter`**: This class encapsulates the logic for handling Go files. It extends `BaseAdapter` and overrides methods to provide Go-specific functionality.

**Important Functions:**

- **`can_handle(file_path: Path) -> bool`**:  Determines if the adapter can process a given file based on its extension. It returns `True` if the file has a `.go` extension, and `False` otherwise. The `file_path` argument is a `Path` object representing the file's location.
- **`parse(file_path: Path, content: str) -> List[str]`**: Parses the content of a Go file into a list of string chunks.  The function attempts to split the content at semantic boundaries, such as function or type declarations, to maintain context within each chunk. If the file is small enough, it returns a single chunk containing the entire content.  If semantic splitting isn't possible or the file is very large, it falls back to line-based chunking. The `file_path` argument is a `Path` object, and `content` is the string content of the file. The return value is a list of strings, where each string represents a chunk of the original content.
- **`validate_content(content: str) -> List[str]`**: Validates the Go code syntax using the `go fmt` tool. It checks if the `go` executable is available in the system's PATH. If `go fmt` reports errors, the function returns a list of error messages. If `go` is not found, it returns a message indicating that validation was skipped. The `content` argument is the string content of the Go file. The return value is a list of strings, where each string represents a validation error message.
- **`_format_chunk(file_path: Path, content: str, part: int = None) -> str`**: Formats a chunk of Go code with file context. It adds the file path and an optional part number to the beginning of the chunk, and wraps the code in a markdown code block. The `file_path` argument is a `Path` object, `content` is the string content of the chunk, and `part` is an optional integer representing the chunk number. The return value is a formatted string.
- **`get_prompt(file_path: Path, parsed_content: str) -> str`**: Constructs a prompt for a large language model to generate documentation for the given Go code. The prompt includes instructions on the desired documentation style, focus areas (package purpose, key types, functions, error handling, concurrency), and strict rules for the output format. The `file_path` argument is a `Path` object, and `parsed_content` is the string content of the chunk to be documented. The return value is a string containing the prompt.

**Type Hints:**

The code makes extensive use of type hints (e.g., `file_path: Path`, `content: str`, `-> List[str]`). These hints improve code readability and allow for static analysis, helping to catch potential errors during development.

**Notable Patterns and Design Decisions:**

- **Adapter Pattern:** The `GoAdapter` class follows the adapter pattern, allowing the documentation generation system to work with Go files in a consistent manner, regardless of the specific file format.
- **Chunking Strategy:** The `parse` function employs a sophisticated chunking strategy that attempts to split the code at semantic boundaries to preserve context. This is important for generating accurate and meaningful documentation.
- **External Tool Validation:** The `validate_content` function leverages the `go fmt` tool for syntax validation, ensuring that only valid Go code is processed.
- **Prompt Engineering:** The `get_prompt` function carefully crafts a prompt that guides the large language model to generate high-quality documentation.
- **Error Handling:** The `validate_content` function includes robust error handling to gracefully handle cases where the `go` executable is not found or when `go fmt` encounters errors.