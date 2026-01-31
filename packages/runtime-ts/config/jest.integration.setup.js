/**
 * Integration test setup
 * Sets up environment variables and cleanup for integration tests
 */

// Set default environment variables for tests
process.env.GITHUB_WORKSPACE = process.env.GITHUB_WORKSPACE || process.cwd();
process.env.GITHUB_ACTOR = process.env.GITHUB_ACTOR || "integration-test";
process.env.GITHUB_REPOSITORY =
  process.env.GITHUB_REPOSITORY || "test/integration";
process.env.GITHUB_RUN_ID = process.env.GITHUB_RUN_ID || "12345";
process.env.GITHUB_SHA = process.env.GITHUB_SHA || "abc123def456";
process.env.GITHUB_EVENT_NAME = process.env.GITHUB_EVENT_NAME || "push";

// Increase test timeout for integration tests
jest.setTimeout(30000);

// Console output helper for debugging integration tests
global.console = {
  ...console,
  // Uncomment to suppress console output during tests
  // log: jest.fn(),
  // info: jest.fn(),
  // warn: jest.fn(),
  // error: jest.fn(),
};
