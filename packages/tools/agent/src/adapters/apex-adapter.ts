import { AgentAdapter } from './interface.js';
import path from 'path';

export class ApexAdapter implements AgentAdapter {
  canHandle(fileName: string): boolean {
    const extension = path.extname(fileName);
    return extension === '.cls' || extension === '.trigger';
  }

  async parse(filePath: string, content: string): Promise<string[]> {
    return [`File: ${filePath}\n\nApex Code Content:\n\`\`\`apex\n${content}\n\`\`\``];
  }

  generatePrompt(filePath: string, parsedContent: string): string {
    return `You are a helper for a Salesforce Architect.
Generate technical documentation for the following Apex class/trigger.
Focus on business logic, triggers, and security implications.

${parsedContent}`;
  }

  postProcess(filePath: string, outputs: string[]): string {
    return outputs.join('\n\n');
  }
}
