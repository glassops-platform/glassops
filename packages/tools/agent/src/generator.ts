import fs from 'fs/promises';
import { existsSync } from 'fs';
import path from 'path';
import { GeminiClient } from './llm-client.js';
import { Validator } from './validator.js';
import { Scanner } from './scanner.js';
import { AgentAdapter } from './adapters/interface.js';
import { TSAdapter } from './adapters/ts-adapter.js';
import { PyAdapter } from './adapters/py-adapter.js';
import { ApexAdapter } from './adapters/apex-adapter.js';
import { LWCAdapter } from './adapters/lwc-adapter.js';
import { TerraformAdapter } from './adapters/terraform-adapter.js';
import { DockerAdapter } from './adapters/docker-adapter.js';
import { MarkdownAdapter } from './adapters/md-adapter.js';
import { YMLAdapter } from './adapters/yml-adapter.js';
import { JSONAdapter } from './adapters/json-adapter.js';
import chalk from 'chalk';
import yaml from 'js-yaml';
import { createHash } from 'crypto';

interface DocCache {
  [filepath: string]: {
    hash: string;
    generatedFiles: string[];
    timestamp: string;
  };
}

interface PromptConfig {
  [key: string]: {
    system: string;
    user: string;
  };
}

export class Generator {
  private client: GeminiClient;
  private scanner: Scanner;
  private adapters: AgentAdapter[] = [];
  private rootDir: string;
  private cachePath: string;
  private cache: DocCache = {};
  private prompts: PromptConfig = {};

  constructor(rootDir: string) {
    this.rootDir = rootDir;
    // Move cache to config directory
    this.cachePath = path.join(rootDir, 'config', 'doc-cache.json');
    this.client = new GeminiClient();
    this.scanner = new Scanner(rootDir);
    
    // Initialize adapters
    this.adapters.push(new TSAdapter());
    this.adapters.push(new PyAdapter());
    this.adapters.push(new ApexAdapter());
    this.adapters.push(new LWCAdapter());
    this.adapters.push(new TerraformAdapter());
    this.adapters.push(new DockerAdapter());
    // this.adapters.push(new MarkdownAdapter());
    this.adapters.push(new YMLAdapter());
    this.adapters.push(new JSONAdapter());
  }

  private async loadCache() {
    try {
      const cacheDir = path.dirname(this.cachePath);
      if (!existsSync(cacheDir)) {
          await fs.mkdir(cacheDir, { recursive: true });
      }

      if (existsSync(this.cachePath)) {
        const content = await fs.readFile(this.cachePath, 'utf-8');
        this.cache = JSON.parse(content);
        console.log(chalk.gray(`Loaded cache from ${this.cachePath}`));
      }
    } catch (e) {
      console.warn(chalk.yellow('Failed to load cache, starting fresh.'));
    }
  }

  private async saveCache() {
    await fs.writeFile(this.cachePath, JSON.stringify(this.cache, null, 2));
  }

  private async loadPrompts() {
    try {
      // Assuming prompts.yml is in the package source or copied to dist
      // Strategy: Try loading from src relative to project root since we are in a monorepo tool context
      const promptPath = path.join(this.rootDir, 'packages/tools/agent/src/prompts.yml');
      
      if (existsSync(promptPath)) {
        const content = await fs.readFile(promptPath, 'utf-8');
        const parsed = yaml.load(content) as any;
        this.prompts = parsed.prompts || parsed; // Handle nesting if present
        console.log(chalk.gray(`Loaded prompts from ${promptPath}`));
      } else {
        console.warn(chalk.yellow(`prompts.yml not found at ${promptPath}, using defaults if available.`));
      }
    } catch (e: any) {
      console.error(chalk.red('Failed to load prompts.yml:'), e.message);
    }
  }

