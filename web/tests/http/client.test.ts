import { describe, expect, it, vi } from "vitest";
import { createHttpClient } from "@api-boilerplate-core/http/client";

const okJson = (payload: unknown, init?: ResponseInit) =>
  new Response(JSON.stringify(payload), {
    status: 200,
    headers: { "Content-Type": "application/json" },
    ...init,
  });

const problemJson = (payload: unknown, init?: ResponseInit) =>
  new Response(JSON.stringify(payload), {
    status: init?.status ?? 400,
    headers: { "Content-Type": "application/problem+json" },
    ...init,
  });

describe("createHttpClient", () => {
  it("retries on retryable status and succeeds", async () => {
    const fetchImpl = vi
      .fn()
      .mockResolvedValueOnce(new Response("server error", { status: 500 }))
      .mockResolvedValueOnce(okJson({ hello: "world" }));

    const client = createHttpClient({
      baseUrl: "http://example.com",
      retries: 1,
      retryBackoffMs: 0,
      fetchImpl,
    });

    const result = await client.request<{ hello: string }>({ path: "/ping" });

    expect(result.data.hello).toBe("world");
    expect(fetchImpl).toHaveBeenCalledTimes(2);
  });

  it("parses problem+json into HttpError", async () => {
    const fetchImpl = vi
      .fn()
      .mockResolvedValueOnce(
        problemJson(
          { title: "Bad Request", detail: "Invalid payload" },
          { status: 400 }
        )
      );
    const client = createHttpClient({
      baseUrl: "http://example.com",
      retries: 0,
      fetchImpl,
    });

    await expect(() =>
      client.request({ path: "/ping", method: "POST", body: { a: 1 } })
    ).rejects.toMatchObject({
      problem: { detail: "Invalid payload" },
      status: 400,
    });
  });

  it("marks timeouts as HttpError with isTimeout flag", async () => {
    vi.useFakeTimers();
    const fetchImpl = vi.fn().mockImplementation((_, init?: RequestInit) => {
      return new Promise<Response>((resolve, reject) => {
        init?.signal?.addEventListener("abort", () => {
          reject(new DOMException("Aborted", "AbortError"));
        });
        setTimeout(() => resolve(okJson({})), 1000);
      });
    });

    const client = createHttpClient({
      baseUrl: "http://example.com",
      retries: 0,
      timeoutMs: 5,
      fetchImpl,
    });
    const promise = client.request({ path: "/slow" });
    const assertion = expect(promise).rejects.toMatchObject({
      isTimeout: true,
    });
    await vi.advanceTimersByTimeAsync(10);
    await assertion;
    vi.useRealTimers();
  });
});
