import * as core from "@actions/core";
import * as fs from "fs";
import * as path from "path";
import { z } from "zod";

const ConfigSchema = z.object({
  governance: z.object({
    enabled: z.boolean().default(true),
    freeze_windows: z
      .array(
        z.object({
          day: z.enum([
            "Monday",
            "Tuesday",
            "Wednesday",
            "Thursday",
            "Friday",
            "Saturday",
            "Sunday",
          ]),
          start: z.string().regex(/^\d{2}:\d{2}$/),
          end: z.string().regex(/^\d{2}:\d{2}$/),
        }),
      )
      .optional(),
    plugin_whitelist: z
      .array(z.string())
      .optional()
      .describe(
        "List of allowed Salesforce CLI plugins with optional version constraints (e.g., ['sfdx-hardis@^4.0.0', '@salesforce/plugin-deploy-retrieve'])",
      ),
    analyzer: z
      .object({
        enabled: z.boolean().default(false),
        severity_threshold: z.number().min(1).max(3).default(1),
        rulesets: z.array(z.string()).optional(),
        opinionated: z.boolean().default(true), // Enforce sf code-analyzer over sf scanner
      })
      .optional(),
  }),
  runtime: z.object({
    cli_version: z.string().default("latest"),
    node_version: z.string().default("20"),
  }),
});

export type ProtocolConfig = z.infer<typeof ConfigSchema>;

export class ProtocolPolicy {
  private configPath: string;

  constructor() {
    const configPathInput =
      process.env.GLASSOPS_CONFIG_PATH || "config/devops-config.json";
    this.configPath = path.join(
      process.env.GITHUB_WORKSPACE || ".",
      configPathInput,
    );
  }

  public async load(): Promise<ProtocolConfig> {
    if (!fs.existsSync(this.configPath)) {
      core.warning(
        "⚠️ No devops-config.json found. Using default unsafe policy.",
      );
      return ConfigSchema.parse({
        governance: { enabled: false },
        runtime: {},
      });
    }

    try {
      const raw = fs.readFileSync(this.configPath, "utf-8");
      const json = JSON.parse(raw);
      return ConfigSchema.parse(json);
    } catch (error) {
      throw new Error(
        `❌ Invalid Governance Policy: ${error instanceof Error ? error.message : String(error)}`,
      );
    }
  }

  public checkFreeze(config: ProtocolConfig): void {
    if (!config.governance.freeze_windows) return;

    const now = new Date();

    // Use UTC to ensure deterministic behavior across runners
    const days = [
      "Sunday",
      "Monday",
      "Tuesday",
      "Wednesday",
      "Thursday",
      "Friday",
      "Saturday",
    ];
    const currentDay = days[now.getUTCDay()];

    const currentTime = `${String(now.getUTCHours()).padStart(2, "0")}:${String(now.getUTCMinutes()).padStart(2, "0")}`;

    for (const window of config.governance.freeze_windows) {
      if (
        window.day === currentDay &&
        currentTime >= window.start &&
        currentTime <= window.end
      ) {
        throw new Error(
          `❄️ FROZEN: Deployment blocked by governance window (${window.day} ${window.start}-${window.end})`,
        );
      }
    }
  }

  public validatePluginWhitelist(
    config: ProtocolConfig,
    pluginName: string,
  ): boolean {
    if (
      !config.governance.plugin_whitelist ||
      config.governance.plugin_whitelist.length === 0
    ) {
      // If no whitelist is configured, allow all plugins (backward compatibility)
      return true;
    }

    // Check if plugin name matches any whitelist entry (with or without version constraints)
    return config.governance.plugin_whitelist.some((whitelistedPlugin) => {
      const extractedName = this.extractPluginName(whitelistedPlugin);
      return pluginName === extractedName;
    });
  }

  public getPluginVersionConstraint(
    config: ProtocolConfig,
    pluginName: string,
  ): string | null {
    if (
      !config.governance.plugin_whitelist ||
      config.governance.plugin_whitelist.length === 0
    ) {
      return null;
    }

    // Find the exact whitelist entry for this plugin
    const whitelistEntry = config.governance.plugin_whitelist.find(
      (whitelistedPlugin) => {
        // For scoped packages like @scope/package@version, extract @scope/package
        // For regular packages like package@version, extract package
        const extractedName = this.extractPluginName(whitelistedPlugin);
        return pluginName === extractedName;
      },
    );

    if (!whitelistEntry) {
      return null;
    }

    // Extract version constraint
    return this.extractVersionConstraint(whitelistEntry);
  }

  private extractPluginName(whitelistEntry: string): string {
    // Handle scoped packages: @scope/package@version -> @scope/package
    // Handle regular packages: package@version -> package
    if (whitelistEntry.startsWith("@")) {
      // Scoped package: find the last @ that's not at position 0
      const lastAtIndex = whitelistEntry.lastIndexOf("@");
      if (lastAtIndex > 0) {
        return whitelistEntry.substring(0, lastAtIndex);
      }
      return whitelistEntry; // No version, return as-is

    }
    
    // Regular package: split on @ and take first part
    const atIndex = whitelistEntry.indexOf("@");
    if (atIndex > 0) {
      return whitelistEntry.substring(0, atIndex);
    }
    return whitelistEntry; // No version, return as-is
  }

  private extractVersionConstraint(whitelistEntry: string): string | null {
    if (whitelistEntry.startsWith("@")) {
      // Scoped package: @scope/package@version -> version
      const lastAtIndex = whitelistEntry.lastIndexOf("@");
      if (lastAtIndex > 0) {
        return whitelistEntry.substring(lastAtIndex + 1);
      }
      return null;

    }
    
    // Regular package: package@version -> version
    const atIndex = whitelistEntry.indexOf("@");
    if (atIndex > 0) {
      return whitelistEntry.substring(atIndex + 1);
    }
    return null;
  }
}
