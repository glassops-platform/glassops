---
type: Documentation
domain: agent
origin: packages/tools/agent/src/adapters/apex-adapter.ts
last_modified: 2026-01-26
generated: true
source: packages/tools/agent/src/adapters/apex-adapter.ts
generated_at: 2026-01-26T14:22:08.874Z
hash: 63fe439758f42e08865f73c4f1675055324908a4942fc3cdf9f58150c0387ccf
---

## Apex Adapter Documentation

**1. Introduction**

This document details the functionality of the Apex Adapter, a component designed to process Salesforce Apex code files (.cls and .trigger) and prepare them for analysis and documentation generation. It serves as a bridge between raw code and the agent responsible for creating meaningful documentation.

**2. Purpose**

The Apex Adapter enables the automated creation of technical documentation for Apex classes and triggers. It identifies relevant files, extracts their content, and formats it into a prompt suitable for a language model to generate comprehensive documentation.

**3. Core Functionality**

The adapter provides four key functions:

*   **File Handling (canHandle):** Determines if the adapter can process a given file based on its extension. It supports files with the extensions `.cls` (Apex class) and `.trigger` (Apex trigger).
*   **Content Parsing (parse):** Reads the content of an Apex file and formats it for inclusion in a prompt. The output is a string array containing the file path and the code content, enclosed in a code block.
*   **Prompt Generation (generatePrompt):** Constructs a prompt for a language model, instructing it to act as a Salesforce Architect and generate technical documentation. The prompt includes the parsed Apex code content and specific instructions to focus on business logic, triggers, and security considerations.
*   **Output Processing (postProcess):** Combines multiple outputs from the language model into a single, formatted string, separated by double newlines.

**4. Workflow**

1.  The system identifies an Apex file (.cls or .trigger).
2.  The Apex Adapter’s `canHandle` function verifies file compatibility.
3.  If compatible, the `parse` function extracts and formats the file’s content.
4.  The `generatePrompt` function creates a prompt containing the formatted code and instructions for documentation generation.
5.  This prompt is sent to a language model.
6.  The language model generates documentation.
7.  The `postProcess` function combines and formats the generated documentation into a final output.

**5. Configuration**

The Apex Adapter does not require explicit configuration. It operates based on file extensions and the predefined prompt template. 

**6. Dependencies**

*   `path`: Node.js module for handling file paths.
*   `AgentAdapter` interface: Defines the standard interface for all adapters.

**7. Future Considerations**

Potential enhancements include:

*   Support for additional Apex file types (e.g., .page, .component).
*   Customizable prompt templates.
*   Integration with Salesforce metadata APIs for enhanced code analysis.