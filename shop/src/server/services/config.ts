export type AppConfig = {
  recsysBaseUrl: string;
  recsysNamespace: string;
  databaseUrl: string;
};

export function loadConfig(): AppConfig {
  const recsysBaseUrl =
    process.env.RECSYS_API_BASE_URL ?? "http://localhost:8000";
  const recsysNamespace = process.env.RECSYS_NAMESPACE ?? "default";
  const databaseUrl = process.env.DATABASE_URL ?? "file:./dev.db";
  return { recsysBaseUrl, recsysNamespace, databaseUrl };
}
