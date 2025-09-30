import { useEffect, useMemo, useState } from "react";
import {
  ExplainService,
  type types_ExplainLLMRequest,
  type types_ExplainLLMResponse,
} from "../../lib/api-client";
import { Button, Section, Label, TextInput } from "../primitives/UIComponents";
import { SafeMarkdown } from "../SafeMarkdown";
import { useStringStorage } from "../../hooks/useStorage";

type TargetType = "item" | "banner" | "surface" | "segment";

interface ExplainLLMViewProps {
  namespace: string;
}

export function ExplainLLMView({ namespace }: ExplainLLMViewProps) {
  const [targetType, setTargetType] = useState<TargetType>("item");
  const [targetId, setTargetId] = useState("");
  const [surface, setSurface] = useState("");
  const [segmentId, setSegmentId] = useState("");
  const [from, setFrom] = useState("");
  const [to, setTo] = useState("");
  const [question, setQuestion] = useState(
    "Why is this not working as expected?"
  );

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [response, setResponse] = useState<types_ExplainLLMResponse | null>(
    null
  );

  // Local settings using storage abstraction
  const { value: storedLlmEnabled, setValue: setStoredLlmEnabled } =
    useStringStorage("LLM_ENABLED", "true");
  const { value: llmProvider, setValue: setLlmProvider } = useStringStorage(
    "LLM_PROVIDER",
    "openai"
  );
  const { value: llmPrimary, setValue: setLlmPrimary } = useStringStorage(
    "LLM_MODEL_PRIMARY",
    "o4-mini"
  );
  const { value: llmEscalate, setValue: setLlmEscalate } = useStringStorage(
    "LLM_MODEL_ESCALATE",
    "o3"
  );

  const [llmEnabled, setLlmEnabled] = useState<boolean>(
    storedLlmEnabled === "true"
  );

  useEffect(() => {
    setLlmEnabled(storedLlmEnabled === "true");
  }, [storedLlmEnabled]);

  const canSubmit = useMemo(() => {
    if (!namespace) return false;
    if (targetType === "segment") return Boolean(segmentId);
    if (targetType === "surface") return Boolean(surface);
    return Boolean(targetId);
  }, [namespace, targetType, targetId, surface, segmentId]);

  const handleGenerate = async () => {
    setLoading(true);
    setError(null);
    setResponse(null);
    try {
      const payload: types_ExplainLLMRequest = {
        namespace,
        target_type: targetType,
        target_id: targetType === "segment" ? undefined : targetId || undefined,
        segment_id:
          targetType === "segment" ? segmentId || undefined : undefined,
        surface: surface || undefined,
        from: from || undefined,
        to: to || undefined,
        question: question || undefined,
      };
      const res = await ExplainService.postV1ExplainLlm(payload);
      // Some generators may return string; normalize
      const normalized: types_ExplainLLMResponse =
        typeof res === "string" ? (JSON.parse(res) as any) : (res as any);
      setResponse(normalized);
    } catch (e: any) {
      const msg = e?.message || "Failed to generate explanation";
      setError(msg);
    } finally {
      setLoading(false);
    }
  };

  // Removed renderMarkdown function - now using SafeMarkdown component

  const extractSections = (md?: string) => {
    if (!md) return null;
    const names = [
      "Summary",
      "Status",
      "Key findings",
      "Likely causes",
      "Suggested fix",
    ];
    const map: Record<string, string> = {};
    const lines = md.split(/\n/);
    let current: string | null = null;
    let buf: string[] = [];
    const flush = () => {
      if (current) {
        map[current] = buf.join("\n").trim();
      }
      buf = [];
    };
    for (const line of lines) {
      const m = line.match(/^##\s+(.+)$/);
      if (m && m[1]) {
        const title = m[1].trim();
        if (
          names.some((n) => title.toLowerCase().startsWith(n.toLowerCase()))
        ) {
          flush();
          // Normalize to canonical name
          const canonical =
            names.find((n) =>
              title.toLowerCase().startsWith(n.toLowerCase())
            ) || title;
          current = canonical;
          continue;
        }
      }
      if (current) buf.push(line);
    }
    flush();
    return map;
  };

  const factsOnlyPanel = (facts: Record<string, any> | undefined) => {
    if (!facts) return null;
    return (
      <div
        style={{
          border: "1px solid #e0e0e0",
          borderRadius: 6,
          padding: 12,
          background: "#fafafa",
        }}
      >
        <h3 style={{ marginTop: 0, color: "#1976d2" }}>Facts</h3>
        <pre
          style={{
            whiteSpace: "pre-wrap",
            wordBreak: "break-word",
            fontSize: 12,
            background: "#f5f5f5",
            padding: 12,
            borderRadius: 4,
            maxHeight: 360,
            overflow: "auto",
          }}
        >
          {JSON.stringify(facts, null, 2)}
        </pre>
      </div>
    );
  };

  return (
    <Section title="Explain (LLM)">
      <div style={{ display: "flex", gap: 16, alignItems: "flex-end" }}>
        <div style={{ display: "flex", flexDirection: "column", gap: 6 }}>
          <label style={{ fontSize: 12, color: "#666" }}>Target Type</label>
          <select
            value={targetType}
            onChange={(e) => setTargetType(e.target.value as TargetType)}
            style={{ padding: 8, border: "1px solid #ddd", borderRadius: 6 }}
          >
            <option value="item">Item</option>
            <option value="banner">Banner</option>
            <option value="surface">Surface</option>
            <option value="segment">Segment</option>
          </select>
        </div>

        {targetType !== "segment" && (
          <Label text={targetType === "surface" ? "Surface" : "Target ID"}>
            <TextInput
              value={targetType === "surface" ? surface : targetId}
              onChange={(e) =>
                targetType === "surface"
                  ? setSurface(e.target.value)
                  : setTargetId(e.target.value)
              }
              placeholder={
                targetType === "surface" ? "home, feed, ..." : "item-123"
              }
              style={{ width: 200 }}
            />
          </Label>
        )}

        {targetType === "segment" && (
          <Label text="Segment ID">
            <TextInput
              value={segmentId}
              onChange={(e) => setSegmentId(e.target.value)}
              placeholder="segment-abc"
              style={{ width: 200 }}
            />
          </Label>
        )}

        <Label text="Namespace">
          <TextInput value={namespace} disabled style={{ width: 160 }} />
        </Label>

        <Label text="From (ISO)">
          <TextInput
            value={from}
            onChange={(e) => setFrom(e.target.value)}
            placeholder="2025-09-01T00:00:00Z"
            style={{ width: 220 }}
          />
        </Label>

        <Label text="To (ISO)">
          <TextInput
            value={to}
            onChange={(e) => setTo(e.target.value)}
            placeholder="2025-09-21T23:59:59Z"
            style={{ width: 220 }}
          />
        </Label>

        <div style={{ flex: 1 }} />

        <Button
          type="button"
          onClick={handleGenerate}
          disabled={!canSubmit || loading}
        >
          {loading ? "Generating..." : "Generate"}
        </Button>
      </div>

      <div style={{ height: 8 }} />

      <div style={{ display: "flex", gap: 16 }}>
        <div style={{ flex: 1 }}>
          <label style={{ fontSize: 12, color: "#666" }}>Question</label>
          <textarea
            value={question}
            onChange={(e) => setQuestion(e.target.value)}
            placeholder="Describe what's not working or what you want to know"
            rows={4}
            style={{
              width: "100%",
              padding: 8,
              border: "1px solid #ddd",
              borderRadius: 6,
              fontFamily: "inherit",
            }}
          />
        </div>

        <div style={{ width: 340 }}>
          <div
            style={{
              border: "1px solid #e0e0e0",
              borderRadius: 6,
              padding: 12,
              background: "#fafafa",
            }}
          >
            <h3 style={{ marginTop: 0, fontSize: 14, color: "#1976d2" }}>
              Settings
            </h3>
            <div style={{ display: "grid", gap: 8 }}>
              <label style={{ display: "flex", alignItems: "center", gap: 8 }}>
                <input
                  type="checkbox"
                  checked={llmEnabled}
                  onChange={(e) => {
                    const v = e.target.checked;
                    setLlmEnabled(v);
                    setStoredLlmEnabled(String(v));
                  }}
                />
                <span>Enable LLM</span>
              </label>
              <Label text="LLM Provider">
                <select
                  value={llmProvider}
                  onChange={(e) => {
                    setLlmProvider(e.target.value);
                  }}
                  style={{
                    padding: 8,
                    border: "1px solid #ddd",
                    borderRadius: 6,
                  }}
                >
                  <option value="openai">openai</option>
                  <option value="none">none</option>
                </select>
              </Label>
              <Label text="Primary model">
                <select
                  value={llmPrimary}
                  onChange={(e) => {
                    setLlmPrimary(e.target.value);
                  }}
                  style={{
                    padding: 8,
                    border: "1px solid #ddd",
                    borderRadius: 6,
                  }}
                >
                  <option value="o4-mini">o4-mini</option>
                  <option value="gpt-4o-mini">gpt-4o-mini</option>
                </select>
              </Label>
              <Label text="Escalate model">
                <select
                  value={llmEscalate}
                  onChange={(e) => {
                    setLlmEscalate(e.target.value);
                  }}
                  style={{
                    padding: 8,
                    border: "1px solid #ddd",
                    borderRadius: 6,
                  }}
                >
                  <option value="o3">o3</option>
                  <option value="gpt-4o">gpt-4o</option>
                </select>
              </Label>
            </div>
          </div>
        </div>
      </div>

      {error && (
        <div
          style={{
            marginTop: 12,
            color: "#d32f2f",
            background: "#ffebee",
            border: "1px solid #ffcdd2",
            borderRadius: 6,
            padding: 12,
          }}
        >
          {error}
        </div>
      )}

      {response && (
        <div style={{ marginTop: 16, display: "grid", gap: 16 }}>
          {response.markdown && llmEnabled
            ? (() => {
                const sections = extractSections(response.markdown);
                if (!sections || Object.keys(sections).length === 0) {
                  return (
                    <div
                      style={{
                        border: "1px solid #e0e0e0",
                        borderRadius: 6,
                        padding: 12,
                        background: "#fff",
                      }}
                    >
                      <SafeMarkdown
                        content={response.markdown || ""}
                        allowHtml={false}
                        allowLinks={true}
                        allowImages={true}
                        allowCode={true}
                      />
                    </div>
                  );
                }
                return (
                  <div style={{ display: "grid", gap: 12 }}>
                    {[
                      "Summary",
                      "Status",
                      "Key findings",
                      "Likely causes",
                      "Suggested fix",
                    ].map((name) =>
                      sections[name] ? (
                        <div
                          key={name}
                          style={{
                            border: "1px solid #e0e0e0",
                            borderRadius: 6,
                            padding: 12,
                            background: "#fff",
                          }}
                        >
                          <h3
                            style={{
                              marginTop: 0,
                              color: "#1976d2",
                              fontSize: 16,
                            }}
                          >
                            {name}
                          </h3>
                          <SafeMarkdown
                            content={sections[name] || ""}
                            allowHtml={false}
                            allowLinks={true}
                            allowImages={true}
                            allowCode={true}
                          />
                        </div>
                      ) : null
                    )}
                  </div>
                );
              })()
            : factsOnlyPanel(response.facts)}

          {response.warnings && response.warnings.length > 0 && (
            <div
              style={{
                color: "#9a6700",
                background: "#fff8c5",
                border: "1px solid #e6db74",
                borderRadius: 6,
                padding: 12,
              }}
            >
              <strong>Warnings:</strong> {response.warnings.join("; ")}
            </div>
          )}

          {response.facts && response.markdown && llmEnabled && (
            <div>{factsOnlyPanel(response.facts)}</div>
          )}
        </div>
      )}
    </Section>
  );
}
