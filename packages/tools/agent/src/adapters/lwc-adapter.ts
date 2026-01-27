import { AgentAdapter } from './interface.js';
import fs from 'fs/promises';
import path from 'path';

export class LWCAdapter implements AgentAdapter {
  canHandle(fileName: string): boolean {
    const extension = path.extname(fileName);
    // Only trigger processing on the main JS controller to avoid duplicate docs
    // We will bundle HTML/CSS manually
    const inLWC = fileName.includes('lwc') || fileName.includes('lwc/');
    return extension === '.js' && inLWC;
  }

  async parse(filePath: string, content: string): Promise<string[]> {
    // Check if this is likely an LWC (has sibling .html or is in lwc folder)
    const dir = path.dirname(filePath);
    const baseName = path.basename(filePath, '.js');
    const htmlPath = path.join(dir, `${baseName}.html`);
    const cssPath = path.join(dir, `${baseName}.css`);
    
    // Safety check (redundant if canHandle is strict, but good for robustness)
    const isLWC = filePath.includes('lwc') || await this.exists(htmlPath);

    if (!isLWC) {
        return []; 
    }

    let bundle = `File: ${filePath}\n\nController (JS):\n\`\`\`javascript\n${content}\n\`\`\`\n`;

    if (await this.exists(htmlPath)) {
        const html = await fs.readFile(htmlPath, 'utf-8');
        bundle += `\nTemplate (HTML):\n\`\`\`html\n${html}\n\`\`\`\n`;
    }

    if (await this.exists(cssPath)) {
        const css = await fs.readFile(cssPath, 'utf-8');
        bundle += `\nStyles (CSS):\n\`\`\`css\n${css}\n\`\`\`\n`;
    }

    return [bundle];
  }

  generatePrompt(filePath: string, parsedContent: string): string {
    return `You are a Salesforce Developer Expert.
Generate documentation for the following Lightning Web Component (LWC).
Describe the component's purpose, public properties (@api), events, and UI behavior.

${parsedContent}`;
  }

  postProcess(filePath: string, outputs: string[]): string {
    return outputs.join('\n\n');
  }

  private async exists(p: string): Promise<boolean> {
      try {
          await fs.access(p);
          return true;
      } catch {
          return false;
      }
  }
}
