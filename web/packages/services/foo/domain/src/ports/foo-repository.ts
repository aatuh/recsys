import type {
  Foo,
  FooCreateInput,
  FooListRequest,
  FooListResponse,
  FooUpdateInput,
} from "../models/foo";

export type FooRepository = {
  list: (filters: FooListRequest) => Promise<FooListResponse>;
  get: (id: string) => Promise<Foo>;
  create: (input: FooCreateInput) => Promise<Foo>;
  update: (id: string, input: FooUpdateInput) => Promise<Foo>;
  remove: (id: string) => Promise<void>;
};
