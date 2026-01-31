// Commented: Kept for reference
// import { describe, expect, it, vi } from "vitest";
// import { createFooRepository } from "@foo/domain-adapters";
// import type {
//   FooDTO,
//   FooListResponse,
// } from "@foo/api-client/foo";
// import { createFooApi } from "@foo/api-client/foo";

// describe("createFooRepository", () => {
//   it("maps API DTOs into domain models", async () => {
//     const dto: FooDTO = {
//       id: "foo1",
//       org_id: "org-demo",
//       namespace: "default",
//       name: "Demo",
//       created_at: "2025-01-01T00:00:00Z",
//       updated_at: "2025-01-01T01:00:00Z",
//     };
//     const listResponse: FooListResponse = {
//       data: [dto],
//       meta: { total: 1, count: 1, limit: 50, offset: 0 },
//     };

//     const api = {
//       list: vi.fn().mockResolvedValue(listResponse),
//       get: vi.fn().mockResolvedValue(dto),
//       create: vi.fn().mockResolvedValue(dto),
//       update: vi.fn().mockResolvedValue(dto),
//       remove: vi.fn().mockResolvedValue(undefined),
//     } as ReturnType<typeof createFooApi>;

//     const repo = createFooRepository({ api });
//     const result = await repo.list({ orgId: "org-demo", namespace: "default" });

//     expect(result.data[0]).toEqual({
//       id: "foo1",
//       orgId: "org-demo",
//       namespace: "default",
//       name: "Demo",
//       createdAt: "2025-01-01T00:00:00Z",
//       updatedAt: "2025-01-01T01:00:00Z",
//     });
//     expect(result.meta.total).toBe(1);
//   });
// });
