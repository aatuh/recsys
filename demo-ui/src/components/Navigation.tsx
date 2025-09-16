import { Button } from "./UIComponents";

export type ViewType =
  | "namespace-seed"
  | "recommendations-playground"
  | "bandit-playground"
  | "user-session"
  | "data-management"
  | "documentation"
  | "privacy-policy";

interface NavigationProps {
  activeView: ViewType;
  onViewChange: (view: ViewType) => void;
  apiBase: string;
  swaggerUrl: string;
  customChatGptUrl?: string;
  namespace: string;
}

export function Navigation({
  activeView,
  onViewChange,
  apiBase,
  swaggerUrl,
  customChatGptUrl,
  namespace,
}: NavigationProps) {
  // Helper function to generate URL for a view
  const getViewUrl = (view: ViewType) => {
    const url = new URL(window.location.href);
    url.searchParams.set("view", view);
    url.searchParams.set("namespace", namespace);
    return url.toString();
  };

  // Handle middle-click to open in new tab
  const handleMouseDown = (view: ViewType, event: React.MouseEvent) => {
    if (event.button === 1) {
      // Middle mouse button
      event.preventDefault();
      window.open(getViewUrl(view), "_blank", "noopener,noreferrer");
    }
  };
  return (
    <div
      style={{
        display: "flex",
        justifyContent: "space-between",
        alignItems: "center",
        marginBottom: 16,
        padding: "12px 0",
        borderBottom: "1px solid #e0e0e0",
      }}
    >
      <div style={{ display: "flex", alignItems: "center", gap: 16 }}>
        <h1
          style={{
            margin: 0,
            fontSize: 24,
            fontWeight: 600,
            cursor: "pointer",
            color: "#1976d2",
          }}
          onClick={() => onViewChange("namespace-seed")}
          onMouseDown={(e) => handleMouseDown("namespace-seed", e)}
          title="Click to go to Namespace & Seed view (middle-click to open in new tab)"
        >
          RecSys Demo UI
        </h1>
        <div style={{ display: "flex", gap: 8, flexWrap: "wrap" }}>
          <Button
            type="button"
            onClick={() => onViewChange("namespace-seed")}
            onMouseDown={(e) => handleMouseDown("namespace-seed", e)}
            style={{
              padding: "8px 16px",
              fontSize: 14,
              backgroundColor:
                activeView === "namespace-seed" ? "#1976d2" : "#fff",
              color: activeView === "namespace-seed" ? "#fff" : "#666",
              border: "1px solid #ddd",
              borderRadius: 6,
              cursor: "pointer",
              fontWeight: activeView === "namespace-seed" ? 600 : 400,
            }}
            title="Middle-click to open in new tab"
          >
            Namespace & Seed
          </Button>
          <Button
            type="button"
            onClick={() => onViewChange("recommendations-playground")}
            onMouseDown={(e) =>
              handleMouseDown("recommendations-playground", e)
            }
            style={{
              padding: "8px 16px",
              fontSize: 14,
              backgroundColor:
                activeView === "recommendations-playground"
                  ? "#1976d2"
                  : "#fff",
              color:
                activeView === "recommendations-playground" ? "#fff" : "#666",
              border: "1px solid #ddd",
              borderRadius: 6,
              cursor: "pointer",
              fontWeight:
                activeView === "recommendations-playground" ? 600 : 400,
            }}
            title="Middle-click to open in new tab"
          >
            Recommendations Playground
          </Button>
          <Button
            type="button"
            onClick={() => onViewChange("bandit-playground")}
            onMouseDown={(e) => handleMouseDown("bandit-playground", e)}
            style={{
              padding: "8px 16px",
              fontSize: 14,
              backgroundColor:
                activeView === "bandit-playground" ? "#1976d2" : "#fff",
              color: activeView === "bandit-playground" ? "#fff" : "#666",
              border: "1px solid #ddd",
              borderRadius: 6,
              cursor: "pointer",
              fontWeight: activeView === "bandit-playground" ? 600 : 400,
            }}
            title="Middle-click to open in new tab"
          >
            Bandit Playground
          </Button>
          <Button
            type="button"
            onClick={() => onViewChange("user-session")}
            onMouseDown={(e) => handleMouseDown("user-session", e)}
            style={{
              padding: "8px 16px",
              fontSize: 14,
              backgroundColor:
                activeView === "user-session" ? "#1976d2" : "#fff",
              color: activeView === "user-session" ? "#fff" : "#666",
              border: "1px solid #ddd",
              borderRadius: 6,
              cursor: "pointer",
              fontWeight: activeView === "user-session" ? 600 : 400,
            }}
            title="Middle-click to open in new tab"
          >
            User Session
          </Button>
          <Button
            type="button"
            onClick={() => onViewChange("data-management")}
            onMouseDown={(e) => handleMouseDown("data-management", e)}
            style={{
              padding: "8px 16px",
              fontSize: 14,
              backgroundColor:
                activeView === "data-management" ? "#1976d2" : "#fff",
              color: activeView === "data-management" ? "#fff" : "#666",
              border: "1px solid #ddd",
              borderRadius: 6,
              cursor: "pointer",
              fontWeight: activeView === "data-management" ? 600 : 400,
            }}
            title="Middle-click to open in new tab"
          >
            Data Management
          </Button>
          <Button
            type="button"
            onClick={() => onViewChange("documentation")}
            onMouseDown={(e) => handleMouseDown("documentation", e)}
            style={{
              padding: "8px 16px",
              fontSize: 14,
              backgroundColor:
                activeView === "documentation" ? "#1976d2" : "#fff",
              color: activeView === "documentation" ? "#fff" : "#666",
              border: "1px solid #ddd",
              borderRadius: 6,
              cursor: "pointer",
              fontWeight: activeView === "documentation" ? 600 : 400,
            }}
            title="Middle-click to open in new tab"
          >
            Documentation
          </Button>
          <Button
            type="button"
            onClick={() => onViewChange("privacy-policy")}
            onMouseDown={(e) => handleMouseDown("privacy-policy", e)}
            style={{
              padding: "8px 16px",
              fontSize: 14,
              backgroundColor:
                activeView === "privacy-policy" ? "#1976d2" : "#fff",
              color: activeView === "privacy-policy" ? "#fff" : "#666",
              border: "1px solid #ddd",
              borderRadius: 6,
              cursor: "pointer",
              fontWeight: activeView === "privacy-policy" ? 600 : 400,
            }}
            title="Middle-click to open in new tab"
          >
            Privacy Policy
          </Button>
        </div>
      </div>
      <div style={{ display: "flex", gap: 8, alignItems: "center" }}>
        {customChatGptUrl && (
          <Button
            type="button"
            style={{
              padding: "6px 12px",
              fontSize: 12,
              backgroundColor: "#f5f5f5",
              color: "#666",
              border: "1px solid #666",
              borderRadius: 4,
            }}
            onClick={() => {
              window.open(customChatGptUrl, "_blank", "noopener,noreferrer");
            }}
            onMouseDown={(e) => {
              if (e.button === 1) {
                // Middle mouse button
                e.preventDefault();
                window.open(customChatGptUrl, "_blank", "noopener,noreferrer");
              }
            }}
            title="Middle-click to open in new tab"
          >
            Ask ChatGPT
          </Button>
        )}
        <Button
          type="button"
          style={{
            padding: "6px 12px",
            fontSize: 12,
            backgroundColor: "#f5f5f5",
            color: "#666",
            border: "1px solid #666",
            borderRadius: 4,
          }}
          onClick={() => {
            window.open(`${swaggerUrl}/docs/`, "_blank", "noopener,noreferrer");
          }}
          onMouseDown={(e) => {
            if (e.button === 1) {
              // Middle mouse button
              e.preventDefault();
              window.open(
                `${swaggerUrl}/docs/`,
                "_blank",
                "noopener,noreferrer"
              );
            }
          }}
          title="Middle-click to open in new tab"
        >
          Explore API
        </Button>
      </div>
    </div>
  );
}
