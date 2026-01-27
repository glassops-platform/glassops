---
type: Documentation
domain: agent
origin: packages/tools/agent/src/adapters/ts-adapter.ts
last_modified: 2026-01-26
generated: true
source: packages/tools/agent/src/adapters/ts-adapter.ts
generated_at: 2026-01-26T14:24:25.314Z
hash: 9d12943eed8c08b252ed5e6341f96bb028e590ba43940d908bcafa3bb8e37ffd
---

## TypeScript Adapter Documentation

**Overview**

This adapter facilitates the processing of TypeScript, JavaScript, and MJS files within the agent system. It handles file type identification, content parsing, prompt generation, and output consolidation.

**Functionality**

The TSAdapter provides a standardized interface for interacting with TypeScript/JavaScript source code. It prepares the code for analysis and documentation generation by a language model.

**Key Components**

*   **File Handling:** The adapter determines if it can process a given file based on its extension (.ts, .js, .mjs).
*   **Parsing:** Currently, the parsing stage passes the file content along with its path as a single string. Future iterations may incorporate Abstract Syntax Tree (AST) parsing for more detailed docstring extraction.
*   **Prompt Generation:** A prompt is constructed, instructing the language model to act as a principal architect and generate high-quality documentation from the provided code content. The prompt emphasizes conciseness, clarity, and professional presentation, while specifically requesting only the documentation content itself as output.
*   **Post-Processing:** The adapter combines multiple outputs from the language model into a single, formatted string, separated by double newlines.

**Configuration**

No specific configuration is required for the TSAdapter. It operates based on file extensions and the defined prompt.

**Usage**

The adapter is automatically selected when processing files with the extensions .ts, .js, or .mjs. You can integrate this adapter into the agent workflow to automatically document your TypeScript/JavaScript codebase.

**Future Enhancements**

Planned improvements include:

*   Implementation of AST parsing to enable more precise extraction of documentation from docstrings and comments.
*   Advanced content filtering and formatting.
*   Support for additional JavaScript/TypeScript related file types.