/**
 * Policy Integration Tests
 *
 * These tests verify the ProtocolPolicy class works correctly
 * with real file system operations and configuration loading.
 *
 * RTM Coverage:
 * - BR-001: Additive Policy Merge (TC-001-P01 to TC-001-N04)
 * - BR-007: Pass/Fail Adjudication - Freeze Windows (TC-007-P02, TC-007-N01)
 * - BR-020: Plugin Whitelist Management (TC-020-P04 to TC-020-N07)
 */

import * as fs from "fs";
import * as path from "path";
import { ProtocolPolicy, ProtocolConfig } from "../protocol/policy";

// Mock only external GitHub Actions modules
jest.mock("@actions/core", () => ({
  warning: jest.fn(),
  info: jest.fn(),
  error: jest.fn(),
  startGroup: jest.fn(),
  endGroup: jest.fn(),
}));

describe("Policy Integration Tests", () => {
  const originalWorkspace = process.env.GITHUB_WORKSPACE;
  const testDir = path.join(__dirname, "test-workspace");
  const configPath = path.join(testDir, "devops-config.json");

  beforeAll(() => {
    // Create test workspace directory
    if (!fs.existsSync(testDir)) {
      fs.mkdirSync(testDir, { recursive: true });
    }
  });

  beforeEach(() => {
    // Point GITHUB_WORKSPACE to test directory
    process.env.GITHUB_WORKSPACE = testDir;

    // Clean up any existing config
    if (fs.existsSync(configPath)) {
      fs.unlinkSync(configPath);
    }
  });

  afterEach(() => {
    // Restore original workspace
    process.env.GITHUB_WORKSPACE = originalWorkspace;
  });

  afterAll(() => {
    // Clean up test workspace
    if (fs.existsSync(testDir)) {
      fs.rmSync(testDir, { recursive: true, force: true });
    }
  });

  // RTM: TC-001-P01, TC-001-P02, TC-001-N04
  describe("Configuration Loading", () => {
    // RTM: TC-001-P01 - Load and merge policy from file
    it("should load configuration from real file", async () => {
      const config = {
        governance: {
          enabled: true,
          freeze_windows: [{ day: "Saturday", start: "00:00", end: "23:59" }],
          plugin_whitelist: [
            "sfdx-hardis@^4.0.0",
            "@salesforce/plugin-deploy-retrieve",
          ],
        },
        runtime: {
          cli_version: "2.30.0",
          node_version: "20",
        },
      };

      fs.writeFileSync(configPath, JSON.stringify(config, null, 2));

      const policy = new ProtocolPolicy();
      const loadedConfig = await policy.load();

      expect(loadedConfig.governance.enabled).toBe(true);
      expect(loadedConfig.governance.freeze_windows).toHaveLength(1);
      expect(loadedConfig.governance.plugin_whitelist).toHaveLength(2);
      expect(loadedConfig.runtime.cli_version).toBe("2.30.0");
    });

    it("should return default config when file is missing", async () => {
      const policy = new ProtocolPolicy();
      const loadedConfig = await policy.load();

      expect(loadedConfig.governance.enabled).toBe(false);
      expect(loadedConfig.runtime.cli_version).toBe("latest");
    });

    it("should throw error for invalid JSON", async () => {
      fs.writeFileSync(configPath, "{ invalid json }");

      const policy = new ProtocolPolicy();
      await expect(policy.load()).rejects.toThrow("Invalid Governance Policy");
    });

    it("should throw error for invalid schema", async () => {
      const invalidConfig = {
        governance: {
          enabled: "not-a-boolean", // Should be boolean
        },
        runtime: {},
      };

      fs.writeFileSync(configPath, JSON.stringify(invalidConfig));

      const policy = new ProtocolPolicy();
      await expect(policy.load()).rejects.toThrow("Invalid Governance Policy");
    });
  });

  describe("Plugin Whitelist Validation", () => {
    let policy: ProtocolPolicy;

    beforeEach(() => {
      policy = new ProtocolPolicy();
    });

    it("should validate plugin against whitelist from loaded config", async () => {
      const config = {
        governance: {
          enabled: true,
          plugin_whitelist: [
            "sfdx-hardis@^4.0.0",
            "@salesforce/plugin-deploy-retrieve@latest",
            "sf-metadata-scanner",
          ],
        },
        runtime: { cli_version: "latest" },
      };

      fs.writeFileSync(configPath, JSON.stringify(config));

      const loadedConfig = await policy.load();

      // Test whitelisted plugins
      expect(policy.validatePluginWhitelist(loadedConfig, "sfdx-hardis")).toBe(
        true,
      );
      expect(
        policy.validatePluginWhitelist(
          loadedConfig,
          "@salesforce/plugin-deploy-retrieve",
        ),
      ).toBe(true);
      expect(
        policy.validatePluginWhitelist(loadedConfig, "sf-metadata-scanner"),
      ).toBe(true);

      // Test non-whitelisted plugins
      expect(
        policy.validatePluginWhitelist(loadedConfig, "malicious-plugin"),
      ).toBe(false);
      expect(
        policy.validatePluginWhitelist(loadedConfig, "unknown-plugin"),
      ).toBe(false);
    });

    it("should extract version constraints from whitelist", async () => {
      const config = {
        governance: {
          enabled: true,
          plugin_whitelist: [
            "sfdx-hardis@^4.0.0",
            "@salesforce/plugin-deploy-retrieve@latest",
            "sf-metadata-scanner", // No version constraint
          ],
        },
        runtime: { cli_version: "latest" },
      };

      fs.writeFileSync(configPath, JSON.stringify(config));

      const loadedConfig = await policy.load();

      // Plugins with version constraints
      expect(
        policy.getPluginVersionConstraint(loadedConfig, "sfdx-hardis"),
      ).toBe("^4.0.0");
      expect(
        policy.getPluginVersionConstraint(
          loadedConfig,
          "@salesforce/plugin-deploy-retrieve",
        ),
      ).toBe("latest");

      // Plugin without version constraint
      expect(
        policy.getPluginVersionConstraint(loadedConfig, "sf-metadata-scanner"),
      ).toBeNull();

      // Unknown plugin
      expect(
        policy.getPluginVersionConstraint(loadedConfig, "unknown"),
      ).toBeNull();
    });

    it("should allow all plugins when no whitelist is configured", async () => {
      const config: ProtocolConfig = {
        governance: { enabled: true },
        runtime: { cli_version: "latest", node_version: "20" },
      };

      expect(policy.validatePluginWhitelist(config, "any-plugin")).toBe(true);
      expect(
        policy.getPluginVersionConstraint(config, "any-plugin"),
      ).toBeNull();
    });

    it("should handle scoped packages without version constraints", async () => {
      const config: ProtocolConfig = {
        governance: {
          enabled: true,
          plugin_whitelist: ["@salesforce/plugin-deploy-retrieve"], // No version
        },
        runtime: { cli_version: "latest", node_version: "20" },
      };

      // Should validate the scoped package (line 162)
      expect(
        policy.validatePluginWhitelist(
          config,
          "@salesforce/plugin-deploy-retrieve",
        ),
      ).toBe(true);

      // Should return null for version constraint (line 180)
      expect(
        policy.getPluginVersionConstraint(
          config,
          "@salesforce/plugin-deploy-retrieve",
        ),
      ).toBeNull();
    });
  });

  describe("Freeze Window Validation", () => {
    let policy: ProtocolPolicy;

    beforeEach(() => {
      policy = new ProtocolPolicy();
    });

    it("should block deployment during configured freeze window", async () => {
      // Create a config with the current day as a freeze window
      const now = new Date();
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

      const config: ProtocolConfig = {
        governance: {
          enabled: true,
          freeze_windows: [
            {
              day: currentDay as
                | "Monday"
                | "Tuesday"
                | "Wednesday"
                | "Thursday"
                | "Friday"
                | "Saturday"
                | "Sunday",
              start: "00:00",
              end: "23:59",
            },
          ],
        },
        runtime: { cli_version: "latest", node_version: "20" },
      };

      expect(() => policy.checkFreeze(config)).toThrow("FROZEN");
    });

    it("should allow deployment outside freeze windows", () => {
      // Create a config with a freeze window NOT including today
      const now = new Date();
      const days = [
        "Sunday",
        "Monday",
        "Tuesday",
        "Wednesday",
        "Thursday",
        "Friday",
        "Saturday",
      ];
      const currentDayIndex = now.getUTCDay();
      const otherDayIndex = (currentDayIndex + 3) % 7; // Pick a different day
      const otherDay = days[otherDayIndex];

      const config: ProtocolConfig = {
        governance: {
          enabled: true,
          freeze_windows: [
            {
              day: otherDay as
                | "Monday"
                | "Tuesday"
                | "Wednesday"
                | "Thursday"
                | "Friday"
                | "Saturday"
                | "Sunday",
              start: "00:00",
              end: "23:59",
            },
          ],
        },
        runtime: { cli_version: "latest", node_version: "20" },
      };

      expect(() => policy.checkFreeze(config)).not.toThrow();
    });

    it("should pass when no freeze windows are configured", () => {
      const config: ProtocolConfig = {
        governance: { enabled: true },
        runtime: { cli_version: "latest", node_version: "20" },
      };

      expect(() => policy.checkFreeze(config)).not.toThrow();
    });
  });
});
