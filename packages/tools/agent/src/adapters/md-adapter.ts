import { AgentAdapter } from './interface.js';

export class MarkdownAdapter implements AgentAdapter {
  canHandle(extension: string): boolean {
    return extension === '.md';
  }

  async parse(filePath: string, content: string): Promise<string[]> {
    const headerPrefix = `File: ${filePath}\n\nMarkdown Content:\n`;
    const targetChunkSize = 10000; // ~10KB chunks

    if (content.length < targetChunkSize) {
      return [`${headerPrefix}${content}`];
    }

    const sections = content.split(/(?=^#+ .*$)/m).filter(s => s.trim().length > 0);
    const chunks: string[] = [];
    let currentChunk = "";

    for (const section of sections) {
      if ((currentChunk + section).length > targetChunkSize && currentChunk.length > 0) {
        chunks.push(`${headerPrefix}${currentChunk}`);
        currentChunk = section;
      } else {
        currentChunk += section;
      }
    }

    if (currentChunk.length > 0) {
      chunks.push(`${headerPrefix}${currentChunk}`);
    }

    return chunks;
  }

  generatePrompt(filePath: string, parsedContent: string): string {
    return `You are a principal architect. Your job is to refine and improve the provided documentation while preserving all of its original intent, technical depth, and core messaging. This is NOT a summary; it should be a comprehensive, pristine, and coherent version of the original that is easily understood by both highly technical and non-technical audiences.

IMPORTANT: Generate ONLY the document content itself. Do NOT include any conversational filler, preambles, post-generation suggestions, or follow-up questions.

Audit and improve the following technical documentation. Check for clarity, broken-link patterns (relative paths), and consistency with the "Container-First" platform vision. Provide an improved version of the document.

${parsedContent}`;
  }

  postProcess(filePath: string, outputs: string[]): string {
    const combined = outputs.join('\n\n');
    return combined.replace(/^```markdown\n/gm, '').replace(/\n```$/gm, '');
  }
}
