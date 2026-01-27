import { AgentAdapter } from './interface.js';
import path from 'path';

export class TerraformAdapter implements AgentAdapter {
  canHandle(fileName: string): boolean {
    const extension = path.extname(fileName);
    return extension === '.tf';
  }

  async parse(filePath: string, content: string): Promise<string[]> {
    return [`File: ${filePath}\n\nTerraform HCL:\n\`\`\`hcl\n${content}\n\`\`\``];
  }

  generatePrompt(filePath: string, parsedContent: string): string {
    return `You are a DevOps Engineer / Terraform Expert.
Generate documentation for the following Terraform infrastructure code.
Identify resources, variables, outputs, and dependencies.

${parsedContent}`;
  }

  postProcess(filePath: string, outputs: string[]): string {
    return outputs.join('\n\n');
  }
}
