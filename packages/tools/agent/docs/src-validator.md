---
type: Documentation
domain: agent
origin: packages/tools/agent/src/validator.ts
last_modified: 2026-01-29
generated: true
source: packages/tools/agent/src/validator.ts
generated_at: 2026-01-29T21:08:25.299758
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

1. **Frontmatter Check:** Verifies that the content begins with a frontmatter block, denoted by ‘---’. Frontmatter is commonly used to store metadata about the content.

2. **Conversational Phrase Detection:** Identifies and flags the presence of predefined conversational phrases that are often undesirable in formal documentation or technical writing. The following phrases are currently checked:
    - "Here is the document"
    - "I hope this helps"
    - "Let me know if"
    - "Feel free to"
    - "As requested"
    - "Sure, here is"
    - "Here's the"
      The check is case-insensitive.

3. **Relative Link Check (Placeholder):** Currently, this check identifies potential relative links within the content using a regular expression. It does not currently validate the existence of the linked files. Future development will include more robust link validation, potentially involving file system access.

**Usage**

You can use the `Validator.validate()` method to check your content.

```typescript
import { Validator } from './validator';

const content = `---
title: My Document
---

Here is the document.  This is some content with a link to [another page](#).`;

const errors = Validator.validate(content, 'path/to/document.md');

if (errors.length > 0) {
    console.log('Validation Errors:');
    for (const error of errors) {
        console.log(error);
    }
} else {
    console.log('Content is valid.');
}
```

**Output**

The `validate` method returns an array of strings. Each string represents a specific error found within the content. An empty array indicates that no errors were detected.

**Future Enhancements**

Planned improvements include:

- Rigorous validation of relative links by checking file system existence.
- Expansion of the list of banned conversational phrases.
- Integration with more sophisticated linting tools.
- Support for additional content validation rules.
