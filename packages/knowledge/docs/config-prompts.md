---
type: Documentation
domain: knowledge
origin: packages/knowledge/config/prompts.yml
last_modified: 2026-02-01
generated: true
source: packages/knowledge/config/prompts.yml
generated_at: 2026-02-01T19:27:45.708644
hash: fa0450b4b022cb08c826f4a7a5d3d82f9e04c2559269e31a838da7444eedfee5
---

# Prompts Configuration Documentation

This file configures the prompts used by the GlassOps documentation agent. It defines different system and user prompts for various adapter types, allowing the agent to generate documentation tailored to specific languages and formats.

## Structure

The YAML file is structured around a top-level key `prompts`. This key contains a dictionary where each key represents an adapter type (e.g., `go`, `py`, `ts`).  A special key `_shared_rules` defines common instructions applied to all prompts. A `default` adapter is also provided as a fallback.

## Key Breakdown

### `prompts`

- **`_shared_rules`**:  A string containing a set of strict rules that are prepended to all system prompts. These rules govern the style and content of the generated documentation, prohibiting the use of certain words, emojis, and formatting conventions.  The value is a multi-line string defining these constraints.
- **`go`**: Configuration for Go source code documentation.
    - **`system`**: A string defining the system prompt for the Go adapter. This prompt instructs the agent to act as a principal architect and platform engineer, translating Go code into comprehensive documentation. It includes the `{{shared_rules}}` variable to incorporate the shared rules.
    - **`user`**: A string defining the user prompt for the Go adapter. This prompt instructs the agent to generate documentation for the provided Go file, using the `{{content}}` variable to represent the file's content.
- **`py`**: Configuration for Python source code documentation.  Similar structure to `go`, but tailored for Python and an AI/ML expert role.
    - **`system`**: System prompt for Python documentation.
    - **`user`**: User prompt for Python documentation.
- **`ts`**: Configuration for TypeScript/JavaScript source code documentation.
    - **`system`**: System prompt for TypeScript/JavaScript documentation.
    - **`user`**: User prompt for TypeScript/JavaScript documentation.
- **`yml`**: Configuration for YAML configuration file documentation.
    - **`system`**: System prompt for YAML documentation.
    - **`user`**: User prompt for YAML documentation.
- **`json`**: Configuration for JSON schema/data documentation.
    - **`system`**: System prompt for JSON documentation.
    - **`user`**: User prompt for JSON documentation.
- **`dockerfile`**: Configuration for Dockerfile documentation.
    - **`system`**: System prompt for Dockerfile documentation.
    - **`user`**: User prompt for Dockerfile documentation.
- **`tf`**: Configuration for Terraform configuration documentation.
    - **`system`**: System prompt for Terraform documentation.
    - **`user`**: User prompt for Terraform documentation.
- **`apex`**: Configuration for Salesforce Apex code documentation.
    - **`system`**: System prompt for Apex documentation.
    - **`user`**: User prompt for Apex documentation.
- **`lwc`**: Configuration for Salesforce Lightning Web Component documentation.
    - **`system`**: System prompt for LWC documentation.
    - **`user`**: User prompt for LWC documentation.
- **`default`**:  A fallback configuration used when the adapter type is not explicitly defined.
    - **`system`**: System prompt for the default adapter.
    - **`user`**: User prompt for the default adapter.

## Variables

The prompts use the following variables:

- **`{{shared_rules}}`**:  This variable is replaced with the content of the `_shared_rules` key.
- **`{{content}}`**: This variable is replaced with the actual content of the file being documented.