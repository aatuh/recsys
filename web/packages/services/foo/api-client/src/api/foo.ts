import { createHttpClient, type HttpClient } from "@api-boilerplate-core/http";
import { apiBase } from "@foo/config";
import type { components } from "./types";

export type FooDTO = components["schemas"]["types.FooDTO"];
export type CreateFooDTO = components["schemas"]["types.CreateFooDTO"];
export type UpdateFooDTO = components["schemas"]["types.UpdateFooDTO"];
export type FooListResponse = components["schemas"]["types.FooListResponse"];

export type FooListParams = {
  orgId: string;
  namespace: string;
  limit?: number;
  offset?: number;
  search?: string;
};

export function createFooApi(
  client: HttpClient = createHttpClient({ baseUrl: apiBase })
) {
  return {
    async create(input: CreateFooDTO): Promise<FooDTO> {
      const { data } = await client.request<FooDTO>({
        path: "/foo",
        method: "POST",
        body: input,
      });
      return data;
    },
    async get(id: string): Promise<FooDTO> {
      const { data } = await client.request<FooDTO>({
        path: `/foo/${id}`,
        method: "GET",
      });
      return data;
    },
    async update(id: string, input: UpdateFooDTO): Promise<FooDTO> {
      const { data } = await client.request<FooDTO>({
        path: `/foo/${id}`,
        method: "PUT",
        body: input,
      });
      return data;
    },
    async remove(id: string): Promise<void> {
      await client.request({
        path: `/foo/${id}`,
        method: "DELETE",
      });
    },
    async list(params: FooListParams): Promise<FooListResponse> {
      const searchParams = new URLSearchParams();
      searchParams.set("org_id", params.orgId);
      searchParams.set("namespace", params.namespace);
      if (typeof params.limit === "number") {
        searchParams.set("limit", String(params.limit));
      }
      if (typeof params.offset === "number") {
        searchParams.set("offset", String(params.offset));
      }
      if (params.search) {
        searchParams.set("search", params.search);
      }
      const qs = searchParams.toString();
      const { data } = await client.request<FooListResponse>({
        path: `/foo${qs ? `?${qs}` : ""}`,
        method: "GET",
      });
      return data;
    },
  };
}
