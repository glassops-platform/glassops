module.exports = {
  preset: "ts-jest",
  testEnvironment: "node",
  rootDir: "../",
  testMatch: ["<rootDir>/src/**/*.test.ts"],
  testPathIgnorePatterns: [
    "<rootDir>/src/integration/",
    ".*\\.integration\\.test\\.ts$",
  ],
  moduleFileExtensions: ["ts", "js", "json", "node"],
  collectCoverage: true,
  coverageDirectory: "coverage",
  coverageReporters: ["text", "lcov", "html", "json-summary"],
  coverageThreshold: {
    global: {
      branches: 80,
      functions: 80,
      lines: 80,
      statements: 80,
    },
  },
};
