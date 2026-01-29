import { ProtocolPolicy, ProtocolConfig } from "./policy";
import * as core from "@actions/core";
import * as fs from "fs";

// Partial Mock: Keeps fs.promises intact for @actions/core
jest.mock("fs", () => ({
  ...jest.requireActual("fs"),
  existsSync: jest.fn(),
  readFileSync: jest.fn(),
}));

jest.mock("@actions/core");

describe("ProtocolPolicy", () => {
  let policy: ProtocolPolicy;
  const mockedExistsSync = fs.existsSync as jest.Mock;
  const mockedReadFileSync = fs.readFileSync as jest.Mock;

  beforeEach(() => {
    policy = new ProtocolPolicy();
    jest.clearAllMocks();
  });

  describe("checkFreeze", () => {
    const mockConfig = {
      governance: {
        enabled: true,
        freeze_windows: [
          {
            day: "Friday",
            start: "09:00",
            end: "17:00",
          },
        ],
      },
      runtime: { cli_version: "latest", node_version: "20" },
    };

    it("should throw an error if the current time is within a freeze window", () => {
      // Jan 23, 2026 is a Friday
      // 10:00:00Z is within the 09:00-17:00 window
      const mockDate = new Date("2026-01-23T10:00:00Z");
      jest.useFakeTimers().setSystemTime(mockDate);

      expect(() => policy.checkFreeze(mockConfig as ProtocolConfig)).toThrow(
        /FROZEN/,
      );
    });

    it("should allow deployment if the current time is outside freeze windows", () => {
      // Jan 23, 2026 at 18:00:00Z (6:00 PM) is outside the 5:00 PM cutoff
      const mockDate = new Date("2026-01-23T18:00:00Z");
      jest.useFakeTimers().setSystemTime(mockDate);

      expect(() =>
        policy.checkFreeze(mockConfig as ProtocolConfig),
      ).not.toThrow();
    });

    it("should do nothing when freeze_windows is undefined", () => {
      const configWithoutWindows = {
        governance: {
          enabled: true,
          // No freeze_windows defined
        },
        runtime: { cli_version: "latest", node_version: "20" },
      };

      expect(() =>
        policy.checkFreeze(configWithoutWindows as ProtocolConfig),
      ).not.toThrow();
    });
  });

  describe("validatePluginWhitelist", () => {
    it("should return true when no whitelist is configured", () => {
      const configWithoutWhitelist = {
        governance: {
          enabled: true,
          // No plugin_whitelist defined
        },
        runtime: { cli_version: "latest", node_version: "20" },
      };

      const result = policy.validatePluginWhitelist(
        configWithoutWhitelist as ProtocolConfig,
        "any-plugin",
      );
      expect(result).toBe(true);
    });

    it("should return true when plugin is in whitelist", () => {
      const configWithWhitelist = {
        governance: {
          enabled: true,
          plugin_whitelist: [
            "sfdx-hardis",
            "@salesforce/plugin-deploy-retrieve",
          ],
        },
        runtime: { cli_version: "latest", node_version: "20" },
      };

      const result = policy.validatePluginWhitelist(
        configWithWhitelist as ProtocolConfig,
        "sfdx-hardis",
      );
      expect(result).toBe(true);
    });

    it("should return true when plugin with version constraint is in whitelist", () => {
      const configWithVersionConstraint = {
        governance: {
          enabled: true,
          plugin_whitelist: [
            "sfdx-hardis@^6.0.0",
            "@salesforce/plugin-deploy-retrieve@latest",
          ],
        },
        runtime: { cli_version: "latest", node_version: "20" },
      };

      const result = policy.validatePluginWhitelist(
        configWithVersionConstraint as ProtocolConfig,
        "sfdx-hardis",
      );
      expect(result).toBe(true);
    });

    it("should return false when plugin is not in whitelist", () => {
      const configWithWhitelist = {
        governance: {
          enabled: true,
          plugin_whitelist: [
            "sfdx-hardis",
            "@salesforce/plugin-deploy-retrieve",
          ],
        },
        runtime: { cli_version: "latest", node_version: "20" },
      };

      const result = policy.validatePluginWhitelist(
        configWithWhitelist as ProtocolConfig,
        "malicious-plugin",
      );
      expect(result).toBe(false);
    });
  });

  describe("getPluginVersionConstraint", () => {
    it("should return null when no whitelist is configured", () => {
      const configWithoutWhitelist = {
        governance: {
          enabled: true,
          // No plugin_whitelist defined
        },
        runtime: { cli_version: "latest", node_version: "20" },
      };

      const result = policy.getPluginVersionConstraint(
        configWithoutWhitelist as ProtocolConfig,
        "any-plugin",
      );
      expect(result).toBeNull();
    });

    it("should return null when plugin is not in whitelist", () => {
      const configWithWhitelist = {
        governance: {
          enabled: true,
          plugin_whitelist: [
            "sfdx-hardis",
            "@salesforce/plugin-deploy-retrieve",
          ],
        },
        runtime: { cli_version: "latest", node_version: "20" },
      };

      const result = policy.getPluginVersionConstraint(
        configWithWhitelist as ProtocolConfig,
        "unknown-plugin",
      );
      expect(result).toBeNull();
    });

    it("should return version constraint when specified", () => {
      const configWithVersionConstraint = {
        governance: {
          enabled: true,
          plugin_whitelist: [
            "sfdx-hardis@^6.0.0",
            "@salesforce/plugin-deploy-retrieve@latest",
          ],
        },
        runtime: { cli_version: "latest", node_version: "20" },
      };

      const result = policy.getPluginVersionConstraint(
        configWithVersionConstraint as ProtocolConfig,
        "sfdx-hardis",
      );
      expect(result).toBe("^6.0.0");
    });

    it("should return null when plugin has no version constraint", () => {
      const configMixedConstraints = {
        governance: {
          enabled: true,
          plugin_whitelist: [
            "sfdx-hardis",
            "@salesforce/plugin-deploy-retrieve@latest",
          ],
        },
        runtime: { cli_version: "latest", node_version: "20" },
      };

      const result = policy.getPluginVersionConstraint(
        configMixedConstraints as ProtocolConfig,
        "sfdx-hardis",
      );
      expect(result).toBeNull();
    });

    it("should return version constraint for scoped packages", () => {
      const configScopedPackage = {
        governance: {
          enabled: true,
          plugin_whitelist: ["@salesforce/plugin-deploy-retrieve@^1.0.0"],
        },
        runtime: { cli_version: "latest", node_version: "20" },
      };

      const result = policy.getPluginVersionConstraint(
        configScopedPackage as ProtocolConfig,
        "@salesforce/plugin-deploy-retrieve",
      );
      expect(result).toBe("^1.0.0");
    });

    it("should return null for scoped package without version constraint", () => {
      const configScopedNoVersion = {
        governance: {
          enabled: true,
          plugin_whitelist: ["@salesforce/plugin-deploy-retrieve"], // No version
        },
        runtime: { cli_version: "latest", node_version: "20" },
      };

      const result = policy.getPluginVersionConstraint(
        configScopedNoVersion as ProtocolConfig,
        "@salesforce/plugin-deploy-retrieve",
      );
      expect(result).toBeNull();
    });
  });

  describe("load", () => {
    it("should return a default unsafe policy if devops-config.json is missing", async () => {
      mockedExistsSync.mockReturnValue(false);

      const config = await policy.load();

      expect(config.governance.enabled).toBe(false);
      expect(core.warning).toHaveBeenCalledWith(
        expect.stringContaining("Using default unsafe policy"),
      );
    });

    it("should successfully load and parse valid devops-config.json", async () => {
      const validConfig = {
        governance: {
          enabled: true,
          freeze_windows: [
            {
              day: "Friday",
              start: "17:00",
              end: "23:59",
            },
          ],
        },
        runtime: {
          cli_version: "latest",
          node_version: "20",
        },
      };

      mockedExistsSync.mockReturnValue(true);
      mockedReadFileSync.mockReturnValue(JSON.stringify(validConfig));

      const config = await policy.load();

      expect(config.governance.enabled).toBe(true);
      expect(config.governance.freeze_windows).toHaveLength(1);
      expect(config.runtime.cli_version).toBe("latest");
    });

    it("should throw an error for invalid JSON in devops-config.json", async () => {
      mockedExistsSync.mockReturnValue(true);
      mockedReadFileSync.mockReturnValue("invalid json");

      await expect(policy.load()).rejects.toThrow("Invalid Governance Policy");
    });

    it("should throw an error for invalid schema in devops-config.json", async () => {
      const invalidConfig = {
        governance: {
          enabled: "not-a-boolean", // Invalid type
        },
      };

      mockedExistsSync.mockReturnValue(true);
      mockedReadFileSync.mockReturnValue(JSON.stringify(invalidConfig));

      await expect(policy.load()).rejects.toThrow("Invalid Governance Policy");
    });

    it("should handle non-Error exceptions when reading config", async () => {
      mockedExistsSync.mockReturnValue(true);
      mockedReadFileSync.mockImplementation(() => {
        throw "string error"; // Non-Error throw to cover String(error) branch
      });

      await expect(policy.load()).rejects.toThrow("Invalid Governance Policy");
    });
  });
});
