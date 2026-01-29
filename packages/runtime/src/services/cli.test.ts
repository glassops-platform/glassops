import { RuntimeEnvironment } from "./cli";
import * as exec from "@actions/exec";
import * as io from "@actions/io";
import * as core from "@actions/core";
import { ProtocolConfig } from "../protocol/policy";

jest.mock("@actions/exec", () => ({
  exec: jest.fn(),
  getExecOutput: jest.fn(),
}));
jest.mock("@actions/io");
jest.mock("@actions/core");

describe("RuntimeEnvironment", () => {
  let runtime: RuntimeEnvironment;
  const mockedExec = exec.exec as jest.Mock;
  const mockedWhich = io.which as jest.Mock;
  const mockedInfo = core.info as jest.Mock;

  beforeEach(() => {
    runtime = new RuntimeEnvironment();
    jest.clearAllMocks();
  });

  describe("install", () => {
    it("should skip install if sf CLI is already available", async () => {
      mockedWhich.mockResolvedValue("/usr/local/bin/sf");

      await runtime.install("latest");

      expect(mockedWhich).toHaveBeenCalledWith("sf", false);
      expect(mockedExec).not.toHaveBeenCalled();
      expect(mockedInfo).toHaveBeenCalledWith(
        "⚡ Salesforce CLI detected in environment. skipping install.",
      );
    });

    it("should install sf CLI if not found", async () => {
      mockedWhich.mockResolvedValue(null);
      mockedExec.mockResolvedValue(0);

      await runtime.install("latest");

      expect(mockedWhich).toHaveBeenCalledWith("sf", false);
      expect(mockedExec).toHaveBeenCalledWith("npm", [
        "install",
        "-g",
        "@salesforce/cli@latest",
      ]);
      expect(mockedExec).toHaveBeenCalledWith("sf", ["version"]);
    });

    it("should install specific version when provided", async () => {
      mockedWhich.mockResolvedValue(null);
      mockedExec.mockResolvedValue(0);

      await runtime.install("2.50.0");

      expect(mockedExec).toHaveBeenCalledWith("npm", [
        "install",
        "-g",
        "@salesforce/cli@2.50.0",
      ]);
    });

    it("should install latest version when no version specified", async () => {
      mockedWhich.mockResolvedValue(null);
      mockedExec.mockResolvedValue(0);

      await runtime.install();

      expect(mockedExec).toHaveBeenCalledWith("npm", [
        "install",
        "-g",
        "@salesforce/cli@latest",
      ]);
    });

    it("should throw error on npm install failure", async () => {
      mockedWhich.mockResolvedValue(null);
      mockedExec.mockRejectedValue(new Error("npm install failed"));

      await expect(runtime.install("latest")).rejects.toThrow(
        "Failed to bootstrap runtime. NPM registry might be down.",
      );
    });

    it("should throw error on sf version check failure", async () => {
      mockedWhich.mockResolvedValue(null);
      mockedExec
        .mockResolvedValueOnce(0) // npm install succeeds
        .mockRejectedValueOnce(new Error("sf version failed")); // sf version fails

      await expect(runtime.install("latest")).rejects.toThrow(
        "Failed to bootstrap runtime. NPM registry might be down.",
      );
    });

    it("should call startGroup and endGroup", async () => {
      mockedWhich.mockResolvedValue(null);
      mockedExec.mockResolvedValue(0);
      const mockedStartGroup = core.startGroup as jest.Mock;
      const mockedEndGroup = core.endGroup as jest.Mock;

      await runtime.install("latest");

      expect(mockedStartGroup).toHaveBeenCalledWith(
        "🔧 Bootstrapping GlassOps Runtime",
      );
      expect(mockedEndGroup).toHaveBeenCalled();
    });
  });

  describe("installPlugins", () => {
    beforeEach(() => {
      jest.dontMock("../protocol/policy");
    });

    const mockConfig: ProtocolConfig = {
      governance: {
        enabled: true,
        plugin_whitelist: ["sfdx-hardis", "@salesforce/plugin-deploy-retrieve"],
      },
      runtime: {
        cli_version: "latest",
        node_version: "20",
      },
    };

    // Remove the beforeEach mock setup - each test will set up its own mocks

    it("should skip installation when no plugins specified", async () => {
      await runtime.installPlugins(mockConfig, []);

      expect(mockedInfo).toHaveBeenCalledWith(
        "ℹ️ No plugins specified for installation.",
      );
      expect(mockedExec).not.toHaveBeenCalled();
    });

    it("should install whitelisted plugins successfully", async () => {
      const mockedStartGroup = core.startGroup as jest.Mock;
      const mockedEndGroup = core.endGroup as jest.Mock;
      const mockedGetExecOutput = exec.getExecOutput as jest.Mock;
      mockedGetExecOutput.mockResolvedValue({
        stdout: JSON.stringify({
          result: [{ name: "sfdx-hardis", version: "1.0.0" }],
        }),
        stderr: "",
        exitCode: 0,
      });

      await runtime.installPlugins(mockConfig, ["sfdx-hardis"]);

      expect(mockedStartGroup).toHaveBeenCalledWith(
        "🔌 Installing Salesforce CLI Plugins",
      );
      expect(mockedInfo).toHaveBeenCalledWith(
        "🔍 Validating plugin: sfdx-hardis",
      );
      expect(mockedInfo).toHaveBeenCalledWith(
        "⬇️ Installing plugin: sfdx-hardis",
      );
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
      expect(mockedEndGroup).toHaveBeenCalled();
    });

    it("should reject non-whitelisted plugins", async () => {
      await expect(
        runtime.installPlugins(mockConfig, ["malicious-plugin"]),
      ).rejects.toThrow("🚫 Plugin 'malicious-plugin' is not in the whitelist");

      expect(mockedExec).not.toHaveBeenCalled();
    });

    it("should allow all plugins when no whitelist is configured", async () => {
      const configWithoutWhitelist: ProtocolConfig = {
        ...mockConfig,
        governance: {
          ...mockConfig.governance,
          plugin_whitelist: undefined,
        },
      };

      const mockedWarning = core.warning as jest.Mock;
      const mockedGetExecOutput = exec.getExecOutput as jest.Mock;
      mockedGetExecOutput.mockResolvedValue({
        stdout: JSON.stringify({
          result: [
            { name: "any-plugin", version: "1.0.0" }, // Mock successful verification
          ],
        }),
        stderr: "",
        exitCode: 0,
      });

      await runtime.installPlugins(configWithoutWhitelist, ["any-plugin"]);

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

    it("should install multiple plugins", async () => {
      const mockedGetExecOutput = exec.getExecOutput as jest.Mock;
      mockedGetExecOutput.mockResolvedValue({
        stdout: JSON.stringify({
          result: [
            { name: "sfdx-hardis", version: "1.0.0" },
            { name: "@salesforce/plugin-deploy-retrieve", version: "2.0.0" },
          ],
        }),
        stderr: "",
        exitCode: 0,
      });

      await runtime.installPlugins(mockConfig, [
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
          `${expectedEchoPrefix}sf "plugins" "install" "sfdx-hardis"`,
        ),
      ]);
      expect(mockedExec).toHaveBeenCalledWith(expectedShell, [
        expectedFlag,
        expect.stringContaining(
          `${expectedEchoPrefix}sf "plugins" "install" "@salesforce/plugin-deploy-retrieve"`,
        ),
      ]);
    });

    it("should verify plugin installation", async () => {
      const mockedGetExecOutput = exec.getExecOutput as jest.Mock;
      mockedGetExecOutput.mockResolvedValue({
        stdout: JSON.stringify({
          result: [{ name: "sfdx-hardis", version: "1.0.0" }],
        }),
        stderr: "",
        exitCode: 0,
      });

      await runtime.installPlugins(mockConfig, ["sfdx-hardis"]);

      expect(exec.getExecOutput).toHaveBeenCalledWith("sf", [
        "plugins",
        "--json",
      ]);
    });

    it("should verify plugin installation with Array output format", async () => {
      const mockedGetExecOutput = exec.getExecOutput as jest.Mock;
      mockedGetExecOutput.mockResolvedValue({
        stdout: JSON.stringify([{ name: "sfdx-hardis", version: "1.0.0" }]),
        stderr: "",
        exitCode: 0,
      });

      await runtime.installPlugins(mockConfig, ["sfdx-hardis"]);

      expect(exec.getExecOutput).toHaveBeenCalledWith("sf", [
        "plugins",
        "--json",
      ]);
    });

    it("should throw error when plugin verification returns unexpected JSON structure", async () => {
      // Ensure installation succeeds
      mockedExec.mockResolvedValue(0);

      const mockedGetExecOutput = exec.getExecOutput as jest.Mock;
      mockedGetExecOutput.mockResolvedValue({
        stdout: JSON.stringify({ some: "random", json: "object" }), // Valid JSON, invalid schema
        stderr: "",
        exitCode: 0,
      });

      await expect(
        runtime.installPlugins(mockConfig, ["sfdx-hardis"]),
      ).rejects.toThrow(
        "Plugin installation failed: Unexpected output format from 'sf plugins --json'",
      );
    });

    it("should throw error on installation failure", async () => {
      mockedExec.mockRejectedValue(new Error("Plugin installation failed"));

      await expect(
        runtime.installPlugins(mockConfig, ["sfdx-hardis"]),
      ).rejects.toThrow(
        "Plugin installation failed: Plugin installation failed",
      );
    });

    it("should throw error when plugin verification returns unexpected format", async () => {
      // Ensure installation succeeds so we get to verification
      mockedExec.mockResolvedValue(0);

      const mockedGetExecOutput = exec.getExecOutput as jest.Mock;
      mockedGetExecOutput.mockResolvedValueOnce({
        stdout: JSON.stringify({ unexpected: "format" }),
        stderr: "",
        exitCode: 0,
      });

      await expect(
        runtime.installPlugins(mockConfig, ["sfdx-hardis"]),
      ).rejects.toThrow(
        "Plugin installation failed: Unexpected output format from 'sf plugins --json'",
      );
    });

    it("should throw error on verification failure", async () => {
      // Reset the exec mock to succeed for installation but fail verification
      mockedExec.mockResolvedValue(0); // Installation succeeds

      const mockedGetExecOutput = exec.getExecOutput as jest.Mock;
      mockedGetExecOutput.mockResolvedValueOnce({
        stdout: JSON.stringify({
          result: [], // Plugin not found after installation
        }),
        stderr: "",
        exitCode: 0,
      });

      await expect(
        runtime.installPlugins(mockConfig, ["sfdx-hardis"]),
      ).rejects.toThrow(
        "Plugin installation failed: Plugin 'sfdx-hardis' installation verification failed",
      );
    });

    it("should install plugin with version constraint when whitelist specifies version", async () => {
      // Reset mocks for this test
      mockedExec.mockResolvedValue(0);
      const mockedGetExecOutput = exec.getExecOutput as jest.Mock;
      mockedGetExecOutput.mockResolvedValue({
        stdout: JSON.stringify({
          result: [{ name: "sfdx-hardis", version: "6.0.0" }],
        }),
        stderr: "",
        exitCode: 0,
      });

      const configWithVersionConstraint: ProtocolConfig = {
        governance: {
          enabled: true,
          plugin_whitelist: ["sfdx-hardis@^6.0.0"],
        },
        runtime: {
          cli_version: "latest",
          node_version: "20",
        },
      };

      await runtime.installPlugins(configWithVersionConstraint, [
        "sfdx-hardis",
      ]);

      // Check that the exec call was made (version constraint is applied in the real implementation)
      // The test verifies the branch is exercised - see mockedInfo calls for "Installing plugin"
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
        expect.stringContaining("Installing plugin:"),
      );
    });

    it("should handle non-Error exceptions", async () => {
      mockedExec.mockImplementation(() => {
        throw "string error"; // Non-Error throw
      });

      await expect(
        runtime.installPlugins(mockConfig, ["sfdx-hardis"]),
      ).rejects.toThrow("Plugin installation failed: string error");
    });

    it("should use sh shell for plugin installation on non-Windows platforms", async () => {
      // Ensure exec succeeds
      mockedExec.mockResolvedValue(0);
      const mockedGetExecOutput = exec.getExecOutput as jest.Mock;
      mockedGetExecOutput.mockResolvedValue({
        stdout: JSON.stringify({
          result: [{ name: "sfdx-hardis", version: "1.0.0" }],
        }),
        stderr: "",
        exitCode: 0,
      });
      // Force platform to linux for this test
      Object.defineProperty(runtime, "platform", { value: "linux" });

      await runtime.installPlugins(mockConfig, ["sfdx-hardis"]);

      expect(mockedExec).toHaveBeenCalledWith("sh", [
        "-c",
        expect.stringMatching(
          /echo y \| sf.*"plugins".*"install".*"sfdx-hardis"/,
        ),
      ]);
    });
  });
});
