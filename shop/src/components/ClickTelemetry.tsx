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
        const isRecommended = a.hasAttribute("data-recommended");
        const widget = a.getAttribute("data-widget") || undefined;
        const rank = a.getAttribute("data-rank")
          ? parseInt(a.getAttribute("data-rank")!)
          : undefined;

        void emit({
          type: "click",
          productId,
          meta: {
            surface: path.startsWith("/products")
              ? "pdp"
              : path === "/cart"
              ? "cart"
              : path === "/orders"
              ? "checkout"
              : "home",
            widget,
            recommended: isRecommended,
            rank,
            href: path,
            text: a.textContent?.trim(),
          },
        });
      } catch {}
    }
    document.addEventListener("click", handler, { capture: true });
    return () =>
      document.removeEventListener("click", handler, { capture: true });
  }, [emit]);
  return null;
}
