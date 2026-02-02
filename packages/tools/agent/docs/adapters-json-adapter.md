---
type: Documentation
domain: agent
origin: packages/tools/agent/src/adapters/json-adapter.ts
last_modified: 2026-02-01
generated: true
source: packages/tools/agent/src/adapters/json-adapter.ts
generated_at: 2026-02-01T19:48:05.408519
hash: fbdff4a5a66631702c17a404f4b360020230038017433e82b6b0aa0f16ab50d2
---

## JSON Adapter Documentation

**1. Introduction**

This document details the functionality of the JSON Adapter, a component designed to process files with the `.json` extension within the agent system. It handles parsing, prompt generation, and post-processing of JSON content for analysis.

**2. Overview**

The JSON Adapter conforms to the `AgentAdapter` interface, providing a standardized way to interact with JSON files. It is intended for use cases where the agent needs to understand and reason about data structured in JSON format.

**3. Capabilities**

*   **File Type Handling:** The adapter specifically handles files ending with the `.json` extension.
*   **Parsing:**  It reads the content of a JSON file and formats it for inclusion in a prompt.
*   **Prompt Generation:** It constructs a prompt tailored for a principal architect role, instructing the language model to analyze the JSON content.
*   **Post-Processing:** It combines multiple outputs from the language model into a single, formatted string.

**4. Interface Implementation**

The `JSONAdapter` class implements the following methods from the `AgentAdapter` interface:

*   **`canHandle(extension: string): boolean`**:
    Determines if the adapter can process a file based on its extension. Returns `true` if the extension is `.json`, and `false` otherwise.

    Example:
    ```typescript
    const canHandleJson = adapter.canHandle('.json'); // Returns true
    const canHandleTxt = adapter.canHandle('.txt');   // Returns false
    ```

*   **`async parse(filePath: string, content: string): Promise<string[]>`**:
    Parses the content of a JSON file and returns an array of strings. Currently, it formats the file path and content within a code block for clarity.

    Example:
    ```typescript
    const filePath = 'path/to/data.json';
    const fileContent = '{"key": "value"}';
    const parsedOutput = await adapter.parse(filePath, fileContent);
    // Returns: ['File: path/to/data.json\n\nJSON Content:\n```json\n{"key": "value"}\n```']
    ```

*   **`generatePrompt(filePath: string, parsedContent: string): string`**:
    Generates a prompt for the language model, including the parsed JSON content. The prompt instructs the model to act as a principal architect and analyze the provided JSON data.

    Example:
    ```typescript
    const prompt = adapter.generatePrompt('path/to/data.json', '{"key": "value"}');
    // Returns a prompt string designed for a language model.
    ```

*   **`postProcess(filePath: string, outputs: string[]): string`**:
    Combines the outputs from the language model into a single string, separated by double newlines.

    Example:
    ```typescript
    const outputs = ['Output 1', 'Output 2'];
    const processedOutput = adapter.postProcess('path/to/data.json', outputs);
    // Returns: 'Output 1\n\nOutput 2'
    ```

**5. Usage**

You can integrate this adapter into the agent system to enable processing of JSON files. The agent will call the `canHandle` method to determine if the adapter is suitable for a given file, and if so, will proceed to use the `parse`, `generatePrompt`, and `postProcess` methods to analyze the file's content.

**6. Future Considerations**

Potential enhancements include:

*   Adding JSON schema validation.
*   Implementing more sophisticated parsing logic to extract specific data elements.
*   Allowing customization of the prompt template.