/**
 * Shared Test Helpers for Integration Tests
 *
 * Provides reusable utilities for setting up test environments,
 * mocking external dependencies, and creating test data.
 */

import * as fs from "fs";
import * as path from "path";
import * as os from "os";

// Test environment constants
export const TEST_WORKSPACE = path.join(
  __dirname,
  "..",
  "..",
  "test-workspace",
);
export const DEFAULT_CONFIG = {
  governance: { enabled: true },
  runtime: { cli_version: "latest", node_version: "20" },
};

/**
 * Creates a test workspace directory and config file
 */
export function setupTestWorkspace(config = DEFAULT_CONFIG): string {
  if (!fs.existsSync(TEST_WORKSPACE)) {
    fs.mkdirSync(TEST_WORKSPACE, { recursive: true });
  }

  const configPath = path.join(TEST_WORKSPACE, "devops-config.json");
  fs.writeFileSync(configPath, JSON.stringify(config, null, 2));

  return TEST_WORKSPACE;
}

/**
 * Cleans up test workspace
 */
export function cleanupTestWorkspace(): void {
  if (fs.existsSync(TEST_WORKSPACE)) {
    fs.rmSync(TEST_WORKSPACE, { recursive: true, force: true });
  }
}

/**
 * Creates mock environment variables for testing
 */
export function createMockEnvironment(
  overrides: Record<string, string> = {},
): Record<string, string> {
  return {
    GITHUB_WORKSPACE: TEST_WORKSPACE,
    GITHUB_ACTOR: "test-actor",
    GITHUB_REPOSITORY: "test-org/test-repo",
    GITHUB_SHA: "abc123",
    GITHUB_EVENT_NAME: "push",
    GITHUB_HEAD_REF: "feature-branch",
    HOME: TEST_WORKSPACE,
    ...overrides,
  };
}

/**
 * Creates mock GitHub Action inputs
 */
export function createMockInputs(
  overrides: Record<string, string> = {},
): Record<string, string> {
  return {
    client_id: "test-client-id",
    jwt_key:
      "-----BEGIN RSA PRIVATE KEY-----\ntest\n-----END RSA PRIVATE KEY-----",
    username: "test@example.com",
    instance_url: "https://login.salesforce.com",
    enforce_policy: "false",
    skip_auth: "true",
    plugins: "",
    test_results: "",
    coverage_percentage: "",
    coverage_required: "",
    ...overrides,
  };
}

/**
 * Creates a temporary JWT key file
 */
export function createTempJwtKey(
  content = "-----BEGIN RSA PRIVATE KEY-----\ntest-key\n-----END RSA PRIVATE KEY-----",
): string {
  const keyPath = path.join(os.tmpdir(), `test-jwt-${Date.now()}.key`);
  fs.writeFileSync(keyPath, content, { mode: 0o600 });
  return keyPath;
}

/**
 * Mocks successful Salesforce CLI execution
 */
export function mockSuccessfulCliExecution(
  mockExec: jest.Mock,
  response?: any,
): void {
  const defaultResponse = {
    result: { orgId: "00D123456789012345", accessToken: "test-token" },
  };

  mockExec.mockImplementation(
    async (command: string, args: string[], options: any) => {
      // Simulate stdout listener for commands that expect JSON output
      if (options?.listeners?.stdout) {
        const mockResponse =
          response || (args.includes("--json") ? defaultResponse : {});
        options.listeners.stdout(Buffer.from(JSON.stringify(mockResponse)));
      }
      return 0; // Success exit code
    },
  );
}

/**
 * Mocks failed Salesforce CLI execution
 */
export function mockFailedCliExecution(
  mockExec: jest.Mock,
  errorMessage = "CLI command failed",
): void {
  mockExec.mockRejectedValue(new Error(errorMessage));
}

/**
 * Mocks cache operations
 */
export function mockCacheOperations(
  mockRestoreCache: jest.Mock,
  shouldRestore = false,
): void {
  mockRestoreCache.mockResolvedValue(shouldRestore ? "cache-key" : null);
}

/**
 * Mocks file system which operations
 */
export function mockWhichOperations(
  mockWhich: jest.Mock,
  availableCommands: string[] = ["sf"],
): void {
  mockWhich.mockImplementation((command: string) => {
    return availableCommands.includes(command)
      ? `/usr/local/bin/${command}`
      : "";
  });
}

/**
 * Test data factories for common scenarios
 */
