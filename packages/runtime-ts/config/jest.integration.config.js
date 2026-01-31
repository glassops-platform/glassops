/** @type {import('jest').Config} */
module.exports = {
  preset: "ts-jest",
  testEnvironment: "node",
  rootDir: "../src",
  testMatch: ["**/*.integration.test.ts"],
  // Longer timeout for integration tests
  testTimeout: 30000,
  // Don't transform node_modules except for specific packages
  transformIgnorePatterns: ["/node_modules/"],
  // Coverage settings for integration tests
  collectCoverage: true,
  coverageDirectory: "../coverage/integration",
  coverageReporters: ["text", "lcov", "html", "json-summary"],
  // Collect coverage from all source files
  collectCoverageFrom: [
    "**/*.ts",
    "!**/*.test.ts",
    "!**/*.integration.test.ts",
    "!**/node_modules/**",
    "!**/index.ts", // Entry point tested via E2E, not integration tests
    "!**/integration/test-helpers.ts", // Test utilities, not production code
  ],
  // Setup file for integration test environment
  setupFilesAfterEnv: ["<rootDir>/../config/jest.integration.setup.js"],
};
