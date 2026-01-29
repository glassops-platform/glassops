import * as core from "@actions/core";
import * as exec from "@actions/exec";
import * as io from "@actions/io";
import { ProtocolConfig } from "../protocol/policy";

export class RuntimeEnvironment {
  protected platform = process.platform;

  public async install(version: string = "latest"): Promise<void> {
    core.startGroup("🔧 Bootstrapping GlassOps Runtime");

    const isInstalled = await io.which("sf", false);
    if (isInstalled) {
      core.info("⚡ Salesforce CLI detected in environment. skipping install.");
      core.endGroup();
      return;
    }

    core.info(`⬇️ Installing @salesforce/cli@${version}...`);
    try {
      await exec.exec("npm", ["install", "-g", `@salesforce/cli@${version}`]);

      await exec.exec("sf", ["version"]);
    } catch {
      throw new Error(
        "❌ Failed to bootstrap runtime. NPM registry might be down.",
      );
    }

    core.endGroup();
  }

  public async installPlugins(
    config: ProtocolConfig,
    plugins: string[],
  ): Promise<void> {
    if (plugins.length === 0) {
      core.info("ℹ️ No plugins specified for installation.");
      return;
    }

    core.startGroup("🔌 Installing Salesforce CLI Plugins");

    for (const plugin of plugins) {
      try {
        core.info(`🔍 Validating plugin: ${plugin}`);

        // Check against whitelist if configured
        if (
          !config.governance.plugin_whitelist ||
          config.governance.plugin_whitelist.length === 0
        ) {
          core.warning(
            `⚠️ No plugin whitelist configured. Installing ${plugin} without validation.`,
          );
          await this.execWithAutoConfirm("sf", ["plugins", "install", plugin]);
        } else {
          // Check if plugin is in whitelist
          const policyEngine = new (
            await import("../protocol/policy")
          ).ProtocolPolicy();
          if (!policyEngine.validatePluginWhitelist(config, plugin)) {
            throw new Error(
              `🚫 Plugin '${plugin}' is not in the whitelist. Allowed plugins: ${config.governance.plugin_whitelist.join(", ")}`,
            );
          }

          // Get version constraint if specified
          const versionConstraint = policyEngine.getPluginVersionConstraint(
            config,
            plugin,
          );
          const installCommand = versionConstraint
            ? `${plugin}@${versionConstraint}`
            : plugin;

          core.info(`⬇️ Installing plugin: ${installCommand}`);
          await this.execWithAutoConfirm("sf", [
            "plugins",
            "install",
            installCommand,
          ]);
        }

        // Verify installation
        const result = await exec.getExecOutput("sf", ["plugins", "--json"]);
        const parsed = JSON.parse(result.stdout);
        let installedPlugins: Array<{ name: string }> | undefined;

        if (Array.isArray(parsed)) {
          installedPlugins = parsed;
        } else if (
          parsed &&
          typeof parsed === "object" &&
          "result" in parsed &&
          Array.isArray(parsed.result)
        ) {
          installedPlugins = parsed.result;
        }

        if (!installedPlugins) {
          throw new Error("Unexpected output format from 'sf plugins --json'");
        }

        const isInstalled = installedPlugins.some(
          (p: { name: string }) => p.name === plugin,
        );

        if (!isInstalled) {
          throw new Error(
            `Plugin '${plugin}' installation verification failed`,
          );
        }

        core.info(`✅ Plugin '${plugin}' installed and verified successfully`);
      } catch (error) {
        const errorMessage =
          error instanceof Error ? error.message : String(error);
        core.error(`❌ Failed to install plugin '${plugin}': ${errorMessage}`);
        throw new Error(`Plugin installation failed: ${errorMessage}`);
      }
    }

    core.endGroup();
  }

  private async execWithAutoConfirm(
    command: string,
    args: string[],
  ): Promise<void> {
    const joinedArgs = args.map((arg) => `"${arg}"`).join(" ");
    const fullCommand = `${command} ${joinedArgs}`;
    if (this.platform === "win32") {
      await exec.exec("cmd", ["/c", `echo y|${fullCommand}`]);
    } else {
      await exec.exec("sh", ["-c", `echo y | ${fullCommand}`]);
    }
  }
}
