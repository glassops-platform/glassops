---
type: Documentation
domain: agent
origin: packages/tools/agent/src/prompts.yml
last_modified: 2026-01-31
generated: true
source: packages/tools/agent/src/prompts.yml
generated_at: 2026-01-31T11:10:54.036798
hash: f15f01e56e84a848ffe263878ad181c1ded7b012e586901035152f6acb5572e9
---

# Prompts Configuration Documentation

This document details the `prompts.yml` file, which defines instructions for a language model used to generate and refine technical documentation. We designed this configuration to support documentation creation for various file types.

## Structure

The `prompts` key contains a dictionary. Each key within this dictionary represents a different file type or documentation task. Currently, the supported types are:

*   `ts`: TypeScript/JavaScript documentation generation.
*   `md`: Markdown documentation refinement.
*   `json`: JSON schema/data structure documentation.
*   `yml`: YAML configuration documentation.

Each file type entry contains two keys: `system` and `user`.

## Key Details

### `ts` (TypeScript Documentation)

*   **`system`**: This key holds the system prompt. It instructs the language model to act as a principal architect and translate TypeScript/JavaScript code into high-level, concise documentation suitable for both technical and non-technical audiences. It emphasizes the need for a pristine, coherent, and professional document. The prompt explicitly instructs the model to *only* output the document content and to avoid conversational elements or code block wrapping. It also contains a list of prohibited words and guidelines for pronoun usage.
*   **`user`**: This key holds the user prompt. It instructs the model to generate documentation for the provided TypeScript/JavaScript code, which is passed in as the `{{content}}` variable.

### `md` (Markdown Refinement)

*   **`system`**: This key holds the system prompt. It instructs the language model to act as a principal architect and refine existing Markdown documentation, preserving its original intent and technical depth. It emphasizes creating a comprehensive, pristine, and coherent version of the original document. Similar to the `ts` prompt, it specifies output formatting and stylistic constraints.
*   **`user`**: This key holds the user prompt. It instructs the model to audit and improve the provided Markdown documentation, focusing on clarity, broken links (specifically relative paths), and consistency with a "Container-First" platform vision. The original document is passed in as the `{{content}}` variable.

### `json` (JSON Documentation)

*   **`system`**: This key holds the system prompt. It instructs the language model to act as a technical documentation expert and document a provided JSON schema or data structure. The prompt emphasizes explaining the data's architectural role, identifying required and optional fields, and outlining common use cases. It also specifies output formatting and stylistic constraints.
*   **`user`**: This key holds the user prompt. It instructs the model to generate documentation for the provided JSON content, which is passed in as the `{{content}}` variable.

### `yml` (YAML Documentation)

*   **`system`**: This key holds the system prompt. It instructs the language model to act as a DevOps engineer and technical writer, documenting provided YAML configuration strings. The prompt emphasizes explaining the configuration's purpose, structure, and the function of each key. It also specifies output formatting and stylistic constraints.
*   **`user`**: This key holds the user prompt. It instructs the model to generate documentation for the provided YAML content, which is passed in as the `{{content}}` variable.

## Variable Substitution

The `{{content}}` variable in the `user` prompts is a placeholder. You should replace this with the actual content of the file you want to document.