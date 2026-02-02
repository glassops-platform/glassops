---
type: Documentation
domain: agent
origin: packages/tools/agent/src/adapters/interface.ts
last_modified: 2026-02-01
generated: true
source: packages/tools/agent/src/adapters/interface.ts
generated_at: 2026-02-01T19:47:48.522342
hash: f2768e9d45e206d7f00f3b4c1b4f1de47b4278b6f233c741457ec05cefe83842
---

## Agent Adapter Interface Documentation

This document details the `AgentAdapter` interface, defining the contract for components responsible for processing files within the system. Adapters enable the system to work with diverse file types and content structures.

### Overview

The `AgentAdapter` interface specifies a set of methods that any adapter must implement to integrate with the core agent functionality. These methods handle file type identification, content parsing, prompt generation for language models, and post-processing of model outputs.

### Interface Definition

```typescript
export interface AgentAdapter {
  canHandle(fileName: string): boolean;
  parse(filePath: string, content: string): Promise<string[]>;
  generatePrompt(filePath: string, parsedContent: string): string;
  postProcess(filePath: string, outputs: string[]): string;
}
```

### Method Descriptions

#### `canHandle(fileName: string): boolean`

This method determines if the adapter is capable of processing a file based on its name (including extension). 

*   **Parameters:**
    *   `fileName`: A string representing the name of the file.
*   **Return Value:** A boolean value. `true` indicates the adapter can handle the file; `false` indicates it cannot.

#### `parse(filePath: string, content: string): Promise<string[]>`

This method parses the content of a file and transforms it into a structured format suitable for use with a language model. The method returns a Promise that resolves to an array of strings, allowing for the handling of large files through chunking.

*   **Parameters:**
    *   `filePath`: A string representing the path to the file.
    *   `content`: A string containing the file's content.
*   **Return Value:** A Promise resolving to a string array. Each string in the array represents a chunk of the parsed content.

#### `generatePrompt(filePath: string, parsedContent: string): string`

This method constructs a prompt to be sent to a language model, based on the file path and the parsed content. The prompt guides the language model to perform the desired operation on the file content.

*   **Parameters:**
    *   `filePath`: A string representing the path to the file.
    *   `parsedContent`: A string representing the parsed content of the file.
*   **Return Value:** A string containing the generated prompt.

#### `postProcess(filePath: string, outputs: string[]): string`

This method processes the output received from a language model. This includes tasks such as formatting the output, resolving relative links, and joining chunks if the input file was processed in parts.

*   **Parameters:**
    *   `filePath`: A string representing the path to the file.
    *   `outputs`: A string array containing the outputs from the language model.
*   **Return Value:** A string containing the final, post-processed output.

### Implementation Guidance

When implementing the `AgentAdapter` interface, you should ensure that:

*   The `canHandle` method accurately reflects the file types your adapter supports.
*   The `parse` method handles potential errors during file parsing and returns meaningful chunks.
*   The `generatePrompt` method creates prompts that are clear and effective for the intended language model task.
*   The `postProcess` method correctly handles any necessary formatting or adjustments to the language model's output.