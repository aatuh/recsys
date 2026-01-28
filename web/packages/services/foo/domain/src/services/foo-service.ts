import type {
  Foo,
  FooCreateInput,
  FooListRequest,
  FooListResponse,
  FooUpdateInput,
} from "../models/foo";
import type { FooRepository } from "../ports/foo-repository";

export type FooService = {
  list: (filters: FooListRequest) => Promise<FooListResponse>;
  get: (id: string) => Promise<Foo>;
  create: (input: FooCreateInput) => Promise<Foo>;
  update: (id: string, input: FooUpdateInput) => Promise<Foo>;
  remove: (id: string) => Promise<void>;
};

function requireField(value: string, label: string): string {
  const trimmed = value.trim();
  if (!trimmed) {
    throw new Error(`${label} is required`);
  }
  return trimmed;
}

export function createFooService(repo: FooRepository): FooService {
  return {
    async list(filters: FooListRequest) {
      const normalizedSearch = filters.search?.trim();
      return repo.list({
        orgId: requireField(filters.orgId, "Org ID"),
        namespace: requireField(filters.namespace, "Namespace"),
        ...(normalizedSearch ? { search: normalizedSearch } : {}),
        ...(filters.limit !== undefined ? { limit: filters.limit } : {}),
        ...(filters.offset !== undefined ? { offset: filters.offset } : {}),
      });
    },
    async get(id: string) {
      return repo.get(requireField(id, "Foo ID"));
    },
    async create(input: FooCreateInput) {
      return repo.create({
        orgId: requireField(input.orgId, "Org ID"),
        namespace: requireField(input.namespace, "Namespace"),
        name: requireField(input.name, "Name"),
      });
    },
    async update(id: string, input: FooUpdateInput) {
      return repo.update(requireField(id, "Foo ID"), {
        name: requireField(input.name, "Name"),
      });
    },
    async remove(id: string) {
      return repo.remove(requireField(id, "Foo ID"));
    },
  };
}
