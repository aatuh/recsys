"use client";
import { useEffect } from "react";

export function ViewEventOnMount({ productId }: { productId: string }) {
  useEffect(() => {
    const userId = window.localStorage.getItem("shop_user_id");
    if (!userId) return;
    const url = new URL(window.location.href);
    const recommended = url.searchParams.get("rec") === "1";
    const payload = {
      userId,
      productId,
      type: "view",
      ts: new Date().toISOString(),
      meta: recommended ? { recommended: true } : undefined,
    };
    if (navigator.sendBeacon) {
      const blob = new Blob([JSON.stringify(payload)], {
        type: "application/json",
      });
      navigator.sendBeacon("/api/events", blob);
    } else {
      fetch("/api/events", {
        method: "POST",
        headers: { "content-type": "application/json" },
        body: JSON.stringify(payload),
      }).catch(() => {});
    }
  }, [productId]);
  return null;
}
