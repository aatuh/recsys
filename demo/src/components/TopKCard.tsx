import { useEffect, useState, useRef } from "react";

// Signal mapping for better display names and explanations
const SIGNAL_DISPLAY_NAMES: Record<string, string> = {
  recent_popularity: "Popular now",
  co_visitation: "Viewed together",
  embedding: "Similar items",
  personalization: "Personalized fit",
  diversity: "Balanced brands",
};

const SIGNAL_EXPLANATIONS: Record<string, string> = {
  recent_popularity:
    "This item is trending and popular with many users right now",
  co_visitation:
    "People who viewed this item also looked at the recommended item",
  embedding: "This item is similar to others you've shown interest in",
  personalization:
    "Recommended based on your personal preferences and behavior",
  diversity: "Helps balance the mix of brands in your recommendations",
};

function getSignalDisplayName(signal: string): string {
  return SIGNAL_DISPLAY_NAMES[signal] || signal;
}

function getSignalExplanation(signal: string): string {
  return (
    SIGNAL_EXPLANATIONS[signal] ||
    `This item was recommended based on ${signal}`
  );
}

type Props = {
  id: string;
  title: string;
  reasons?: string[];
  delta?: number;
  annotation?: string;
  muted?: boolean;
  previousPosition?: number;
  currentPosition?: number;
  brand?: string;
  busy?: boolean;
  score?: number;
  price?: number;
  tags?: string[];
  available?: boolean;
  position?: number;
  onPin?: (id: string) => void;
  onBlockBrand?: (brand: string) => void;
};

