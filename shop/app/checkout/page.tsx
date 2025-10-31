"use client";
import { useEffect, useState } from "react";

export default function CheckoutPage() {
  const [userId, setUserId] = useState<string>("");
  const [result, setResult] = useState<any>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    const u = window.localStorage.getItem("shop_user_id") || "";
    setUserId(u);
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
