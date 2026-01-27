import { Command } from 'commander';
import { Generator } from './generator.js';
import { fileURLToPath } from 'url';
import path from 'path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const program = new Command();

program
  .name('glassops-agent')
  .description('AI-powered documentation and metadata agent for GlassOps')
  .version('1.0.0');

program
  .command('generate [patterns...]')
  .description('Generate documentation for the repository')
  .action(async (patterns) => {
    try {
      // Robust root detection: dist/index.js is 4 levels deep from repo root
      // dist -> agent -> tools -> packages -> ROOT
      const rootDir = path.resolve(__dirname, '../../../../');
      const generator = new Generator(rootDir);
      
      const activePatterns = patterns && patterns.length > 0 ? patterns : [
        'packages/**/*.ts',
        'packages/**/*.py',
        'packages/**/*.mjs',
        'packages/**/*.cls',
        'packages/**/*.trigger',
        'packages/**/*.js', // For LWC
        'packages/**/*.tf',
        'packages/**/Dockerfile',
        'docs/**/*.md',
        'packages/**/*.yml',
        'packages/**/*.yaml',
        'packages/**/*.json',
        'scripts/**/*.py',
        'scripts/**/*.ts',
        '*.py'
      ];

      await generator.run(activePatterns);
    } catch (err: unknown) {
      const error = err as Error;
      console.error('💥 Terminal Error:', error.message);
      if (error.stack) console.error(error.stack);
      process.exit(1);
    }
  });

program.parse();
