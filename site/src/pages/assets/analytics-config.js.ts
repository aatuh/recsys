import type { APIRoute } from "astro";

const measurementId = import.meta.env.PUBLIC_GA_MEASUREMENT_ID?.trim() ?? "";

export const GET: APIRoute = () =>
  new Response(`window.RECSYS_ANALYTICS_CONFIG = ${JSON.stringify({ measurementId })};\n`, {
    headers: {
      "Cache-Control": "public, max-age=300",
      "Content-Type": "application/javascript; charset=utf-8",
    },
  });
