---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/config/jest.integration.setup.js
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/config/jest.integration.setup.js
generated_at: 2026-01-31T11:09:29.158371
hash: 50ee3e4481b84677818a5aa4389662eeb26e01b9753a4b71e7d38555636d2959
---

## Integration Test Configuration

This document details the configuration applied specifically for integration tests within the runtime environment. It outlines environment variable setup and adjustments to the testing framework to support comprehensive and reliable integration testing.

### Purpose

The primary goal of this configuration is to establish a consistent and predictable environment for integration tests, mimicking conditions found in a typical continuous integration (CI) pipeline. This ensures tests accurately reflect real-world scenarios and reduces the likelihood of false positives or negatives.

### Environment Variables

The configuration sets default values for several environment variables commonly used by the runtime and its integrations. These variables are used to simulate the GitHub Actions environment. If these variables are already defined in the execution environment, their existing values are preserved. The following variables are configured:

*   `GITHUB_WORKSPACE`:  Defaults to the current working directory (`process.cwd()`). Represents the root directory of the repository.
*   `GITHUB_ACTOR`: Defaults to “integration-test”. Represents the GitHub user or application that triggered the workflow.
*   `GITHUB_REPOSITORY`: Defaults to “test/integration”. Represents the owner and repository name.
*   `GITHUB_RUN_ID`: Defaults to “12345”. Represents the unique identifier for the current workflow run.
*   `GITHUB_SHA`: Defaults to “abc123def456”. Represents the commit SHA that triggered the workflow.
*   `GITHUB_EVENT_NAME`: Defaults to “push”. Represents the event that triggered the workflow.

You can override these default values by setting the corresponding environment variables before running the tests.

### Test Timeout

The default timeout for integration tests is increased to 30,000 milliseconds (30 seconds) using `jest.setTimeout(30000)`. Integration tests often involve interactions with external resources or more complex operations, requiring a longer execution time than unit tests.

### Console Output Management

We provide a mechanism to control console output during test execution. By default, all console methods (`log`, `info`, `warn`, `error`) function as expected. However, you can suppress console output by uncommenting the corresponding `jest.fn()` assignments within the `global.console` object. This can be useful for reducing noise in test results.

```typescript
global.console = {
  ...console,
  // Uncomment to suppress console output during tests
  // log: jest.fn(),
  // info: jest.fn(),
  // warn: jest.fn(),
  // error: jest.fn(),
};
```

This configuration ensures that integration tests are executed in a controlled and representative environment, leading to more reliable and accurate results.