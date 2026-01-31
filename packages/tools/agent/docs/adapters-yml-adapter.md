---
type: Documentation
domain: agent
origin: packages/tools/agent/src/adapters/yml-adapter.ts
last_modified: 2026-01-31
generated: true
source: packages/tools/agent/src/adapters/yml-adapter.ts
generated_at: 2026-01-31T10:19:39.593986
hash: ca031af7b7ed139863454c2b9d49b1bdcbfe1037cfe30ed64c73e2f9d88dd15a
---

## YML Adapter Documentation

**1. Introduction**

This document details the functionality of the YML Adapter, a component designed to process YAML files within a larger agent system. The adapter handles file identification, content parsing, prompt generation for analysis, and output consolidation.

**2. Purpose**

The YML Adapter enables the agent to understand and work with YAML configuration files. It prepares the file content for review and analysis by a language model, and then formats the results for presentation.

**3. Core Functionality**

The adapter provides four primary functions:

*   **File Type Handling:** Determines if the adapter can process a given file based on its extension.
*   **Content Parsing:** Reads the file content and prepares it for analysis.
*   **Prompt Generation:** Constructs a prompt to instruct a language model to analyze the file content.
*   **Output Processing:** Combines multiple outputs into a single, formatted string.

**4. Adapter Details**

#### 4.1. `canHandle(extension: string): boolean`

This function checks if the adapter supports a file with the given extension. 

*   **Parameter:**
    *   `extension`: A string representing the file extension (e.g., ".yml", ".yaml").
*   **Return Value:**
    *   `true` if the extension is ".yml" or ".yaml", `false` otherwise.

#### 4.2. `parse(filePath: string, content: string): Promise<string[]>`

This asynchronous function takes a file path and its content as input and formats the content into a string array.

*   **Parameters:**
    *   `filePath`: A string representing the path to the YAML file.
    *   `content`: A string containing the YAML file's content.
*   **Return Value:**
    *   A `Promise` that resolves to a string array containing the formatted file content, enclosed in a code block. Example:
        ```
        [
          "File: /path/to/file.yml\n\nYAML Content:\n\`\`\`yaml\nfile_content_here\n\`\`\`"
        ]
        ```

#### 4.3. `generatePrompt(filePath: string, parsedContent: string): string`

This function creates a prompt for a language model, instructing it to analyze the provided YAML content.

*   **Parameters:**
    *   `filePath`: A string representing the path to the YAML file.
    *   `parsedContent`: A string containing the parsed YAML content.
*   **Return Value:**
    *   A string representing the prompt. The prompt instructs the language model to act as a principal architect and validate/explain the provided YAML configuration.

#### 4.4. `postProcess(filePath: string, outputs: string[]): string`

This function combines the outputs from the language model into a single string, separated by double newlines.

*   **Parameters:**
    *   `filePath`: A string representing the path to the YAML file.
    *   `outputs`: A string array containing the outputs from the language model.
*   **Return Value:**
    *   A string containing the combined outputs, separated by double newlines.

**5. Usage**

You can integrate this adapter into a larger system by instantiating the `YMLAdapter` class and calling its methods in sequence. First, check if the adapter can handle the file using `canHandle()`. If it can, use `parse()` to prepare the content, then `generatePrompt()` to create a prompt for analysis. Finally, after receiving outputs from the language model, use `postProcess()` to format the results.

**6. Error Handling**

The `parse` function is asynchronous and may encounter errors during file reading or processing. Implement appropriate error handling mechanisms in your application to manage potential exceptions.