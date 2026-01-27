import * as exec from "@actions/exec";
import * as core from "@actions/core";

export interface AnalyzerResult {
  violations: Violation[];
  exitCode: number;
}

export interface Violation {
  rule: string;
  description: string;
  severity: number;
  file: string;
  line: number;
}

export class Analyzer {
  /**
   * Runs the Salesforce Code Analyzer
   * @param paths Directories or files to scan
   * @param ruleset Optional ruleset to enforce
   */
  public async scan(
    paths: string[],
    ruleset?: string,
  ): Promise<AnalyzerResult> {
    await this.ensureCompatibility();

    const args = [
      "code-analyzer",
      "run",
      "--normalize-severity",
      "--output-format",
      "json",
    ];

    // Add paths
    // sf code-analyzer run --target "src,test"
    args.push("--target", paths.join(","));

    if (ruleset) {
      args.push("--ruleset", ruleset);
    }

    let stdout = "";
    let stderr = "";

    try {
      const exitCode = await exec.exec("sf", args, {
        listeners: {
          stdout: (data: Buffer) => {
            stdout += data.toString();
          },
          stderr: (data: Buffer) => {
            stderr += data.toString();
          },
        },
        silent: true,
        ignoreReturnCode: true, // Analyzer returns non-zero on violations
      });

      return this.parseOutput(stdout, exitCode);
    } catch (error) {
      core.error(
        `Analyzer execution failed: ${stderr || (error as Error).message}`,
      );
      throw error;
    }
  }

  /**
   * ENFORCE OPINIONATED POLICY:
   * We explicitly reject "sf scanner" command usage to force migration to code-analyzer.
   */
  public async ensureCompatibility(): Promise<void> {
    // In a real runtime, we might check if 'scanner' is installed and warn/fail.
    // For now, this is a placeholder for the "Opinionated Policy" check.
    // We assume the environment has 'sf code-analyzer' installed.
  }

  private parseOutput(jsonOutput: string, exitCode: number): AnalyzerResult {
    try {
      // Find the JSON array in stdout (it might have some clutter)
      const jsonStart = jsonOutput.indexOf("[");
      const jsonEnd = jsonOutput.lastIndexOf("]");

      if (jsonStart === -1 || jsonEnd === -1) {
        // If clean execution but no JSON, maybe no violations or empty
        return { violations: [], exitCode };
      }

      const cleanJson = jsonOutput.substring(jsonStart, jsonEnd + 1);
      const rawResults = JSON.parse(cleanJson);

      // Map raw analyzer output to simplified violations
      // Note: Actual schema depends on sf code-analyzer version, implies generic mapping here
      const violations: Violation[] = [];

      // Expected schema: Array of { engine: string, violations: [...] } or direct violations
      if (Array.isArray(rawResults)) {
        for (const fileResult of rawResults) {
          // Handle sfdx-scanner / code-analyzer output format
          // Often: { engine: "eslint", fileName: "...", violations: [...] }
          if (fileResult.violations) {
            for (const v of fileResult.violations) {
              violations.push({
                rule: v.ruleName,
                description: v.message,
                severity: v.severity, // 1 (high) to 3 (low) roughly
                file: fileResult.fileName,
                line: v.line,
              });
            }
          }
        }
      }

      return { violations, exitCode };
    } catch (e) {
      core.warning(`Failed to parse analyzer output: ${e}. Raw: ${jsonOutput}`);
      return { violations: [], exitCode };
    }
  }
}
