import { z } from "zod";

export const DeploymentContractSchema = z.object({
  schemaVersion: z.string().default("1.0"),
  meta: z.object({
    adapter: z.string(),
    engine: z.enum(["native", "hardis", "custom"]),
    timestamp: z.string().datetime(),
    trigger: z.string(),
  }),
  status: z.enum(["Succeeded", "Failed", "Blocked"]),
  quality: z.object({
    coverage: z.object({
      actual: z.number().min(0).max(100),
      required: z.number().min(0).max(100),
      met: z.boolean(),
    }),
    tests: z.object({
      total: z.number().min(0),
      passed: z.number().min(0),
      failed: z.number().min(0),
    }),
  }),
  audit: z.object({
    triggeredBy: z.string(),
    orgId: z.string(),
    repository: z.string(),
    commit: z.string(),
  }),
});

export type DeploymentContract = z.infer<typeof DeploymentContractSchema>;
