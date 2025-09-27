// Centralized runtime configuration for the demo UI.
// Now uses the new config module with schema validation.

import { OpenAPI } from "./lib/api-client";
import { config as appConfig } from "./config/index";

// Configure the generated API client base once.
OpenAPI.BASE = appConfig.api.baseUrl;

// Re-export for backward compatibility
export const apiBase = appConfig.api.baseUrl;
export const swaggerUiUrl = appConfig.api.swaggerUiUrl;
export const customChatGptUrl = appConfig.openai?.customUrl;

// Re-export the full config
export { appConfig as config };
