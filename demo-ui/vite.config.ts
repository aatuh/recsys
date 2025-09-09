import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

const apiTarget = "http://recsys-api:8000";

// Parse allowed hosts from environment variable
const getAllowedHosts = (): string[] => {
  const allowedHosts = process.env.VITE_ALLOWED_HOSTS;
  if (!allowedHosts) {
    throw new Error("VITE_ALLOWED_HOSTS is not set");
  }
  return allowedHosts.split(",").map((host) => host.trim());
};

export default defineConfig({
  plugins: [react()],
  server: {
    host: true,
    port: 3000,
    strictPort: true,
    allowedHosts: getAllowedHosts(),
    proxy: {
      "/api": {
        target: apiTarget,
        changeOrigin: true,
        secure: false,
      },
    },
  },
});
