import { z } from "zod";
declare const ConfigSchema: z.ZodObject<{
    governance: z.ZodObject<{
        enabled: z.ZodDefault<z.ZodBoolean>;
        freeze_windows: z.ZodOptional<z.ZodArray<z.ZodObject<{
            day: z.ZodEnum<{
                Monday: "Monday";
                Tuesday: "Tuesday";
                Wednesday: "Wednesday";
                Thursday: "Thursday";
                Friday: "Friday";
                Saturday: "Saturday";
                Sunday: "Sunday";
            }>;
            start: z.ZodString;
            end: z.ZodString;
        }, z.core.$strip>>>;
        plugin_whitelist: z.ZodOptional<z.ZodArray<z.ZodString>>;
        analyzer: z.ZodOptional<z.ZodObject<{
            enabled: z.ZodDefault<z.ZodBoolean>;
            severity_threshold: z.ZodDefault<z.ZodNumber>;
            rulesets: z.ZodOptional<z.ZodArray<z.ZodString>>;
            opinionated: z.ZodDefault<z.ZodBoolean>;
        }, z.core.$strip>>;
    }, z.core.$strip>;
    runtime: z.ZodObject<{
        cli_version: z.ZodDefault<z.ZodString>;
        node_version: z.ZodDefault<z.ZodString>;
    }, z.core.$strip>;
}, z.core.$strip>;
export type ProtocolConfig = z.infer<typeof ConfigSchema>;
export declare class ProtocolPolicy {
    private configPath;
    constructor();
    load(): Promise<ProtocolConfig>;
    checkFreeze(config: ProtocolConfig): void;
    validatePluginWhitelist(config: ProtocolConfig, pluginName: string): boolean;
    getPluginVersionConstraint(config: ProtocolConfig, pluginName: string): string | null;
    private extractPluginName;
    private extractVersionConstraint;
}
export {};
//# sourceMappingURL=policy.d.ts.map