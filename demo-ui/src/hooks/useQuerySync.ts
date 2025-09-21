import { useEffect, useState } from "react";

export function useQuerySync<T extends string>(
  key: string,
  initial: T,
  valid?: readonly T[],
  opts?: { storageKey?: string }
) {
  const [value, setValue] = useState<T>(() => {
    const url = new URL(window.location.href);
    const q = url.searchParams.get(key) as T | null;
    if (q && (!valid || valid.includes(q))) return q;
    if (opts?.storageKey) {
      try {
        const saved = localStorage.getItem(opts.storageKey) as T | null;
        if (saved && (!valid || valid.includes(saved))) return saved;
      } catch {
        // ignore persistence errors
      }
    }
    return initial;
  });

  useEffect(() => {
    const url = new URL(window.location.href);
    const q = url.searchParams.get(key) as T | null;
    if (q && (!valid || valid.includes(q)) && q !== value) {
      setValue(q);
    }
  }, []);

  useEffect(() => {
    const url = new URL(window.location.href);
    url.searchParams.set(key, value);
    window.history.replaceState({}, "", url.toString());
    if (opts?.storageKey) {
      try {
        localStorage.setItem(opts.storageKey, value);
      } catch {
        // ignore persistence errors
      }
    }
  }, [key, value, opts?.storageKey]);

  return [value, setValue] as const;
}
