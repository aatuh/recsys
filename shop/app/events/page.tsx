import { prisma } from "@/server/db/client";

export default async function EventsPage() {
  const events = await prisma.event.findMany({
    orderBy: { ts: "desc" },
    take: 50,
    include: { user: true, product: true },
  });
  return (
    <main className="space-y-4">
      <h1 className="text-xl font-semibold">Events</h1>
      <form action="/api/events/flush" method="post">
        <button className="border rounded px-2 py-1 text-xs">
          Flush pending
        </button>
      </form>
      <form action="/api/events/retry-failed" method="post">
        <button className="border rounded px-2 py-1 text-xs">
          Retry failed
        </button>
      </form>
      <ul className="space-y-2">
        {events.map((e: any) => (
          <li key={e.id} className="border rounded p-3">
            <div className="text-xs text-gray-600">
              {e.ts.toISOString()} • {e.type} • {e.recsysStatus}
            </div>
            <div className="text-sm">
              <a className="underline" href={`/events/${e.id}`}>
                {e.user?.displayName} {e.product ? `→ ${e.product.name}` : ""}
              </a>
            </div>
          </li>
        ))}
      </ul>
    </main>
  );
}
