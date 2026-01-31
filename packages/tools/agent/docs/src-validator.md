---
type: Documentation
domain: agent
origin: packages/tools/agent/src/validator.ts
last_modified: 2026-01-31
generated: true
source: packages/tools/agent/src/validator.ts
generated_at: 2026-01-31T10:21:17.919667
hash: 16ed4cfc2c65069c38113989639e5896ce05a75ddd3ff8f3a87cefccf6974f0b
---

## Agent Validator Documentation

This document details the functionality of the Agent Validator, a tool designed to assess the quality and adherence to specific standards of text-based content. It identifies potential issues related to formatting, style, and link integrity.

### Overview

The Agent Validator provides static analysis of content, returning a list of identified errors. It is intended to be integrated into content pipelines to ensure consistency and quality before deployment. We focus on identifying common issues that detract from professional documentation.

### Functionality

The core functionality is provided by the `Validator` class and its static `validate` method.

#### `Validator.validate(content: string, _filePath: string): string[]`

This method accepts content as a string and an optional file path (currently unused) and returns an array of strings, where each string represents a detected error.

**Parameters:**

- `content`: The string content to be validated.
- `_filePath`: The path to the file containing the content (not currently used for validation).

**Return Value:**

An array of strings representing validation errors. An empty array indicates no errors were found.

**Validation Checks Performed:**

1.  **Frontmatter Check:** Verifies that the content begins with a frontmatter block (`---`). This is a common convention for metadata in documentation files.

2.  **Conversational Phrase Detection:** Identifies and flags the presence of common conversational phrases that are generally undesirable in formal documentation. The following phrases are currently checked:
    - "Here is the document"
    - "I hope this helps"
    - "Let me know if"
    - "Feel free to"
    - "As requested"
    - "Sure, here is"
    - "Here's the"

    The check is case-insensitive.

3.  **Relative Link Check (Placeholder):** Currently, this check identifies potential relative links using a regular expression: `\[.*?\]\((\.\.?\/.*?)\)`. It does _not_ currently validate the existence of the linked files. This functionality is reserved for future development and will involve asynchronous file system checks.

### Usage

You can use the `Validator.validate` method to check content. For example:

```typescript
import { Validator } from './validator';

const myContent = `---
title: My Document
---

Here is the document.  This is some content. [Link to another page] (./another-page.md)`;

const errors = Validator.validate(myContent, 'path/to/my/document.md');

if (errors.length > 0) {
    console.log('Validation Errors:');
    for (const error of errors) {
        console.log(error);
    }
} else {
    console.log('Content is valid.');
}
```

This will output the following errors:

```
Validation Errors:
Conversational phrase detected: "Here is the document"
```

### Future Enhancements

- Implement robust relative link validation by checking file system existence.
- Expand the list of banned conversational phrases.
- Add support for validating other content aspects, such as heading structure and image alt text.
- Provide more detailed error messages with specific line numbers.
