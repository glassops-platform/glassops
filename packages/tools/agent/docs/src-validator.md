---
type: Documentation
domain: agent
origin: packages/tools/agent/src/validator.ts
last_modified: 2026-01-31
generated: true
source: packages/tools/agent/src/validator.ts
generated_at: 2026-01-31T09:24:07.533081
hash: 16ed4cfc2c65069c38113989639e5896ce05a75ddd3ff8f3a87cefccf6974f0b
---

## Agent Validator Documentation

**Introduction**

This document details the functionality of the Agent Validator, a tool designed to assess the quality and adherence to specific standards of text-based content. It identifies potential issues within content, helping to ensure clarity and professionalism.

**Purpose**

The Agent Validator is intended to be integrated into content pipelines to automatically check for common problems before publication or further processing. It focuses on structural elements and stylistic concerns.

**Functionality**

The core function of this tool is the `validate` method. This method accepts content as a string and a file path (currently unused) as input, and returns an array of strings representing any identified errors. 

The validation process consists of the following checks:

* **Frontmatter Check:** Verifies that the content begins with a frontmatter block, denoted by ‘---’. This is a common convention for metadata separation in many content formats.
* **Conversational Phrase Detection:** Identifies and flags the presence of predefined conversational phrases that are often undesirable in formal documentation or technical writing. A list of banned phrases is maintained within the tool. The check is case-insensitive.
* **Relative Link Check (Placeholder):** Currently, this feature performs a basic pattern match for relative links within the content (e.g., `[text](./path)` or `[text](../path)`).  Future development will include more robust validation, potentially verifying the existence of linked files.

**Usage**

To use the validator, you must call the `Validator.validate()` method, providing the content string as an argument. 

```typescript
import { Validator } from './validator';

const content = `---
title: My Document
---

Here is the document you requested. 

[Link to another page](./another-page.md)`

const errors = Validator.validate(content, 'path/to/document.md');

if (errors.length > 0) {
  console.log('Validation Errors:');
  errors.forEach(error => console.log(error));
} else {
  console.log('Content is valid.');
}
```

**Output**

The `validate` method returns an array of strings. Each string in the array represents a specific error found within the content. If no errors are found, an empty array is returned.

**Future Enhancements**

Planned improvements include:

* Implementation of robust relative link validation, including file system checks.
* Expansion of the list of banned conversational phrases.
* Addition of checks for other common content issues, such as excessive passive voice or inconsistent formatting.
* Support for configurable validation rules.