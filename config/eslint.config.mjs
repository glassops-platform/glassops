import { createRequire } from 'module';
import path from 'path';
import { fileURLToPath } from 'url';
import js from '@eslint/js';
import globals from 'globals';
import jestPlugin from 'eslint-plugin-jest'; // Native Jest Plugin
import tseslint from 'typescript-eslint';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// Initialize require for CJS compatibility
const require = createRequire(import.meta.url);

// 1. Load Salesforce Config Directly (No FlatCompat)
// The error log showed this is ALREADY an array, so we require it directly.
const lwcConfig = require('@salesforce/eslint-config-lwc/recommended');

export default [
    // --- GLOBAL IGNORES ---
    {
        ignores: ['**/node_modules/**', '**/dist/**', '**/coverage/**']
    },

    // --- BASE JS/TS CONFIGURATION ---
    {
        files: ['**/*.js', '**/*.ts'],
        languageOptions: {
            sourceType: 'module',
            ecmaVersion: 'latest',
            globals: {
                ...globals.browser,
                ...globals.node
            }
        }
    },

    // --- SALESFORCE RULES ---
    // Since lwcConfig is an array (Flat Config), we just spread it here.
    ...lwcConfig,
    
    // --- APP-SPECIFIC OVERRIDES ---
    // Restore overrides for portfolio/demo specific patterns
    {
        files: ['**/lwc/**/*.js'],
        rules: {
            '@lwc/lwc-platform/no-inline-disable': 'off',
            'no-await-in-loop': 'off',
            // Allow setTimeout and requestAnimationFrame for animations
            '@lwc/lwc/no-async-operation': 'off'
        }
    },

    // --- JEST NATIVE CONFIG ---
    // We explicitly load the modern Jest config here to ensure test files work.
    {
        files: ['**/__tests__/**'],
        ...jestPlugin.configs['flat/recommended'],
        languageOptions: {
            globals: {
                ...globals.jest
            }
        }
    },

    // --- TYPESCRIPT CONFIGURATION ---
    // Spread recommended configs (they are an array)
    ...tseslint.configs.recommended,
    
    // TS Specific overrides
    {
        files: ['**/*.ts'],
        languageOptions: {
            parser: tseslint.parser,
            parserOptions: {
                projectService: true,
                tsconfigRootDir: import.meta.dirname,
            },
        }
    }
];