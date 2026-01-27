import { AgentAdapter } from './interface.js';

export class JSONAdapter implements AgentAdapter {
  canHandle(extension: string): boolean {
    return extension === '.json';
  }

  async parse(filePath: string, content: string): Promise<string[]> {
    return [`File: ${filePath}\n\nJSON Content:\n\`\`\`json\n${content}\n\`\`\``];
  }

  generatePrompt(filePath: string, parsedContent: string): string {
    return `You are a principal architect. Your job is to translate the provided content into a high-level, concise, but all-inclusive document that is easily understood by both highly technical and non-technical audiences. The document must be pristine, coherent, and professional.

IMPORTANT: Generate ONLY the document content itself. Do NOT include any conversational filler, preambles, post-generation suggestions, or follow-up questions.

Analyze the following JSON file:
${parsedContent}`;
  }

  postProcess(filePath: string, outputs: string[]): string {
    return outputs.join('\n\n');
  }
}
