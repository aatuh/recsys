import { describe, expect, it, vi } from "vitest";
import { createFooService } from "@foo/domain";
import type { FooRepository } from "@foo/domain";

describe("createFooService", () => {
  it("rejects missing required fields", async () => {
    const repo: FooRepository = {
      list: vi.fn(),
      get: vi.fn(),
      create: vi.fn(),
      update: vi.fn(),
      remove: vi.fn(),
    };

    const service = createFooService(repo);

    await expect(
      service.create({ orgId: "", namespace: "default", name: "Demo" })
    ).rejects.toThrow("Org ID is required");

    await expect(
      service.create({ orgId: "org-demo", namespace: "", name: "Demo" })
    ).rejects.toThrow("Namespace is required");
  });

  it("trims and forwards inputs", async () => {
    const repo: FooRepository = {
      list: vi.fn().mockResolvedValue({ data: [], meta: { total: 0, count: 0, limit: 50, offset: 0 } }),
      get: vi.fn(),
      create: vi.fn().mockResolvedValue({
        id: "foo1",
        orgId: "org-demo",
        namespace: "default",
        name: "Demo",
        createdAt: "2025-01-01T00:00:00Z",
        updatedAt: "2025-01-01T00:00:00Z",
      }),
      update: vi.fn(),
      remove: vi.fn(),
    };

    const service = createFooService(repo);

    await service.create({ orgId: " org-demo ", namespace: " default ", name: " Demo " });

    expect(repo.create).toHaveBeenCalledWith({
      orgId: "org-demo",
      namespace: "default",
      name: "Demo",
    });
  });
});
