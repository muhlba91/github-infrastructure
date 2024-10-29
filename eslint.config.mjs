// @ts-check

import { fixupConfigRules, fixupPluginRules } from '@eslint/compat';
import { FlatCompat } from "@eslint/eslintrc";

import eslint from '@eslint/js';
import tseslint from 'typescript-eslint';
import eslintFunctional from 'eslint-plugin-functional';
import eslintImport from 'eslint-plugin-import';
import importRecommendedConfig from "eslint-plugin-import/config/recommended.js";
import importTypeScriptConfig from "eslint-plugin-import/config/typescript.js";
import eslintPrettier from 'eslint-config-prettier';

const flatCompat = new FlatCompat();

export default tseslint.config(
  eslint.configs.recommended,
  ...tseslint.configs.recommended,
  ...tseslint.configs.stylistic,
  {
    files: [
      "src/**/*.ts",
      "src/**/*.tsx",
    ],
    extends: [
      eslintFunctional.configs.externalTypeScriptRecommended,
      eslintFunctional.configs.lite,
      // eslintFunctional.configs.recommended,
      // eslintFunctional.configs.stylistic,
      eslintPrettier,
      ...fixupConfigRules(flatCompat.config(importRecommendedConfig)),
      ...fixupConfigRules(flatCompat.config(importTypeScriptConfig)),
    ],
    languageOptions: {
      parserOptions: {
        projectService: true,
      },
    },
    plugins: {
      import: fixupPluginRules(eslintImport),
    },
    rules: {
      "import/order": [
        "error",
        {
          "newlines-between": "always",
          "alphabetize": {
            "order": "asc",
          },
        },
      ],
      "sort-imports": [
        "error",
        {
          "ignoreDeclarationSort": true,
          "ignoreCase": true,
        },
      ],
      "functional/prefer-immutable-types": "off",
      "functional/no-return-void": "off",
    }
  },
);
