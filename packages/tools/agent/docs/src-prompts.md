---
type: Documentation
domain: agent
origin: packages/tools/agent/src/prompts.yml
last_modified: 2026-01-26
generated: true
source: packages/tools/agent/src/prompts.yml
generated_at: 2026-01-26T05:19:58.589Z
hash: 0abfaf7a606997a2061e4b43d5b4fd45d537610dd09ed5414a61edb910d03a04
---

# `prompts.yml` Documentation

This file defines a set of prompts used by an agent to generate or refine technical documentation. Each prompt is tailored to a specific input file type and desired output. The structure consists of a `prompts` key, which contains a dictionary of prompts categorized by file extension.

## `prompts`

This is the root key, containing all prompt definitions.

### `ts` (TypeScript/JavaScript Documentation Generation)

This section defines a prompt for generating documentation from TypeScript or JavaScript source code.

*   **`system`**:  This key holds the system message, which instructs the agent on its role and behavior.  It specifies that the agent acts as a principal architect, translating code into high-level documentation suitable for both technical and non-technical audiences. It emphasizes generating *only* the document content, avoiding conversational elements or code blocks. Strict rules are enforced regarding language (no emojis, specific words prohibited) and pronoun usage ("We"/"I" for maintainers, "You" for user instructions). The system message is duplicated in the provided YAML.
*   **`user`**: This key contains the user message, which provides the input to the agent. It instructs the agent to generate documentation for the provided TypeScript/JavaScript content, passed via the `{{content}}` variable.

### `md` (Markdown Documentation Refinement)

This section defines a prompt for refining existing Markdown documentation.

*   **`system`**: This key holds the system message, instructing the agent to act as a principal architect and improve existing documentation while preserving its original intent and technical depth. It emphasizes a comprehensive, pristine, and coherent output. Similar to the `ts` prompt, it specifies generating only the document content and enforces strict language and pronoun usage rules.
*   **`user`**: This key contains the user message, instructing the agent to audit and improve the provided Markdown documentation. It specifically requests checks for clarity, broken links (relative paths), and consistency with a "Container-First" platform vision. The content to be audited is passed via the `{{content}}` variable.

### `json` (JSON Schema/Data Documentation)

This section defines a prompt for documenting JSON schemas or data structures.

*   **`system`**: This key holds the system message, instructing the agent to act as a technical documentation expert. It focuses on explaining the data's architectural role, required/optional fields, and common use cases. It emphasizes outputting valid Markdown only, avoiding conversational text or code blocks, and prohibits mentioning specific terms.
*   **`user`**: This key contains the user message, instructing the agent to generate documentation for the provided JSON content, passed via the `{{content}}` variable.

### `yml` (YAML Configuration Documentation)

This section defines a prompt for documenting YAML configuration strings.

*   **`system`**: This key holds the system message, instructing the agent to act as a DevOps engineer and technical writer. It focuses on explaining the purpose, structure, and key controls within the provided YAML configuration. It emphasizes outputting valid Markdown only, avoiding conversational text or code blocks, and prohibits mentioning specific terms.
*   **`user`**: This key contains the user message, instructing the agent to generate documentation for the provided YAML content, passed via the `{{content}}` variable.