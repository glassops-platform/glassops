---
type: Documentation
domain: knowledge
last_modified: 2026-02-02
generated: true
source: packages/knowledge/generation/validator.py
generated_at: 2026-02-02T22:29:43.482366
hash: 211f4ed3b51a8bef6cf2581464bd112fdbefc42ec04a9dd621375a587353962e
---

## Documentation Validator Module

This module provides functionality to validate generated documentation content for quality and syntax issues. It aims to ensure documentation is professional, concise, and free of common problems like conversational filler or banned terminology.

**Key Classes:**

*   **Validator:** This class contains the core validation logic. It is designed as a collection of class methods, offering a centralized point for performing various checks on the documentation content.  It does not require instantiation.

**Important Functions:**

*   **`get_adapter_for_lang(cls, lang: str) -> Optional[BaseAdapter]`:** This class method acts as a factory, returning an appropriate adapter object based on the detected programming language of a code block. The `lang` parameter is a string representing the language (e.g., "python", "go"). It returns `None` if no adapter is found for the given language. Type hinting ensures the input is a string and the output is either a `BaseAdapter` object or `None`.
*   **`extract_code_blocks(cls, content: str) -> List[tuple[str, str]]`:** This class method extracts code blocks from markdown content using regular expressions. It identifies blocks enclosed in triple backticks (```) and returns a list of tuples, where each tuple contains the language of the code block and the code itself. The `content` parameter is the markdown string to parse. The return value is a list of tuples, each containing a language string and a code string.
*   **`validate(cls, content: str, file_path: str = "") -> dict`:** This is the primary validation function. It takes the documentation `content` as input, along with an optional `file_path` for context. It performs a series of checks, including:
    *   Frontmatter presence
    *   Detection of banned conversational phrases
    *   Detection of banned words
    *   Detection of prohibited terms
    *   Delegation of code block validation to language-specific adapters.
    It returns a dictionary containing three lists: `passes`, `warnings`, and `errors`. These lists store the results of each validation check.
*   **`print_report(results: dict)`:** This static method takes the dictionary returned by the `validate` function and prints a formatted report to the console, clearly indicating any errors, warnings, or successful passes.

**Type Hints:**

The code extensively uses type hints (e.g., `lang: str`, `-> List[tuple[str, str]]`) to improve code readability and maintainability. These hints specify the expected data types for function parameters and return values, enabling static analysis and helping to prevent type-related errors.

**Notable Patterns and Design Decisions:**

*   **Class Methods:** The validation logic is implemented using class methods, allowing access to class-level constants (like `BANNED_PHRASES` and `BANNED_WORDS`) without requiring an instance of the `Validator` class.
*   **Adapter Pattern:** The use of adapters (e.g., `PythonAdapter`, `GoAdapter`) promotes loose coupling and allows for easy extension to support additional languages. Each adapter is responsible for validating code blocks in its specific language.
*   **Regular Expressions:** Regular expressions are used for extracting code blocks from the markdown content.
*   **Dictionary-Based Results:** The `validate` function returns a dictionary to provide a structured and comprehensive report of the validation results.
*   **Banned Phrase/Word Lists:** The use of lists for banned phrases and words makes it easy to maintain and update the validation rules.
*   **Static Method for Reporting:** The `print_report` method is static, meaning it doesn't require an instance of the class to be called, and is solely responsible for formatting and displaying the validation results.

**Adapters:**

The module depends on several adapter classes (defined in `glassops.knowledge.adapters`) to handle language-specific validation:

*   `BaseAdapter`:  The base class for all adapters.
*   `GoAdapter`: Validates Go code.
*   `PythonAdapter`: Validates Python code.
*   `LWCAdapter`: Validates HTML/XML content.
*   `ApexAdapter`: Validates Apex code.
*   `YAMLAdapter`: Validates YAML content.
*   `JSONAdapter`: Validates JSON content.
*   `DockerAdapter`: Validates Dockerfile content.
*   `TerraformAdapter`: Validates Terraform code.

You can extend the functionality of this module by creating new adapters for additional languages or file types.