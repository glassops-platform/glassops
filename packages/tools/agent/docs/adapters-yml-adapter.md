---
type: Documentation
domain: agent
origin: packages/tools/agent/src/adapters/yml-adapter.ts
last_modified: 2026-01-26
generated: true
source: packages/tools/agent/src/adapters/yml-adapter.ts
generated_at: 2026-01-26T14:24:40.753Z
hash: bb54d48830802f3f6516f166fc530539e8ce0e4c3ebdf8c267a50a814eae62a6
---

## YML Adapter Documentation

**Overview**

This adapter enables the agent to process YAML files. It handles file extension recognition, content parsing, prompt generation for analysis, and output consolidation. I am designed to integrate seamlessly with the agent framework to provide structured understanding of YAML configurations.

**Functionality**

The YML Adapter provides the following core functions:

*   **File Type Handling:** I determine if a file can be processed based on its extension. I support both `.yml` and `.yaml` file types.
*   **Content Parsing:** When provided with a file path and content, I format the content into a structured string suitable for analysis. This includes the file path and the YAML content itself, enclosed in a code block for clarity.
*   **Prompt Generation:** I construct a prompt designed to instruct a language model to analyze the YAML content. The prompt directs the model to act as a principal architect, producing a high-level, concise, and professional document explaining the configuration.
*   **Output Consolidation:** I combine multiple outputs from the language model into a single, coherent string, separated by double newlines for readability.

**Technical Details**

*   **Interface Implementation:** I implement the `AgentAdapter` interface, ensuring compatibility with the agent framework.
*   **Asynchronous Operation:** The `parse` function is asynchronous, allowing for non-blocking file processing.
*   **String-Based Input/Output:** I operate on strings for both input (file content) and output (parsed content and generated prompts).

**Configuration**

You do not directly configure this adapter. Its behavior is determined by the agent framework and the language model it employs. 

**Usage**

The agent framework automatically selects this adapter when processing files with `.yml` or `.yaml` extensions. You do not need to explicitly invoke this adapter. The agent will handle the parsing, prompt generation, and output processing automatically.

**Output Format**

The final output is a single string containing the analysis of the YAML configuration, formatted as a document. The structure and content of this document are determined by the language model and the prompt I generate.