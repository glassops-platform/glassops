import * as exec from "@actions/exec";

export interface HealthCheckResult {
  healthy: boolean;
  version?: string;
  error?: string;
}

/**
 * Perform a health check by verifying the Salesforce CLI is available.
 */
export async function healthCheck(): Promise<HealthCheckResult> {
  try {
    let output = "";
    await exec.exec("sf", ["version", "--json"], {
      silent: true,
      listeners: {
        stdout: (data) => (output += data.toString()),
      },
    });

    const result = JSON.parse(output);
    const version = result?.cliVersion ?? result?.result?.cliVersion ?? "unknown";

    return {
      healthy: true,
      version,
    };
  } catch (error) {
    return {
      healthy: false,
      error: error instanceof Error ? error.message : String(error),
    };
  }
}
