---
type: Documentation
domain: agent
origin: packages/tools/agent/src/prompts.yml
last_modified: 2026-02-01
generated: true
source: packages/tools/agent/src/prompts.yml
generated_at: 2026-02-01T19:51:05.417901
hash: f15f01e56e84a848ffe263878ad181c1ded7b012e586901035152f6acb5572e9
---

# Prompts Configuration Documentation

This document details the `prompts.yml` file, which defines prompts used for generating and refining technical documentation. We designed this configuration to support documentation creation for various file types.

## Structure

The `prompts` key contains a dictionary. Each key within this dictionary represents a different file type or documentation task. Currently, the supported types are:

*   `ts`: TypeScript/JavaScript documentation generation.
*   `md`: Markdown documentation refinement.
*   `json`: JSON schema/data structure documentation.
*   `yml`: YAML configuration documentation.

Each file type entry contains two keys: `system` and `user`.

## Key Details

### `ts` (TypeScript Documentation)

*   **`system`**: This key holds the system prompt for the language model when generating documentation from TypeScript/JavaScript code. The prompt instructs the model to act as a principal architect and produce a high-level, concise, and inclusive document suitable for both technical and non-technical audiences. It includes strict rules regarding output format and content restrictions. The prompt is repeated for emphasis.
*   **`user`**: This key contains the user prompt. It instructs the model to generate documentation for the provided TypeScript/JavaScript code, which is passed in as the `{{content}}` variable.

### `md` (Markdown Refinement)

*   **`system`**: This key holds the system prompt for refining existing Markdown documentation. The prompt instructs the model to act as a principal architect and improve the documentation while preserving its original intent and technical depth. It emphasizes creating a comprehensive and coherent version. It also includes strict rules regarding output format and content restrictions.
*   **`user`**: This key contains the user prompt. It instructs the model to audit and improve the provided Markdown documentation, focusing on clarity, broken links, and consistency with the "Container-First" platform vision. The documentation to be improved is passed in as the `{{content}}` variable.

### `json` (JSON Documentation)

*   **`system`**: This key holds the system prompt for documenting JSON schemas or data structures. The prompt instructs the model to act as a technical documentation expert and explain the data's architectural role, required/optional fields, and common use cases. It specifies the output format requirements.
*   **`user`**: This key contains the user prompt. It instructs the model to generate documentation for the provided JSON content, which is passed in as the `{{content}}` variable.

### `yml` (YAML Documentation)

*   **`system`**: This key holds the system prompt for documenting YAML configurations. The prompt instructs the model to act as a DevOps engineer and technical writer, explaining the configuration's purpose, structure, and key controls. It specifies the output format requirements.
*   **`user`**: This key contains the user prompt. It instructs the model to generate documentation for the provided YAML content, which is passed in as the `{{content}}` variable.

## Variable Substitution

The `{{content}}` variable in the `user` prompts is a placeholder. You should replace this with the actual content of the file you want to document.