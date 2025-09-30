import { envSchema, createAppConfig, type AppConfig } from "./schema";

/**
 * Load and validate environment variables at startup.
 * Throws descriptive errors for missing or invalid values.
 */
function loadConfig(): AppConfig {
  try {
    // Extract VITE_ prefixed environment variables
    const envVars = Object.fromEntries(
      Object.entries((import.meta as any).env || {}).filter(([key]) =>
        key.startsWith("VITE_")
      )
    );

    // Validate environment variables
    const validatedEnv = envSchema.parse(envVars);

    // Transform into application configuration
    return createAppConfig(validatedEnv);
  } catch (error) {
    if (error instanceof Error) {
      throw new Error(
        `Configuration validation failed: ${error.message}. ` +
          "Please check your .env file and ensure all required variables are set."
      );
    }
    throw error;
  }
}

/**
 * Application configuration loaded at startup.
 * Contains validated environment variables and derived settings.
 */
export const config: AppConfig = loadConfig();

// Re-export types for convenience
export type { AppConfig } from "./schema";
