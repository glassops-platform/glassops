---
type: Documentation
domain: agent
origin: packages/tools/agent/src/validator.ts
last_modified: 2026-01-26
generated: true
source: packages/tools/agent/src/validator.ts
generated_at: 2026-01-26T14:13:38.856Z
hash: 9280486d450e51646da8376017615d9a70b9c8993abb21f99ab51457a1b40820
---

## Agent Validator Documentation

**Overview**

The Agent Validator is a tool designed to assess the quality and suitability of content intended for use with the agent system. It performs a series of checks to identify common issues that can negatively impact performance or user experience. This document details the validation process.

**Functionality**

The `Validator` class provides a static `validate` method that accepts content as a string and a file path (currently unused). The method returns an array of strings, where each string represents a detected error or issue. 

**Validation Checks**

The `validate` method currently performs the following checks:

1. **Frontmatter Presence:**  It verifies that the content begins with a frontmatter block (`---`). Frontmatter is essential for defining metadata associated with the content.  If missing, an error is reported.

2. **Conversational Filler Detection:** The tool identifies and flags instances of common conversational phrases that are generally undesirable in agent-processed content. These phrases often add unnecessary verbosity and can detract from clarity. The following phrases are currently checked:
    * "Here is the document"
    * "I hope this helps"
    * "Let me know if"
    * "Feel free to"
    * "As requested"
    * "Sure, here is"
    * "Here's the"
    
    When a banned phrase is found, an error message including the phrase is added to the results. The check is case-insensitive.

3. **Relative Link Check (Placeholder):**  A regular expression is used to identify potential relative links within the content (links starting with `./` or `../`). Currently, this check does not verify the existence of the linked files. It serves as a placeholder for future implementation of more robust link validation. 

**Usage**

To validate content, call the `Validator.validate()` method, passing the content string and the file path as arguments.  

```typescript
import { Validator } from './validator';

const content = "--- \n title: My Document \n ---\n Here is the document you requested.";
const filePath = "path/to/document.md";
const errors = Validator.validate(content, filePath);

if (errors.length > 0) {
  console.log("Validation Errors:");
  for (const error of errors) {
    console.log(error);
  }
} else {
  console.log("Content is valid.");
}
```

**Future Enhancements**

Planned improvements to the validator include:

*   Rigorous relative link validation (checking file existence).
*   Additional checks for content quality, such as readability scores and keyword density.
*   Customizable validation rules.
*   Support for different content formats beyond plain text.