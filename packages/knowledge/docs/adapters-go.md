---
type: Documentation
domain: knowledge
origin: packages/knowledge/generation/adapters/go.py
last_modified: 2026-02-01
generated: true
source: packages/knowledge/generation/adapters/go.py
generated_at: 2026-02-01T19:30:52.047993
hash: e352c93ef3e7fa4196bbc8dcd488f396a6f7ff1fd9edb8a91019e9a4c7ea986b
---

## Go Language Adapter Documentation

This module provides an adapter for generating documentation from Go source files. It is designed to be part of a larger documentation generation system, responsible for parsing Go code into manageable chunks and preparing prompts for a language model.

**Key Classes:**

*   **`GoAdapter`**: This class inherits from `BaseAdapter` and implements the logic specific to Go files. It handles file type checking, parsing, and prompt generation.

**Important Functions:**

*   **`can_handle(file_path: Path) -> bool`**: This function determines if the adapter can process a given file based on its extension. It returns `True` if the file has a `.go` extension, and `False` otherwise. The `file_path` argument is a `Path` object representing the file's location.

*   **`parse(file_path: Path, content: str) -> List[str]`**: This function parses the content of a Go file into a list of string chunks. The goal is to create semantically meaningful chunks for better documentation generation.
    *   If the file content is smaller than `TARGET_CHUNK_SIZE`, it returns a single chunk containing the entire file content.
    *   Otherwise, it attempts to split the content based on top-level declarations (functions, types, constants, variables) using a regular expression. This aims to preserve logical boundaries within the code.
    *   If splitting by declarations isn't sufficient (e.g., for very large files), it falls back to line-based chunking to ensure that no chunk exceeds `TARGET_CHUNK_SIZE`.
    *   The `file_path` argument is a `Path` object, and `content` is a string containing the file's content. The function returns a list of strings, where each string represents a chunk of the original content.

*   **`_format_chunk(file_path: Path, content: str, part: int = None) -> str`**: This is a helper function that formats a chunk of Go code with file context. It adds the file path and an optional part number to the beginning of the chunk, and wraps the code in a markdown code block. The `file_path` argument is a `Path` object, `content` is the chunk's content, and `part` is an optional integer indicating the chunk number. It returns a formatted string.

*   **`get_prompt(file_path: Path, parsed_content: str) -> str`**: This function constructs a prompt for a language model, providing instructions and context for generating documentation. It includes a detailed role description for the language model (principal architect), specific instructions on the desired output format, and a list of focus areas (package purpose, key types, functions, error handling, concurrency, design decisions). The `file_path` argument is a `Path` object, and `parsed_content` is the chunk of code to be documented. It returns a string containing the prompt.

**Type Hints:**

The code makes extensive use of type hints (e.g., `file_path: Path`, `content: str`, `-> List[str]`). These hints improve code readability and allow for static analysis, helping to catch potential errors early in the development process.

**Notable Patterns and Design Decisions:**

*   **Adapter Pattern:** The `GoAdapter` class follows the adapter pattern, allowing the documentation generation system to work with different languages without modifying the core logic.
*   **Chunking Strategy:** The `parse` function employs a hybrid chunking strategy, attempting to split code based on semantic boundaries (declarations) while ensuring that chunks remain within a manageable size. This balances semantic coherence with the limitations of language models.
*   **Prompt Engineering:** The `get_prompt` function demonstrates careful prompt engineering, providing clear instructions and context to the language model to guide the documentation generation process.
*   **`TARGET_CHUNK_SIZE` Constant:** The `TARGET_CHUNK_SIZE` constant defines the maximum size of a chunk in tokens. This value is chosen to be conservative, ensuring that the language model can process the chunks effectively.