export const TestData = {
  validJwtKey:
    "-----BEGIN RSA PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC...\n-----END RSA PRIVATE KEY-----",

  invalidJwtKey: "not-a-valid-jwt-key",

  testResults: {
    valid: { total: 100, passed: 95, failed: 5 },
    empty: { total: 0, passed: 0, failed: 0 },
    allPassed: { total: 50, passed: 50, failed: 0 },
    allFailed: { total: 20, passed: 0, failed: 20 },
  },

  coverageData: {
    good: { actual: 92, required: 80 },
    borderline: { actual: 80, required: 80 },
    failing: { actual: 75, required: 80 },
    perfect: { actual: 100, required: 90 },
  },

  freezeWindows: {
    weekend: [{ day: "Saturday", start: "00:00", end: "23:59" }],
    weekday: [{ day: "Monday", start: "09:00", end: "17:00" }],
    multiple: [
      { day: "Friday", start: "17:00", end: "23:59" },
      { day: "Saturday", start: "00:00", end: "23:59" },
    ],
  },

  pluginConfigs: {
    whitelist: ["sfdx-hardis@^4.0.0", "@salesforce/plugin-deploy-retrieve"],
    noWhitelist: [],
    versioned: ["sfdx-hardis@^6.0.0", "sf-metadata-scanner@1.2.3"],
    scoped: ["@salesforce/plugin-deploy-retrieve@latest"],
  },

  salesforceEnvironments: {
    production: "https://login.salesforce.com",
    sandbox: "https://test.salesforce.com",
    custom: "https://mycompany.my.salesforce.com",
  },

  repositoryFormats: {
    valid: "owner/repo",
    invalid: "invalid-repo-format",
    nested: "org/team/project",
  },
};

/**
 * Assertion helpers for common test patterns
 */
export const Assertions = {
  expectSuccessfulExecution: (
    mockSetOutput: jest.Mock,
    expectedOutputs: Record<string, any>,
  ) => {
    Object.entries(expectedOutputs).forEach(([key, value]) => {
      if (value === "any-string") {
        expect(mockSetOutput).toHaveBeenCalledWith(key, expect.any(String));
      } else {
        expect(mockSetOutput).toHaveBeenCalledWith(key, value);
      }
    });
  },

  expectErrorLogged: (mockSetFailed: jest.Mock, expectedMessage: string) => {
    expect(mockSetFailed).toHaveBeenCalledWith(
      expect.stringContaining(expectedMessage),
    );
  },

  expectWarningLogged: (mockWarning: jest.Mock, expectedMessage: string) => {
    expect(mockWarning).toHaveBeenCalledWith(
      expect.stringContaining(expectedMessage),
    );
  },

  expectInfoLogged: (mockInfo: jest.Mock, expectedMessage: string) => {
    expect(mockInfo).toHaveBeenCalledWith(expectedMessage);
  },

  expectCliCommandExecuted: (mockExec: jest.Mock, expectedArgs: string[]) => {
    expect(mockExec).toHaveBeenCalledWith(
      "sf",
      expect.arrayContaining(expectedArgs),
      expect.any(Object),
    );
  },

  expectFileExists: (filePath: string) => {
    expect(fs.existsSync(filePath)).toBe(true);
    return JSON.parse(fs.readFileSync(filePath, "utf-8"));
  },

  expectFileNotExists: (filePath: string) => {
    expect(fs.existsSync(filePath)).toBe(false);
  },
};

/**
 * Test scenario builders for complex integration tests
 */
export class TestScenarioBuilder {
  private env: Record<string, string> = {};
  private inputs: Record<string, string> = {};
  private config: any = DEFAULT_CONFIG;
  private cacheRestored = false;
  private availableCommands: string[] = ["sf"];

  withEnvironment(overrides: Record<string, string>): this {
    this.env = { ...createMockEnvironment(), ...overrides };
    return this;
  }

  withInputs(overrides: Record<string, string>): this {
    this.inputs = { ...createMockInputs(), ...overrides };
    return this;
  }

  withConfig(config: any): this {
    this.config = config;
    return this;
  }

  withCacheRestored(restored = true): this {
    this.cacheRestored = restored;
    return this;
  }

  withAvailableCommands(commands: string[]): this {
    this.availableCommands = commands;
    return this;
  }

  build(): {
    env: Record<string, string>;
    inputs: Record<string, string>;
    config: any;
    cacheRestored: boolean;
    availableCommands: string[];
  } {
    return {
      env: this.env,
      inputs: this.inputs,
      config: this.config,
      cacheRestored: this.cacheRestored,
      availableCommands: this.availableCommands,
    };
  }
}
