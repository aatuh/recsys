"use client";
import { useState } from "react";
import { useTelemetry } from "@/lib/telemetry/useTelemetry";

export function AddToCartButton({
  productId,
  surface = "home",
  widget,
  recommended = false,
  rank,
  unitPrice,
  currency = "USD",
}: {
  productId: string;
  surface?: "home" | "pdp" | "cart" | "checkout" | "products";
  widget?: string;
  recommended?: boolean;
  rank?: number;
  unitPrice?: number;
  currency?: string;
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
