"use client";
import { useEffect, useState } from "react";
import {
  getStoredShopUserId,
  setStoredShopUserId,
} from "@/lib/shopUser/client";

export function UserPicker() {
  const [users, setUsers] = useState<
    Array<{ id: string; displayName: string }>
  >([]);
  const [userId, setUserId] = useState<string>("");

  useEffect(() => {
    fetch(`/api/users?limit=100`)
      .then((r) => r.json())
      .then((d) => setUsers(d.items || []))
      .catch(() => setUsers([]));
    const saved = getStoredShopUserId();
    setUserId(saved);
  }, []);

  const onChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const val = e.target.value;
    setUserId(val);
    setStoredShopUserId(val);
    if (val) {
      fetch(`/api/users/${encodeURIComponent(val)}/upsert`, {
        method: "POST",
      }).catch(() => {});
    }
  };

  return (
    <select
      className="border rounded px-2 py-1 text-sm"
      value={userId}
      onChange={onChange}
    >
      <option value="">Select user</option>
      {users.map((u) => (
        <option key={u.id} value={u.id}>
          {u.displayName}
        </option>
      ))}
    </select>
  );
}
