---
type: Documentation
domain: knowledge
origin: packages/knowledge/generation/adapters/typescript.py
last_modified: 2026-01-31
generated: true
source: packages/knowledge/generation/adapters/typescript.py
generated_at: 2026-01-31T09:53:03.409674
hash: 38edc897ad829164a4fa8e31f1d033c31f53ae2e1924bbe7f308e5d525489af4
---

## TypeScript Adapter Documentation

This document details the functionality of the TypeScript Adapter, a component designed for documentation generation from TypeScript and JavaScript source code. It is a port of functionality originally present in a related project.

**Module Purpose:**

The primary responsibility of this adapter is to read TypeScript or JavaScript files, split their content into manageable chunks, and format those chunks for processing by a language model. This prepares the code for documentation generation.

**Key Classes:**

*   **`TypeScriptAdapter`**: This class inherits from `BaseAdapter` and implements the specific logic for handling TypeScript and JavaScript files. It determines if a file can be processed, parses the file content into chunks, and formats those chunks with relevant context.

**Important Functions:**

*   **`can_handle(file_path: Path) -> bool`**: This function checks if the adapter can process a given file based on its extension. It returns `True` if the file has a `.ts`, `.js`, `.mjs`, `.tsx`, or `.jsx` extension; otherwise, it returns `False`. The `file_path` argument is a `Path` object representing the file's location.

*   **`parse(file_path: Path, content: str) -> List[str]`**: This function takes the file path and content as input and splits the content into a list of strings (chunks). Each chunk is designed to be within a target size, approximately 6000 tokens, to accommodate language model input limits. The function intelligently splits the content by lines to avoid breaking code mid-statement. If the entire content is smaller than the target size, it returns a list containing a single chunk with the entire content. The `file_path` argument is a `Path` object, and `content` is a string representing the file's content. The return value is a `List` of strings, where each string is a chunk of the original content.

*   **`_format_chunk(file_path: Path, content: str, part: int = None) -> str`**: This is a helper function that formats a single chunk of code with file context. It prepends the file path and an optional part number to the chunk, and wraps the content in a markdown code block with `typescript` syntax highlighting. The `file_path` argument is a `Path` object, `content` is the string representing the code chunk, and `part` is an optional integer indicating the chunk number. The function returns a formatted string.

*   **`get_prompt(file_path: Path, parsed_content: str) -> str`**: This function constructs a prompt to be sent to a language model. The prompt instructs the model to act as a principal architect and generate high-level documentation from the provided code content. It includes specific instructions regarding the desired output format and constraints, such as avoiding certain words and phrases. The `file_path` argument is a `Path` object, and `parsed_content` is the string representing the code chunk. The function returns a string containing the prompt.

**Type Hints:**

The code extensively uses type hints (e.g., `file_path: Path`, `content: str`, `-> List[str]`). These hints improve code readability and allow for static analysis, helping to catch potential errors during development. They also clarify the expected input and output types for each function.

**Notable Patterns and Design Decisions:**

*   **Adapter Pattern:** The `TypeScriptAdapter` follows the Adapter pattern, inheriting from a `BaseAdapter` class. This allows for easy addition of support for other languages by creating new adapter classes that implement the same interface.
*   **Chunking Strategy:** The `parse` function implements a line-based chunking strategy to avoid splitting code in the middle of statements. This improves the quality of the generated documentation.
*   **Target Chunk Size:** The `TARGET_CHUNK_SIZE` constant defines the maximum size of each chunk. This value is chosen to balance the need for manageable chunks with the desire to preserve code context.
*   **Prompt Engineering:** The `get_prompt` function carefully crafts a prompt to guide the language model towards generating the desired documentation style and content. The prompt includes specific instructions and constraints to ensure high-quality output.