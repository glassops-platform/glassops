import rootConfig from '../../../config/eslint.config.mjs';

export default [
    ...rootConfig,
    {
        ignores: ['dist/**', 'node_modules/**']
    },
    {
        files: ['**/*.ts'],
        rules: {
            '@lwc/lwc/no-async-operation': 'off',
            'no-await-in-loop': 'off',
            '@typescript-eslint/no-explicit-any': 'warn',
            '@typescript-eslint/no-unused-vars': ['warn', { "argsIgnorePattern": "^_" }]
        }
    }
];
