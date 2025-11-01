"use client";
import { useEffect, useState } from "react";
import {
  getStoredShopUserId,
  SHOP_USER_CHANGED_EVENT,
  SHOP_USER_STORAGE_KEY,
  ShopUserChangeDetail,
} from "@/lib/shopUser/client";

export default function CheckoutPage() {
  const [userId, setUserId] = useState<string>("");
  const [result, setResult] = useState<{
    orderId: string;
    total: number;
  } | null>(null);
  const [loading, setLoading] = useState(false);

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

  const onCheckout = async () => {
    if (!userId) {
      alert("Select a user first");
      return;
    }
    setLoading(true);
    setResult(null);
    try {
      const res = await fetch("/api/checkout", {
        method: "POST",
        headers: { "content-type": "application/json" },
        body: JSON.stringify({ userId }),
      });
      const data = await res.json();
      setResult(data);
    } finally {
      setLoading(false);
    }
  };

  return (
    <main className="space-y-4">
      <h1 className="text-xl font-semibold">Checkout</h1>
      <button
        className="border rounded px-3 py-2 text-sm"
        disabled={loading}
        onClick={onCheckout}
      >
        {loading ? "Processing..." : "Confirm purchase"}
      </button>
      {result && (
        <div className="text-sm">
          Order {result.orderId}, total ${result.total}
        </div>
      )}
    </main>
  );
}
