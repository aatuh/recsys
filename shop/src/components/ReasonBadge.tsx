"use client";
import { useState } from "react";

export function ReasonBadge({ reasons }: { reasons?: string[] }) {
  if (!reasons || reasons.length === 0) return null;
  const [open, setOpen] = useState(false);
  return (
    <span className="inline-block">
      <button
        type="button"
        className="ml-2 text-[10px] px-1.5 py-0.5 rounded bg-gray-200 hover:bg-gray-300"
        onClick={() => setOpen((v) => !v)}
        aria-expanded={open}
        title={reasons.join(", ")}
      >
        Why?
      </button>
      {open && (
        <div className="mt-1 p-2 border rounded bg-white shadow text-xs max-w-xs">
          <div className="font-medium mb-1">Reasons</div>
          <ul className="list-disc pl-4">
            {reasons.map((r, i) => (
              <li key={i}>{r}</li>
            ))}
          </ul>
        </div>
      )}
    </span>
  );
}
