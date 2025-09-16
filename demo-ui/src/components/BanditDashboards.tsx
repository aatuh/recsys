import React, { useMemo, useState } from "react";
import { Section, Row, Label, Th, Td } from "./UIComponents";
import { useViewState } from "../contexts/ViewStateContext";

// Simple inline bar for visualization
function Bar({ value, max }: { value: number; max: number }) {
  const pct = max > 0 ? Math.min(100, (value / max) * 100) : 0;
  return (
    <div style={{ background: "#eee", borderRadius: 4, height: 8 }}>
      <div
        style={{
          width: `${pct  }%`,
          height: 8,
          background: "#4caf50",
          borderRadius: 4,
        }}
      />
    </div>
  );
}

export function BanditDashboards(props: {
  namespace: string;
  availablePolicies: Array<{ policy_id?: string; name?: string }>;
}) {
  const { availablePolicies } = props;
  const { banditPlayground } = useViewState();
  const { decisionHistory, rewardHistory } = banditPlayground;

  const [contextKey, setContextKey] = useState("device");

  // Map policy id to display name
  const policyName = (id: string) =>
    availablePolicies.find((p) => p.policy_id === id)?.name || id;

  // Per-policy stats
  const perPolicy = useMemo(() => {
    const byPolicy: Record<string, { trials: number; successes: number }> = {};
    for (const d of decisionHistory) {
      if (!byPolicy[d.policyId]) {
        byPolicy[d.policyId] = { trials: 0, successes: 0 };
      }
      byPolicy[d.policyId]!.trials += 1;
    }
    for (const r of rewardHistory) {
      if (!byPolicy[r.policyId]) {
        byPolicy[r.policyId] = { trials: 0, successes: 0 };
      }
      if (r.success && r.reward) byPolicy[r.policyId]!.successes += 1;
    }
    return byPolicy;
  }, [decisionHistory, rewardHistory]);

  // Exploration rate over time (rolling buckets)
  const exploreSeries = useMemo(() => {
    // Build a simple time buckets of last N decisions
    const N = 50;
    const recent = decisionHistory.slice(0, N).slice().reverse(); // oldest -> newest
    return recent.map((d, i) => ({
      idx: i + 1,
      explore: d.explore ? 1 : 0,
    }));
  }, [decisionHistory]);

  const exploreRate = useMemo(() => {
    if (exploreSeries.length === 0) return 0;
    const sum = exploreSeries.reduce((a, b) => a + b.explore, 0);
    return sum / exploreSeries.length;
  }, [exploreSeries]);

  // Context buckets split (e.g., device ios vs android)
  const contextBuckets = useMemo(() => {
    const buckets: Record<string, { decisions: number; explores: number }> = {};
    for (const d of decisionHistory) {
      const key = String(d.context?.[contextKey] ?? "(missing)");
      buckets[key] ||= { decisions: 0, explores: 0 };
      buckets[key].decisions += 1;
      if (d.explore) buckets[key].explores += 1;
    }
    return buckets;
  }, [decisionHistory, contextKey]);

  // Max values for bars
  const maxTrials = useMemo(
    () => Math.max(1, ...Object.values(perPolicy).map((x) => x.trials || 0)),
    [perPolicy]
  );
  const maxBucketDec = useMemo(
    () => Math.max(1, ...Object.values(contextBuckets).map((x) => x.decisions)),
    [contextBuckets]
  );

  return (
    <Section title="Bandit Dashboards">
      <div style={{ color: "#666", fontSize: 14, marginBottom: 12 }}>
        Live, client-side dashboards based on your recent decisions and rewards.
      </div>

      {/* Per-policy stats */}
      <div style={{ marginBottom: 16 }}>
        <Label text="Per-policy stats (successes / trials)">
          <div style={{ overflowX: "auto" }}>
            <table style={{ borderCollapse: "collapse", minWidth: 520 }}>
              <thead>
                <tr>
                  <Th>policy</Th>
                  <Th>trials</Th>
                  <Th>successes</Th>
                  <Th>win rate</Th>
                  <Th>viz</Th>
                </tr>
              </thead>
              <tbody>
                {Object.entries(perPolicy).length === 0 && (
                  <tr>
                    <Td>No data yet. Make decisions and send rewards.</Td>
                  </tr>
                )}
                {Object.entries(perPolicy).map(([pid, v]) => {
                  const win = v.trials > 0 ? v.successes / v.trials : 0;
                  return (
                    <tr key={pid}>
                      <Td>{policyName(pid)}</Td>
                      <Td>{v.trials}</Td>
                      <Td>{v.successes}</Td>
                      <Td>{(win * 100).toFixed(1)}%</Td>
                      <Td>
                        <Bar value={v.trials} max={maxTrials} />
                      </Td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </Label>
      </div>

      {/* Exploration rate over time */}
      <div style={{ marginBottom: 16 }}>
        <Label
          text={`Exploration rate (last ${exploreSeries.length} decisions)`}
        >
          <div style={{ fontSize: 12, marginBottom: 8 }}>
            Current explore rate: {(exploreRate * 100).toFixed(1)}%
          </div>
          <div style={{ fontSize: 12, color: "#6c757d", marginBottom: 8 }}>
            <strong>Exploration</strong> is when the chosen policy differs from
            the current empirical best; it helps discover better policies.
            <br /> <strong> Exploitation</strong> chooses the empirical best to
            maximize reward now.
          </div>
          <div
            style={{
              display: "flex",
              gap: 2,
              alignItems: "flex-end",
              height: 60,
              border: "1px solid #eee",
              padding: 6,
              borderRadius: 4,
            }}
          >
            {exploreSeries.map((p) => (
              <div
                key={p.idx}
                title={`#${p.idx}: ${p.explore ? "explore" : "exploit"}`}
                style={{
                  width: 6,
                  height: p.explore ? 52 : 18,
                  background: p.explore ? "#ff9800" : "#4caf50",
                  borderRadius: 2,
                }}
              />
            ))}
          </div>
        </Label>
      </div>

      {/* Context buckets split */}
      <div style={{ marginBottom: 8 }}>
        <Row>
          <Label text="Context key">
            <select
              value={contextKey}
              onChange={(e) => setContextKey(e.target.value)}
              style={{ padding: "6px 8px", border: "1px solid #ddd" }}
            >
              <option value="device">device</option>
              <option value="locale">locale</option>
            </select>
          </Label>
        </Row>
      </div>
      <div>
        <Label text={`Context buckets by ${contextKey}`}>
          <div style={{ overflowX: "auto" }}>
            <table style={{ borderCollapse: "collapse", minWidth: 520 }}>
              <thead>
                <tr>
                  <Th>bucket</Th>
                  <Th>decisions</Th>
                  <Th>explore%</Th>
                  <Th>viz</Th>
                </tr>
              </thead>
              <tbody>
                {Object.entries(contextBuckets).length === 0 && (
                  <tr>
                    <Td>No data yet. Make decisions.</Td>
                  </tr>
                )}
                {Object.entries(contextBuckets).map(([b, v]) => {
                  const rate = v.decisions ? v.explores / v.decisions : 0;
                  return (
                    <tr key={b}>
                      <Td>{b}</Td>
                      <Td>{v.decisions}</Td>
                      <Td>{(rate * 100).toFixed(1)}%</Td>
                      <Td>
                        <Bar value={v.decisions} max={maxBucketDec} />
                      </Td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </Label>
      </div>
    </Section>
  );
}
