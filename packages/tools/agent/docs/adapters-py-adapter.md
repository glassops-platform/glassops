---
type: Documentation
domain: agent
origin: packages/tools/agent/src/adapters/py-adapter.ts
last_modified: 2026-01-31
generated: true
source: packages/tools/agent/src/adapters/py-adapter.ts
generated_at: 2026-01-31T10:18:26.123986
hash: 8d2e85d94103c29d72984c3f34046684fc1ec3b7c96954d4cc1e5cd60a1b7e5f
---

## Python File Adapter Documentation

This document details the functionality of the Python File Adapter, a component designed to process Python files within a larger system. It serves as an interface between the core system and Python source code, preparing the code for analysis and documentation generation.

### Overview

The Python File Adapter is responsible for identifying, parsing, and preparing Python files for further processing. It determines if a given file is a Python file, extracts its content, and formats it into a prompt suitable for a language model. Finally, it consolidates the outputs from the language model into a single, coherent document.

### Core Functionality

The adapter operates through four primary functions: `canHandle`, `parse`, `generatePrompt`, and `postProcess`.

#### 1. `canHandle(fileName: string): boolean`

This function determines whether the adapter can process a given file based on its file extension. 

*   **Input:** `fileName` – A string representing the name of the file.
*   **Output:** A boolean value. Returns `true` if the file has a `.py` extension, indicating it is a Python file; otherwise, returns `false`.

   ```typescript
   const isPythonFile = adapter.canHandle('my_script.py'); // Returns true
   const isJavaScriptFile = adapter.canHandle('script.js'); // Returns false
   ```

#### 2. `parse(filePath: string, content: string): Promise<string[]>`

This function takes the file path and content of a Python file and prepares it for prompt generation. Currently, it performs a simple pass-through, preserving the file path and content. Future versions will incorporate Python Abstract Syntax Tree (AST) parsing to identify and separate classes and functions within the code.

*   **Input:**
    *   `filePath` – A string representing the path to the Python file.
    *   `content` – A string containing the content of the Python file.
*   **Output:** A Promise resolving to an array of strings. Currently, this array contains a single string that includes the file path and the complete Python code content, formatted for inclusion in a prompt.

   ```typescript
   const parsedContent = await adapter.parse('/path/to/my_script.py', 'print("Hello, world!")');
   // Returns: ['File: /path/to/my_script.py\n\nPython Code Content:\n```python\nprint("Hello, world!")\n```']
   ```

#### 3. `generatePrompt(filePath: string, parsedContent: string): string`

This function constructs a prompt for a language model, instructing it to generate documentation for the provided Python code. The prompt frames the language model as a principal engineer and Python expert, emphasizing the need for clear, concise, and professional documentation suitable for both technical and non-technical audiences.

*   **Input:**
    *   `filePath` – A string representing the path to the Python file.
    *   `parsedContent` – A string containing the parsed content of the Python file (output from the `parse` function).
*   **Output:** A string representing the prompt to be sent to the language model.

   ```typescript
   const prompt = adapter.generatePrompt('/path/to/my_script.py', 'File: ...\n\nPython Code Content:\n```python\n...\n```');
   // Returns a string containing the prompt instructions and the parsed content.
   ```

#### 4. `postProcess(filePath: string, outputs: string[]): string`

This function consolidates the outputs received from the language model into a single string, separated by double newlines.

*   **Input:**
    *   `filePath` – A string representing the path to the Python file.
    *   `outputs` – An array of strings, where each string represents a separate output from the language model.
*   **Output:** A string containing the combined outputs, separated by double newlines.

   ```typescript
   const finalDocumentation = adapter.postProcess('/path/to/my_script.py', ['Section 1', 'Section 2']);
   // Returns: 'Section 1\n\nSection 2'
   ```

### Dependencies

*   The `path` module is used for extracting file extensions.
*   The `AgentAdapter` interface defines the contract that this adapter implements.

### Future Enhancements

*   Implement Python AST parsing within the `parse` function to enable more granular analysis and documentation generation, specifically separating classes and functions.
*   Add error handling and logging to improve robustness and debugging capabilities.