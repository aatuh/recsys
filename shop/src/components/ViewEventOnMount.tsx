"use client";
import { useEffect } from "react";
import { useTelemetry } from "@/lib/telemetry/useTelemetry";

export function ViewEventOnMount({ 
  productId, 
  surface = "pdp",
  widget,
  recommended = false,
  rank 
}: { 
  productId: string;
  surface?: "home" | "pdp" | "cart" | "checkout";
  widget?: string;
  recommended?: boolean;
  rank?: number;
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
      }
    });
  }, [emit, productId, surface, widget, recommended, rank]);

  return null;
}