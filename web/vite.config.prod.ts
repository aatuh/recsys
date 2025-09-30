import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

const apiTarget = process.env.VITE_API_HOST || "http://localhost:8081";
const isVitest =
  process.env.VITEST === "true" || process.env.NODE_ENV === "test";

// Parse allowed hosts from environment variable
const getAllowedHosts = (): string[] => {
  const allowedHosts = process.env.VITE_ALLOWED_HOSTS;
  if (!allowedHosts) {
    if (isVitest) {
      return ["localhost"];
    }
    // Default allowed hosts for production
    return ["localhost", "0.0.0.0", "127.0.0.1"];
  }
  return allowedHosts.split(",").map((host) => host.trim());
};

export default defineConfig({
  plugins: [react()],
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
        // Production server configuration
        headers: {
          // Content Security Policy
          "Content-Security-Policy": [
            "default-src 'self'",
            "script-src 'self' 'unsafe-inline' 'unsafe-eval'",
            "style-src 'self' 'unsafe-inline'",
            "img-src 'self' data: blob: https:",
            "font-src 'self' data:",
            "connect-src 'self' ws: wss: https:",
            "object-src 'none'",
            "base-uri 'self'",
            "form-action 'self'",
            "frame-ancestors 'none'",
          ].join("; "),
          // Security headers
          "X-Content-Type-Options": "nosniff",
          "X-Frame-Options": "DENY",
          "X-XSS-Protection": "1; mode=block",
          "Referrer-Policy": "strict-origin-when-cross-origin",
          // Caching headers for static assets
          "Cache-Control": "public, max-age=31536000, immutable",
        },
      },
  // Production build optimizations
  build: {
    // Enable source maps for production debugging
    sourcemap: true,
    // Optimize chunk splitting
    rollupOptions: {
      output: {
        manualChunks: {
          // Vendor chunks for better caching
          vendor: ["react", "react-dom"],
          query: ["@tanstack/react-query"],
          ui: ["react-markdown", "dompurify", "marked"],
          ml: ["@xenova/transformers"],
        },
      },
    },
    // Enable minification
    minify: "terser",
    terserOptions: {
      compress: {
        drop_console: true, // Remove console.log in production
        drop_debugger: true,
      },
    },
    // Set chunk size warning limit
    chunkSizeWarningLimit: 1000,
  },
  // Preview server configuration (for production)
  preview: {
    host: true,
    port: 3000,
    strictPort: true,
    headers: {
      // Production headers for preview server
      "Content-Security-Policy": [
        "default-src 'self'",
        "script-src 'self' 'unsafe-inline' 'unsafe-eval'",
        "style-src 'self' 'unsafe-inline'",
        "img-src 'self' data: blob: https:",
        "font-src 'self' data:",
        "connect-src 'self' ws: wss: https:",
        "object-src 'none'",
        "base-uri 'self'",
        "form-action 'self'",
        "frame-ancestors 'none'",
      ].join("; "),
      "X-Content-Type-Options": "nosniff",
      "X-Frame-Options": "DENY",
      "X-XSS-Protection": "1; mode=block",
      "Referrer-Policy": "strict-origin-when-cross-origin",
      // Caching headers
      "Cache-Control": "public, max-age=3600",
    },
  },
});
