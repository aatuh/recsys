import type { ReactNode } from "react";
import { useEffect, useMemo, useRef } from "react";
import Modal from "./Modal";
import type { ScoredItem } from "../types/recommendations";

type Props = {
  open: boolean;
  status: "running" | "ready" | "error";
  persona: string;
  before: ScoredItem[];
  after?: ScoredItem[];
  message?: string;
  onClose: () => void;
  onReplay: () => void;
  busy?: boolean;
};

type DiffRow = {
  id: string;
  title: string;
  reasons: string[];
  rank: number;
  delta?: number | null;
};

export default function AdaptDiffOverlay({
  open,
  status,
  persona,
  before,
  after,
  message,
  onClose,
  onReplay,
  busy = false,
}: Props) {
  const replayButtonRef = useRef<HTMLButtonElement | null>(null);
  const beforeRows: DiffRow[] = useMemo(
    () =>
      before.map((item, idx) => ({
        id: item.id,
        title: item.title || item.id,
        reasons: (item.reasons || []).slice(0, 2),
        rank: idx + 1,
        delta: null,
      })),
    [before]
  );

  const afterRows: DiffRow[] = useMemo(() => {
    if (!after) return [];
    return after.map((item, idx) => {
      const previousIndex = before.findIndex((b) => b.id === item.id);
      const delta = previousIndex >= 0 ? previousIndex - idx : null;
      return {
        id: item.id,
        title: item.title || item.id,
        reasons: (item.reasons || []).slice(0, 2),
        rank: idx + 1,
        delta,
      };
    });
  }, [after, before]);

  const canReplay = status === "ready" && !busy;

  useEffect(() => {
    if (!open || !canReplay) return;
    replayButtonRef.current?.focus();
  }, [open, canReplay]);

  const replayLabel = busy ? "Replaying…" : "Replay persona";

  const headerActions = (
    <button
      ref={replayButtonRef}
      className="btn btn-secondary btn-small modal-replay"
      onClick={onReplay}
      disabled={!canReplay}
      type="button"
    >
      {replayLabel}
    </button>
  );

  return (
    <Modal
      title={`Watch it adapt — ${persona}`}
      open={open}
      onClose={onClose}
      actions={headerActions}
    >
      {status === "running" ? (
        <RunningCopy persona={persona} />
      ) : status === "error" ? (
        <ErrorCopy message={message} onClose={onClose} />
      ) : (
        <ReadyState
          beforeRows={beforeRows}
          afterRows={afterRows}
          onClose={onClose}
        />
      )}
    </Modal>
  );
}

type ReadyProps = {
  beforeRows: DiffRow[];
  afterRows: DiffRow[];
  onClose: () => void;
};

function ReadyState({ beforeRows, afterRows, onClose }: ReadyProps) {
  return (
    <div>
      <p className="diff-caption">
        We simulated view → view → click → add → purchase. The right column shows
        how the list shifts seconds later.
      </p>
      <div className="diff-columns">
        <section className="diff-column">
          <header>Before</header>
          <ul className="diff-list">
            {beforeRows.map((row, idx) => (
              <li
                key={`before-${row.id}`}
                className="diff-card"
                style={{ animationDelay: `${idx * 40}ms` }}
              >
                <RankPill>{row.rank}</RankPill>
                <div className="diff-card-body">
                  <div className="diff-card-title">{row.title}</div>
                  {row.reasons.length ? (
                    <div className="diff-card-reasons">
                      {row.reasons.join(" · ")}
                    </div>
                  ) : null}
                </div>
              </li>
            ))}
          </ul>
        </section>
        <section className="diff-column">
          <header>After</header>
          <ul className="diff-list">
            {afterRows.map((row, idx) => {
              const improved = typeof row.delta === "number" && row.delta > 0;
              const worsened = typeof row.delta === "number" && row.delta < 0;
              const toneClass = improved
                ? "diff-card-after--improved"
                : worsened
                ? "diff-card-after--worsened"
                : "diff-card-after--neutral";
              return (
                <li
                  key={`after-${row.id}`}
                  className={`diff-card diff-card-after ${toneClass}`}
                  style={{ animationDelay: `${idx * 40}ms` }}
                >
                  <RankPill>{row.rank}</RankPill>
                  <div className="diff-card-body">
                    <div className="diff-card-title">{row.title}</div>
                    {row.reasons.length ? (
                      <div className="diff-card-reasons">
                        {row.reasons.join(" · ")}
                      </div>
                    ) : null}
                  </div>
                  {typeof row.delta === "number" && row.delta !== 0 ? (
                    <DeltaBadge delta={row.delta} />
                  ) : null}
                </li>
              );
            })}
          </ul>
        </section>
      </div>
      <div className="diff-actions">
        <button onClick={onClose} className="diff-secondary" type="button">
          Close
        </button>
      </div>
    </div>
  );
}

type RankProps = { children: ReactNode };
function RankPill({ children }: RankProps) {
  return <span className="diff-rank">{children}</span>;
}

type DeltaProps = { delta: number };
function DeltaBadge({ delta }: DeltaProps) {
  const label = delta > 0 ? `↑${delta}` : `↓${Math.abs(delta)}`;
  const toneClass = delta > 0 ? "diff-delta-up" : "diff-delta-down";
  return <span className={`diff-delta ${toneClass}`}>{label}</span>;
}

type RunningProps = { persona: string };
function RunningCopy({ persona }: RunningProps) {
  return (
    <div className="diff-running">
      <div className="diff-spinner" aria-hidden />
      <div>
        <div className="diff-running-title">Simulating activity…</div>
        <div className="diff-running-copy">
          We’re streaming a five-event burst for {persona}. Updated
          recommendations arrive momentarily.
        </div>
      </div>
    </div>
  );
}

type ErrorProps = { message?: string; onClose: () => void };
function ErrorCopy({ message, onClose }: ErrorProps) {
  return (
    <div>
      <p className="diff-error">
        Couldn’t replay this scenario: {message || "Unexpected error."}
      </p>
      <div className="diff-actions">
        <button onClick={onClose} className="diff-secondary" type="button">
          Close
        </button>
      </div>
    </div>
  );
}
