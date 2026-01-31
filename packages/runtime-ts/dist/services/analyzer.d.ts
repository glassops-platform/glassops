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
export declare class Analyzer {
    /**
     * Runs the Salesforce Code Analyzer
     * @param paths Directories or files to scan
     * @param ruleset Optional ruleset to enforce
     */
    scan(paths: string[], ruleset?: string): Promise<AnalyzerResult>;
    /**
     * ENFORCE OPINIONATED POLICY:
     * We explicitly reject "sf scanner" command usage to force migration to code-analyzer.
     */
    ensureCompatibility(): Promise<void>;
    private parseOutput;
}
//# sourceMappingURL=analyzer.d.ts.map