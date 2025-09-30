import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

const isVitest =
  process.env.VITEST === "true" || process.env.NODE_ENV === "test";

const getAllowedHosts = (): string[] => {
  const allowedHosts = process.env.VITE_ALLOWED_HOSTS;
  if (!allowedHosts) {
    if (isVitest) return ["localhost"];
    return ["localhost", "0.0.0.0", "127.0.0.1"];
  }
  return allowedHosts.split(",").map((host) => host.trim());
};

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  // Cast to unknown then to object to avoid any usage in TS config
  server: (isVitest
    ? ({} as unknown)
    : {
        host: true,
        port: 3001,
        strictPort: true,
        allowedHosts: getAllowedHosts(),
      }) as Record<string, unknown>,
});
