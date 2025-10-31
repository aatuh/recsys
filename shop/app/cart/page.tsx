"use client";
import { useTelemetry } from "@/lib/telemetry/useTelemetry";
import { useEffect, useMemo, useState } from "react";

export default function CartPage() {
  const [userId, setUserId] = useState<string>("");
  const [data, setData] = useState<any>(null);
  const { emit } = useTelemetry();

  useEffect(() => {
    const u = window.localStorage.getItem("shop_user_id") || "";
    setUserId(u);
  }, []);

  useEffect(() => {
    if (!userId) return;
    fetch(`/api/cart?userId=${encodeURIComponent(userId)}`)
      .then((r) => r.json())
      .then(setData)
      .catch(() => setData(null));
  }, [userId]);

  const total = useMemo(() => data?.total ?? 0, [data]);

  if (!userId) return <p className="p-4">Select a user first.</p>;

  return (
    <main className="space-y-4">
      <h1 className="text-xl font-semibold">Cart</h1>
      <div className="text-sm">User: {userId}</div>
      <ul className="space-y-2">
        {data?.cart?.items?.map((it: any) => (
          <li
            key={it.id}
            className="border rounded p-3 flex items-center justify-between gap-3"
          >
            <div>
              <div className="font-medium">{it.product?.name}</div>
              <div className="text-xs text-gray-600">Qty: {it.qty}</div>
            </div>
            <div className="flex items-center gap-2">
              <div>${(it.qty * it.unitPrice).toFixed(2)}</div>
              <button
                className="border rounded px-2 py-1 text-xs"
                onClick={async () => {
                  await fetch(`/api/cart`, {
                    method: "PATCH",
                    headers: { "content-type": "application/json" },
                    body: JSON.stringify({
                      userId,
                      productId: it.productId,
                      qty: it.qty - 1,
                    }),
                  });
                  fetch(`/api/cart?userId=${encodeURIComponent(userId)}`)
                    .then((r) => r.json())
                    .then(setData);
                }}
              >
                -
              </button>
              <button
                className="border rounded px-2 py-1 text-xs"
                onClick={async () => {
                  await fetch(`/api/cart`, {
                    method: "PATCH",
                    headers: { "content-type": "application/json" },
                    body: JSON.stringify({
                      userId,
                      productId: it.productId,
                      qty: it.qty + 1,
                    }),
                  });
                  fetch(`/api/cart?userId=${encodeURIComponent(userId)}`)
                    .then((r) => r.json())
                    .then(setData);
                }}
              >
                +
              </button>
              <button
                className="border rounded px-2 py-1 text-xs"
                onClick={async () => {
                  await fetch(`/api/cart`, {
                    method: "PATCH",
                    headers: { "content-type": "application/json" },
                    body: JSON.stringify({
                      userId,
                      productId: it.productId,
                      qty: 0,
                    }),
                  });
                  fetch(`/api/cart?userId=${encodeURIComponent(userId)}`)
                    .then((r) => r.json())
                    .then(setData);
                }}
              >
                Remove
              </button>
            </div>
          </li>
        ))}
      </ul>
      <div className="font-semibold">Total: ${total.toFixed(2)}</div>
      {data?.cart?.items?.length ? (
        <form
          action="/api/checkout"
          method="post"
          onSubmit={async (e) => {
            e.preventDefault();
            await fetch("/api/checkout", {
              method: "POST",
              headers: { "content-type": "application/json" },
              body: JSON.stringify({ userId }),
            });
            void emit({ type: "purchase", value: total });
            // refresh
            fetch(`/api/cart?userId=${encodeURIComponent(userId)}`)
              .then((r) => r.json())
              .then(setData);
          }}
        >
          <button className="border rounded px-3 py-2 text-sm">Checkout</button>
        </form>
      ) : null}
    </main>
  );
}
