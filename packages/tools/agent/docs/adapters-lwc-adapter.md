---
type: Documentation
domain: agent
origin: packages/tools/agent/src/adapters/lwc-adapter.ts
last_modified: 2026-01-31
generated: true
source: packages/tools/agent/src/adapters/lwc-adapter.ts
generated_at: 2026-01-31T09:20:53.467865
hash: b49e1c6a1867343593bd52f58aca9365ea3ee9aa67113aa6132a5a858aa83fb5
---

## Lightning Web Component (LWC) Adapter Documentation

**1. Introduction**

This document details the functionality of the LWC Adapter, a component designed to process Lightning Web Component files and prepare them for documentation generation. It identifies LWC JavaScript files, retrieves associated HTML and CSS, and formats the content for input into a documentation model.

**2. Functionality Overview**

The LWC Adapter operates as part of a larger agent system. Its primary role is to recognize, parse, and prepare LWC source code for documentation. It specifically targets JavaScript files associated with LWC structures.

**3. Core Components**

*   **`canHandle(fileName: string): boolean`**: This function determines if the adapter should process a given file. It checks for the `.js` extension and the presence of “lwc” in the filename or path, ensuring it focuses on LWC JavaScript controller files. This prevents processing of duplicate documentation for HTML or CSS.
*   **`parse(filePath: string, content: string): Promise<string[]>`**: This asynchronous function is the core of the adapter. It takes the file path and content of a JavaScript file as input. It then:
    *   Determines if the file is part of an LWC by checking for corresponding HTML or CSS files, or if the file path contains “lwc”.
    *   Reads the content of associated HTML and CSS files (if they exist).
    *   Constructs a bundled string containing the JavaScript, HTML, and CSS content, formatted for documentation.
    *   Returns an array containing this bundled string.
*   **`generatePrompt(filePath: string, parsedContent: string): string`**: This function creates a prompt for a documentation model. It instructs the model to act as a Salesforce Developer Expert and generate documentation for the provided LWC content, focusing on purpose, public properties, events, and UI behavior.
*   **`postProcess(filePath: string, outputs: string[]): string`**: This function combines multiple outputs from the documentation model into a single string, separated by double newlines.
*   **`exists(p: string): Promise<boolean>`**: A private helper function that asynchronously checks if a file exists at the given path.

**4. Workflow**

1.  The agent system identifies a JavaScript file.
2.  The `canHandle` function determines if the LWC Adapter should process the file.
3.  If `canHandle` returns true, the `parse` function is called to extract and bundle the LWC’s JavaScript, HTML, and CSS content.
4.  The `generatePrompt` function creates a prompt for the documentation model, including the bundled content.
5.  The documentation model processes the prompt and generates documentation.
6.  The `postProcess` function combines the documentation outputs into a final, formatted string.

**5. Dependencies**

*   `fs/promises`: For asynchronous file system operations.
*   `path`: For manipulating file paths.

**6. Usage Notes**

You should ensure that the agent system correctly identifies and passes LWC JavaScript files to this adapter. The adapter expects associated HTML and CSS files to have the same base name as the JavaScript file. 

**7. Error Handling**

The `exists` function includes error handling to gracefully manage cases where files are not found. The `parse` function includes a safety check to ensure it only processes valid LWC files.