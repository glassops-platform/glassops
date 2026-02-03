---
type: Documentation
domain: knowledge
last_modified: 2026-02-02
generated: true
source: packages/knowledge/tests/test_validator.py
generated_at: 2026-02-02T22:32:34.073083
hash: 7bca9ccdc760ee3264d389580b867a46ba11ab0ff95ae0b47105233dc7cf7c8c
---

## Knowledge Validator Documentation

This document details the functionality of the Knowledge Validator, a component designed to assess the quality and adherence to guidelines of knowledge content. It ensures content meets defined standards before being incorporated into a knowledge base.

### Module Responsibilities

The primary responsibility of this module is to validate knowledge content, identifying potential issues such as missing metadata, prohibited language, and code quality concerns. The validator returns a structured report detailing any errors, warnings, or successful validations.

### Key Classes

**Validator:** This class serves as the central point for content validation. It orchestrates the validation process, including frontmatter checks, banned phrase/word detection, and code block validation via language-specific adapters.

### Important Functions

**Validator.validate(content: str) -> dict:**

This is the core function of the module. It accepts a string `content` representing the knowledge article and returns a dictionary containing validation results. The dictionary has the following keys:

*   `errors`: A list of strings, each representing a validation error.
*   `warnings`: A list of strings, each representing a potential issue or guideline violation.
*   `passes`: A list of strings, indicating successful validations (e.g., code block validation).

The function performs the following checks:

1.  **Frontmatter Check:** Verifies the presence of a frontmatter block (delimited by `---`).  If missing, an error is reported.
2.  **Banned Phrase Detection:**  Scans the content for conversational phrases. If found, a warning is added to the results.
3.  **Banned Word Detection:**  Scans the content for prohibited words. If found, a warning is added to the results.
4.  **Banned Term Detection:** Scans the content for specific prohibited terms. If found, a warning is added to the results.
5.  **Code Block Validation:** Identifies code blocks within the content and delegates validation to a language-specific adapter.

### Design Patterns and Decisions

*   **Adapter Pattern:** The validator employs an adapter pattern for code block validation.  The `get_adapter_for_lang` method (used in testing) is intended to retrieve an appropriate adapter based on the programming language detected in the code block. This allows for easy extension to support new languages without modifying the core validator logic.
*   **Structured Reporting:** The use of a dictionary to return validation results provides a clear and organized way to communicate the outcome of the validation process.
*   **Type Hints:** Type hints (e.g., `content: str -> dict`) are used to improve code readability and maintainability, and to enable static analysis.

### Test Cases Overview

The test suite covers the following scenarios:

*   **Basic Validation:**  Confirms that valid content passes validation without errors or warnings.
*   **Missing Frontmatter:**  Verifies that content lacking a frontmatter block is flagged with an error.
*   **Banned Phrases/Words:**  Tests the detection of prohibited phrases and words, resulting in warnings.
*   **Banned Terms:** Tests the detection of prohibited terms, resulting in warnings.
*   **Code Block Validation (Success):**  Demonstrates successful delegation to a code adapter and the reporting of a successful validation.
*   **Code Block Validation (Failure):**  Tests the handling of errors returned by a code adapter, resulting in an error in the validation report.
*   **Unknown Language:** Confirms that code blocks with unknown languages are ignored without causing errors.

You can extend the functionality by creating new adapters for different languages and adding more sophisticated validation rules.