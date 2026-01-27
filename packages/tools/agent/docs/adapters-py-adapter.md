---
type: Documentation
domain: agent
origin: packages/tools/agent/src/adapters/py-adapter.ts
last_modified: 2026-01-26
generated: true
source: packages/tools/agent/src/adapters/py-adapter.ts
generated_at: 2026-01-26T14:23:53.193Z
hash: 0469c7ef8c254e35322e240f556d429bab07581e353b960d884886f97a7daaaf
---

## Python Adapter Documentation

**Overview**

This adapter enables the agent to process Python (.py) files. It handles file identification, content parsing, prompt generation for documentation creation, and output consolidation.

**Functionality**

The Python Adapter provides the following core functions:

* **File Handling:** Identifies files with the `.py` extension.
* **Content Parsing:** Currently, the adapter performs a simple pass-through of the file content. Future development will include parsing the Python Abstract Syntax Tree (AST) to enable more granular splitting of code into classes and functions.
* **Prompt Generation:** Constructs a prompt for a language model, instructing it to act as a principal engineer and Python expert to generate high-quality documentation. The prompt includes the parsed Python file content.
* **Output Processing:** Combines multiple outputs from the language model into a single, formatted string, separated by double newlines.

**Technical Details**

* **Interface Implementation:** Implements the `AgentAdapter` interface, ensuring compatibility with the agent framework.
* **Input:** Accepts a file path (string) and file content (string).
* **Output:** Returns a string containing the generated documentation.

**Configuration**

Currently, this adapter does not require specific configuration. 

**Future Enhancements**

Planned improvements include:

* **AST Parsing:** Implementing Python AST parsing to improve content segmentation for documentation.
* **Advanced Prompting:** Refining the prompt to further optimize documentation quality and format.