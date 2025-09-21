import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

const apiTarget = process.env.VITE_API_HOST;
const isVitest =
  process.env.VITEST === "true" || process.env.NODE_ENV === "test";

// Parse allowed hosts from environment variable
const getAllowedHosts = (): string[] => {
  const allowedHosts = process.env.VITE_ALLOWED_HOSTS;
  if (!allowedHosts) {
    if (isVitest) {
      return ["localhost"];
    }
    throw new Error("VITE_ALLOWED_HOSTS is not set");
  }
  return allowedHosts.split(",").map((host) => host.trim());
};

export default defineConfig({
  plugins: [react()],
  test: {
    environment: "jsdom",
    setupFiles: ["src/test/setup.ts"],
  },
  server: isVitest
    ? ({} as any)
    : {
        host: true,
        port: 3000,
        strictPort: true,
        allowedHosts: getAllowedHosts(),
        proxy: {
          "/api": {
            target: apiTarget,
            changeOrigin: true,
            secure: false,
            rewrite: (path) => path.replace(/^\/api/, ""),
          },
        },
      },
});
