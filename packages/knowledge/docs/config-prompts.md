---
type: Documentation
domain: knowledge
last_modified: 2026-02-02
generated: true
source: packages/knowledge/config/prompts.yml
generated_at: 2026-02-02T22:27:13.874010
hash: c809ab9a381397a3b92f80840e9ea06a056a6aab7b4a737819df290867ee5e0e
---

# Prompts Configuration Documentation

This YAML file configures the prompts used by the GlassOps documentation agent. It defines different system and user prompts for various adapter types, allowing the agent to generate documentation tailored to specific file types and technologies.

## Structure

The file is structured around a top-level key `prompts`. This key contains a dictionary where each key represents an adapter type (e.g., `go`, `py`, `ts`).  A special key `_shared_rules` defines common instructions applied to all prompts. A `default` adapter is also provided as a fallback.

## Key Definitions

### `prompts`

The main container for all prompt configurations.

### `_shared_rules`

This key holds a string containing a set of strict rules that are prepended to every system prompt. These rules govern the style and content of the generated documentation.  The rules include constraints on language, formatting, and attribution.  The value is a multi-line string.

### Adapter Types (`go`, `py`, `ts`, `yml`, `json`, `dockerfile`, `tf`, `apex`, `lwc`, `default`)

Each adapter type defines a `system` and `user` prompt.

#### `system`

A string containing instructions for the language model. This prompt defines the role the model should assume (e.g., "principal architect", "DevOps engineer") and the specific tasks it should perform.  It also includes the `{{shared_rules}}` placeholder, which is replaced with the content of the `_shared_rules` key during prompt construction.  The `system` prompt emphasizes generating only the document content, avoiding conversational elements.

#### `user`

A string containing the instructions for the user. This prompt provides the context for the documentation task, typically including a placeholder `{{content}}` where the file content to be documented will be inserted.

### `go`

Configures the documentation generation for Go source code. The system prompt instructs the model to act as a principal architect and platform engineer, focusing on package purpose, key types, functions, error handling, and concurrency.

### `py`

Configures documentation for Python source code. The system prompt instructs the model to act as a principal architect and AI/ML expert, focusing on module purpose, classes, functions, type hints, and design patterns.

### `ts`

Configures documentation for TypeScript/JavaScript files. The system prompt instructs the model to act as a principal architect, emphasizing a pristine and coherent document.

### `yml`

Configures documentation for YAML configuration files. The system prompt instructs the model to act as a DevOps engineer and technical writer, focusing on the purpose, structure, and key controls of the configuration.

### `json`

Configures documentation for JSON schemas or data structures. The system prompt instructs the model to act as a technical documentation expert, focusing on data representation, required fields, and use cases.

### `dockerfile`

Configures documentation for Dockerfiles. The system prompt instructs the model to act as a DevOps expert, focusing on the base image, stages, instructions, security, and build/run processes.

### `tf`

Configures documentation for Terraform configurations. The system prompt instructs the model to act as an Infrastructure as Code expert, focusing on resources, variables, outputs, dependencies, and security.

### `apex`

Configures documentation for Salesforce Apex code. The system prompt instructs the model to act as a Salesforce architect, focusing on class/trigger purpose, methods, governor limits, and integration points.

### `lwc`

Configures documentation for Salesforce Lightning Web Components. The system prompt instructs the model to act as a Salesforce Lightning expert, focusing on component purpose, properties, wire adapters, event handling, and lifecycle hooks.

### `default`

Provides a fallback configuration for file types not explicitly defined. It instructs the model to act as a principal architect and generate a high-level document.