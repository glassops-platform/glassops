---
type: Documentation
domain: knowledge
last_modified: 2026-03-08
generated: true
source: packages/knowledge/generation/validator.py
generated_at: 2026-03-08T22:45:49.620276
hash: 753cecf576fa02e8dbb106251a2fbba39f348caba02e1caf5a89e791178db609
---

## Documentation Validator Module

This module provides functionality to validate generated documentation content for quality and syntax issues. It aims to ensure consistency and adherence to style guidelines. We designed it to be extensible through adapters for different programming languages and file formats.

**Key Classes:**

*   **Validator:** This class contains the core validation logic. It is responsible for checking the content against a set of predefined rules and delegating code block validation to language-specific adapters. It operates as a collection of class methods, providing a centralized point for validation tasks.

**Important Functions:**

*   **`get_adapter_for_lang(lang: str) -> Optional[BaseAdapter]`:** This class method acts as a factory, returning the appropriate adapter instance based on the provided language string. The `lang` parameter is case-insensitive. It supports Python, Go, HTML/XML (through LWCAdapter), and Apex. If no adapter is found for the given language, it returns `None`. The type hint `Optional[BaseAdapter]` indicates that the function may return either a `BaseAdapter` object or `None`.
*   **`extract_code_blocks(content: str) -> List[tuple[str, str]]`:** This class method extracts code blocks from markdown content using a regular expression. It returns a list of tuples, where each tuple contains the language of the code block and the code itself. The `re.DOTALL` flag ensures that the regular expression matches across multiple lines.
*   **`validate(content: str, file_path: str = "") -> dict`:** This is the main validation function. It takes the documentation content and an optional file path as input. It performs several checks, including:
    *   Presence of a frontmatter block.
    *   Absence of banned conversational phrases.
    *   Absence of banned words.
    *   Absence of specific terms.
    *   Validation of code blocks using appropriate adapters.
    It returns a dictionary containing lists of passes, warnings, and errors.
*   **`print_report(results: dict)`:** This static method takes the validation results dictionary and prints a formatted report to the console, clearly indicating any errors or warnings found.

**Type Hints:**

The code makes extensive use of type hints (e.g., `lang: str`, `-> Optional[BaseAdapter]`, `content: str`, `-> dict`). These hints improve code readability and allow for static analysis, helping to catch potential errors during development.

**Notable Patterns and Design Decisions:**

*   **Adapter Pattern:** The use of adapters (`BaseAdapter`, `PythonAdapter`, `GoAdapter`, etc.) allows for easy extension to support new languages and file formats without modifying the core `Validator` class. You can add new adapters by implementing the `BaseAdapter` interface.
*   **Class Methods:** The validation functions are implemented as class methods, which allows them to be called without creating an instance of the `Validator` class.
*   **Banned Phrase/Word Lists:** The use of `BANNED_PHRASES` and `BANNED_WORDS` lists makes it easy to customize the validation rules.
*   **Clear Result Structure:** The `validate` function returns a dictionary with a well-defined structure (`passes`, `warnings`, `errors`), making it easy to process the validation results programmatically.
*   **Regular Expressions:** Regular expressions are used for extracting code blocks, providing a flexible and powerful way to parse the markdown content.