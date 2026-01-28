import { describe, expect, it, vi, beforeEach } from "vitest";
import { createFooApi } from "@foo/api-client/foo";
import type { HttpClient } from "@api-boilerplate-core/http/client";

describe("foo API client", () => {
  const requestMock = vi.fn();
  const fakeClient: HttpClient = {
    request: requestMock as HttpClient["request"],
  };
  const api = createFooApi(fakeClient);

  beforeEach(() => {
    requestMock.mockReset();
  });

  it("calls POST /foo with body", async () => {
    requestMock.mockResolvedValue({
      data: { id: "foo1" },
      status: 201,
      headers: new Headers(),
    });

    await api.create({
      org_id: "org-demo",
      namespace: "default",
      name: "Demo",
    });

    expect(requestMock).toHaveBeenCalledWith(
      expect.objectContaining({
        path: "/foo",
        method: "POST",
        body: { org_id: "org-demo", namespace: "default", name: "Demo" },
      })
    );
  });

  it("calls GET /foo with filters", async () => {
    requestMock.mockResolvedValue({
      data: { data: [], meta: { total: 0, count: 0, limit: 10, offset: 5 } },
      status: 200,
      headers: new Headers(),
    });

    await api.list({
      orgId: "org-demo",
      namespace: "default",
      search: "demo",
      limit: 10,
      offset: 5,
    });

    expect(requestMock).toHaveBeenCalledWith(
      expect.objectContaining({
        path: "/foo?org_id=org-demo&namespace=default&limit=10&offset=5&search=demo",
        method: "GET",
      })
    );
  });

  it("calls DELETE /foo/{id}", async () => {
    requestMock.mockResolvedValue({
      data: null,
      status: 204,
      headers: new Headers(),
    });

    await api.remove("foo123");

    expect(requestMock).toHaveBeenCalledWith(
      expect.objectContaining({
        path: "/foo/foo123",
        method: "DELETE",
      })
    );
  });
});
