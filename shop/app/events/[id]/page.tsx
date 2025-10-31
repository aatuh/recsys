import { prisma } from "@/server/db/client";

export default async function EventDetail({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  const e = await prisma.event.findUnique({
    where: { id },
    include: { user: true, product: true },
  });
  if (!e) return <div className="p-6">Not found</div>;
  return (
    <main className="space-y-4 p-6">
      <h1 className="text-xl font-semibold">Event {e.id}</h1>
      <div className="text-sm">{e.ts.toISOString()}</div>
      <div className="text-sm">Type: {e.type}</div>
      <div className="text-sm">User: {e.user?.displayName}</div>
      <div className="text-sm">Product: {e.product?.name}</div>
      <pre className="text-xs p-3 border rounded bg-gray-50">
        {e.metaText || "{}"}
      </pre>
    </main>
  );
}
