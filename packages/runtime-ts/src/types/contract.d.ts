export interface DeploymentContract {
    schemaVersion: string;

    meta: {
        adapter: string;
        engine: string;
        engineVersion: string;
        timestamp: string;
        trigger: string;
    };

    status: "Succeeded" | "Failed" | "Canceled" | "Partial";

    deployment: {
        id: string;
        url?: string;
        mode: "validate" | "deploy" | "quick_deploy";
        validationId?: string;
        metrics: {
            componentsDeployed: number;
            componentsFailed: number;
            testsRun: number;
            durationMs: number;
        };
    };

    quality: {
        coverage: {
            actual: number;
            required: number;
            met: boolean;
            details?: {
                orgWideCoverage: number;
                packageCoverage?: number;
            };
        };

        /** 
         * Architectural Validity Results (Phase 1.5)
         * Enforces invariants (e.g. "One Trigger Per Object")
         */
        staticAnalysis?: {
            toolsUsed: string[];        // e.g. ["pmd", "glassops-fs-scanner"]
            score: number;              // Informational health score
            criticalViolations: number; // The Enforcer (e.g. SOQL in loop)
            warningViolations: number;  // The Technical Debt
            
            // Optional: Top offenders for the audit trail
            topIssues?: Array<{
                file: string;
                rule: string;             // e.g. "MultipleTriggersPerObject"
                severity: "Critical" | "Warning";
                line: number;
                message: string;
            }>;
        };
    };

    policy: {
        source: {
            githubFloor: number;
            configFile?: number;
            salesforceCmdt?: number;
        };
        effective: number;
        overrides?: Array<{
            reason: string;
            approver: string;
        }>;
    };

    audit: {
        triggeredBy: string;
        repository: string;
        ref: string;
        commit: string;
        runUrl: string;
    };

    errors?: Array<{
        code: string;
        message: string;
        severity: "Critical" | "Warning";
        component?: string;
    }>;

    extensions?: {
        [key: string]: any;
    };
}