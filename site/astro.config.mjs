import sitemap from "@astrojs/sitemap";
import { defineConfig } from "astro/config";

const site = "https://recsys.app";

const technicalDocsPages = [
  "/documentation/technical/",
  "/documentation/technical/developer-quickstart/",
  "/documentation/technical/local-end-to-end/",
  "/documentation/technical/architecture/",
  "/documentation/technical/artifacts-and-pipelines/",
  "/documentation/technical/integration/",
  "/documentation/technical/evaluation-decisions/",
  "/documentation/technical/operations/",
  "/documentation/technical/reference/",
  "/documentation/technical/reference/api/",
  "/documentation/technical/reference/config/",
  "/documentation/technical/reference/data-contracts/",
  "/documentation/technical/security/",
  "/documentation/technical/commercial/licensing/",
  "/documentation/technical/commercial/pricing/",
  "/documentation/technical/commercial/procurement/",
].map((path) => new URL(path, site).toString());

export default defineConfig({
  site,
  output: "static",
  outDir: "../.site",
  trailingSlash: "always",
  i18n: {
    defaultLocale: "en",
    locales: ["en", "fi"],
    routing: {
      prefixDefaultLocale: false,
      redirectToDefaultLocale: false,
    },
  },
  integrations: [
    sitemap({
      customPages: technicalDocsPages,
      i18n: {
        defaultLocale: "en",
        locales: {
          en: "en",
          fi: "fi",
        },
      },
    }),
  ],
});
