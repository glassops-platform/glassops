---
type: Documentation
domain: agent
origin: packages/tools/agent/src/adapters/docker-adapter.ts
last_modified: 2026-01-31
generated: true
source: packages/tools/agent/src/adapters/docker-adapter.ts
generated_at: 2026-01-31T10:16:34.118619
hash: 098372ed82eb7d58e32e7edba32f2300351a37a4344df625a5893b16e48fc3bc
---

## Docker Adapter Documentation

**1. Introduction**

This document details the Docker Adapter, a component designed to process Dockerfile content and generate documentation. It functions as an adapter within a larger agent system, responsible for handling files specifically identified as Dockerfiles.

**2. Purpose**

The Docker Adapter enables automated documentation generation for Dockerfiles. It parses the Dockerfile content, prepares a prompt for a language model, and formats the resulting output for presentation.

**3. Core Functionality**

The adapter provides four primary functions:

*   **`canHandle(fileName: string): boolean`**: Determines if the adapter is capable of processing a given file based on its name. It returns `true` if the file name (case-insensitive) ends with “dockerfile”, and `false` otherwise.

*   **`parse(filePath: string, content: string): Promise<string[]>`**:  Parses the Dockerfile content. This function takes the file path and the file’s content as input. It formats the content into a string array suitable for prompt generation. The output includes the file path and the Dockerfile content enclosed in a code block.

*   **`generatePrompt(filePath: string, parsedContent: string): string`**: Constructs a prompt for a language model. This prompt instructs the model to act as a DevOps Engineer and provide documentation for the provided Dockerfile content. The prompt includes the parsed content from the `parse` function.

*   **`postProcess(filePath: string, outputs: string[]): string`**:  Combines the outputs from the language model into a single string, separated by double newlines. This function takes the file path and the array of outputs as input and returns a single, formatted string.

**4. Adapter Interface**

This adapter implements the `AgentAdapter` interface. This interface defines a standard contract for adapters within the agent system, ensuring consistency and interoperability.

**5. Usage**

You integrate this adapter into the agent system. The system will call `canHandle` to determine if the adapter should process a file. If it can, the system will sequentially call `parse`, `generatePrompt`, and `postProcess` to obtain the final documentation.

**6. Example**

Consider a Dockerfile located at `/path/to/Dockerfile` with the following content:

```dockerfile
FROM ubuntu:latest
RUN apt-get update && apt-get install -y nginx
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

The `parse` function would produce the following output:

```
[
  "File: /path/to/Dockerfile\n\nDockerfile Content:\n\`\`\`dockerfile\nFROM ubuntu:latest\nRUN apt-get update && apt-get install -y nginx\nEXPOSE 80\nCMD [\"nginx\", \"-g\", \"daemon off;\"]\n\`\`\`"
]
```

This output is then passed to `generatePrompt` to create a prompt for the language model. The `postProcess` function then combines the language model’s responses into a single documentation string.