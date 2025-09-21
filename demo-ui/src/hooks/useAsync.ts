import { useCallback, useMemo, useState } from "react";
import type { AsyncState } from "../types/ui";

export function useAsync<TArgs extends any[], TRes>(
  fn: (...args: TArgs) => Promise<TRes>
) {
  const [state, setState] = useState<AsyncState<TRes>>({
    loading: false,
    error: null,
    data: null,
  });

  const run = useCallback(
    async (...args: TArgs): Promise<TRes | null> => {
      setState((s) => ({ ...s, loading: true, error: null }));
      try {
        const res = await fn(...args);
        setState({ loading: false, error: null, data: res });
        return res;
      } catch (e: any) {
        const err = e instanceof Error ? e : new Error(String(e));
        setState({ loading: false, error: err, data: null });
        return null;
      }
    },
    [fn]
  );

  return useMemo(() => ({ ...state, run }), [state, run]);
}
