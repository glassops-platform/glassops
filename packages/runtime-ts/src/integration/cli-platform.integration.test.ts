/**
 * CLI Platform Integration Tests
 *
 * These tests verify platform-specific CLI operations work correctly
 * across Windows, macOS, and Linux environments.
 *
 * RTM Coverage:
 * - BR-009: Adapter Interface - Cross-platform CLI execution (TC-009-P01)
 * - BR-020: Plugin Installation across platforms (TC-020-P01)
 */

import { RuntimeEnvironment } from "../services/cli";
import { ProtocolConfig } from "../protocol/policy";
import * as exec from "@actions/exec";
import * as io from "@actions/io";

// Mock external GitHub Actions modules
jest.mock("@actions/exec");
jest.mock("@actions/core");
jest.mock("@actions/io");

describe("CLI Platform Integration Tests", () => {
  const mockedExec = exec.exec as jest.Mock;
  const mockedGetExecOutput = exec.getExecOutput as jest.Mock;
  const mockedWhich = io.which as jest.Mock;

  // Store original platform
  const originalPlatform = process.platform;

  beforeEach(() => {
    jest.clearAllMocks();

    // Default mock implementations
    mockedExec.mockResolvedValue(0);
    mockedWhich.mockResolvedValue("/usr/local/bin/sf");
    mockedGetExecOutput.mockResolvedValue({
      stdout: JSON.stringify({
        result: [{ name: "sfdx-hardis", version: "4.5.0" }],
      }),
      stderr: "",
      exitCode: 0,
    });
  });

  afterEach(() => {
    // Restore platform after each test
    Object.defineProperty(process, "platform", {
      value: originalPlatform,
      writable: true,
    });
  });

  afterAll(() => {
    // Final cleanup
    Object.defineProperty(process, "platform", {
      value: originalPlatform,
      writable: true,
    });
  });

  describe("Platform-Specific Plugin Installation", () => {
    it("should use Windows echo prefix on Windows platform", async () => {
      // Set platform BEFORE creating instance
      Object.defineProperty(process, "platform", {
        value: "win32",
        writable: true,
      });

      // Create instance AFTER platform is set
      const runtime = new RuntimeEnvironment();

      const config: ProtocolConfig = {
        governance: {
          enabled: true,
          plugin_whitelist: ["sfdx-hardis@^4.0.0"],
        },
        runtime: { cli_version: "latest", node_version: "20" },
      };

      await runtime.installPlugins(config, ["sfdx-hardis"]);

      expect(mockedExec).toHaveBeenCalledWith("cmd", [
        "/c",
        expect.stringContaining("echo y|"),
      ]);
    });

    it("should use Unix echo prefix on Linux platform", async () => {
      Object.defineProperty(process, "platform", {
        value: "linux",
        writable: true,
      });

      const runtime = new RuntimeEnvironment();

      const config: ProtocolConfig = {
        governance: {
          enabled: true,
          plugin_whitelist: ["sfdx-hardis@^4.0.0"],
        },
        runtime: { cli_version: "latest", node_version: "20" },
      };

      await runtime.installPlugins(config, ["sfdx-hardis"]);

      expect(mockedExec).toHaveBeenCalledWith("sh", [
        "-c",
        expect.stringContaining("echo y |"),
      ]);
    });

    it("should use Unix echo prefix on macOS platform", async () => {
      Object.defineProperty(process, "platform", {
        value: "darwin",
        writable: true,
      });

      const runtime = new RuntimeEnvironment();

      const config: ProtocolConfig = {
        governance: {
          enabled: true,
          plugin_whitelist: ["sfdx-hardis@^4.0.0"],
        },
        runtime: { cli_version: "latest", node_version: "20" },
      };

      await runtime.installPlugins(config, ["sfdx-hardis"]);

      expect(mockedExec).toHaveBeenCalledWith("sh", [
        "-c",
        expect.stringContaining("echo y |"),
      ]);
    });

    it("should handle plugin installation failure on Windows", async () => {
      Object.defineProperty(process, "platform", {
        value: "win32",
        writable: true,
      });

      const runtime = new RuntimeEnvironment();
      mockedExec.mockRejectedValue(new Error("Plugin installation failed"));

      const config: ProtocolConfig = {
        governance: {
          enabled: true,
          plugin_whitelist: ["sfdx-hardis"],
        },
        runtime: { cli_version: "latest", node_version: "20" },
      };

      await expect(
        runtime.installPlugins(config, ["sfdx-hardis"]),
      ).rejects.toThrow("Plugin installation failed");
    });

    it("should handle plugin installation failure on Unix", async () => {
      Object.defineProperty(process, "platform", {
        value: "linux",
        writable: true,
      });

      const runtime = new RuntimeEnvironment();
      mockedExec.mockRejectedValue(new Error("Plugin installation failed"));

      const config: ProtocolConfig = {
        governance: {
          enabled: true,
          plugin_whitelist: ["sfdx-hardis"],
        },
        runtime: { cli_version: "latest", node_version: "20" },
      };

      await expect(
        runtime.installPlugins(config, ["sfdx-hardis"]),
      ).rejects.toThrow("Plugin installation failed");
    });
  });

  describe("CLI Installation Across Platforms", () => {
    it("should install CLI with version when not present", async () => {
      const runtime = new RuntimeEnvironment();
      mockedWhich.mockResolvedValue(""); // CLI not present

      await runtime.install("2.30.0");

      expect(mockedExec).toHaveBeenCalledWith("npm", [
        "install",
        "-g",
        "@salesforce/cli@2.30.0",
      ]);
      expect(mockedExec).toHaveBeenCalledWith("sf", ["version"]);
    });

    it("should handle CLI installation failure", async () => {
      const runtime = new RuntimeEnvironment();
      mockedWhich.mockResolvedValue(""); // CLI not present
      mockedExec.mockRejectedValue(new Error("npm install failed"));

      await expect(runtime.install()).rejects.toThrow(
        "Failed to bootstrap runtime",
      );
    });

    it("should skip installation when CLI is already present", async () => {
      const runtime = new RuntimeEnvironment();
      mockedWhich.mockResolvedValue("/usr/local/bin/sf");

      await runtime.install();

      expect(mockedExec).not.toHaveBeenCalledWith(
        "npm",
        expect.arrayContaining(["install"]),
      );
    });
  });

  describe("Cross-Platform Command Execution", () => {
    it("should execute commands with correct shell on Windows", async () => {
      Object.defineProperty(process, "platform", {
        value: "win32",
        writable: true,
      });

      const runtime = new RuntimeEnvironment();

      // Setup mock for test-plugin verification
      mockedGetExecOutput.mockResolvedValue({
        stdout: JSON.stringify({
          result: [{ name: "test-plugin", version: "1.0.0" }],
        }),
        stderr: "",
        exitCode: 0,
      });

      const config: ProtocolConfig = {
        governance: {
          enabled: true,
          plugin_whitelist: ["test-plugin"],
        },
        runtime: { cli_version: "latest", node_version: "20" },
      };

      await runtime.installPlugins(config, ["test-plugin"]);

      expect(mockedExec).toHaveBeenCalledWith("cmd", expect.any(Array));
    });

    it("should execute commands with correct shell on Unix", async () => {
      Object.defineProperty(process, "platform", {
        value: "linux",
        writable: true,
      });

      const runtime = new RuntimeEnvironment();

      // Setup mock for test-plugin verification
      mockedGetExecOutput.mockResolvedValue({
        stdout: JSON.stringify({
          result: [{ name: "test-plugin", version: "1.0.0" }],
        }),
        stderr: "",
        exitCode: 0,
      });

      const config: ProtocolConfig = {
        governance: {
          enabled: true,
          plugin_whitelist: ["test-plugin"],
        },
        runtime: { cli_version: "latest", node_version: "20" },
      };

      await runtime.installPlugins(config, ["test-plugin"]);

      expect(mockedExec).toHaveBeenCalledWith("sh", expect.any(Array));
    });
  });
});
