---
type: Documentation
domain: knowledge
origin: packages/knowledge/generation/adapters/yaml_adapter.py
last_modified: 2026-01-31
generated: true
source: packages/knowledge/generation/adapters/yaml_adapter.py
generated_at: 2026-01-31T09:53:21.187505
hash: 02479b1d1aa043a20a103d0efcc3bdbdb68628a7e2e1134784aada793175c79b
---

## YAML Adapter Documentation

This document describes the YAML Adapter, a component responsible for processing YAML configuration files as input for documentation generation. It handles parsing, chunking, and prompt creation for large YAML files to work effectively with language models.

**Module Responsibilities:**

The primary responsibility of this module is to adapt YAML files into a format suitable for documentation generation. This involves determining if a file is a YAML file, splitting large files into smaller chunks, and constructing a prompt that instructs a language model to document the YAML content.

**Key Classes:**

*   **`YAMLAdapter`**: This class inherits from `BaseAdapter` and implements the specific logic for handling YAML files. It defines methods for identifying YAML files, parsing their content, formatting chunks, and generating prompts.

**Important Functions:**

*   **`can_handle(file_path: Path) -> bool`**: This function determines whether the adapter can process a given file based on its extension. It returns `True` if the file path has a `.yml` or `.yaml` extension, and `False` otherwise. The `file_path` argument is a `Path` object representing the file's location.
*   **`parse(file_path: Path, content: str) -> List[str]`**: This function parses the content of a YAML file and splits it into chunks if the content exceeds a predefined size (`TARGET_CHUNK_SIZE`). It returns a list of strings, where each string represents a chunk of the YAML content, formatted for use with a language model. The `file_path` argument is a `Path` object, and `content` is a string containing the YAML file's content.
*   **`_format_chunk(file_path: Path, content: str, part: int = None) -> str`**: This private helper function formats a single chunk of YAML content. It adds metadata such as the file path and chunk number (if applicable) and wraps the content in a Markdown code block. The `file_path` argument is a `Path` object, `content` is the YAML chunk as a string, and `part` is an optional integer representing the chunk number.
*   **`get_prompt(file_path: Path, parsed_content: str) -> str`**: This function constructs a prompt that instructs a language model to document the provided YAML content. The prompt includes instructions on the desired output format (Markdown), specific constraints (no conversational text, no mention of certain terms), and a clear task description. The `file_path` argument is a `Path` object, and `parsed_content` is a string containing the YAML content to be documented.

**Type Hints:**

The code extensively uses type hints (e.g., `file_path: Path`, `content: str`, `-> List[str]`). These hints improve code readability and allow for static analysis, helping to catch potential errors during development. They clearly define the expected data types for function arguments and return values.

**Notable Patterns and Design Decisions:**

*   **Adapter Pattern:** The `YAMLAdapter` class follows the Adapter pattern, allowing it to interface with a generic documentation generation pipeline while handling the specific details of YAML file processing.
*   **Chunking:** Large YAML files are split into smaller chunks to avoid exceeding the input limits of language models. The `TARGET_CHUNK_SIZE` constant defines the maximum size of each chunk.
*   **Prompt Engineering:** The `get_prompt` function demonstrates careful prompt engineering to guide the language model towards generating the desired documentation output. The prompt includes specific instructions and constraints to ensure quality and consistency.
*   **Markdown Formatting:** The adapter consistently formats YAML content within Markdown code blocks to ensure proper rendering in documentation.