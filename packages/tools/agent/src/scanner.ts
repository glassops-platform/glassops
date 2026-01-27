import fg from 'fast-glob';
import path from 'path';
import fs from 'fs';
import ignore from 'ignore';

export class Scanner {
  private rootDir: string;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  private ig: any;

  constructor(rootDir: string) {
    this.rootDir = rootDir;
    this.ig = ignore();
    
    // Load .gitignore if it exists
    const gitignorePath = path.join(this.rootDir, '.gitignore');
    if (fs.existsSync(gitignorePath)) {
      const gitignoreContent = fs.readFileSync(gitignorePath, 'utf8');
      this.ig.add(gitignoreContent);
    }
    
    // Always ignore typical build/lock files
    this.ig.add(['node_modules/**', 'dist/**', 'package-lock.json', '.env', 'docs/generated/**', 'venv/**', '__pycache__/**']);
  }

  async findFiles(patterns: string[]): Promise<string[]> {
    const entries = await fg(patterns, {
      cwd: this.rootDir,
      dot: true
    });

    // Filter results using the ignore logic
    const filtered = entries.filter(file => !this.ig.ignores(file));

    return filtered.map(file => path.resolve(this.rootDir, file));
  }
}
