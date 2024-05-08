module.exports = require("typescript-eslint").config(
  {
    files: [
      "src/**/*.ts",
      "src/**/*.tsx",
    ],
    extends: [
      require("@eslint/js").configs.recommended,
      ...require("typescript-eslint").configs.recommended,
      ...require("typescript-eslint").configs.stylistic,
      require("eslint-plugin-functional/flat").configs.lite,
      require("eslint-plugin-import").configs.typescript,
      require("eslint-config-prettier"),
    ],
    languageOptions: {
      parser: require("typescript-eslint").parser,
      parserOptions: {
        project: "./tsconfig.json",
      },
      globals: {
        BigInt: true,
        console: true,
        WebAssembly: true,
      },
    },
    ignores: [
      "**/node_modules/**",
      "**/build/**",
    ],
    plugins: {
      "import": require("eslint-plugin-import"),
      "functional": require("eslint-plugin-functional/flat"),
      "@typescript-eslint": require("typescript-eslint").plugin,
    },
    rules: {
      "@typescript-eslint/explicit-module-boundary-types": "off",
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
    },
  },
);