  async run(targetPatterns: string[]) {
    console.log(chalk.blue('GlassOps Agent starting...'));
    await this.loadCache();
    await this.loadPrompts();

    const files = await this.scanner.findFiles(targetPatterns);
    console.log(chalk.gray(`Found ${files.length} files to process.`));

    try {
      for (const file of files) {
        const ext = path.extname(file);
        // Find adapter that can handle this file
        // Priority: First match wins.
        const adapter = this.adapters.find(a => a.canHandle(file));

        if (!adapter) {
          continue; // Silent skip for non-matching files
        }

        // Hash check
        const content = await fs.readFile(file, 'utf-8');
        const hash = createHash('sha256').update(content).digest('hex');
        const relativePathRaw = path.relative(this.rootDir, file);
        const relativePath = relativePathRaw.split(path.sep).join('/');

        if (this.cache[relativePath]?.hash === hash) {
          console.log(chalk.gray(`Skipped (unchanged): ${relativePath}`));
          continue;
        }

        console.log(chalk.cyan(`Processing ${relativePath}...`));

        try {
          // Inject prompts if available
          // We'll need to cast or update adapter interface to accept prompts
          // For now, let's assume valid prompts exist
          const adapterKey = ext.replace('.', ''); // simple mapping: .ts -> ts, .md -> md
          const promptConfig = this.prompts[adapterKey] || this.prompts.default;

          // Pass relativePath to adapter so the LLM context sees the clean path
          const chunks = await adapter.parse(relativePath, content);
          const chunkOutputs: string[] = [];
          
          for (let i = 0; i < chunks.length; i++) {
            if (chunks.length > 1) console.log(chalk.gray(`  └─ Processing chunk ${i + 1}/${chunks.length}...`));
            
            // Generate Prompt
            // Using the promptConfig if available, otherwise using adapter default logic
            let prompt = '';
            if (promptConfig) {
               // Simple variable replacement
               prompt = promptConfig.system + '\n\n' + `File Path: ${relativePath}\n\n` + promptConfig.user.replace('{{content}}', chunks[i]);
            } else {
               prompt = adapter.generatePrompt(file, chunks[i]);
            }

            let output = await this.client.generateContent(prompt);
            
            // Clean up Markdown code blocks if the LLM wraps the entire response
            // Regex matches ```markdown ... ``` or just ``` ... ``` at the start/end
            // Also strips any aggressive whitespace padding
            output = output.replace(/^\s*```[a-z]*\s*/i, '').replace(/\s*```\s*$/, '').trim();
            
            chunkOutputs.push(output);
          }

          const processed = adapter.postProcess(file, chunkOutputs);

          
          const relativeDir = path.dirname(relativePath);
          const baseName = path.basename(file, ext);

          // Infer Metadata
          let type = 'Documentation';
          const normalizedPath = relativePath.split('/').map(p => p.toLowerCase());
          if (normalizedPath.includes('adr') || baseName.toLowerCase().startsWith('adr')) {
              type = 'ADR';
          }

          let domain = 'global';
          const parts = relativePath.split('/');
          let packageDocsRoot = path.join(this.rootDir, 'docs');
          let internalPath = ''; // Flatten by default per user request

          if (parts[0] === 'packages') {
              // Handle package-level federated docs
              // Structure: packages/<name> or packages/<group>/<name>
              // Groups: adapters, tools
              const groups = ['adapters', 'tools'];
              let packageRootParts = 0;

              if (groups.includes(parts[1]) && parts.length > 2) {
                  packageRootParts = 3; // packages/adapters/name
                  domain = parts[2];
              } else if (parts.length > 1) {
                  packageRootParts = 2; // packages/runtime
                  domain = parts[1];
              }

              if (packageRootParts > 0) {
                  const packageRoot = parts.slice(0, packageRootParts).join('/');
                  const fullPackageRoot = path.join(this.rootDir, packageRoot);
                  
                  // Verification: ensure the inferred package root is a directory
                  if (existsSync(fullPackageRoot)) {
                      const stats = await fs.stat(fullPackageRoot);
                      if (stats.isDirectory()) {
                          packageDocsRoot = path.join(fullPackageRoot, 'docs');
                      }
                  }
              }
          }
          
          const outDir = packageDocsRoot;
          await fs.mkdir(outDir, { recursive: true });
          
          // Append parent directory name to filename only if not in the same directory as the docs folder
          const sourceDirFull = path.resolve(this.rootDir, relativeDir);
          const docsParentFull = path.dirname(packageDocsRoot);
          
          let fileName = `${baseName}.md`;
          if (sourceDirFull !== docsParentFull) {
              const parentName = path.basename(sourceDirFull);
              if (parentName && parentName !== '.' && parentName !== baseName) {
                  fileName = `${parentName}-${baseName}.md`;
              }
          }
            
          const outPath = path.join(outDir, fileName);
          
          const today = new Date().toISOString().split('T')[0];

          // Strict Frontmatter Injection
          const frontmatter = `---
type: ${type}
domain: ${domain}
origin: ${relativePath}
last_modified: ${today}
generated: true
source: ${relativePath}
generated_at: ${new Date().toISOString()}
hash: ${hash}
---

`;
          const finalContent = frontmatter + processed;

          // Validation
          const validationErrors = Validator.validate(finalContent, outPath);
          if (validationErrors.length > 0) {
            console.warn(chalk.yellow(`Validation Warnings for ${baseName}:`));
            validationErrors.forEach(err => console.warn(chalk.yellow(`  - ${err}`)));
            // We currently log warnings but still write the file. 
            // In strict mode (CI), we might want to throw or delete.
          }

          await fs.writeFile(outPath, finalContent, 'utf-8');

          // Update Cache
          this.cache[relativePath] = {
            hash,
            generatedFiles: [path.relative(this.rootDir, outPath).split(path.sep).join('/')],
            timestamp: new Date().toISOString()
          };

          console.log(chalk.green(`Generated: ${path.relative(this.rootDir, outPath)}`));
        } catch (err: unknown) {
          const error = err as Error;
          console.error(chalk.red(`Error processing ${file}:`), error.message);
        }
      }
    } finally {
      await this.saveCache();
    }
  }
}
