---
type: Documentation
domain: agent
origin: packages/tools/agent/src/adapters/json-adapter.ts
last_modified: 2026-01-26
generated: true
source: packages/tools/agent/src/adapters/json-adapter.ts
generated_at: 2026-01-26T14:22:58.348Z
hash: 6e7fa918d0f9d854c033d03ac35ea4226c0fdf1592d90c6fe7900d0258957ff0
---

## JSON Adapter Documentation

**Overview**

This document details the functionality of the JSON Adapter, a component designed to process files with the `.json` extension within an agent system. The adapter’s purpose is to prepare JSON file content for analysis and subsequent use by the agent.

**Functionality**

The JSON Adapter provides four core functions: determining file compatibility, parsing file content, generating prompts for analysis, and post-processing outputs.

**1. File Compatibility (canHandle)**

The `canHandle` function verifies if the adapter is suitable for a given file based on its extension. 

*   **Input:** `extension` (string) – The file extension.
*   **Output:** `boolean` – Returns `true` if the extension is `.json`, otherwise `false`.

**2. Content Parsing (parse)**

The `parse` function takes the file path and content of a JSON file and formats it for inclusion in a prompt.

*   **Input:**
    *   `filePath` (string) – The path to the JSON file.
    *   `content` (string) – The content of the JSON file.
*   **Output:** `string[]` – An array containing a single string. This string includes the file path and the JSON content, formatted within a code block for clarity.

**3. Prompt Generation (generatePrompt)**

The `generatePrompt` function constructs a prompt to be sent to a language model for analysis of the parsed JSON content. 

*   **Input:**
    *   `filePath` (string) – The path to the JSON file.
    *   `parsedContent` (string) – The formatted JSON content (output from the `parse` function).
*   **Output:** `string` – A prompt instructing the language model to act as a principal architect and create a high-level document from the provided JSON content. The prompt emphasizes clarity, conciseness, and professional presentation.

**4. Output Post-Processing (postProcess)**

The `postProcess` function combines multiple outputs from the language model into a single, formatted string.

*   **Input:**
    *   `filePath` (string) – The path to the JSON file.
    *   `outputs` (string[]) – An array of strings representing the outputs from the language model.
*   **Output:** `string` – A single string containing all outputs, separated by double newlines for readability.

**Integration**

I am designed to be integrated into a larger agent system. You can configure the system to use this adapter by ensuring that files with the `.json` extension are routed to this component for processing. The output of this adapter serves as input for subsequent stages in the agent’s workflow, such as analysis and documentation generation.