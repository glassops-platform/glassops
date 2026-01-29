import * as core from "@actions/core";
import * as exec from "@actions/exec";
import * as fs from "fs";
import * as path from "path";
import * as os from "os";

interface AuthRequest {
  clientId: string;
  jwtKey: string;
  username: string;
  instanceUrl?: string;
}

export class IdentityResolver {
  public async authenticate(req: AuthRequest): Promise<string> {
    core.startGroup("🔑 Authenticating Identity");

    const keyPath = path.join(os.tmpdir(), `glassops-jwt-${Date.now()}.key`);
    fs.writeFileSync(keyPath, req.jwtKey, { mode: 0o600 });

    try {
      const args = [
        "org",
        "login",
        "jwt",
        "--client-id",
        req.clientId,
        "--jwt-key-file",
        keyPath,
        "--username",
        req.username,
        "--set-default",
        "--json",
      ];

      if (req.instanceUrl) {
        args.push("--instance-url", req.instanceUrl);
      }

      interface LoginResult {
        result: { orgId: string; accessToken: string };
      }

      let output = "";
      await exec.exec("sf", args, {
        listeners: { stdout: (data) => (output += data.toString()) },
        silent: true,
      });

      const result = JSON.parse(output) as LoginResult;
      core.info(`✅ Authenticated as ${req.username} (${result.result.orgId})`);

      return result.result.orgId;
    } catch {
      throw new Error("❌ Authentication Failed. Check Client ID and JWT Key.");
    } finally {
      if (fs.existsSync(keyPath)) {
        fs.unlinkSync(keyPath);
      }
      core.endGroup();
    }
  }
}
