---
type: Documentation
domain: agent
origin: packages/tools/agent/src/adapters/lwc-adapter.ts
last_modified: 2026-01-31
generated: true
source: packages/tools/agent/src/adapters/lwc-adapter.ts
generated_at: 2026-01-31T10:17:33.409617
hash: b49e1c6a1867343593bd52f58aca9365ea3ee9aa67113aa6132a5a858aa83fb5
---

## Lightning Web Component (LWC) Adapter Documentation

**Introduction**

This document details the functionality of the LWC Adapter, a component designed to process Lightning Web Component files and prepare them for documentation generation. It identifies LWC JavaScript files, extracts associated HTML and CSS, and formats the content for input into a large language model.

**Functionality**

The LWC Adapter operates as part of a larger system for automatically generating documentation from source code. It specifically targets Salesforce’s Lightning Web Component framework. The adapter’s core functions are: file identification, content parsing, prompt generation, and output post-processing.

**Key Features**

*   **File Identification:** The adapter determines if a file should be processed based on its extension and directory structure. It specifically looks for `.js` files containing “lwc” in the filename or path.
*   **Content Parsing:**  Given a file path and its content, the adapter attempts to locate corresponding HTML and CSS files. It then bundles the JavaScript, HTML (if found), and CSS (if found) into a single string.
*   **Prompt Generation:** The adapter constructs a prompt for a language model, instructing it to act as a Salesforce Developer Expert and document the provided LWC. The prompt includes the bundled component content.
*   **Post-Processing:** The adapter combines multiple outputs from the language model into a single, formatted string.

**Adapter Interface & Methods**

The LWC Adapter implements the `AgentAdapter` interface, requiring the following methods:

*   `canHandle(fileName: string): boolean`
    *   **Purpose:** Determines if the adapter can process the given file.
    *   **Input:** `fileName` (string) - The name of the file to check.
    *   **Output:** `boolean` - `true` if the adapter can handle the file, `false` otherwise.
*   `parse(filePath: string, content: string): Promise<string[]>`
    *   **Purpose:** Parses the file content and extracts related assets.
    *   **Input:**
        *   `filePath` (string) - The path to the file.
        *   `content` (string) - The content of the file.
    *   **Output:** `Promise<string[]>` - A promise that resolves to an array of strings, each representing a bundled component section.
*   `generatePrompt(filePath: string, parsedContent: string): string`
    *   **Purpose:** Creates a prompt for the language model.
    *   **Input:**
        *   `filePath` (string) - The path to the file.
        *   `parsedContent` (string) - The parsed content of the file.
    *   **Output:** `string` - The prompt string.
*   `postProcess(filePath: string, outputs: string[]): string`
    *   **Purpose:** Processes the outputs from the language model.
    *   **Input:**
        *   `filePath` (string) - The path to the file.
        *   `outputs` (string[]) - An array of strings representing the outputs from the language model.
    *   **Output:** `string` - The combined and formatted output string.

**Workflow**

1.  The system identifies a JavaScript file with a name or path containing “lwc”.
2.  The LWC Adapter’s `canHandle` method confirms the file type.
3.  The `parse` method reads the file content and attempts to locate corresponding HTML and CSS files in the same directory.
4.  The adapter bundles the JavaScript, HTML, and CSS content into a single string.
5.  The `generatePrompt` method creates a prompt instructing the language model to document the LWC.
6.  The language model processes the prompt and generates documentation.
7.  The `postProcess` method combines the documentation segments into a final output.

**Example**

Consider a Lightning Web Component with the following files:

*   `myComponent.js`
*   `myComponent.html`
*   `myComponent.css`

The adapter will bundle the contents of all three files into a single string, which will then be used as input for the language model.

**Error Handling**

The adapter includes a safety check to ensure it is processing a valid LWC. The `exists` method handles potential file access errors gracefully, returning `false` if a file is not found. This prevents the process from crashing if a corresponding HTML or CSS file is missing.