---
type: Documentation
domain: agent
origin: packages/tools/agent/src/adapters/terraform-adapter.ts
last_modified: 2026-01-26
generated: true
source: packages/tools/agent/src/adapters/terraform-adapter.ts
generated_at: 2026-01-26T14:24:12.619Z
hash: d7206fb8f67a9090f0b60bf44a52a07a3caa8ca137a00d4a2cdb65cdea5df875
---

## Terraform Adapter Documentation

**1. Introduction**

This document details the Terraform Adapter, a component designed to integrate Terraform configuration files into a larger system for analysis and documentation. It enables processing of Terraform code to extract meaningful information and generate comprehensive documentation.

**2. Purpose**

The Terraform Adapter serves as a bridge between Terraform infrastructure-as-code and tools requiring structured understanding of that code. It identifies Terraform files, parses their content, and prepares prompts for language models to generate documentation.

**3. Core Functionality**

The adapter provides four primary functions:

*   **File Handling (canHandle):** Determines if the adapter can process a given file based on its extension. It specifically recognizes files with the `.tf` extension, indicating Terraform configuration files.
*   **Parsing (parse):** Reads the content of a Terraform file and formats it for use in subsequent steps. The output is a string array containing the file path and the Terraform code itself, enclosed in a code block for clarity.
*   **Prompt Generation (generatePrompt):** Constructs a prompt tailored for a language model. This prompt instructs the model to act as a DevOps Engineer/Terraform Expert and generate documentation for the provided Terraform code, focusing on resources, variables, outputs, and dependencies.
*   **Post-Processing (postProcess):** Combines the outputs from the language model into a single, formatted string, separated by double newlines for readability.

**4. Technical Details**

*   **Interface:** Implements the `AgentAdapter` interface, ensuring compatibility with the broader system.
*   **Dependencies:** Relies on the `path` module for file extension handling.
*   **Input:** Accepts a file path and the content of a Terraform file.
*   **Output:** Produces a formatted string containing documentation generated from the Terraform code.

**5. Usage**

To use the Terraform Adapter:

1.  Ensure the input file has a `.tf` extension.
2.  Provide the file path and content to the adapter.
3.  The adapter will prepare the content and generate a prompt for a language model.
4.  After receiving the language model’s output, the adapter will format it into a readable documentation string.

**6. Maintainability**

I designed this adapter to be modular and easily extensible. Future enhancements could include:

*   More sophisticated parsing to extract specific elements from Terraform code.
*   Support for additional Terraform-related file types (e.g., `.tfvars`).
*   Customizable prompt templates to control the documentation style.