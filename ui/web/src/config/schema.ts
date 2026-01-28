import { z } from "zod";

/**
 * Environment variable schema with runtime validation.
 * All VITE_ prefixed variables are available in the browser.
 */
export const envSchema = z.object({
  // API Configuration
  VITE_API_BASE_URL: z.string().default("/api"),
  VITE_API_HOST: z.string().url().optional(),
  VITE_SWAGGER_UI_URL: z.string().url().default("http://localhost:8081"),

  // OpenAI Configuration (optional)
  VITE_CUSTOM_CHATGPT_URL: z.string().url().optional(),

  // Development Configuration
  VITE_ALLOWED_HOSTS: z.string().default("localhost,127.0.0.1"),

  // Logging Configuration
  VITE_LOG_LEVEL: z.enum(["debug", "info", "warn", "error"]).default("info"),
  VITE_ENABLE_ANALYTICS: z
    .string()
    .transform((val: string) => val === "true")
    .default("false"),

  // Embeddings Configuration
  VITE_EMBEDDING_MODEL: z.string().default("Xenova/all-MiniLM-L6-v2"),
  VITE_EMBEDDING_DIMENSION: z
    .string()
    .transform((val: string) => parseInt(val, 10))
    .default("384"),
});

export type EnvSchema = z.infer<typeof envSchema>;

/**
 * Application configuration derived from environment variables.
 */
export interface AppConfig {
  api: {
    baseUrl: string;
    host?: string;
    swaggerUiUrl: string;
  };
  openai?: {
    customUrl: string;
  };
  development: {
    allowedHosts: string[];
  };
  logging: {
    level: "debug" | "info" | "warn" | "error";
    enableAnalytics: boolean;
  };
  embeddings: {
    model: string;
    dimension: number;
  };
}

/**
 * Transform validated environment variables into typed configuration.
 */
export function createAppConfig(env: EnvSchema): AppConfig {
  return {
    api: {
      baseUrl: env.VITE_API_BASE_URL,
      host: env.VITE_API_HOST,
      swaggerUiUrl: env.VITE_SWAGGER_UI_URL,
    },
    openai: env.VITE_CUSTOM_CHATGPT_URL
      ? { customUrl: env.VITE_CUSTOM_CHATGPT_URL }
      : undefined,
    development: {
      allowedHosts: env.VITE_ALLOWED_HOSTS.split(",").map((host: string) =>
        host.trim()
      ),
    },
    logging: {
      level: env.VITE_LOG_LEVEL,
      enableAnalytics: env.VITE_ENABLE_ANALYTICS,
    },
    embeddings: {
      model: env.VITE_EMBEDDING_MODEL,
      dimension: env.VITE_EMBEDDING_DIMENSION,
    },
  };
}
