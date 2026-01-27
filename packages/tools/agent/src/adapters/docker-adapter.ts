import { AgentAdapter } from './interface.js';
import path from 'path';

export class DockerAdapter implements AgentAdapter {
  canHandle(fileName: string): boolean {
    return path.basename(fileName).toLowerCase() === 'dockerfile';
  }

  async parse(filePath: string, content: string): Promise<string[]> {
    return [`File: ${filePath}\n\nDockerfile Content:\n\`\`\`dockerfile\n${content}\n\`\`\``];
  }

  generatePrompt(filePath: string, parsedContent: string): string {
    return `You are a DevOps Engineer.
Generate documentation for the following Dockerfile.
Explain the base image, build stages, exposed ports, and runtime configuration.

${parsedContent}`;
  }

  postProcess(filePath: string, outputs: string[]): string {
    return outputs.join('\n\n');
  }
}
