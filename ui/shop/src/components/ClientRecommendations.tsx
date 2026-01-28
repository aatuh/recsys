"use client";
import { useEffect, useState } from "react";
import { RecommendationWidget } from "@/components/RecommendationWidget";
import {
  getStoredShopUserId,
  SHOP_USER_CHANGED_EVENT,
  SHOP_USER_STORAGE_KEY,
  ShopUserChangeDetail,
} from "@/lib/shopUser/client";

type Props = {
  surface: "home" | "pdp" | "cart" | "checkout";
  widget: string;
  k?: number;
  className?: string;
};

export function ClientRecommendations({
  surface,
  widget,
  k = 8,
  className = "",
}: Props) {
  const [userId, setUserId] = useState<string>("");

  useEffect(() => {
    const readUserId = () => {
      const id = getStoredShopUserId();
      setUserId(id);
    };

    const handleUserChange = (event: Event) => {
      const custom = event as CustomEvent<ShopUserChangeDetail>;
      const nextId = custom.detail?.userId ?? "";
      setUserId(nextId);
    };

    const handleStorage = (event: StorageEvent) => {
      if (event.key === SHOP_USER_STORAGE_KEY) {
        setUserId(event.newValue ?? "");
      }
    };

    readUserId();
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

  if (!userId) {
    return (
      <div className={`space-y-3 ${className}`}>
        <h3 className="text-lg font-semibold">Recommended for you</h3>
        <div className="text-sm text-gray-500">
          Select a user to see personalized recommendations.
        </div>
      </div>
    );
  }

  return (
    <RecommendationWidget
      userId={userId}
      surface={surface}
      widget={widget}
      k={k}
      className={className}
    />
  );
}
