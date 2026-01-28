"use client";
import { useEffect } from "react";
import { useTelemetry } from "@/lib/telemetry/useTelemetry";

export function ViewEventOnMount({ 
  productId, 
  surface = "pdp",
  widget,
  recommended = false,
  rank,
  coldStart = false,
}: { 
  productId: string;
  surface?: "home" | "pdp" | "cart" | "checkout";
  widget?: string;
  recommended?: boolean;
  rank?: number;
  coldStart?: boolean;
}) {
  const { emit } = useTelemetry();

  useEffect(() => {
    void emit({
      type: "view",
      productId,
      value: 1,
      meta: {
        surface,
        widget,
        recommended,
        rank,
        cold_start: coldStart || undefined,
      }
    });
  }, [emit, productId, surface, widget, recommended, rank, coldStart]);

  return null;
}
