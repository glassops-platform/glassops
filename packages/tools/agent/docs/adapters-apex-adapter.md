---
type: Documentation
domain: agent
origin: packages/tools/agent/src/adapters/apex-adapter.ts
last_modified: 2026-02-01
generated: true
source: packages/tools/agent/src/adapters/apex-adapter.ts
generated_at: 2026-02-01T19:47:11.450868
hash: bf5aa959825ccecf68ff871bcacc77fb8428a3e5e8906ab4e51a0b6bce76c81d
---

## Apex Adapter Documentation

**1. Introduction**

This document details the Apex Adapter, a component designed to process Salesforce Apex code files (.cls and .trigger) within a larger agent system. It handles file identification, content parsing, prompt generation for analysis, and output formatting.

**2. Functionality Overview**

The Apex Adapter enables the system to understand and work with Apex code. It prepares the code for review by creating a structured prompt that guides a language model to generate relevant documentation.

**3. Core Components**

*   **File Handling:** The adapter determines if a given file should be processed based on its extension.
*   **Parsing:** It extracts the file content and formats it for inclusion in a prompt.
*   **Prompt Engineering:** It constructs a prompt tailored for Apex code analysis, instructing the language model to focus on key areas like business logic, triggers, and security.
*   **Output Formatting:** It combines the outputs from the language model into a readable format.

**4. Adapter Interface & Methods**

This adapter implements the `AgentAdapter` interface, providing the following methods:

*   **`canHandle(fileName: string): boolean`**
    *   Purpose: Determines if the adapter can process a file based on its name.
    *   Input: `fileName` (string) – The name of the file.
    *   Return Value: `boolean` – `true` if the adapter can handle the file (extension is `.cls` or `.trigger`), `false` otherwise.
    *   Example: `canHandle('MyTrigger.trigger')` returns `true`.

*   **`async parse(filePath: string, content: string): Promise<string[]>`**
    *   Purpose: Parses the file content and prepares it for prompt generation.
    *   Input:
        *   `filePath` (string) – The path to the file.
        *   `content` (string) – The content of the file.
    *   Return Value: `Promise<string[]>` – A promise that resolves to an array of strings containing the parsed content. The content is formatted as a code block.
    *   Example:
        ```typescript
        parse('/path/to/MyClass.cls', 'public class MyClass { ... }')
        // Returns: ['File: /path/to/MyClass.cls\n\nApex Code Content:\n\`\`\`apex\npublic class MyClass { ... }\n\`\`\`']
        ```

*   **`generatePrompt(filePath: string, parsedContent: string): string`**
    *   Purpose: Generates a prompt for the language model, incorporating the parsed content.
    *   Input:
        *   `filePath` (string) – The path to the file.
        *   `parsedContent` (string) – The parsed content of the file.
    *   Return Value: `string` – The generated prompt.
    *   Example: The prompt instructs the language model to act as a Salesforce Architect and focus on business logic, triggers, and security implications.

*   **`postProcess(filePath: string, outputs: string[]): string`**
    *   Purpose: Combines the outputs from the language model into a single string.
    *   Input:
        *   `filePath` (string) – The path to the file.
        *   `outputs` (string[]) – An array of strings representing the outputs from the language model.
    *   Return Value: `string` – A single string containing the combined outputs, separated by double newlines.
    *   Example: `postProcess('/path/to/MyTrigger.trigger', ['Output 1', 'Output 2'])` returns `"Output 1\n\nOutput 2"`.

**5. Usage**

You integrate this adapter into the agent system to enable processing of Apex code files. The system will call `canHandle` to determine if the adapter should be used for a given file. If it can handle the file, the system will then call `parse`, `generatePrompt`, and `postProcess` to prepare the code for analysis and format the results.

**6. Dependencies**

*   `path` (Node.js built-in module) – Used for extracting file extensions.
*   `AgentAdapter` interface – Defines the contract for all adapters.