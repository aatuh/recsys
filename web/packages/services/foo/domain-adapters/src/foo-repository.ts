// Commented: Kept for reference
// import {
//   createFooApi,
//   type CreateFooDTO,
//   type FooDTO,
//   type FooListResponse as ApiFooListResponse,
//   type UpdateFooDTO,
// } from "@foo/api-client/foo";
// import type {
//   Foo,
//   FooCreateInput,
//   FooListRequest,
//   FooListResponse,
//   FooRepository,
//   FooUpdateInput,
// } from "@foo/domain";

// function mapFoo(dto: FooDTO): Foo {
//   return {
//     id: requireField(dto.id, "foo.id"),
//     orgId: requireField(dto.org_id, "foo.org_id"),
//     namespace: requireField(dto.namespace, "foo.namespace"),
//     name: requireField(dto.name, "foo.name"),
//     createdAt: requireField(dto.created_at, "foo.created_at"),
//     updatedAt: requireField(dto.updated_at, "foo.updated_at"),
//   };
// }

// function toCreateDto(input: FooCreateInput): CreateFooDTO {
//   return {
//     org_id: input.orgId,
//     namespace: input.namespace,
//     name: input.name,
//   };
// }

// function toUpdateDto(input: FooUpdateInput): UpdateFooDTO {
//   return {
//     name: input.name,
//   };
// }

// function mapList(response: ApiFooListResponse): FooListResponse {
//   const data = response.data ?? [];
//   const meta = response.meta ?? {};
//   return {
//     data: data.map(mapFoo),
//     meta: {
//       total: meta.total ?? data.length,
//       count: meta.count ?? data.length,
//       limit: meta.limit ?? data.length,
//       offset: meta.offset ?? 0,
//       ...(meta.filters !== undefined ? { filters: meta.filters } : {}),
//       ...(meta.search !== undefined ? { search: meta.search } : {}),
//     },
//   };
// }

// function requireField(value: string | undefined, label: string): string {
//   const trimmed = value?.trim();
//   if (!trimmed) {
//     throw new Error(`[foo] Missing ${label}`);
//   }
//   return trimmed;
// }

// type FooRepoDeps = {
//   api?: ReturnType<typeof createFooApi>;
// };

// export function createFooRepository(deps: FooRepoDeps = {}): FooRepository {
//   const api = deps.api ?? createFooApi();
//   return {
//     async list(filters: FooListRequest): Promise<FooListResponse> {
//       const response = await api.list({
//         orgId: filters.orgId,
//         namespace: filters.namespace,
//         ...(filters.limit !== undefined ? { limit: filters.limit } : {}),
//         ...(filters.offset !== undefined ? { offset: filters.offset } : {}),
//         ...(filters.search !== undefined ? { search: filters.search } : {}),
//       });
//       return mapList(response);
//     },
//     async get(id: string): Promise<Foo> {
//       const dto = await api.get(id);
//       return mapFoo(dto);
//     },
//     async create(input: FooCreateInput): Promise<Foo> {
//       const dto = await api.create(toCreateDto(input));
//       return mapFoo(dto);
//     },
//     async update(id: string, input: FooUpdateInput): Promise<Foo> {
//       const dto = await api.update(id, toUpdateDto(input));
//       return mapFoo(dto);
//     },
//     async remove(id: string): Promise<void> {
//       await api.remove(id);
//     },
//   };
// }
