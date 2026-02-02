---
type: Documentation
domain: agent
origin: packages/tools/agent/src/adapters/yml-adapter.ts
last_modified: 2026-02-01
generated: true
source: packages/tools/agent/src/adapters/yml-adapter.ts
generated_at: 2026-02-01T19:49:53.762638
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
    *   `true` if the extension is ".yml" or ".yaml", indicating the adapter can handle the file.
    *   `false` otherwise.

#### 4.2. `parse(filePath: string, content: string): Promise<string[]>`

This asynchronous function takes the file path and content as input and formats the content into a string array.

*   **Parameters:**
    *   `filePath`: A string representing the path to the YAML file.
    *   `content`: A string containing the YAML file's content.
*   **Return Value:**
    *   A `Promise` that resolves to a string array containing a formatted representation of the file path and YAML content, enclosed in backticks for code formatting. Example:
        ```
        ["File: /path/to/file.yml\n\nYAML Content:\n```yaml\nfile_content\n```"]
        ```

#### 4.3. `generatePrompt(filePath: string, parsedContent: string): string`

This function creates a prompt for a language model, instructing it to analyze the provided YAML content.

*   **Parameters:**
    *   `filePath`: A string representing the path to the YAML file.
    *   `parsedContent`: A string containing the parsed YAML content.
*   **Return Value:**
    *   A string containing the prompt. The prompt instructs the language model to act as a principal architect and validate and explain the provided YAML configuration. The `parsedContent` is embedded directly within the prompt.

#### 4.4. `postProcess(filePath: string, outputs: string[]): string`

This function combines the outputs from the language model into a single string, separated by double newlines.

*   **Parameters:**
    *   `filePath`: A string representing the path to the YAML file. (Currently unused)
    *   `outputs`: An array of strings, where each string represents an output from the language model.
*   **Return Value:**
    *   A string containing all outputs joined by double newlines (`\n\n`).

**5. Usage**

You would integrate this adapter into a system that processes files based on their extension. When a ".yml" or ".yaml" file is encountered, this adapter will be used to parse the content, generate a prompt, and process the resulting output.

**6. Dependencies**

This adapter has a dependency on the `AgentAdapter` interface, defined in `./interface.js`.