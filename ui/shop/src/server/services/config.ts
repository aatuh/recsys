export type AppConfig = {
  recsysBaseUrl: string;
  recsysNamespace: string;
  recsysOrgId: string;
  databaseUrl: string;
};

export function loadConfig(): AppConfig {
  const recsysBaseUrl =
    process.env.RECSYS_API_BASE_URL ?? "http://localhost:8000";
  const recsysNamespace = process.env.RECSYS_NAMESPACE ?? "default";
  const recsysOrgId =
    process.env.RECSYS_ORG_ID ??
    process.env.ORG_ID ??
    "00000000-0000-0000-0000-000000000001";
  const databaseUrl = process.env.DATABASE_URL ?? "file:./dev.db";
  return { recsysBaseUrl, recsysNamespace, recsysOrgId, databaseUrl };
}
