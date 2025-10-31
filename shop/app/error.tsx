"use client";
import { useEffect } from "react";

export default function GlobalError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    console.error(error);
  }, [error]);

  return (
    <html>
      <body>
        <div className="p-6">
          <h2 className="text-xl font-semibold">Something went wrong</h2>
          <p className="text-sm text-gray-600 mt-2">{error.message}</p>
          <button
            className="mt-4 border rounded px-3 py-2 text-sm"
            onClick={() => reset()}
          >
            Try again
          </button>
        </div>
      </body>
    </html>
  );
}
