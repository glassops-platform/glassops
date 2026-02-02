---
type: Documentation
domain: agent
origin: packages/tools/agent/src/validator.ts
last_modified: 2026-02-01
generated: true
source: packages/tools/agent/src/validator.ts
generated_at: 2026-02-01T19:51:37.419592
hash: 16ed4cfc2c65069c38113989639e5896ce05a75ddd3ff8f3a87cefccf6974f0b
---

## Agent Validator Documentation

This document details the functionality of the Agent Validator, a tool designed to assess the quality and adherence to specific standards of text-based content. It identifies potential issues related to formatting, style, and link integrity.

### Overview

The Agent Validator provides static analysis of content, returning a list of identified errors. It is intended to be integrated into content pipelines to ensure consistency and improve the overall quality of generated documentation or text assets.

### Functionality

The core function of this tool is the `validate` method. You provide the content as a string, and optionally a file path for context. The function returns an array of strings, where each string represents a detected error.

#### `validate(content: string, _filePath: string): string[]`

This static method performs the validation checks.

*   **`content`**:  The string containing the content to be validated.
*   **`_filePath`**: The path to the file containing the content. This parameter is currently unused but is reserved for future functionality, such as providing more context-aware error messages.
*   **Return Value**: An array of strings. Each string in the array describes an error found within the content. If no errors are found, an empty array is returned.

### Validation Checks

The `validate` method currently performs the following checks:

1.  **Frontmatter Check**:  Verifies that the content begins with a frontmatter block, denoted by `---`.  This is a common convention for metadata in many documentation formats.

2.  **Conversational Filler Check**:  Identifies and flags the presence of common conversational phrases that are often undesirable in technical documentation. The following phrases are currently checked:
    *   "Here is the document"
    *   "I hope this helps"
    *   "Let me know if"
    *   "Feel free to"
    *   "As requested"
    *   "Sure, here is"
    *   "Here's the"

    The check is case-insensitive.

3.  **Relative Link Check (Placeholder)**:  Currently, this check identifies potential relative links within the content using a regular expression: `\[.*?\]\((\.\.?\/.*?)\)`.  However, it does *not* currently validate the existence of the linked files.  Future development will include robust link validation, potentially involving asynchronous file system access.

### Example Usage

```typescript
import { Validator } from './validator';

const content = `---
title: My Document
---

Here is the document.  This is some content.
[Link to another document](./another_document.md)
`;

const errors = Validator.validate(content, 'my_document.md');

if (errors.length > 0) {
  console.log('Validation Errors:');
  for (const error of errors) {
    console.log(error);
  }
} else {
  console.log('Content is valid.');
}
```

### Future Enhancements

We plan to expand the capabilities of the Agent Validator to include:

*   More comprehensive link validation, including checking for broken links and ensuring correct link targets.
*   Support for additional validation checks, such as grammar and spelling checks.
*   Customizable validation rules, allowing users to define their own criteria for content quality.
*   Integration with other content management and publishing tools.