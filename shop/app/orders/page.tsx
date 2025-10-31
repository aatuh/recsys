import { prisma } from "@/server/db/client";

export default async function OrdersPage() {
  const orders = await prisma.order.findMany({
    orderBy: { createdAt: "desc" },
    include: { user: true, items: { include: { product: true } } },
    take: 50,
  });
  return (
    <main className="space-y-4">
      <h1 className="text-xl font-semibold">Orders</h1>
      {orders.length === 0 ? (
        <div className="text-sm text-muted-foreground">No orders yet.</div>
      ) : (
        <ul className="space-y-2">
          {orders.map((o: any) => (
            <li key={o.id} className="border rounded p-3">
              <div className="text-sm font-medium">
                {o.id} • ${o.total} {o.currency}
              </div>
              <div className="text-xs text-gray-600">
                {o.user?.displayName} • {o.createdAt.toISOString()}
              </div>
            </li>
          ))}
        </ul>
      )}
    </main>
  );
}
