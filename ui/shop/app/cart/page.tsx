"use client";
import { useTelemetry } from "@/lib/telemetry/useTelemetry";
import { useEffect, useMemo, useState } from "react";
import {
  getStoredShopUserId,
  SHOP_USER_CHANGED_EVENT,
  SHOP_USER_STORAGE_KEY,
  ShopUserChangeDetail,
} from "@/lib/shopUser/client";

export default function CartPage() {
  const [userId, setUserId] = useState<string>("");
  const [data, setData] = useState<{
    cart: {
      items: Array<{
        id: string;
        product: { id: string; name: string; price: number; imageUrl?: string };
        qty: number;
        unitPrice: number;
        productId: string;
      }>;
    };
    total: number;
  } | null>(null);
  const { emit } = useTelemetry();

  useEffect(() => {
    const readUser = () => setUserId(getStoredShopUserId());

    const handleUserChange = (event: Event) => {
      const custom = event as CustomEvent<ShopUserChangeDetail>;
      setUserId(custom.detail?.userId ?? "");
    };

    const handleStorage = (event: StorageEvent) => {
      if (event.key === SHOP_USER_STORAGE_KEY) {
        setUserId(event.newValue ?? "");
      }
    };

    readUser();
    window.addEventListener(
      SHOP_USER_CHANGED_EVENT,
      handleUserChange as EventListener
    );
    window.addEventListener("storage", handleStorage);

    return () => {
      window.removeEventListener(
        SHOP_USER_CHANGED_EVENT,
        handleUserChange as EventListener
      );
      window.removeEventListener("storage", handleStorage);
    };
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
        {data?.cart?.items?.map(
          (it: {
            id: string;
            product: {
              id: string;
              name: string;
              price: number;
              imageUrl?: string;
            };
            qty: number;
            unitPrice: number;
            productId: string;
          }) => (
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
          )
        )}
      </ul>
      <div className="font-semibold">Total: ${total.toFixed(2)}</div>
      {data?.cart?.items?.length ? (
        <form
          action="/api/checkout"
          method="post"
          onSubmit={async (e) => {
            e.preventDefault();
            const orderRes = await fetch("/api/checkout", {
              method: "POST",
              headers: { "content-type": "application/json" },
              body: JSON.stringify({ userId }),
            });

            if (orderRes.ok) {
              const orderData = await orderRes.json();
              const orderId = orderData.order?.id;

              // Emit line-item purchase events
              for (const item of data.cart.items) {
                void emit({
                  type: "purchase",
                  productId: item.productId,
                  value: item.qty,
                  meta: {
                    surface: "checkout",
                    order_id: orderId,
                    unit_price: item.unitPrice,
                    currency: "USD",
                    line_item_id: item.id,
                  },
                });
              }

              // Emit order summary event (custom type)
              void emit({
                type: "custom",
                value: total,
                meta: {
                  surface: "checkout",
                  kind: "order",
                  order_id: orderId,
                  currency: "USD",
                  items: data.cart.items.map(
                    (item: {
                      productId: string;
                      qty: number;
                      unitPrice: number;
                    }) => ({
                      item_id: item.productId,
                      qty: item.qty,
                      unit_price: item.unitPrice,
                    })
                  ),
                },
              });
            }

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
