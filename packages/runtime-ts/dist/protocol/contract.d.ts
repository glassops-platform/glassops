import { z } from "zod";
export declare const DeploymentContractSchema: z.ZodObject<{
    schemaVersion: z.ZodDefault<z.ZodString>;
    meta: z.ZodObject<{
        adapter: z.ZodString;
        engine: z.ZodEnum<{
            custom: "custom";
            native: "native";
            hardis: "hardis";
        }>;
        timestamp: z.ZodString;
        trigger: z.ZodString;
    }, z.core.$strip>;
    status: z.ZodEnum<{
        Succeeded: "Succeeded";
        Failed: "Failed";
        Blocked: "Blocked";
    }>;
    quality: z.ZodObject<{
        coverage: z.ZodObject<{
            actual: z.ZodNumber;
            required: z.ZodNumber;
            met: z.ZodBoolean;
        }, z.core.$strip>;
        tests: z.ZodObject<{
            total: z.ZodNumber;
            passed: z.ZodNumber;
            failed: z.ZodNumber;
        }, z.core.$strip>;
    }, z.core.$strip>;
    audit: z.ZodObject<{
        triggeredBy: z.ZodString;
        orgId: z.ZodString;
        repository: z.ZodString;
        commit: z.ZodString;
    }, z.core.$strip>;
}, z.core.$strip>;
export type DeploymentContract = z.infer<typeof DeploymentContractSchema>;
//# sourceMappingURL=contract.d.ts.map