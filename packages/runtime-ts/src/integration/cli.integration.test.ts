/**
 * CLI Integration Tests
 *
 * These tests verify the RuntimeEnvironment class works correctly
 * with real module loading (especially dynamic imports) while only
 * mocking external I/O operations.
 */

import { RuntimeEnvironment } from "../services/cli";
import { ProtocolConfig } from "../protocol/policy";
import * as exec from "@actions/exec";
import * as io from "@actions/io";
import * as core from "@actions/core";

// Only mock external GitHub Actions modules for I/O
jest.mock("@actions/exec");
jest.mock("@actions/io");
jest.mock("@actions/core");

describe("CLI Integration Tests", () => {
  let runtime: RuntimeEnvironment;
  const mockedExec = exec.exec as jest.Mock;
  const mockedGetExecOutput = exec.getExecOutput as jest.Mock;
  const mockedWhich = io.which as jest.Mock;
  const mockedInfo = core.info as jest.Mock;
  const mockedWarning = core.warning as jest.Mock;
  const mockedError = core.error as jest.Mock;
  const mockedStartGroup = core.startGroup as jest.Mock;
  const mockedEndGroup = core.endGroup as jest.Mock;

  beforeEach(() => {
    runtime = new RuntimeEnvironment();
    jest.clearAllMocks();

    // Default mock implementations
    mockedExec.mockResolvedValue(0);
    mockedWhich.mockResolvedValue("");
  });

  describe("CLI Installation", () => {
    it("should skip installation when CLI is already present", async () => {
      mockedWhich.mockResolvedValue("/usr/local/bin/sf");

      await runtime.install();

      expect(mockedWhich).toHaveBeenCalledWith("sf", false);
      expect(mockedInfo).toHaveBeenCalledWith(
        "⚡ Salesforce CLI detected in environment. skipping install.",
      );
      expect(mockedExec).not.toHaveBeenCalled();
    });

    it("should install CLI when not present", async () => {
      mockedWhich.mockResolvedValue("");

      await runtime.install("2.30.0");

      expect(mockedExec).toHaveBeenCalledWith("npm", [
        "install",
        "-g",
        "@salesforce/cli@2.30.0",
      ]);
      expect(mockedExec).toHaveBeenCalledWith("sf", ["version"]);
    });

    it("should handle installation failure gracefully", async () => {
      mockedWhich.mockResolvedValue("");
      mockedExec.mockRejectedValue(new Error("npm install failed"));

      await expect(runtime.install()).rejects.toThrow(
        "Failed to bootstrap runtime",
      );
    });
  });

  describe("Plugin Installation with Whitelist", () => {
    const mockSuccessfulPluginVerification = () => {
      mockedGetExecOutput.mockResolvedValue({
        stdout: JSON.stringify({
          result: [
            { name: "sfdx-hardis", version: "4.5.0" },
            { name: "@salesforce/plugin-deploy-retrieve", version: "2.0.0" },
          ],
        }),
        stderr: "",
        exitCode: 0,
      });
    };

    beforeEach(() => {
      mockSuccessfulPluginVerification();
    });

    it("should install plugin without validation when no whitelist configured", async () => {
      // Setup mock for any-plugin verification
      mockedGetExecOutput.mockResolvedValue({
        stdout: JSON.stringify({
          result: [{ name: "any-plugin", version: "1.0.0" }],
        }),
        stderr: "",
        exitCode: 0,
      });

      const config: ProtocolConfig = {
        governance: { enabled: true },
        runtime: { cli_version: "latest", node_version: "20" },
      };

      await runtime.installPlugins(config, ["any-plugin"]);

      expect(mockedWarning).toHaveBeenCalledWith(
        "⚠️ No plugin whitelist configured. Installing any-plugin without validation.",
      );
      const expectedShell = process.platform === "win32" ? "cmd" : "sh";
      const expectedFlag = process.platform === "win32" ? "/c" : "-c";
      const expectedEchoPrefix =
        process.platform === "win32" ? "echo y|" : "echo y | ";

      expect(mockedExec).toHaveBeenCalledWith(expectedShell, [
        expectedFlag,
        expect.stringContaining(
          `${expectedEchoPrefix}sf "plugins" "install" "any-plugin"`,
        ),
      ]);
    });

    it("should install whitelisted plugin with version constraint", async () => {
      const config: ProtocolConfig = {
        governance: {
          enabled: true,
          plugin_whitelist: ["sfdx-hardis@^6.0.0"],
        },
        runtime: { cli_version: "latest", node_version: "20" },
      };

      await runtime.installPlugins(config, ["sfdx-hardis"]);

      // This is the key test - verify the version constraint is applied
      const expectedShell = process.platform === "win32" ? "cmd" : "sh";
      const expectedFlag = process.platform === "win32" ? "/c" : "-c";
      const expectedEchoPrefix =
        process.platform === "win32" ? "echo y|" : "echo y | ";

      expect(mockedExec).toHaveBeenCalledWith(expectedShell, [
        expectedFlag,
        expect.stringContaining(
          `${expectedEchoPrefix}sf "plugins" "install" "sfdx-hardis@^6.0.0"`,
        ),
      ]);
      expect(mockedInfo).toHaveBeenCalledWith(
        "⬇️ Installing plugin: sfdx-hardis@^6.0.0",
      );
    });

    it("should install scoped package with version constraint", async () => {
      const config: ProtocolConfig = {
        governance: {
          enabled: true,
          plugin_whitelist: ["@salesforce/plugin-deploy-retrieve@^3.0.0"],
        },
        runtime: { cli_version: "latest", node_version: "20" },
      };

      await runtime.installPlugins(config, [
        "@salesforce/plugin-deploy-retrieve",
      ]);

      const expectedShell = process.platform === "win32" ? "cmd" : "sh";
      const expectedFlag = process.platform === "win32" ? "/c" : "-c";
      const expectedEchoPrefix =
        process.platform === "win32" ? "echo y|" : "echo y | ";

      expect(mockedExec).toHaveBeenCalledWith(expectedShell, [
        expectedFlag,
        expect.stringContaining(
          `${expectedEchoPrefix}sf "plugins" "install" "@salesforce/plugin-deploy-retrieve@^3.0.0"`,
        ),
      ]);
    });

    it("should install whitelisted plugin without version constraint", async () => {
      const config: ProtocolConfig = {
        governance: {
          enabled: true,
          plugin_whitelist: ["sfdx-hardis"], // No version constraint
        },
        runtime: { cli_version: "latest", node_version: "20" },
      };

      await runtime.installPlugins(config, ["sfdx-hardis"]);

      const expectedShell = process.platform === "win32" ? "cmd" : "sh";
      const expectedFlag = process.platform === "win32" ? "/c" : "-c";
      const expectedEchoPrefix =
        process.platform === "win32" ? "echo y|" : "echo y | ";

      expect(mockedExec).toHaveBeenCalledWith(expectedShell, [
        expectedFlag,
        expect.stringContaining(
          `${expectedEchoPrefix}sf "plugins" "install" "sfdx-hardis"`,
        ),
      ]);
    });

    it("should reject non-whitelisted plugin", async () => {
      const config: ProtocolConfig = {
        governance: {
          enabled: true,
          plugin_whitelist: ["sfdx-hardis@^6.0.0"],
        },
        runtime: { cli_version: "latest", node_version: "20" },
      };

      await expect(
        runtime.installPlugins(config, ["malicious-plugin"]),
      ).rejects.toThrow("🚫 Plugin 'malicious-plugin' is not in the whitelist");

      // Verify no exec call was made for the malicious plugin
      expect(mockedExec).not.toHaveBeenCalledWith(
        "sf",
        expect.arrayContaining(["install", "malicious-plugin"]),
      );
    });

    it("should install multiple whitelisted plugins", async () => {
      const config: ProtocolConfig = {
        governance: {
          enabled: true,
          plugin_whitelist: [
            "sfdx-hardis@^6.0.0",
            "@salesforce/plugin-deploy-retrieve@latest",
          ],
        },
        runtime: { cli_version: "latest", node_version: "20" },
      };

      await runtime.installPlugins(config, [
        "sfdx-hardis",
        "@salesforce/plugin-deploy-retrieve",
      ]);

      const expectedShell = process.platform === "win32" ? "cmd" : "sh";
      const expectedFlag = process.platform === "win32" ? "/c" : "-c";
      const expectedEchoPrefix =
        process.platform === "win32" ? "echo y|" : "echo y | ";

      expect(mockedExec).toHaveBeenCalledWith(expectedShell, [
        expectedFlag,
        expect.stringContaining(
          `${expectedEchoPrefix}sf "plugins" "install" "sfdx-hardis@^6.0.0"`,
        ),
      ]);
      expect(mockedExec).toHaveBeenCalledWith(expectedShell, [
        expectedFlag,
        expect.stringContaining(
          `${expectedEchoPrefix}sf "plugins" "install" "@salesforce/plugin-deploy-retrieve@latest"`,
        ),
      ]);
    });

    it("should skip installation when no plugins specified", async () => {
      const config: ProtocolConfig = {
        governance: { enabled: true },
        runtime: { cli_version: "latest", node_version: "20" },
      };

      await runtime.installPlugins(config, []);

      expect(mockedInfo).toHaveBeenCalledWith(
        "ℹ️ No plugins specified for installation.",
      );
      expect(mockedStartGroup).not.toHaveBeenCalled();
    });

    it("should verify plugin installation", async () => {
      const config: ProtocolConfig = {
        governance: {
          enabled: true,
          plugin_whitelist: ["sfdx-hardis"],
        },
        runtime: { cli_version: "latest", node_version: "20" },
      };

      await runtime.installPlugins(config, ["sfdx-hardis"]);

      expect(mockedGetExecOutput).toHaveBeenCalledWith("sf", [
        "plugins",
        "--json",
      ]);
    });

    it("should fail when plugin verification fails", async () => {
      mockedGetExecOutput.mockResolvedValue({
        stdout: JSON.stringify({ result: [] }), // Empty - plugin not found
        stderr: "",
        exitCode: 0,
      });

      const config: ProtocolConfig = {
        governance: {
          enabled: true,
          plugin_whitelist: ["sfdx-hardis"],
        },
        runtime: { cli_version: "latest", node_version: "20" },
      };

      await expect(
        runtime.installPlugins(config, ["sfdx-hardis"]),
      ).rejects.toThrow(
        "Plugin 'sfdx-hardis' installation verification failed",
      );
    });

    it("should handle plugin installation error", async () => {
      mockedExec.mockRejectedValue(new Error("Network error"));

      const config: ProtocolConfig = {
        governance: {
          enabled: true,
          plugin_whitelist: ["sfdx-hardis"],
        },
        runtime: { cli_version: "latest", node_version: "20" },
      };

      await expect(
        runtime.installPlugins(config, ["sfdx-hardis"]),
      ).rejects.toThrow("Plugin installation failed: Network error");

      expect(mockedError).toHaveBeenCalledWith(
        "❌ Failed to install plugin 'sfdx-hardis': Network error",
      );
    });
  });

  describe("End-to-End Plugin Flow", () => {
    it("should complete full plugin installation workflow", async () => {
      // Setup mocks for full workflow
      mockedGetExecOutput.mockResolvedValue({
        stdout: JSON.stringify({
          result: [{ name: "sfdx-hardis", version: "4.5.0" }],
        }),
        stderr: "",
        exitCode: 0,
      });

      const config: ProtocolConfig = {
        governance: {
          enabled: true,
          plugin_whitelist: ["sfdx-hardis@^6.0.0"],
        },
        runtime: { cli_version: "latest", node_version: "20" },
      };

      await runtime.installPlugins(config, ["sfdx-hardis"]);

      // Verify the complete flow
      expect(mockedStartGroup).toHaveBeenCalledWith(
        "🔌 Installing Salesforce CLI Plugins",
      );
      expect(mockedInfo).toHaveBeenCalledWith(
        "🔍 Validating plugin: sfdx-hardis",
      );
      expect(mockedInfo).toHaveBeenCalledWith(
        "⬇️ Installing plugin: sfdx-hardis@^6.0.0",
      );
      const expectedShell = process.platform === "win32" ? "cmd" : "sh";
      const expectedFlag = process.platform === "win32" ? "/c" : "-c";
      const expectedEchoPrefix =
        process.platform === "win32" ? "echo y|" : "echo y | ";

      expect(mockedExec).toHaveBeenCalledWith(expectedShell, [
        expectedFlag,
        expect.stringContaining(
          `${expectedEchoPrefix}sf "plugins" "install" "sfdx-hardis@^6.0.0"`,
        ),
      ]);
      expect(mockedGetExecOutput).toHaveBeenCalledWith("sf", [
        "plugins",
        "--json",
      ]);
      expect(mockedInfo).toHaveBeenCalledWith(
        "✅ Plugin 'sfdx-hardis' installed and verified successfully",
      );
      expect(mockedEndGroup).toHaveBeenCalled();
    });
  });
});
