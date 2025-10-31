"use client";
import { useEffect } from "react";
import { useTelemetry } from "@/lib/telemetry/useTelemetry";

export default function ClickTelemetry() {
  const { emit } = useTelemetry();
  useEffect(() => {
    function handler(e: MouseEvent) {
      const t = e.target as HTMLElement | null;
      if (!t) return;
      const a = t.closest("a") as HTMLAnchorElement | null;
      if (!a) return;
      try {
        const url = new URL(a.href, window.location.origin);
        const path = url.pathname;
        // Only business links
        const isBusiness =
          path.startsWith("/products") ||
          path === "/cart" ||
          path === "/orders";
        if (!isBusiness) return;
        const productId = a.getAttribute("data-product-id") || undefined;
        void emit({
          type: "click",
          productId,
          meta: { href: path, text: a.textContent?.trim() },
        });
      } catch {}
    }
    document.addEventListener("click", handler, { capture: true });
    return () =>
      document.removeEventListener("click", handler, { capture: true } as any);
  }, [emit]);
  return null;
}
