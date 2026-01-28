export const SHOP_USER_STORAGE_KEY = "shop_user_id";
export const SHOP_USER_CHANGED_EVENT = "shop:user-changed";

export type ShopUserChangeDetail = {
  userId: string;
};

export function getStoredShopUserId(): string {
  if (typeof window === "undefined") {
    return "";
  }
  return window.localStorage.getItem(SHOP_USER_STORAGE_KEY) ?? "";
}

export function setStoredShopUserId(userId: string) {
  if (typeof window === "undefined") {
    return;
  }
  window.localStorage.setItem(SHOP_USER_STORAGE_KEY, userId);
  const event = new CustomEvent<ShopUserChangeDetail>(SHOP_USER_CHANGED_EVENT, {
    detail: { userId },
  });
  window.dispatchEvent(event);
}
