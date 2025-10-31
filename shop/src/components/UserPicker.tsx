"use client";
import { useEffect, useState } from "react";

export function UserPicker() {
  const [users, setUsers] = useState<any[]>([]);
  const [userId, setUserId] = useState<string>("");

  useEffect(() => {
    fetch(`/api/users?limit=100`)
      .then((r) => r.json())
      .then((d) => setUsers(d.items || []))
      .catch(() => setUsers([]));
    const saved = window.localStorage.getItem("shop_user_id") || "";
    setUserId(saved);
  }, []);

  const onChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const val = e.target.value;
    setUserId(val);
    window.localStorage.setItem("shop_user_id", val);
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
