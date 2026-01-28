import { publicEnv } from "./public-env";

const defaultAppUrl = "http://localhost:3000";

export const appBaseUrl = publicEnv.NEXT_PUBLIC_APP_URL
  ? publicEnv.NEXT_PUBLIC_APP_URL.replace(/\/+$/, "")
  : defaultAppUrl;

export const appDefaults = {
  orgId: publicEnv.NEXT_PUBLIC_APP_ORG_ID || "org-demo",
  namespace: publicEnv.NEXT_PUBLIC_APP_NAMESPACE || "default",
};
