import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  experimental: {
    externalDir: true,
  },
  transpilePackages: [
    "@api-boilerplate-core/ui",
    "@api-boilerplate-core/theme",
    "@api-boilerplate-core/i18n-shared",
    "@api-boilerplate-core/legal",
    "@api-boilerplate-core/content",
    "@api-boilerplate-core/http",
    "@api-boilerplate-core/env",
    "@api-boilerplate-core/layouts",
    "@api-boilerplate-core/widgets",
    "@foo/config",
    "@foo/api-client",
    "@foo/domain",
    "@foo/domain-adapters",
    "@foo/hooks",
    "@foo/i18n",
    "@foo/legal",
  ],
};

export default nextConfig;
