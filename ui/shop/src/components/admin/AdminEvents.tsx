"use client";
import { useCallback, useEffect, useMemo, useState } from "react";
import { useToast } from "@/components/ToastProvider";

type EventRow = {
  id: string;
  type: string;
  userId: string;
  productId?: string;
  value: number;
  ts: string;
  recsysStatus: string;
  metaText?: string | null;
};

function parseCsv(value: string): string[] {
  return value
    .split(/[,;]/)
    .map((entry) => entry.trim())
    .filter(Boolean);
}

function formatMetaPreview(raw?: string | null, max = 80): string {
  if (!raw) return "";
  try {
    const parsed = JSON.parse(raw) as Record<string, unknown>;
    const flattened = Object.entries(parsed)
      .map(([key, val]) => `${key}:${typeof val === "string" ? val : JSON.stringify(val)}`)
      .join(" ");
    return flattened.length > max ? `${flattened.slice(0, max)}…` : flattened;
  } catch {
    const trimmed = raw.replace(/\s+/g, " ").trim();
    return trimmed.length > max ? `${trimmed.slice(0, max)}…` : trimmed;
  }
}

export default function AdminEvents() {
  const toast = useToast();
  const [items, setItems] = useState<EventRow[]>([]);
  const [total, setTotal] = useState(0);
  const [limit, setLimit] = useState(50);
  const [offset, setOffset] = useState(0);
  const [type, setType] = useState("");
  const [selected, setSelected] = useState<Record<string, boolean>>({});
  const [loading, setLoading] = useState(false);
  const [seedCount, setSeedCount] = useState(40);
  const [seedTypes, setSeedTypes] = useState("view,click,add,purchase");
  const [seedSurfaces, setSeedSurfaces] = useState("home,pdp,cart");
  const [seedWidgets, setSeedWidgets] = useState(
    "home_top_picks,similar_items"
  );
  const [seedIncludeBandit, setSeedIncludeBandit] = useState(true);
  const [seedIncludeColdStart, setSeedIncludeColdStart] = useState(true);
  const [showSeedOptions, setShowSeedOptions] = useState(false);
  const [eventForm, setEventForm] = useState({
    userId: "",
    productId: "",
    type: "view",
    value: "1",
    ts: "",
    meta: "",
  });
  const [submittingEvent, setSubmittingEvent] = useState(false);

  function setEventField<K extends keyof typeof eventForm>(
    key: K,
    value: string
  ) {
    setEventForm((prev) => ({ ...prev, [key]: value }));
  }
  function prefillMeta() {
    const sample = {
      surface: "home",
      widget: "home_top_picks",
      rank: 1,
      request_id: `manual-${Date.now()}`,
      bandit_policy_id: "manual_explore_default",
      recommended: true,
    };
    setEventForm((prev) => ({
      ...prev,
      meta: JSON.stringify(sample, null, 2),
    }));
  }

  const selectedIds = useMemo(
    () => Object.keys(selected).filter((k) => selected[k]),
    [selected]
  );

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const url = new URL(`/api/events`, window.location.origin);
      url.searchParams.set("limit", String(limit));
      url.searchParams.set("offset", String(offset));
      if (type) url.searchParams.set("type", type);
      const res = await fetch(url);
      const data = await res.json();
      setItems(
        (data.items || []).map(
          (e: {
            id: string;
            type: string;
            userId: string;
            productId: string | null;
            ts: string;
            recsysStatus: string;
            value?: number;
            metaText?: string | null;
          }) => ({
            id: e.id,
            type: e.type,
            userId: e.userId,
            productId: e.productId,
            value: typeof e.value === "number" ? e.value : 1,
            ts: e.ts,
            recsysStatus: e.recsysStatus,
            metaText: e.metaText ?? null,
          })
        )
      );
      setTotal(data.total || 0);
      setSelected({});
    } finally {
      setLoading(false);
    }
  }, [limit, offset, type]);

  useEffect(() => {
    load();
  }, [limit, offset, load]);

  async function onDeleteSelected() {
    if (!selectedIds.length) return;
    await fetch("/api/events/batch", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ action: "delete", ids: selectedIds }),
    });
    toast(`Deleted ${selectedIds.length} events`);
    load();
  }
  async function onRetryFailed() {
    await fetch("/api/events/retry-failed", { method: "POST" });
    toast("Retry triggered");
    load();
  }
  async function onFlushPending() {
    await fetch("/api/events/flush", { method: "POST" });
    toast("Flush triggered");
    load();
  }
  async function onDeletePending() {
    if (!confirm("Delete ALL pending events?")) return;
    await fetch("/api/events/batch", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ action: "delete-pending" }),
    });
    toast("All pending events deleted");
    load();
  }

  async function onDeleteFromRecsys() {
    if (!selectedIds.length) {
      alert("Please select events to delete from Recsys");
      return;
    }
    if (!confirm(`Delete ${selectedIds.length} events from Recsys?`)) return;

    try {
      const response = await fetch("/api/events/batch", {
        method: "POST",
        headers: { "content-type": "application/json" },
        body: JSON.stringify({
          action: "delete-from-recsys",
          ids: selectedIds,
        }),
      });

      if (!response.ok) {
        throw new Error(
          `Failed to delete events from Recsys: ${response.status}`
        );
      }

      const result = await response.json();
      toast(`Deleted ${result.deleted || 0} events from Recsys`);
      load();
    } catch (error) {
      console.error("Failed to delete events from Recsys:", error);
      toast("Failed to delete events from Recsys");
    }
  }

  async function onSeedEvents(count: number) {
    const payload: Record<string, unknown> = { count };
    const typeList = parseCsv(seedTypes)
      .map((entry) => entry.toLowerCase())
      .filter((entry) =>
        ["view", "click", "add", "purchase", "custom"].includes(entry)
      );
    if (typeList.length > 0) payload.types = typeList;

    const surfaces = parseCsv(seedSurfaces);
    if (surfaces.length > 0) payload.surfaces = surfaces;

    const widgets = parseCsv(seedWidgets);
    if (widgets.length > 0) payload.widgets = widgets;

    payload.includeBandit = seedIncludeBandit;
    payload.includeColdStart = seedIncludeColdStart;

    try {
      const response = await fetch("/api/events/seed", {
        method: "POST",
        headers: { "content-type": "application/json" },
        body: JSON.stringify(payload),
      });
      if (!response.ok) {
        const result = await response.json().catch(() => ({}));
        throw new Error(
          (result && result.error) ||
            `Failed to seed events (status ${response.status})`
        );
      }
      toast(`Seeded ${count} events`);
      load();
    } catch (error) {
      console.error("Failed to seed events", error);
      toast(error instanceof Error ? error.message : "Failed to seed events");
    }
  }

  async function onCreateEvent(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (!eventForm.userId.trim()) {
      toast("User ID is required");
      return;
    }
    const payload: Record<string, unknown> = {
      userId: eventForm.userId.trim(),
      type: eventForm.type as "view" | "click" | "add" | "purchase" | "custom",
    };
    if (eventForm.productId.trim()) {
      payload.productId = eventForm.productId.trim();
    }
    if (eventForm.value.trim()) {
      const numeric = Number(eventForm.value);
      if (!Number.isFinite(numeric)) {
        toast("Value must be a number");
        return;
      }
      payload.value = numeric;
    }
    if (eventForm.ts.trim()) {
      const timestamp = Date.parse(eventForm.ts);
      if (Number.isNaN(timestamp)) {
        toast("Timestamp must be ISO-8601 or yyyy-mm-ddTHH:MM format");
        return;
      }
      payload.ts = new Date(timestamp).toISOString();
    }
    if (eventForm.meta.trim()) {
      try {
        payload.meta = JSON.parse(eventForm.meta);
      } catch (error) {
        console.error("Invalid meta JSON", error);
        toast("Meta must be valid JSON");
        return;
      }
    }
    setSubmittingEvent(true);
    try {
      const response = await fetch("/api/events", {
        method: "POST",
        headers: { "content-type": "application/json" },
        body: JSON.stringify(payload),
      });
      if (!response.ok) {
        const result = await response.json().catch(() => ({}));
        throw new Error(
          (result && result.error) ||
            `Failed to create event (status ${response.status})`
        );
      }
      toast("Event created");
      setEventForm((prev) => ({
        ...prev,
        productId: "",
        value: "1",
        meta: "",
        ts: "",
      }));
      load();
    } catch (error) {
      console.error("Failed to create event", error);
      toast(
        error instanceof Error ? error.message : "Failed to create event"
      );
    } finally {
      setSubmittingEvent(false);
    }
  }

  async function onNuke() {
    if (!confirm("Delete ALL events?")) return;
    await fetch("/api/admin/nuke", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ tables: ["events"] }),
    });
    toast("All events deleted");
    load();
  }

  return (
    <section className="space-y-4">
      <div className="flex gap-2 items-center">
        <select
          className="border p-2 text-sm"
          value={type}
          onChange={(e) => setType(e.target.value)}
        >
          <option value="">All types</option>
          <option value="view">view</option>
          <option value="click">click</option>
          <option value="add">add</option>
          <option value="purchase">purchase</option>
          <option value="custom">custom</option>
        </select>
        <button
          className="border rounded px-3 py-2 text-sm"
          onClick={() => {
            setOffset(0);
            load();
          }}
          disabled={loading}
        >
          Filter
        </button>
        <div className="ml-auto flex gap-2 items-center">
          <label className="text-xs text-gray-600">Seed count</label>
          <input
            className="border p-1 w-20 text-sm"
            type="number"
            min={1}
            max={500}
            value={seedCount}
            onChange={(e) => setSeedCount(parseInt(e.target.value || "0", 10))}
          />
          <button
            className="border rounded px-3 py-2 text-sm"
            onClick={() =>
              onSeedEvents(Math.max(1, Math.min(500, seedCount || 0)))
            }
          >
            Seed events
          </button>
          <button
            type="button"
            className="border rounded px-3 py-2 text-xs"
            onClick={() => setShowSeedOptions((prev) => !prev)}
          >
            {showSeedOptions ? "Hide seed options" : "Seed options"}
          </button>
          <button
            className="border rounded px-3 py-2 text-sm"
            onClick={onFlushPending}
          >
            Flush pending
          </button>
          <button
            className="border rounded px-3 py-2 text-sm"
            onClick={onRetryFailed}
          >
            Retry failed
          </button>
          <button
            className="border rounded px-3 py-2 text-sm bg-red-50 text-red-700 hover:bg-red-100"
            onClick={onDeletePending}
          >
            Delete pending
          </button>
          <button
            className="border rounded px-3 py-2 text-sm"
            onClick={onDeleteSelected}
            disabled={!selectedIds.length}
          >
            Delete selected
          </button>
          <button
            className="border rounded px-3 py-2 text-sm bg-orange-50 text-orange-700 hover:bg-orange-100"
            onClick={onDeleteFromRecsys}
            disabled={!selectedIds.length}
          >
            Delete from Recsys
          </button>
          <button
            className="border rounded px-3 py-2 text-sm bg-red-50 text-red-700 hover:bg-red-100"
            onClick={onNuke}
          >
            Nuke events
          </button>
        </div>
      </div>

      {showSeedOptions ? (
        <div className="border rounded p-4 bg-gray-50 space-y-3">
          <h3 className="text-sm font-semibold text-gray-700">
            Seed generator options
          </h3>
          <div className="grid gap-3 md:grid-cols-2">
            <label className="text-xs text-gray-700">
              Event types
              <input
                className="mt-1 w-full border rounded p-2 text-sm"
                value={seedTypes}
                onChange={(e) => setSeedTypes(e.target.value)}
              />
            </label>
            <label className="text-xs text-gray-700">
              Surfaces
              <input
                className="mt-1 w-full border rounded p-2 text-sm"
                value={seedSurfaces}
                onChange={(e) => setSeedSurfaces(e.target.value)}
              />
            </label>
            <label className="text-xs text-gray-700">
              Widgets
              <input
                className="mt-1 w-full border rounded p-2 text-sm"
                value={seedWidgets}
                onChange={(e) => setSeedWidgets(e.target.value)}
              />
            </label>
            <label className="flex items-center gap-2 text-xs text-gray-700">
              <input
                type="checkbox"
                checked={seedIncludeBandit}
                onChange={(e) => setSeedIncludeBandit(e.target.checked)}
              />
              Include bandit metadata
            </label>
            <label className="flex items-center gap-2 text-xs text-gray-700">
              <input
                type="checkbox"
                checked={seedIncludeColdStart}
                onChange={(e) => setSeedIncludeColdStart(e.target.checked)}
              />
              Insert cold-start impressions
            </label>
          </div>
        </div>
      ) : null}

      <table className="w-full text-sm border">
        <thead>
          <tr className="bg-gray-50">
            <th className="p-2 border">
              <input
                type="checkbox"
                checked={
                  items.length > 0 && selectedIds.length === items.length
                }
                onChange={(e) => {
                  const next: Record<string, boolean> = {};
                  if (e.target.checked)
                    items.forEach((r) => (next[r.id] = true));
                  setSelected(next);
                }}
              />
            </th>
            <th className="p-2 border">Type</th>
            <th className="p-2 border">User</th>
            <th className="p-2 border">Product</th>
            <th className="p-2 border">Value</th>
            <th className="p-2 border">Meta</th>
            <th className="p-2 border">Status</th>
            <th className="p-2 border">Timestamp</th>
          </tr>
        </thead>
        <tbody>
          {items.map((r) => (
            <tr key={r.id} className="border-b">
              <td className="p-2 border">
                <input
                  type="checkbox"
                  checked={!!selected[r.id]}
                  onChange={(e) =>
                    setSelected((s) => ({ ...s, [r.id]: e.target.checked }))
                  }
                />
              </td>
              <td className="p-2 border">{r.type}</td>
              <td className="p-2 border text-xs">{r.userId}</td>
              <td className="p-2 border text-xs">{r.productId || ""}</td>
              <td className="p-2 border text-xs">{r.value}</td>
              <td className="p-2 border text-xs font-mono text-gray-600">
                {formatMetaPreview(r.metaText)}
              </td>
              <td className="p-2 border">{r.recsysStatus}</td>
              <td className="p-2 border text-xs">
                {new Date(r.ts).toISOString()}
              </td>
            </tr>
          ))}
        </tbody>
      </table>

      <div className="flex items-center gap-2">
        <button
          className="border rounded px-2 py-1 text-sm"
          onClick={() => setOffset(Math.max(0, offset - limit))}
          disabled={offset === 0}
        >
          Prev
        </button>
        <div className="text-xs text-gray-600">
          {offset + 1} - {Math.min(offset + limit, total)} of {total}
        </div>
        <button
          className="border rounded px-2 py-1 text-sm"
          onClick={() => setOffset(offset + limit)}
          disabled={offset + limit >= total}
        >
          Next
        </button>
        <select
          className="border p-1 text-sm ml-2"
          value={limit}
          onChange={(e) => setLimit(parseInt(e.target.value, 10))}
        >
          <option value={20}>20</option>
          <option value={50}>50</option>
          <option value={100}>100</option>
        </select>
      </div>

      <section className="space-y-2 border-t pt-4">
        <h2 className="font-medium">Create manual event</h2>
        <form
          className="grid grid-cols-1 md:grid-cols-3 gap-2"
          onSubmit={onCreateEvent}
        >
          <input
            className="border p-2"
            name="userId"
            placeholder="User ID"
            value={eventForm.userId}
            onChange={(e) => setEventField("userId", e.target.value)}
          />
          <input
            className="border p-2"
            name="productId"
            placeholder="Product ID (optional)"
            value={eventForm.productId}
            onChange={(e) => setEventField("productId", e.target.value)}
          />
          <select
            className="border p-2"
            name="type"
            value={eventForm.type}
            onChange={(e) => setEventField("type", e.target.value)}
          >
            <option value="view">view</option>
            <option value="click">click</option>
            <option value="add">add</option>
            <option value="purchase">purchase</option>
            <option value="custom">custom</option>
          </select>
          <input
            className="border p-2"
            name="value"
            type="number"
            step="0.1"
            placeholder="Value"
            value={eventForm.value}
            onChange={(e) => setEventField("value", e.target.value)}
          />
          <input
            className="border p-2 md:col-span-2"
            name="ts"
            placeholder="Timestamp (ISO, optional)"
            value={eventForm.ts}
            onChange={(e) => setEventField("ts", e.target.value)}
          />
          <textarea
            className="border p-2 md:col-span-3 font-mono text-xs"
            name="meta"
            placeholder='{"surface":"home","widget":"home_top_picks"}'
            value={eventForm.meta}
            onChange={(e) => setEventField("meta", e.target.value)}
          />
          <div className="flex gap-2 md:col-span-3">
            <button
              type="button"
              className="border rounded px-3 py-2 text-sm"
              onClick={prefillMeta}
            >
              Prefill meta
            </button>
            <button
              className="border rounded px-3 py-2 text-sm bg-blue-50 hover:bg-blue-100"
              disabled={submittingEvent}
            >
              {submittingEvent ? "Creating…" : "Create event"}
            </button>
          </div>
        </form>
      </section>
    </section>
  );
}
