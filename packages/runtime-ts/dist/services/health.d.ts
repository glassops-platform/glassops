export interface HealthCheckResult {
    healthy: boolean;
    version?: string;
    error?: string;
}
/**
 * Perform a health check by verifying the Salesforce CLI is available.
 */
export declare function healthCheck(): Promise<HealthCheckResult>;
//# sourceMappingURL=health.d.ts.map