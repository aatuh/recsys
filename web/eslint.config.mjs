import path from "node:path";
import { fileURLToPath } from "node:url";
import { defineConfig, globalIgnores } from "eslint/config";
import nextVitals from "eslint-config-next/core-web-vitals";
import nextTs from "eslint-config-next/typescript";
import reactHooks from "eslint-plugin-react-hooks";

const configDir = path.dirname(fileURLToPath(import.meta.url));

const eslintConfig = defineConfig([
  ...nextVitals,
  ...nextTs,
  // Override default ignores of eslint-config-next.
  globalIgnores([
    // Default ignores of eslint-config-next:
    ".next/**",
    "out/**",
    "build/**",
    "next-env.d.ts",
    ".pnpmfile.cjs",
    "node_modules/**",
    ".pnpm-store/**",
    "pnpm-lock.yaml",
    "coverage/**",
  ]),
  {
    settings: {
      next: {
        rootDir: [configDir],
      },
    },
  },
  {
    plugins: {
      "react-hooks": reactHooks,
    },
    rules: {
      "react-hooks/rules-of-hooks": "error",
      "react-hooks/exhaustive-deps": "warn",
      "react-hooks/set-state-in-effect": "warn",
    },
  },
  {
    files: ["packages/services/foo/domain/**/*.{ts,tsx}"],
    rules: {
      "no-restricted-imports": [
        "error",
        {
          patterns: [
            {
              group: [
                "@foo/api-client",
                "@foo/config",
                "@foo/domain-adapters",
                "@foo/hooks",
                "@foo/i18n",
                "@foo/legal",
                "@api-boilerplate-core/content",
                "@api-boilerplate-core/env",
                "@api-boilerplate-core/http",
                "@api-boilerplate-core/i18n-shared",
                "@api-boilerplate-core/legal",
                "@api-boilerplate-core/layouts",
                "@api-boilerplate-core/theme",
                "@api-boilerplate-core/ui",
                "@api-boilerplate-core/widgets",
              ],
              message:
                "Domain layer must stay framework-agnostic and adapter-free.",
            },
          ],
        },
      ],
    },
  },
  {
    files: ["packages/services/foo/domain-adapters/**/*.{ts,tsx}"],
    rules: {
      "no-restricted-imports": [
        "error",
        {
          patterns: [
            {
              group: [
                "@foo/config",
                "@foo/hooks",
                "@foo/i18n",
                "@foo/legal",
                "@api-boilerplate-core/content",
                "@api-boilerplate-core/env",
                "@api-boilerplate-core/i18n-shared",
                "@api-boilerplate-core/legal",
                "@api-boilerplate-core/layouts",
                "@api-boilerplate-core/theme",
                "@api-boilerplate-core/ui",
                "@api-boilerplate-core/widgets",
              ],
              message:
                "Domain adapters should not depend on UI or app-level packages.",
            },
          ],
        },
      ],
    },
  },
]);

export default eslintConfig;
