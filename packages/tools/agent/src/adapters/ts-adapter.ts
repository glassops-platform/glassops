import { AgentAdapter } from './interface.js';
import path from 'path';

export class TSAdapter implements AgentAdapter {
  canHandle(fileName: string): boolean {
    const extension = path.extname(fileName);
    return extension === '.ts' || extension === '.js' || extension === '.mjs';
  }

  async parse(filePath: string, content: string): Promise<string[]> {
    // Phase 1: Simple content passing with context
    // We can add AST parsing here later for granular doc-strings
    return [`File: ${filePath}\n\nCode Content:\n\`\`\`typescript\n${content}\n\`\`\``];
  }

  generatePrompt(filePath: string, parsedContent: string): string {
    return `You are a principal architect. Your job is to translate the provided content into a high-level, concise, but all-inclusive document that is easily understood by both highly technical and non-technical audiences. The document must be pristine, coherent, and professional.

IMPORTANT: Generate ONLY the document content itself. Do NOT include any conversational filler, preambles (e.g., "Here is the document..."), post-generation suggestions, or follow-up questions.

Generate documentation for the following TypeScript/JavaScript file:
${parsedContent}`;
  }

  postProcess(filePath: string, outputs: string[]): string {
    return outputs.join('\n\n');
  }
}
