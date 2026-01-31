export interface GlassOpsConfiguration {
    version: "1.0";
    metadata?: {
        last_updated?: string;
        schema_version?: string;
    };

    execution: {
        engine: "native" | "hardis";
        fallback?: "native" | "none";
    };

    governance: {
        minCoverage: number;
        requireTests: boolean;
        /**
         * Entity enforcing this policy.
         * Note: Philosophically, enforcement usually belongs to the protocol/policy engine,
         * but this allow explicit annotation in the config source.
         */
        enforcedBy?: string;
    };

    environments: {
        [key: string]: {
            display_name?: string;
            branch_mapping?: string;
            deployment_policy?: {
                test_level?: string;
                wait_time?: string;
                use_delta?: boolean;
                validation_required?: boolean;
                auto_deploy_on_merge?: boolean;
            };
            quality_gates?: {
                minCoverage?: number;
                security_severity_threshold?: number;
                block_on_test_failure?: boolean;
            };
            github_environment?: string;
            notes?: string;
        };
    };

    glassops?: {
        enablePlatformEvents?: boolean;
    };

    notifications?: {
        enabled_by_default?: boolean;
        channels?: {
            slack?: { enabled: boolean; mention_on_failure?: boolean };
            email?: { enabled: boolean };
        };
    };
    
    // Extensions
    [key: string]: any;
}
