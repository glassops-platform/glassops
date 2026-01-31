---
type: Documentation
domain: agent
origin: packages/tools/agent/src/adapters/json-adapter.ts
last_modified: 2026-01-31
generated: true
source: packages/tools/agent/src/adapters/json-adapter.ts
generated_at: 2026-01-31T09:20:25.593790
hash: fbdff4a5a66631702c17a404f4b360020230038017433e82b6b0aa0f16ab50d2
---

## JSON Adapter Documentation

**1. Introduction**

This document details the functionality of the JSON Adapter, a component designed to process files with the `.json` extension within the agent system. It handles parsing, prompt generation, and post-processing of JSON file content.

**2. Purpose**

The JSON Adapter enables the agent to understand and work with JSON data. It prepares the JSON content for analysis by constructing a specific prompt for a language model.

**3. Functionality**

The adapter provides the following core functions:

*   **File Type Handling:** Determines if the adapter can process a given file based on its extension. It specifically handles files ending in `.json`.
*   **Parsing:** Reads the content of a JSON file and formats it into a structured string. This string includes the file path and the JSON content enclosed in a code block for clarity.
*   **Prompt Generation:** Creates a prompt for a language model, instructing it to analyze the parsed JSON content as if it were a principal architect reviewing documentation. The prompt includes the formatted JSON content.
*   **Post-Processing:** Combines multiple outputs from the language model into a single, formatted string, separated by double newlines.

**4. Technical Details**

The JSON Adapter implements the `AgentAdapter` interface. This ensures compatibility and consistent behavior within the agent system. 

*   **Input:**
    *   `filePath`: String representing the path to the JSON file.
    *   `content`: String containing the JSON file’s content.
*   **Output:**
    *   A string containing the processed output, ready for presentation or further action.

**5. Usage**

When the agent encounters a `.json` file, it will automatically route the file to this adapter for processing. You do not need to directly interact with the adapter; it operates as part of the agent’s automated workflow.

**6. Output Format**

The final output consists of the language model’s responses, concatenated with double newlines between each response. The initial parsed content is formatted to clearly present the JSON data to the language model.