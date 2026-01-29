import { AgentAdapter } from './interface.js';
import path from 'path';

export class PyAdapter implements AgentAdapter {
  canHandle(fileName: string): boolean {
    const extension = path.extname(fileName);
    return extension === '.py';
  }

  async parse(filePath: string, content: string): Promise<string[]> {
    // Simple pass-through for now. 
    // Future improvement: Parse python AST to split classes/functions
    return [`File: ${filePath}\n\nPython Code Content:\n\`\`\`python\n${content}\n\`\`\``];
  }

  generatePrompt(filePath: string, parsedContent: string): string {
    return `You are a principal engineer and Python expert. Your job is to translate the provided content into a high-level, concise, but all-inclusive document that is easily understood by both highly technical and non-technical audiences. The document must be pristine, coherent, and professional.

IMPORTANT: Generate ONLY the document content itself. Do NOT include any conversational filler, preambles (e.g., "Here is the document..."), post-generation suggestions, or follow-up questions.

Generate documentation for the following Python file:
${parsedContent}`;
  }

  postProcess(filePath: string, outputs: string[]): string {
    return outputs.join('\n\n');
  }
}
