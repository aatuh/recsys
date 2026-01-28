"use client";
import { useState } from "react";
import { useTelemetry } from "@/lib/telemetry/useTelemetry";
import { BanditMeta } from "@/lib/recommendations/bandit";

export function AddToCartButton({
  productId,
  surface = "home",
  widget,
  recommended = false,
  rank,
  unitPrice,
  currency = "USD",
  coldStart = false,
  banditMeta,
}: {
  productId: string;
  surface?: "home" | "pdp" | "cart" | "checkout" | "products";
  widget?: string;
  recommended?: boolean;
  rank?: number;
  unitPrice?: number;
  currency?: string;
  coldStart?: boolean;
  banditMeta?: BanditMeta;
}) {
  const [loading, setLoading] = useState(false);
  const [ok, setOk] = useState(false);
  const { emit } = useTelemetry();

  const onAdd = async () => {
    const userId = window.localStorage.getItem("shop_user_id");
    if (!userId) {
      alert("Select a user first");
      return;
    }
    setLoading(true);
    setOk(false);
    try {
      const res = await fetch("/api/cart", {
        method: "POST",
        headers: { "content-type": "application/json" },
        body: JSON.stringify({ userId, productId, qty: 1 }),
      });
      if (res.ok) {
        setOk(true);
        // Get cart ID for metadata
        const cartRes = await fetch(
          `/api/cart?userId=${encodeURIComponent(userId)}`
        );
        const cartData = await cartRes.json();

        void emit({
          type: "add",
          productId,
          value: 1,
          meta: {
            surface,
            widget,
            recommended,
            rank,
            cart_id: cartData.cart?.id,
            unit_price: unitPrice,
            currency,
            cold_start: coldStart || undefined,
            bandit_policy_id: banditMeta?.policyId,
            bandit_request_id: banditMeta?.requestId,
            bandit_algorithm: banditMeta?.algorithm,
            bandit_bucket: banditMeta?.bucket,
            bandit_explore:
              banditMeta?.explore !== undefined
                ? banditMeta.explore
                : undefined,
            bandit_experiment: banditMeta?.experiment,
            bandit_variant: banditMeta?.variant,
          },
        });
      }
    } finally {
      setLoading(false);
      setTimeout(() => setOk(false), 1200);
    }
  };

  return (
    <button
      disabled={loading}
      className="text-xs border rounded px-2 py-1 hover:bg-gray-50 disabled:opacity-50"
      onClick={onAdd}
    >
      {ok ? "Added" : loading ? "Adding..." : "Add to cart"}
    </button>
  );
}
