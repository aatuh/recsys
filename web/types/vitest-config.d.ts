declare module "vitest/config" {
  export type UserConfig = Record<string, unknown>;
  export type UserConfigExport =
    | UserConfig
    | Promise<UserConfig>
    | ((env: unknown) => UserConfig | Promise<UserConfig>);

  export function defineConfig(config: UserConfigExport): UserConfigExport;
}
