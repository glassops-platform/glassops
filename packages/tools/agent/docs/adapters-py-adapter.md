---
type: Documentation
domain: agent
origin: packages/tools/agent/src/adapters/py-adapter.ts
last_modified: 2026-02-01
generated: true
source: packages/tools/agent/src/adapters/py-adapter.ts
generated_at: 2026-02-01T19:49:02.505230
hash: 8d2e85d94103c29d72984c3f34046684fc1ec3b7c96954d4cc1e5cd60a1b7e5f
---

## Python File Adapter Documentation

This document details the functionality of the Python File Adapter, a component designed to process Python files within a larger agent system. This adapter handles file identification, content parsing, prompt generation, and output post-processing.

### Overview

The Python File Adapter serves as an interface between the agent and Python source code. It determines if a given file is a Python file, extracts its content, prepares a prompt for a language model to document the code, and then combines the model’s outputs into a cohesive document.

### Core Functionality

The adapter provides four primary functions:

1.  **File Type Identification (`canHandle`)**:
    *   **Purpose**: Determines whether the adapter can process a given file based on its extension.
    *   **Input**: `fileName` (string) – The name of the file to check.
    *   **Output**: `boolean` – Returns `true` if the file has a `.py` extension, indicating it is a Python file; otherwise, returns `false`.
    *   **Example**: `canHandle('my_script.py')` returns `true`. `canHandle('document.txt')` returns `false`.

2.  **Content Parsing (`parse`)**:
    *   **Purpose**: Extracts the content of a Python file. Currently, this function performs a simple pass-through of the file content. Future versions will incorporate Python Abstract Syntax Tree (AST) parsing to identify and separate classes and functions.
    *   **Input**:
        *   `filePath` (string) – The path to the Python file.
        *   `content` (string) – The content of the Python file.
    *   **Output**: `Promise<string[]>` – Returns a promise that resolves to an array of strings. Currently, this array contains a single string formatted with the file path and the complete Python code content enclosed in a code block.
    *   **Example**:
        ```typescript
        parse('/path/to/my_script.py', 'def my_function():\n  print("Hello")')
        ```
        returns a promise resolving to:
        `['File: /path/to/my_script.py\n\nPython Code Content:\n\`\`\`python\ndef my_function():\n  print("Hello")\`\`\`']`

3.  **Prompt Generation (`generatePrompt`)**:
    *   **Purpose**: Creates a prompt for a language model, instructing it to generate documentation for the provided Python code.
    *   **Input**:
        *   `filePath` (string) – The path to the Python file.
        *   `parsedContent` (string) – The parsed content of the Python file (output from the `parse` function).
    *   **Output**: `string` – A formatted prompt string designed for a language model. The prompt instructs the model to act as a principal engineer and Python expert, and to produce high-quality documentation.
    *   **Example**:
        ```typescript
        generatePrompt('/path/to/my_script.py', 'File: /path/to/my_script.py\n\nPython Code Content:\n\`\`\`python\ndef my_function():\n  print("Hello")\`\`\`')
        ```
        returns a string containing a detailed prompt including the file path and code content.

4.  **Output Post-Processing (`postProcess`)**:
    *   **Purpose**: Combines multiple output strings from the language model into a single, formatted document.
    *   **Input**:
        *   `filePath` (string) – The path to the Python file.
        *   `outputs` (string[]) – An array of strings representing the outputs from the language model.
    *   **Output**: `string` – A single string containing all the outputs joined by double newlines.
    *   **Example**:
        ```typescript
        postProcess('/path/to/my_script.py', ['Output 1', 'Output 2'])
        ```
        returns:
        `'Output 1\n\nOutput 2'`

### Usage

You can integrate this adapter into a larger system by instantiating the `PyAdapter` class and calling its methods in sequence. The typical workflow involves using `canHandle` to verify file type, `parse` to extract content, `generatePrompt` to create a prompt for a language model, and finally, `postProcess` to assemble the model’s responses into a final document.

### Future Enhancements

Planned improvements include:

*   Implementing Python AST parsing within the `parse` function to enable more granular analysis and documentation of Python code elements (classes, functions, etc.).
*   Adding support for more complex Python features and coding styles.