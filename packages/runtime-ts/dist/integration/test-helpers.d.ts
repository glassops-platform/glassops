/**
 * Shared Test Helpers for Integration Tests
 *
 * Provides reusable utilities for setting up test environments,
 * mocking external dependencies, and creating test data.
 */
export declare const TEST_WORKSPACE: string;
export declare const DEFAULT_CONFIG: {
    governance: {
        enabled: boolean;
    };
    runtime: {
        cli_version: string;
        node_version: string;
    };
};
/**
 * Creates a test workspace directory and config file
 */
export declare function setupTestWorkspace(config?: {
    governance: {
        enabled: boolean;
    };
    runtime: {
        cli_version: string;
        node_version: string;
    };
}): string;
/**
 * Cleans up test workspace
 */
export declare function cleanupTestWorkspace(): void;
/**
 * Creates mock environment variables for testing
 */
export declare function createMockEnvironment(overrides?: Record<string, string>): Record<string, string>;
/**
 * Creates mock GitHub Action inputs
 */
export declare function createMockInputs(overrides?: Record<string, string>): Record<string, string>;
/**
 * Creates a temporary JWT key file
 */
export declare function createTempJwtKey(content?: string): string;
/**
 * Mocks successful Salesforce CLI execution
 */
export declare function mockSuccessfulCliExecution(mockExec: jest.Mock, response?: any): void;
/**
 * Mocks failed Salesforce CLI execution
 */
export declare function mockFailedCliExecution(mockExec: jest.Mock, errorMessage?: string): void;
/**
 * Mocks cache operations
 */
export declare function mockCacheOperations(mockRestoreCache: jest.Mock, shouldRestore?: boolean): void;
/**
 * Mocks file system which operations
 */
export declare function mockWhichOperations(mockWhich: jest.Mock, availableCommands?: string[]): void;
/**
 * Test data factories for common scenarios
 */
export declare const TestData: {
    validJwtKey: string;
    invalidJwtKey: string;
    testResults: {
        valid: {
            total: number;
            passed: number;
            failed: number;
        };
        empty: {
            total: number;
            passed: number;
            failed: number;
        };
        allPassed: {
            total: number;
            passed: number;
            failed: number;
        };
        allFailed: {
            total: number;
            passed: number;
            failed: number;
        };
    };
    coverageData: {
        good: {
            actual: number;
            required: number;
        };
        borderline: {
            actual: number;
            required: number;
        };
        failing: {
            actual: number;
            required: number;
        };
        perfect: {
            actual: number;
            required: number;
        };
    };
    freezeWindows: {
        weekend: {
            day: string;
            start: string;
            end: string;
        }[];
        weekday: {
            day: string;
            start: string;
            end: string;
        }[];
        multiple: {
            day: string;
            start: string;
            end: string;
        }[];
    };
    pluginConfigs: {
        whitelist: string[];
        noWhitelist: never[];
        versioned: string[];
        scoped: string[];
    };
    salesforceEnvironments: {
        production: string;
        sandbox: string;
        custom: string;
    };
    repositoryFormats: {
        valid: string;
        invalid: string;
        nested: string;
    };
};
/**
 * Assertion helpers for common test patterns
 */
export declare const Assertions: {
    expectSuccessfulExecution: (mockSetOutput: jest.Mock, expectedOutputs: Record<string, any>) => void;
    expectErrorLogged: (mockSetFailed: jest.Mock, expectedMessage: string) => void;
    expectWarningLogged: (mockWarning: jest.Mock, expectedMessage: string) => void;
    expectInfoLogged: (mockInfo: jest.Mock, expectedMessage: string) => void;
    expectCliCommandExecuted: (mockExec: jest.Mock, expectedArgs: string[]) => void;
    expectFileExists: (filePath: string) => any;
    expectFileNotExists: (filePath: string) => void;
};
/**
 * Test scenario builders for complex integration tests
 */
export declare class TestScenarioBuilder {
    private env;
    private inputs;
    private config;
    private cacheRestored;
    private availableCommands;
    withEnvironment(overrides: Record<string, string>): this;
    withInputs(overrides: Record<string, string>): this;
    withConfig(config: any): this;
    withCacheRestored(restored?: boolean): this;
    withAvailableCommands(commands: string[]): this;
    build(): {
        env: Record<string, string>;
        inputs: Record<string, string>;
        config: any;
        cacheRestored: boolean;
        availableCommands: string[];
    };
}
//# sourceMappingURL=test-helpers.d.ts.map