export default function TopKCard({
  id,
  title,
  reasons = [],
  delta,
  annotation,
  muted,
  previousPosition,
  currentPosition,
  brand,
  busy,
  score,
  price,
  tags = [],
  available,
  position,
  onPin,
  onBlockBrand,
}: Props) {
  const [showExplanation, setShowExplanation] = useState<string | null>(null);
  const cardRef = useRef<HTMLDivElement>(null);

  // Close explanation when clicking outside
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (cardRef.current && !cardRef.current.contains(event.target as Node)) {
        setShowExplanation(null);
      }
    }

    if (showExplanation) {
      document.addEventListener("mousedown", handleClickOutside);
      return () =>
        document.removeEventListener("mousedown", handleClickOutside);
    }
  }, [showExplanation]);
  const deltaValue = typeof delta === "number" && delta !== 0 ? delta : null;
  const deltaClass =
    deltaValue !== null
      ? deltaValue > 0
        ? "topk-delta up"
        : "topk-delta down"
      : null;

  // Calculate position animation
  const shouldAnimate =
    typeof previousPosition === "number" &&
    typeof currentPosition === "number" &&
    previousPosition !== currentPosition;

  const animationClass = shouldAnimate ? "topk-card-moving" : "";
  const cardClasses = `topk-card${
    muted ? " topk-card-muted" : ""
  } ${animationClass}`.trim();

  // Simple highlight animation when position changes
  const getAnimationStyle = () => {
    if (!shouldAnimate) {
      return {};
    }

    return {
      animation: "positionChange 0.6s ease-out",
    };
  };

  // Log animation when positions change
  useEffect(() => {
    if (shouldAnimate) {
      console.log(
        `Position changed for ${id}: from ${previousPosition} to ${currentPosition}`
      );
    }
  }, [shouldAnimate, id, previousPosition, currentPosition]);

  const sendEvent = (eventType: number, eventName: string) => {
    try {
      const apiBase =
        (import.meta as { env?: { VITE_API_BASE_URL?: string } }).env
          ?.VITE_API_BASE_URL || "http://localhost:8081";
      const ns = localStorage.getItem("demo:ns") || "default";
      const payload = {
        namespace: ns,
        events: [
          {
            user_id: localStorage.getItem("demo:lastUser") || "user-1",
            item_id: id,
            type: eventType,
            ts: new Date().toISOString(),
            value: 1,
          },
        ],
      };
      fetch(`${apiBase}/v1/events:batch`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      })
        .then(() => {
          console.log(`Simulated ${eventName} for item ${id}`);
        })
        .catch(() => {});
    } catch {
      // best-effort fire-and-forget for demo
    }
  };

  return (
    <div ref={cardRef} className={cardClasses} style={getAnimationStyle()}>
      <div className="topk-row">
        <div className="topk-title">{title}</div>
        {annotation ? (
          <span className="topk-annotation">{annotation}</span>
        ) : null}
        {deltaClass && deltaValue !== null ? (
          <span
            className={deltaClass}
            title={
              deltaValue > 0
                ? `Up ${deltaValue}`
                : `Down ${Math.abs(deltaValue)}`
            }
          >
            {deltaValue > 0 ? "↑" : "↓"}
            {Math.abs(deltaValue)}
          </span>
        ) : null}
        <div className="topk-actions">
          {onPin ? (
            <button
              className="topk-action-btn topk-pin-btn"
              type="button"
              onClick={() => onPin(id)}
              disabled={!!busy}
              title="Pin this item to #1"
            >
              <svg
                width="18"
                height="18"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
              >
                <path d="M12 17l5-5-5-5" />
                <path d="M17 12H7" />
                <circle cx="12" cy="12" r="10" />
              </svg>
            </button>
          ) : null}
          {onBlockBrand ? (
            <button
              className="topk-action-btn topk-block-btn"
              type="button"
              onClick={() => brand && onBlockBrand(brand)}
              disabled={!!busy || !brand}
              title={brand ? `Block brand ${brand}` : "No brand to block"}
            >
              <svg
                width="18"
                height="18"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
              >
                <path d="M18 6L6 18" />
                <path d="M6 6l12 12" />
                <circle cx="12" cy="12" r="10" />
              </svg>
            </button>
          ) : null}
        </div>
      </div>

      {/* Event simulation buttons */}
      <div className="topk-event-buttons">
        <button
          className="topk-event-btn topk-view-btn"
          type="button"
          onClick={(e) => {
            e.stopPropagation();
            sendEvent(1, "view");
          }}
          title="Simulate viewing this item"
        >
          👁️ View
        </button>
        <button
          className="topk-event-btn topk-click-btn"
          type="button"
          onClick={(e) => {
            e.stopPropagation();
            sendEvent(2, "click");
          }}
          title="Simulate clicking this item"
        >
          👆 Click
        </button>
        <button
          className="topk-event-btn topk-purchase-btn"
          type="button"
          onClick={(e) => {
            e.stopPropagation();
            sendEvent(3, "purchase");
          }}
          title="Simulate purchasing this item"
        >
          🛒 Buy
        </button>
      </div>

      <div className="topk-meta">
        <div className="topk-id">#{id}</div>
        {typeof position === "number" && (
          <div className="topk-position">Position {position + 1}</div>
        )}
        {typeof score === "number" && (
          <div className="topk-score">Score: {score.toFixed(3)}</div>
        )}
      </div>

      <div className="topk-details">
        {brand && (
          <div className="topk-brand">
            <span className="topk-brand-label">Brand:</span>
            <span className="topk-brand-name">{brand}</span>
          </div>
        )}
        {typeof price === "number" && (
          <div className="topk-price">${price.toFixed(2)}</div>
        )}
        {typeof available === "boolean" && (
          <div
            className={`topk-availability ${
              available ? "available" : "unavailable"
            }`}
          >
            {available ? "✓ Available" : "✗ Out of stock"}
          </div>
        )}
      </div>

      {tags.length > 0 && (
        <div className="topk-tags">
          {tags.slice(0, 4).map((tag, idx) => (
            <span key={`${id}-tag-${idx}`} className="topk-tag">
              {tag.replace(/^(brand:|cat:)/, "")}
            </span>
          ))}
        </div>
      )}

      {reasons.length ? (
        <div className="topk-reasons">
          {reasons.slice(0, 3).map((r, idx) => {
            const explanation = getSignalExplanation(r);
            const isShowing = showExplanation === r;
            return (
              <span
                key={`${id}-reason-${idx}`}
                className={`topk-reason-pill ${
                  isShowing ? "showing-explanation" : ""
                }`}
                title={explanation}
                onClick={(e) => {
                  e.preventDefault();
                  e.stopPropagation();
                  setShowExplanation(isShowing ? null : r);
                }}
                onMouseEnter={() => {
                  // On desktop, show explanation on hover
                  if (window.innerWidth > 768) {
                    setShowExplanation(r);
                  }
                }}
                onMouseLeave={() => {
                  // On desktop, hide explanation when not hovering
                  if (window.innerWidth > 768) {
                    setShowExplanation(null);
                  }
                }}
              >
                {getSignalDisplayName(r)}
                {isShowing && (
                  <div className="reason-explanation">
                    <div className="reason-explanation-content">
                      {explanation}
                    </div>
                    <button
                      className="reason-explanation-close"
                      onClick={(e) => {
                        e.preventDefault();
                        e.stopPropagation();
                        setShowExplanation(null);
                      }}
                      aria-label="Close explanation"
                    >
                      ×
                    </button>
                  </div>
                )}
              </span>
            );
          })}
        </div>
      ) : null}
    </div>
  );
}
