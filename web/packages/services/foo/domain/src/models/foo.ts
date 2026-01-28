export type Foo = {
  id: string;
  orgId: string;
  namespace: string;
  name: string;
  createdAt: string;
  updatedAt: string;
};

export type FooCreateInput = {
  orgId: string;
  namespace: string;
  name: string;
};

export type FooUpdateInput = {
  name: string;
};

export type FooListRequest = {
  orgId: string;
  namespace: string;
  limit?: number;
  offset?: number;
  search?: string;
};

export type ListMeta = {
  total: number;
  count: number;
  limit: number;
  offset: number;
  filters?: Record<string, string[]>;
  search?: string;
};

export type FooListResponse = {
  data: Foo[];
  meta: ListMeta;
};
