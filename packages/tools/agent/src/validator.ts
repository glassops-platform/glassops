
export class Validator {
  static validate(content: string, _filePath: string): string[] {
    const errors: string[] = [];

    // 1. Check Frontmatter
    if (!content.startsWith('---')) {
      errors.push('Missing frontmatter block');
    }

    // 2. Check for Conversational Filler
    const bannedPhrases = [
      "Here is the document",
      "I hope this helps",
      "Let me know if",
      "Feel free to",
      "As requested",
      "Sure, here is",
      "Here's the"
      // "Here is" is too aggressive, might match "Here is an example"
    ];

    for (const phrase of bannedPhrases) {
      if (content.toLowerCase().includes(phrase.toLowerCase())) {
        errors.push(`Conversational phrase detected: "${phrase}"`);
      }
    }

    // 3. Check for Broken Relative Links (Naive check)
    // Matches [text](./path) or [text](../path)
    const linkRegex = /\[.*?\]\((\.\.?\/.*?)\)/g;
    let match;
    while ((match = linkRegex.exec(content)) !== null) {
      // const _link = match[1]; 
      // We can't easily validate file existence here without async/fs, 
      // but we can flag suspicious links if needed.
      // For now, this is just a placeholder for future rigorous link checking.
    }
    return errors;
  }
}
