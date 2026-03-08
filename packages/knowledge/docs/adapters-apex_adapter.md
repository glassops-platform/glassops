---
type: Documentation
domain: knowledge
last_modified: 2026-03-08
generated: true
source: packages/knowledge/adapters/apex_adapter.py
generated_at: 2026-03-08T22:43:17.417918
hash: b8cae6f14bd4a3c54afb37fe610ff7087405547feb09f526e9cf469d79e95f3c
---

## Salesforce Apex Adapter Documentation

This adapter is designed to process Salesforce Apex code (classes and triggers) for documentation generation. It handles the parsing, chunking, and formatting of Apex source files, preparing them for input into a language model for documentation creation.

**Key Responsibilities:**

*   Identifying Apex files based on their extensions (.cls for classes, .trigger for triggers).
*   Reading and processing the content of Apex files.
*   Splitting large files into smaller chunks to accommodate language model input limits.
*   Formatting the content and associated metadata for consistent input to the documentation generation process.
*   Constructing a prompt for the language model, instructing it to generate comprehensive documentation for the provided Apex code.

**ApexAdapter Class:**

The `ApexAdapter` class inherits from the `BaseAdapter` class and implements the specific logic for handling Apex files.

*   `TARGET_CHUNK_SIZE`:  A constant defining the maximum size (in characters) of a single chunk of Apex code. This value is set to 24000.
*   `can_handle(file_path: Path) -> bool`: This method determines if the adapter can process a given file based on its extension. It returns `True` if the file extension is ".cls" or ".trigger", and `False` otherwise. The `file_path` argument is a `Path` object representing the file.
*   `parse(file_path: Path, content: str) -> List[str]`: This method takes the file path and content of an Apex file as input and returns a list of strings, where each string represents a chunk of the file content. It handles files larger than `TARGET_CHUNK_SIZE` by splitting them into smaller chunks. It also attempts to read a corresponding metadata file (with a "-meta.xml" suffix) and includes it with the first chunk. The `file_path` argument is a `Path` object, and `content` is the file content as a string.
*   `validate_content(content: str) -> List[str]`: This method currently returns an empty list. It is intended for future implementation of content validation checks. The `content` argument is the file content as a string.
*   `_format_chunk(file_path: Path, content: str, part: int = None, metadata: str = None) -> str`: This private method formats a single chunk of Apex code into a string suitable for input to the language model. It includes the file path, file type (Apex Class or Apex Trigger), an optional part number (for chunks), the Apex code itself (wrapped in ```apex```), and optionally the metadata (wrapped in ```xml```). The `file_path` argument is a `Path` object, `content` is the chunk content, `part` is an optional integer representing the chunk number, and `metadata` is the metadata content as a string.
*   `get_prompt(file_path: Path, parsed_content: str) -> str`: This method constructs a prompt for the language model, providing instructions on how to document the Apex code. The prompt specifies the desired documentation elements (purpose, methods, governor limits, integration points, test coverage) and includes strict formatting rules. The `file_path` argument is a `Path` object, and `parsed_content` is the chunk content.

**Type Hints:**

The code makes extensive use of type hints (e.g., `file_path: Path`, `content: str`, `-> List[str]`). These hints improve code readability and allow for static analysis, helping to catch potential errors during development.

**Design Decisions and Patterns:**

*   **Adapter Pattern:** The `ApexAdapter` class follows the Adapter pattern, allowing the documentation generation system to work with different code types (Apex in this case) without modifying the core system.
*   **Chunking:** The implementation splits large files into chunks to avoid exceeding the input limits of the language model. The `TARGET_CHUNK_SIZE` constant controls the maximum size of each chunk.
*   **Metadata Handling:** The adapter attempts to read and include metadata associated with the Apex file, providing additional context for documentation generation.
*   **Prompt Engineering:** The `get_prompt` method carefully crafts a prompt to guide the language model in generating high-quality documentation. The prompt includes specific instructions and formatting requirements.
*   **Error Handling:** The code includes basic error handling when reading the metadata file, logging warnings if the file cannot be read.