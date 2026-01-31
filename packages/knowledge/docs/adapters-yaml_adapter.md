---
type: Documentation
domain: knowledge
origin: packages/knowledge/generation/adapters/yaml_adapter.py
last_modified: 2026-01-31
generated: true
source: packages/knowledge/generation/adapters/yaml_adapter.py
generated_at: 2026-01-31T08:57:47.709493
hash: 02479b1d1aa043a20a103d0efcc3bdbdb68628a7e2e1134784aada793175c79b
---

## YAML Adapter Documentation

This document describes the YAML Adapter, a component responsible for processing YAML configuration files as input for documentation generation. It handles file identification, parsing into manageable chunks, formatting, and constructing prompts for a language model.

**Module Responsibilities:**

The primary responsibility of this module is to adapt YAML files into a format suitable for processing by a language model. This involves determining if a file is a YAML file, splitting large files into smaller parts, and preparing the content along with a prompt that instructs the language model on how to document the YAML configuration.

**Key Classes:**

* **`YAMLAdapter`**: This class inherits from `BaseAdapter` and implements the specific logic for handling YAML files. It defines how to identify YAML files, parse their content, format the content for inclusion in a prompt, and generate the prompt itself.

**Important Functions:**

* **`can_handle(file_path: Path) -> bool`**: This function determines whether the adapter can process a given file based on its extension. It returns `True` if the file path’s suffix is either ".yml" or ".yaml", and `False` otherwise. The `file_path` argument is a `Path` object representing the file's location.
* **`parse(file_path: Path, content: str) -> List[str]`**: This function takes the file path and the file content as input and splits the content into a list of strings (chunks).  It addresses the limitation of language models regarding input size by dividing large YAML files into smaller, more manageable chunks, each not exceeding `TARGET_CHUNK_SIZE` characters. The function returns a list of these formatted chunks.
* **`_format_chunk(file_path: Path, content: str, part: int = None) -> str`**: This private helper function formats a single chunk of YAML content. It prepends the file path and an optional part number to the content, wraps the content in a code block, and returns the formatted string. The `part` argument, if provided, indicates the chunk number within the original file.
* **`get_prompt(file_path: Path, parsed_content: str) -> str`**: This function constructs a prompt to be sent to the language model. The prompt instructs the model to act as a DevOps engineer and technical writer, documenting the provided YAML configuration. It emphasizes the need for valid Markdown output, prohibits conversational text, and specifies stylistic constraints (no emojis, certain words excluded, and pronoun usage). The `parsed_content` argument represents the YAML content to be documented.

**Type Hints:**

The code extensively uses type hints (e.g., `file_path: Path`, `content: str`, `-> List[str]`). These hints improve code readability and allow for static analysis, helping to catch potential errors during development. They clearly define the expected data types for function arguments and return values.

**Notable Patterns and Design Decisions:**

* **Adapter Pattern:** The `YAMLAdapter` class follows the Adapter pattern, allowing the system to work with YAML files in a consistent manner alongside other potential configuration file types.  It provides a common interface (`BaseAdapter`) for different file formats.
* **Chunking:** The `parse` function implements a chunking mechanism to handle large YAML files that exceed the input size limits of the language model. This ensures that even extensive configurations can be documented.
* **Prompt Engineering:** The `get_prompt` function demonstrates careful prompt engineering, providing clear instructions and constraints to the language model to ensure the generation of high-quality documentation.
* **String Formatting:** The code uses f-strings for clear and concise string formatting, improving readability and maintainability.
* **Constant for Chunk Size:** The `TARGET_CHUNK_SIZE` constant allows for easy adjustment of the maximum chunk size without modifying the core logic.