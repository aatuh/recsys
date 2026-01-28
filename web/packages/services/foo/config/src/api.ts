import { publicEnv } from "./public-env";

export const apiBase = publicEnv.NEXT_PUBLIC_API_BASE_URL
  ? publicEnv.NEXT_PUBLIC_API_BASE_URL.replace(/\/+$/, "")
  : "http://localhost:8000/api/v1";
