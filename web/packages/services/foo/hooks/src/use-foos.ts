import { useCallback, useEffect, useMemo, useState } from "react";
import type {
  Foo,
  FooCreateInput,
  FooListRequest,
  FooListResponse,
  FooRepository,
  FooService,
  FooUpdateInput,
  ListMeta,
} from "@foo/domain";
import { createFooService } from "@foo/domain";
import { createFooRepository } from "@foo/domain-adapters";
import { resolveHttpErrorMessage } from "@api-boilerplate-core/http/errors";
import { useLocale } from "@foo/i18n";

export type FooListStatus = "idle" | "loading" | "error";
export type FooActionStatus = "idle" | "saving" | "error";

type FooDeps = {
  repo?: FooRepository;
  service?: FooService;
};

export function useFooList(filters: FooListRequest, deps?: FooDeps) {
  const { t, tRaw } = useLocale();
  const service = useMemo(
    () =>
      deps?.service ?? createFooService(deps?.repo ?? createFooRepository()),
    [deps?.repo, deps?.service]
  );
  const [items, setItems] = useState<Foo[]>([]);
  const [meta, setMeta] = useState<ListMeta | null>(null);
  const [status, setStatus] = useState<FooListStatus>("idle");
  const [error, setError] = useState<string | null>(null);
  const [refreshKey, setRefreshKey] = useState(0);

  const reload = useCallback(() => {
    setRefreshKey((prev) => prev + 1);
  }, []);

  const stableFilters = useMemo(() => {
    const next: FooListRequest = {
      orgId: filters.orgId,
      namespace: filters.namespace,
    };
    if (filters.limit !== undefined) {
      next.limit = filters.limit;
    }
    if (filters.offset !== undefined) {
      next.offset = filters.offset;
    }
    if (filters.search !== undefined) {
      next.search = filters.search;
    }
    return next;
  }, [
    filters.orgId,
    filters.namespace,
    filters.limit,
    filters.offset,
    filters.search,
  ]);

  useEffect(() => {
    let active = true;
    const load = async () => {
      setStatus("loading");
      setError(null);
      try {
        const response: FooListResponse = await service.list(stableFilters);
        if (!active) return;
        setItems(response.data);
        setMeta(response.meta);
        setStatus("idle");
      } catch (err) {
        if (!active) return;
        setStatus("error");
        setError(resolveHttpErrorMessage(err, tRaw, t("app.errors.loadFoos")));
      }
    };
    load();
    return () => {
      active = false;
    };
  }, [refreshKey, service, stableFilters, t, tRaw]);

  return { items, meta, status, error, reload, setItems };
}

type FooActionDeps = FooDeps & {
  onSuccess?: (foo: Foo | null) => void;
};

export function useFooActions(deps?: FooActionDeps) {
  const { t, tRaw } = useLocale();
  const service = useMemo(
    () =>
      deps?.service ?? createFooService(deps?.repo ?? createFooRepository()),
    [deps?.repo, deps?.service]
  );
  const [status, setStatus] = useState<FooActionStatus>("idle");
  const [error, setError] = useState<string | null>(null);

  const createFoo = useCallback(
    async (input: FooCreateInput) => {
      setStatus("saving");
      setError(null);
      try {
        const foo = await service.create(input);
        setStatus("idle");
        deps?.onSuccess?.(foo);
        return foo;
      } catch (err) {
        setStatus("error");
        setError(resolveHttpErrorMessage(err, tRaw, t("app.errors.saveFoo")));
        return null;
      }
    },
    [deps, service, t, tRaw]
  );

  const updateFoo = useCallback(
    async (id: string, input: FooUpdateInput) => {
      setStatus("saving");
      setError(null);
      try {
        const foo = await service.update(id, input);
        setStatus("idle");
        deps?.onSuccess?.(foo);
        return foo;
      } catch (err) {
        setStatus("error");
        setError(resolveHttpErrorMessage(err, tRaw, t("app.errors.saveFoo")));
        return null;
      }
    },
    [deps, service, t, tRaw]
  );

  const removeFoo = useCallback(
    async (id: string) => {
      setStatus("saving");
      setError(null);
      try {
        await service.remove(id);
        setStatus("idle");
        deps?.onSuccess?.(null);
        return true;
      } catch (err) {
        setStatus("error");
        setError(resolveHttpErrorMessage(err, tRaw, t("app.errors.deleteFoo")));
        return false;
      }
    },
    [deps, service, t, tRaw]
  );

  const clearError = useCallback(() => setError(null), []);

  return { createFoo, updateFoo, removeFoo, status, error, clearError };
}
