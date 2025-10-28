// @ts-check
import { includeIgnoreFile } from '@eslint/compat';
import js from '@eslint/js';
import svelte from 'eslint-plugin-svelte';
import unicorn from 'eslint-plugin-unicorn';
import { defineConfig } from 'eslint/config';
import globals from 'globals';
import { fileURLToPath } from 'node:url';
import ts from 'typescript-eslint';

const gitignorePath = fileURLToPath(new globalThis.URL('../../.gitignore', import.meta.url));

/** @type {import('eslint').Linter.Config[]} */
export default defineConfig(
	includeIgnoreFile(gitignorePath),
	js.configs.recommended,
	unicorn.configs.recommended,
	...ts.configs.recommended,
	...svelte.configs.recommended,
	{
		files: ['src/**/*.ts', 'src/**/*.js', 'src/**/*.svelte', 'src/**/*.svelte.ts', 'src/**/*.svelte.js'],
		languageOptions: {
			globals: { ...globals.browser, ...globals.node },
			parserOptions: {
				tsconfigRootDir: import.meta.dirname,
				svelteConfig: {},
				projectService: true,
				extraFileExtensions: ['.svelte'],
				parser: ts.parser
			}
		},
		rules: {
			'no-undef': 'off',
			'@typescript-eslint/no-unused-vars': ['warn', { argsIgnorePattern: '^_', varsIgnorePattern: '^_' }],
			'@typescript-eslint/no-explicit-any': 'off',
			'@typescript-eslint/no-this-alias': 'off',
			'unicorn/prefer-ternary': 'off',
			'no-console': 'off',
			'unicorn/prefer-code-point': 'off',
			'unicorn/filename-case': [
				'warn',
				{ cases: { pascalCase: true, kebabCase: true }, multipleFileExtensions: false }
			],
			'unicorn/no-null': 'off',
			'unicorn/prevent-abbreviations': 'off'
		}
	},
	{ rules: { 'unicorn/prefer-top-level-await': 'off' } }
);
