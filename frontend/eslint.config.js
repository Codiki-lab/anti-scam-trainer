import js from '@eslint/js'
import globals from 'globals'
import reactHooks from 'eslint-plugin-react-hooks'
import reactRefresh from 'eslint-plugin-react-refresh'
import tseslint from 'typescript-eslint'

export default tseslint.config(
  { ignores: ['dist', 'node_modules', '.vite', 'playwright-report', 'test-results'] },
  {
    extends: [js.configs.recommended, ...tseslint.configs.recommended],
    files: ['**/*.{ts,tsx}'],
    languageOptions: { globals: globals.browser },
    plugins: { 'react-hooks': reactHooks, 'react-refresh': reactRefresh },
    rules: {
      ...reactHooks.configs.recommended.rules,
      'react-refresh/only-export-components': ['warn', { allowConstantExport: true }],
    },
  },
  {
    files: ['src/shared/**/*.{ts,tsx}'],
    rules: {
      'no-restricted-imports': [
        'error',
        {
          patterns: [
            {
              group: ['@/entities/**', '@/features/**', '@/widgets/**', '@/pages/**', '@/app/**'],
              message: 'shared не должен зависеть от бизнес-слоёв.',
            },
          ],
        },
      ],
    },
  },
  {
    files: ['src/entities/**/*.{ts,tsx}'],
    rules: {
      'no-restricted-imports': [
        'error',
        {
          patterns: [
            {
              group: ['@/features/**', '@/widgets/**', '@/pages/**', '@/app/**'],
              message: 'entities не должен зависеть от верхних слоёв.',
            },
            {
              group: ['@/entities/*/*', '@/shared/*/*'],
              message: 'Используйте публичный API соседнего слоя или slice.',
            },
          ],
        },
      ],
    },
  },
  {
    files: ['src/features/**/*.{ts,tsx}'],
    rules: {
      'no-restricted-imports': [
        'error',
        {
          patterns: [
            {
              group: ['@/widgets/**', '@/pages/**', '@/app/**'],
              message: 'features не должен зависеть от верхних слоёв.',
            },
            {
              group: ['@/entities/*/*', '@/features/**', '@/shared/*/*'],
              message: 'Используйте публичные API; внутри feature — относительные импорты.',
            },
          ],
        },
      ],
    },
  },
  {
    files: ['src/widgets/**/*.{ts,tsx}'],
    rules: {
      'no-restricted-imports': [
        'error',
        {
          patterns: [
            {
              group: ['@/pages/**', '@/app/**'],
              message: 'widgets не должен зависеть от pages или app.',
            },
            {
              group: ['@/entities/*/*', '@/features/*/*', '@/widgets/**', '@/shared/*/*'],
              message: 'Используйте публичные API; внутри widget — относительные импорты.',
            },
          ],
        },
      ],
    },
  },
  {
    files: ['src/pages/**/*.{ts,tsx}'],
    rules: {
      'no-restricted-imports': [
        'error',
        {
          patterns: [
            { group: ['@/app/**'], message: 'pages не должен зависеть от app.' },
            {
              group: [
                '@/entities/*/*',
                '@/features/*/*',
                '@/widgets/*/*',
                '@/pages/**',
                '@/shared/*/*',
              ],
              message: 'Используйте публичные API; внутри page — относительные импорты.',
            },
          ],
        },
      ],
    },
  },
  {
    files: ['src/app/**/*.{ts,tsx}'],
    rules: {
      'no-restricted-imports': [
        'error',
        {
          patterns: [
            {
              group: [
                '@/entities/*/*',
                '@/features/*/*',
                '@/widgets/*/*',
                '@/pages/*/*',
                '@/shared/*/*',
              ],
              message: 'App собирает приложение только через публичные API модулей.',
            },
          ],
        },
      ],
    },
  },
)
