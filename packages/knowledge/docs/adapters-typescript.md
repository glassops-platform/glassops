---
type: Documentation
domain: knowledge
last_modified: 2026-02-02
generated: true
source: packages/knowledge/adapters/typescript.py
generated_at: 2026-02-02T22:26:24.090726
hash: 9356d32a599402f09de7b513080f979bae64821623b24dc2e9f8614e4b63c7f6
---

## TypeScript Adapter Documentation

This module provides an adapter for processing TypeScript and JavaScript source files during documentation generation. It is designed to split files into manageable chunks and format them for input to a language model. The adapter’s functionality is inspired by a TypeScript implementation.

**Responsibilities:**

*   Determine if a file can be handled based on its extension.
*   Parse file content into chunks of a defined size.
*   Format chunks with file context for inclusion in prompts.
*   Provide a base for content validation (currently a placeholder).
*   Construct a prompt for a language model to generate documentation.

### TypeScriptAdapter Class

The `TypeScriptAdapter` class inherits from `BaseAdapter` and implements the core logic for handling TypeScript and JavaScript files.

**Attributes:**

*   `TARGET_CHUNK_SIZE`: An integer representing the maximum size of a chunk in characters (currently 24000). This value is intended to correspond to approximately 6000 tokens.

**Methods:**

*   `can_handle(file_path: Path) -> bool`:
    This method checks if the adapter can handle a given file based on its extension. It returns `True` if the file extension is one of `.ts`, `.js`, `.mjs`, `.tsx`, or `.jsx`; otherwise, it returns `False`.

*   `parse(file_path: Path, content: str) -> List[str]`:
    This method parses the content of a TypeScript or JavaScript file into a list of chunks. It splits the content into chunks that are no larger than `TARGET_CHUNK_SIZE`. If the entire content is smaller than `TARGET_CHUNK_SIZE`, it returns a list containing a single chunk with the entire content. The method iterates through the lines of the content, adding them to the current chunk until the `TARGET_CHUNK_SIZE` is exceeded. It then creates a new chunk and continues.

*   `validate_content(content: str) -> List[str]`:
    This method is currently a placeholder for content validation. It always returns an empty list. We intend to add functionality to validate the TypeScript content in future versions.

*   `_format_chunk(file_path: Path, content: str, part: int = None) -> str`:
    This private method formats a chunk of content with file context. It adds the file path and an optional part number to the beginning of the chunk, and wraps the content in a code block with the `typescript` language identifier. The `part` argument is used to indicate the chunk number when a file is split into multiple chunks.

*   `get_prompt(file_path: Path, parsed_content: str) -> str`:
    This method constructs a prompt for a language model. The prompt instructs the model to act as a principal architect and translate the provided TypeScript/JavaScript code into high-level documentation. It includes specific instructions regarding the desired output format and constraints, such as avoiding certain words and phrases. The `parsed_content` is inserted directly into the prompt.

### Type Hints

The code makes extensive use of type hints to improve code readability and maintainability. For example:

*   `file_path: Path` indicates that the `file_path` argument should be a `Path` object.
*   `content: str` indicates that the `content` argument should be a string.
*   `-> List[str]` indicates that the method returns a list of strings.

These type hints help to prevent errors and make the code easier to understand.

### Design Decisions

*   **Chunking Strategy:** The `parse` method uses a line-based chunking strategy to avoid splitting lines of code. This helps to maintain the integrity of the code and makes it easier to understand.
*   **File Context:** The `_format_chunk` method adds file context to each chunk, which helps the language model to understand the origin of the code.
*   **Prompt Engineering:** The `get_prompt` method carefully crafts a prompt that instructs the language model to generate high-quality documentation. The prompt includes specific instructions regarding the desired output format and constraints.