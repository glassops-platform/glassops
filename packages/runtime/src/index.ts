import * as cache from "@actions/cache";
import * as core from "@actions/core";
import * as crypto from "crypto";
import * as fs from "fs";
import * as path from "path";
import { ProtocolPolicy, ProtocolConfig } from "./protocol/policy";
import { RuntimeEnvironment } from "./services/cli";
import { IdentityResolver } from "./services/identity";
import { Analyzer } from "./services/analyzer";
import {
  DeploymentContractSchema,
  DeploymentContract,
} from "./protocol/contract";

// Custom error types for better error categorization
class GlassOpsError extends Error {
  constructor(
    message: string,
    public readonly phase: string,
    public readonly code: string,
    public readonly cause?: Error,
  ) {
    super(message);
    this.name = "GlassOpsError";
  }
}

class PolicyError extends GlassOpsError {
  constructor(message: string, cause?: Error) {
    super(message, "Policy", "POLICY_VIOLATION", cause);
    this.name = "PolicyError";
  }
}

class BootstrapError extends GlassOpsError {
  constructor(message: string, cause?: Error) {
    super(message, "Bootstrap", "BOOTSTRAP_FAILED", cause);
    this.name = "BootstrapError";
  }
}

class IdentityError extends GlassOpsError {
  constructor(message: string, cause?: Error) {
    super(message, "Identity", "AUTHENTICATION_FAILED", cause);
    this.name = "IdentityError";
  }
}

class ContractError extends GlassOpsError {
  constructor(message: string, cause?: Error) {
    super(message, "Contract", "CONTRACT_GENERATION_FAILED", cause);
    this.name = "ContractError";
  }
}

