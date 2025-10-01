import "./App.css";
import { useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { seedMinimal } from "./services/seeding";
import { ensureApiBase } from "./lib/api";
import { generateDemoNamespace } from "./lib/namespaces";

export default function App() {
  const navigate = useNavigate();
  const [busy, setBusy] = useState(false);
  const ns = useMemo(() => generateDemoNamespace(), []);

  async function seedAndStart() {
    setBusy(true);
    try {
      const api = import.meta.env.VITE_API_BASE_URL || "http://localhost:8081";
      await seedMinimal(api, ns);
      // Navigate to Stage; persona/surface defaults
      const params = new URLSearchParams({
        ns,
        u: "user-1",
        s: "home_top",
        k: "10",
      });
      ensureApiBase(api);
      navigate(`/stage?${params.toString()}`);
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="app-shell">
      <header className="demo-title-header">
        <h1 className="demo-title">RecSys Demo</h1>
      </header>

      <section className="app-hero" aria-label="Executive demo setup">
        <h2 className="app-hero-title">
          Turn data into decisions that win customers.
        </h2>
        <p className="app-hero-copy">
          Experience how personalized recommendations and clear insights deliver
          both growth and confidence.
        </p>
        <div className="app-hero-actions">
          <button
            onClick={seedAndStart}
            disabled={busy}
            className="btn btn-primary btn-large"
          >
            {busy ? "Seeding demo…" : "See it in action"}
          </button>
        </div>
      </section>

      <section className="pitch-deck">
        <div className="pitch-cards">
          <div className="pitch-card">
            <h3 className="pitch-card-title">Personalized at Scale</h3>
            <p className="pitch-card-hook">
              Every customer feels like your only customer.
            </p>
            <p className="pitch-card-copy">
              Deliver spot-on recommendations that adapt as your catalog grows —
              keeping experiences relevant without extra effort.
            </p>
          </div>

          <div className="pitch-card">
            <h3 className="pitch-card-title">Decisions You Can Trust</h3>
            <p className="pitch-card-hook">No more black boxes.</p>
            <p className="pitch-card-copy">
              Every recommendation comes with a clear reason why — so you can
              explain it to your team, your board, or your customers.
            </p>
          </div>

          <div className="pitch-card">
            <h3 className="pitch-card-title">Balance AI With Business Rules</h3>
            <p className="pitch-card-hook">Smarts with control.</p>
            <p className="pitch-card-copy">
              Combine data-driven recommendations with your business priorities
              — highlight, boost, or block items to match your strategy.
            </p>
          </div>

          <div className="pitch-card">
            <h3 className="pitch-card-title">From Idea to Impact Fast</h3>
            <p className="pitch-card-hook">Prove value quickly.</p>
            <p className="pitch-card-copy">
              Test recommendations in days, not months. Start small, show
              results early, and scale when you're ready.
            </p>
          </div>

          <div className="pitch-card">
            <h3 className="pitch-card-title">Measure What Matters</h3>
            <p className="pitch-card-hook">Beyond clicks.</p>
            <p className="pitch-card-copy">
              Track engagement, retention, and revenue impact — and optimize for
              the outcomes that actually drive growth.
            </p>
          </div>

          <div className="pitch-card">
            <h3 className="pitch-card-title">Built for Trust and Growth</h3>
            <p className="pitch-card-hook">Confidence from day one.</p>
            <p className="pitch-card-copy">
              Privacy safeguards, full visibility into decisions, and proven
              scalability ensure your business is future-ready.
            </p>
          </div>
        </div>

        <div className="pitch-actions">
          <button
            onClick={seedAndStart}
            disabled={busy}
            className="btn btn-primary btn-large"
          >
            {busy ? "Seeding demo…" : "Start the demo"}
          </button>
        </div>
      </section>
    </div>
  );
}
