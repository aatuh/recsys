import { createClientEnv } from "@api-boilerplate-core/env";
import { z } from "zod";

const requireApiBase = process.env["NODE_ENV"] === "production";

export const publicEnv = createClientEnv(
  {
    NEXT_PUBLIC_API_BASE_URL: requireApiBase
      ? z.string().url()
      : z.string().url().optional(),
    NEXT_PUBLIC_APP_URL: z.string().url().optional(),
    NEXT_PUBLIC_APP_ORG_ID: z.string().optional(),
    NEXT_PUBLIC_APP_NAMESPACE: z.string().optional(),
    NEXT_PUBLIC_DEFAULT_LOCALE: z.string().optional(),
    NEXT_PUBLIC_SUPPORTED_LOCALES: z.string().optional(),
    NEXT_PUBLIC_LOCALE_AUTO_DETECT: z.string().optional(),
  },
  {
    label: "API Boilerplate public environment variables",
    ...(requireApiBase ? { strict: true } : {}),
    runtimeEnv: {
      NEXT_PUBLIC_API_BASE_URL: process.env["NEXT_PUBLIC_API_BASE_URL"],
      NEXT_PUBLIC_APP_URL: process.env["NEXT_PUBLIC_APP_URL"],
      NEXT_PUBLIC_APP_ORG_ID: process.env["NEXT_PUBLIC_APP_ORG_ID"],
      NEXT_PUBLIC_APP_NAMESPACE: process.env["NEXT_PUBLIC_APP_NAMESPACE"],
      NEXT_PUBLIC_DEFAULT_LOCALE: process.env["NEXT_PUBLIC_DEFAULT_LOCALE"],
      NEXT_PUBLIC_SUPPORTED_LOCALES:
        process.env["NEXT_PUBLIC_SUPPORTED_LOCALES"],
      NEXT_PUBLIC_LOCALE_AUTO_DETECT:
        process.env["NEXT_PUBLIC_LOCALE_AUTO_DETECT"],
    },
  }
);