async function run(): Promise<void> {
  try {
    // LEVEL 1 PRIMITIVE: Core Governance Controls
    // ============================================

    // 1. Environment Context Validation
    const requiredEnvVars = [
      "GITHUB_WORKSPACE",
      "GITHUB_ACTOR",
      "GITHUB_REPOSITORY",
    ];

    const missingEnvVars = requiredEnvVars.filter(
      (envVar) => !process.env[envVar],
    );
    if (missingEnvVars.length > 0) {
      throw new Error(
        `Missing required environment variables: ${missingEnvVars.join(", ")}`,
      );
    }

    // 2. Input Validation & Sanitization
    const requiredInputs = ["client_id", "jwt_key", "username"];
    const missingInputs = requiredInputs.filter(
      (input) => !core.getInput(input),
    );
    if (missingInputs.length > 0) {
      throw new Error(`Missing required inputs: ${missingInputs.join(", ")}`);
    }

    // Validate JWT key format (basic check)
    const jwtKey = core.getInput("jwt_key");
    if (!jwtKey.includes("BEGIN") || !jwtKey.includes("END")) {
      throw new Error(
        "Invalid JWT key format - must contain BEGIN and END markers",
      );
    }

    // Validate Salesforce instance URL
    const instanceUrl =
      core.getInput("instance_url") || "https://login.salesforce.com";
    try {
      // eslint-disable-next-line no-new
      new URL(instanceUrl);
    } catch {
      throw new Error(`Invalid instance URL: ${instanceUrl}`);
    }

    // 3. Rate Limiting (Basic session tracking - reserved for future implementation)
    // Session tracking and concurrency limits will be implemented in a future version

    // 4. Resource Limits Validation
    const maxExecutionTime = 30 * 60 * 1000; // 30 minutes max
    const startTime = Date.now();

    // Safety check for execution time
    const safetyTimeout = setTimeout(() => {
      if (Date.now() - startTime > maxExecutionTime) {
        core.error("Execution timeout exceeded - terminating session");
        process.exit(1);
      }
    }, maxExecutionTime);
    // Ensure the timeout doesn't prevent the process from exiting
    safetyTimeout.unref();

    // 5. Data Integrity & Compliance Checks
    // Validate GitHub context for security
    if (
      process.env.GITHUB_EVENT_NAME === "pull_request" &&
      !process.env.GITHUB_HEAD_REF
    ) {
      throw new Error("Invalid pull request context - missing GITHUB_HEAD_REF");
    }

    // Basic compliance check - ensure we're not running in forked repositories without proper context
    if (
      process.env.GITHUB_EVENT_NAME === "pull_request" &&
      process.env.GITHUB_HEAD_REF?.includes(":")
    ) {
      core.warning(
        "⚠️ Running on forked repository - additional security validations recommended",
      );
    }

    // Validate repository format
    const repoPattern = /^[a-zA-Z0-9._-]+\/[a-zA-Z0-9._-]+$/;
    if (!repoPattern.test(process.env.GITHUB_REPOSITORY || "")) {
      throw new Error(
        `Invalid repository format: ${process.env.GITHUB_REPOSITORY}`,
      );
    }

    // Generate unique runtime ID for this execution session
    const runtimeId = crypto.randomUUID();
    core.setOutput("runtime_id", runtimeId);

    // Structured logging helper
    const log = {
      info: (message: string, phase?: string) => {
        const prefix = phase ? `[${phase}] ` : "";
        core.info(`${prefix}${message}`);
      },
      warning: (message: string, phase?: string) => {
        const prefix = phase ? `[${phase}] ` : "";
        core.warning(`${prefix}${message}`);
      },
      error: (message: string, phase?: string) => {
        const prefix = phase ? `[${phase}] ` : "";
        core.error(`${prefix}${message}`);
      },
    };

    // 0. Cache Retrieval Phase
    core.startGroup("📦 Restoring GlassOps Runtime Cache");
    const cacheKey = `glassops-runtime-${process.platform}-${process.arch}-${process.version}`;
    const cachePaths = [
      path.join(process.env.HOME || process.env.USERPROFILE || ".", ".npm"),
      path.join(process.env.HOME || process.env.USERPROFILE || ".", ".sf"),
    ];

    const cacheHit = await cache.restoreCache(cachePaths, cacheKey);
    if (cacheHit) {
      core.info("✅ Cache restored successfully");
    } else {
      core.info("⚠️ Cache miss - will cache after bootstrap");
    }
    core.endGroup();

    // 1. Policy Phase
    log.info("Evaluating governance policies...", "Policy");
    let config: ProtocolConfig;
    try {
      const policyEngine = new ProtocolPolicy();
      config = await policyEngine.load();

      // BR-003: Static Analysis Invariants
      if (config.governance.analyzer?.enabled) {
        log.info("Running static code analysis...", "Analyzer");
        const analyzer = new Analyzer();

        // Opinionated Check: Ensure we aren't using deprecated tools
        if (config.governance.analyzer.opinionated) {
          await analyzer.ensureCompatibility();
        }

        const scanResults = await analyzer.scan(
          ["."],
          config.governance.analyzer.rulesets?.[0],
        );

        const criticalViolations = scanResults.violations.filter(
          (v) =>
            v.severity <= (config.governance.analyzer?.severity_threshold || 1),
        );

        if (criticalViolations.length > 0) {
          throw new PolicyError(
            `Static analysis failed: ${criticalViolations.length} critical violations found.`,
          );
        }
        log.info("✅ Static analysis passed", "Analyzer");
      }

      if (core.getInput("enforce_policy") === "true") {
        try {
          policyEngine.checkFreeze(config);
          core.setOutput("is_locked", "false");
          log.info(
            "✅ Policy check passed - no freeze windows active",
            "Policy",
          );
        } catch (freezeError) {
          core.setOutput("is_locked", "true");
          throw new PolicyError(
            freezeError instanceof Error
              ? freezeError.message
              : "Freeze window violation",
            freezeError instanceof Error ? freezeError : undefined,
          );
        }
      } else {
        core.setOutput("is_locked", "false");
        log.warning("Policy enforcement disabled", "Policy");
      }
    } catch (error) {
      if (error instanceof PolicyError) throw error;
      throw new PolicyError(
        "Policy evaluation failed",
        error instanceof Error ? error : undefined,
      );
    }

    // 2. Bootstrap Phase
    log.info("Bootstrapping Salesforce CLI environment...", "Bootstrap");
    try {
      const runtime = new RuntimeEnvironment();
      await runtime.install(config.runtime.cli_version);
      log.info(
        `✅ CLI ${config.runtime.cli_version} installed successfully`,
        "Bootstrap",
      );

      // Install plugins if specified
      const pluginsInput = core.getInput("plugins");
      if (pluginsInput) {
        const plugins = pluginsInput
          .split(",")
          .map((p) => p.trim())
          .filter((p) => p.length > 0);
        if (plugins.length > 0) {
          log.info(
            `Installing ${plugins.length} plugin(s): ${plugins.join(", ")}`,
            "Bootstrap",
          );
          await runtime.installPlugins(config, plugins);
          log.info("✅ All plugins installed successfully", "Bootstrap");
        }
      }
    } catch (error) {
      throw new BootstrapError(
        "Bootstrap phase failed",
        error instanceof Error ? error : undefined,
      );
    }

    // 3. Identity Phase
    log.info("Authenticating with Salesforce...", "Identity");
    let orgId: string;

    if (core.getInput("skip_auth") === "true") {
      log.warning(
        "⚠️ Skipping authentication as requested by configuration",
        "Identity",
      );
      orgId = "00D00000000TEST"; // Dummy Org ID for testing
    } else {
      try {
        const identity = new IdentityResolver();
        orgId = await identity.authenticate({
          clientId: core.getInput("client_id"),
          jwtKey: core.getInput("jwt_key"),
          username: core.getInput("username"),
          instanceUrl: core.getInput("instance_url"),
        });
        log.info(`✅ Authenticated with org ${orgId}`, "Identity");
      } catch (error) {
        throw new IdentityError(
          "Salesforce authentication failed",
          error instanceof Error ? error : undefined,
        );
      }
    }

    // 4. Contract Validation Phase
    log.info("📄 Generating Deployment Contract v1.0...", "Contract");
    let deploymentContract: DeploymentContract;
    try {
      // Parse test results from input
      let testResults = { total: 0, passed: 0, failed: 0 };
      const testResultsInput = core.getInput("test_results");
      if (testResultsInput) {
        try {
          testResults = JSON.parse(testResultsInput);
          log.info(
            `Parsed test results: ${testResults.passed}/${testResults.total} passed`,
            "Contract",
          );
        } catch (parseError) {
          log.warning(
            `⚠️ Invalid test_results JSON, using defaults: ${parseError}`,
            "Contract",
          );
        }
      }

      // Get coverage data from inputs
      const coverageActual =
        parseFloat(core.getInput("coverage_percentage")) || 0;
      const coverageRequired =
        parseFloat(core.getInput("coverage_required")) || 80;

      deploymentContract = DeploymentContractSchema.parse({
        schemaVersion: "1.0",
        meta: {
          adapter: "native",
          engine: "native",
          timestamp: new Date().toISOString(),
          trigger: process.env.GITHUB_EVENT_NAME || "manual",
        },
        status: "Succeeded",
        quality: {
          coverage: {
            actual: coverageActual,
            required: coverageRequired,
            met: coverageActual >= coverageRequired,
          },
          tests: testResults,
        },
        audit: {
          triggeredBy: process.env.GITHUB_ACTOR || "unknown",
          orgId: orgId,
          repository: process.env.GITHUB_REPOSITORY || "unknown",
          commit: process.env.GITHUB_SHA || "unknown",
        },
      });

      // Write contract to file for cryptographic signing/archival
      const contractPath = path.join(
        process.env.GITHUB_WORKSPACE || ".",
        "glassops-contract.json",
      );
      fs.writeFileSync(
        contractPath,
        JSON.stringify(deploymentContract, null, 2),
      );
      core.setOutput("contract_path", contractPath);
      log.info(`✅ Contract written to ${contractPath}`, "Contract");
    } catch (error) {
      throw new ContractError(
        "Contract generation failed",
        error instanceof Error ? error : undefined,
      );
    }

    // 5. Output Session State
    core.setOutput("org_id", deploymentContract.audit.orgId);
    core.setOutput("glassops_ready", "true");
    core.info("✅ GlassOps Runtime is ready for governed execution.");
  } catch (error) {
    if (error instanceof Error) core.setFailed(error.message);
  }
}

export { run };

// Only auto-run when executed directly (not when imported for testing)
if (require.main === module || !process.env.JEST_WORKER_ID) {
  run();
}
