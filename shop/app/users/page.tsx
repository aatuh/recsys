import { prisma } from "@/server/db/client";

export default async function UsersPage() {
  const users = await prisma.user.findMany({ orderBy: { createdAt: "desc" } });
  return (
    <main className="space-y-4">
      <h1 className="text-xl font-semibold">Users</h1>
      <ul className="space-y-2">
        {users.map((u: { id: string; displayName: string }) => (
          <li key={u.id} className="border rounded p-3">
            <div className="font-medium">{u.displayName}</div>
            <div className="text-xs text-gray-600">{u.id}</div>
          </li>
        ))}
      </ul>
    </main>
  );
}
