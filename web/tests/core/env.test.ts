import { describe, expect, it } from "vitest";
import { z } from "zod";
import { createClientEnv } from "@api-boilerplate-core/env";

describe("createClientEnv", () => {
  it("applies defaults from the schema", () => {
    const env = createClientEnv(
      {
        NEXT_PUBLIC_API_URL: z.string().default("https://example.test"),
        NEXT_PUBLIC_FLAG: z.string().optional(),
      },
      {
        runtimeEnv: { NEXT_PUBLIC_FLAG: "on" },
        strict: true,
      }
    );

    expect(env.NEXT_PUBLIC_API_URL).toBe("https://example.test");
    expect(env.NEXT_PUBLIC_FLAG).toBe("on");
    expect(
      (env as Record<string, string | undefined>)["OTHER"]
    ).toBeUndefined();
  });

  it("rejects non-public keys in strict mode", () => {
    expect(() =>
      createClientEnv(
        {
          API_URL: z.string(),
        },
        {
          runtimeEnv: { API_URL: "https://example.test" },
          strict: true,
        }
      )
    ).toThrow(/non-public keys/i);
  });
});
