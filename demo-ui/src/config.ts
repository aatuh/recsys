// Centralized runtime configuration for the demo UI.
// Now uses the new config module with schema validation.

import { config as appConfig } from "./config/index";

function stripTrailingSlash(url: string): string {
  return url.endsWith("/") ? url.slice(0, -1) : url;
}

// Re-export for backward compatibility
export const apiBase = (() => {
  const baseUrl = appConfig.api.baseUrl.startsWith("/")
    ? appConfig.api.baseUrl.substring(1)
    : appConfig.api.baseUrl;

  const absoluteBase = appConfig.api.host
    ? new URL(baseUrl, appConfig.api.host).toString()
    : appConfig.api.baseUrl;

  return stripTrailingSlash(absoluteBase);
})();
export const swaggerUiUrl = appConfig.api.swaggerUiUrl;
export const customChatGptUrl = appConfig.openai?.customUrl;

// Re-export the full config
export { appConfig as config };
