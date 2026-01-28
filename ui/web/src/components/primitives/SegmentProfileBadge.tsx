import React from "react";

interface SegmentProfileBadgeProps {
  segmentId?: string;
  profileId?: string;
  style?: React.CSSProperties;
}

export function SegmentProfileBadge({
  segmentId,
  profileId,
  style = {},
}: SegmentProfileBadgeProps) {
  if (!segmentId && !profileId) {
    return null;
  }

  return (
    <div
      style={{
        display: "flex",
        gap: 8,
        alignItems: "center",
        fontSize: "12px",
        ...style,
      }}
    >
      {segmentId && (
        <span
          style={{
            padding: "2px 6px",
            backgroundColor: "#e3f2fd",
            color: "#1976d2",
            border: "1px solid #bbdefb",
            borderRadius: 3,
            fontWeight: 500,
          }}
        >
          segment={segmentId}
        </span>
      )}
      {profileId && (
        <span
          style={{
            padding: "2px 6px",
            backgroundColor: "#f3e5f5",
            color: "#7b1fa2",
            border: "1px solid #e1bee7",
            borderRadius: 3,
            fontWeight: 500,
          }}
        >
          profile={profileId}
        </span>
      )}
    </div>
  );
}
