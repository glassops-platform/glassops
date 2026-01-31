---
type: Documentation
domain: agent
origin: packages/tools/agent/src/adapters/interface.ts
last_modified: 2026-01-31
generated: true
source: packages/tools/agent/src/adapters/interface.ts
generated_at: 2026-01-31T09:20:09.673704
hash: f2768e9d45e206d7f00f3b4c1b4f1de47b4278b6f233c741457ec05cefe83842
---

## Agent Adapter Interface Specification

**Document Version:** 1.0
**Date:** October 26, 2023

**1. Introduction**

This document details the `AgentAdapter` interface. This interface defines the contract for components responsible for processing files within the agent system. Adapters enable the system to work with diverse file types and formats.

**2. Purpose**

The `AgentAdapter` interface standardizes how file content is ingested, prepared for language model interaction, and finalized into a usable output. It promotes modularity and extensibility, allowing for easy integration of new file type support.

**3. Interface Definition**

The `AgentAdapter` interface requires implementing the following methods:

**3.1. `canHandle(fileName: string): boolean`**

*   **Description:** This method determines if a specific adapter is capable of processing a file based on its name (including extension).
*   **Parameters:**
    *   `fileName`: A string representing the name of the file.
*   **Return Value:** A boolean value. `true` indicates the adapter can handle the file; `false` indicates it cannot.
*   **Usage:** You should implement this method to check the file extension or other relevant characteristics to determine compatibility.

**3.2. `parse(filePath: string, content: string): Promise<string[]>`**

*   **Description:** This method parses the content of a file and transforms it into a structured format suitable for processing by a language model. The method returns a promise that resolves to an array of strings, allowing for the handling of large files through chunking.
*   **Parameters:**
    *   `filePath`: A string representing the path to the file.
    *   `content`: A string containing the file’s content.
*   **Return Value:** A Promise resolving to a string array. Each string in the array represents a chunk of the parsed content.
*   **Usage:** I recommend breaking down large files into smaller, manageable chunks within this method to avoid exceeding language model input limits.

**3.3. `generatePrompt(filePath: string, parsedContent: string): string`**

*   **Description:** This method constructs a prompt to be sent to a language model, using the parsed content of the file.
*   **Parameters:**
    *   `filePath`: A string representing the path to the file.
    *   `parsedContent`: A string representing the parsed content of the file.
*   **Return Value:** A string representing the generated prompt.
*   **Usage:** We expect this method to format the `parsedContent` into a prompt that instructs the language model on the desired task.

**3.4. `postProcess(filePath: string, outputs: string[]): string`**

*   **Description:** This method performs final processing on the output received from the language model. This includes tasks such as formatting, resolving relative links, and joining chunks if the file was processed in parts.
*   **Parameters:**
    *   `filePath`: A string representing the path to the file.
    *   `outputs`: A string array containing the outputs from the language model.
*   **Return Value:** A string representing the final, processed output.
*   **Usage:** If the `parse` method returned multiple chunks, you must combine them into a single coherent output within this method.



**4. Implementation Considerations**

*   Adapters should handle errors gracefully and provide informative error messages.
*   The `parse` method should be designed to handle potentially large files efficiently.
*   The `generatePrompt` method should create prompts that are clear, concise, and effective in guiding the language model.
*   The `postProcess` method should ensure the final output is well-formatted and consistent.