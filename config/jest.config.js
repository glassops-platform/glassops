const { jestConfig } = require('@salesforce/sfdx-lwc-jest/config');
const path = require('path');

const projectRoot = path.resolve(__dirname, '..');

module.exports = {
    ...jestConfig,
    rootDir: projectRoot,
    moduleNameMapper: {
        ...jestConfig.moduleNameMapper,
        '^lightning/(.+)$': path.join(projectRoot, 'test', 'jest-mocks', 'lightning', 'universalMock.js')
    },
    roots: [path.join(projectRoot, 'packages', 'portfolio-site')],
    testMatch: ['**/__tests__/**/*.{js,ts}'],
    modulePathIgnorePatterns: [path.join(projectRoot, 'packages', 'integration-api')],
    testPathIgnorePatterns: ['/node_modules/']
};
