---
type: Documentation
domain: agent
origin: packages/tools/agent/src/adapters/terraform-adapter.ts
last_modified: 2026-02-01
generated: true
source: packages/tools/agent/src/adapters/terraform-adapter.ts
generated_at: 2026-02-01T19:49:22.760453
hash: 7c28ae755942fb80ed2efcfecf4b5ac58629d14bc12bba1935b1c59f82527957
---

## Terraform Adapter Documentation

**1. Introduction**

This document details the Terraform Adapter, a component designed to integrate Terraform configuration files into a larger operational system. It provides capabilities for identifying, parsing, and preparing Terraform code for analysis and documentation generation.

**2. Overview**

The Terraform Adapter functions as an agent within a broader system, responsible for handling files with the `.tf` extension. It extracts content, formats it, and prepares it for processing by language models or other analytical tools. The adapter’s primary function is to transform Terraform code into a format suitable for generating documentation and understanding infrastructure as code.

**3. Functionality**

The adapter provides four core functions:

*   **`canHandle(fileName: string): boolean`**: This function determines if the adapter is capable of processing a given file based on its extension. It returns `true` if the file extension is `.tf`, indicating a Terraform configuration file, and `false` otherwise.

*   **`parse(filePath: string, content: string): Promise<string[]>`**: This asynchronous function takes the file path and content of a Terraform file as input. It formats the content into a structured string array, including the file path and the Terraform code itself, enclosed in a code block. The function returns a Promise that resolves to an array of strings. Example output:

    ```
    [
      "File: /path/to/your/file.tf\n\nTerraform HCL:\n\`\`\`hcl\nresource \"aws_instance\" \"example\" {\n  ami           = \"ami-0c55b2ab9919489a2\"\n  instance_type = \"t2.micro\"\n}\n\`\`\`"
    ]
    ```

*   **`generatePrompt(filePath: string, parsedContent: string): string`**: This function constructs a prompt designed for a language model. The prompt instructs the model to act as a DevOps Engineer/Terraform Expert and generate documentation for the provided Terraform code. It includes the parsed content from the `parse` function.

*   **`postProcess(filePath: string, outputs: string[]): string`**: This function takes the file path and an array of strings (outputs from a language model or other processing step) and concatenates them into a single string, separated by double newlines. This provides a consolidated output for presentation or further processing.

**4. Usage**

You integrate this adapter into a system that processes infrastructure-as-code files. The typical workflow is as follows:

1.  The system identifies a file with a `.tf` extension.
2.  The system invokes the `canHandle` function to confirm the adapter’s compatibility.
3.  If compatible, the system reads the file content and passes the file path and content to the `parse` function.
4.  The output of the `parse` function is then used as input to the `generatePrompt` function to create a prompt for a language model.
5.  The language model processes the prompt and returns documentation as an array of strings.
6.  Finally, the `postProcess` function combines these strings into a single, formatted output.

**5. Dependencies**

*   `path`: Node.js built-in module for handling file paths.
*   `./interface.js`: Defines the `AgentAdapter` interface that this class implements.

**6. Implementation Details**

The `TerraformAdapter` class implements the `AgentAdapter` interface. This ensures a consistent structure for interacting with different types of infrastructure-as-code files. The adapter focuses specifically on Terraform configuration files and provides tailored parsing and prompt generation for this format. I have designed this adapter to be modular and easily extensible to support additional Terraform features or integrations.