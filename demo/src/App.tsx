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
      <section className="app-hero" aria-label="Executive demo setup">
        <header>
          <h1 className="app-hero-title">RecSys Demo</h1>
          <p className="app-hero-copy">
            Explore RecSys with realistic catalog data, and step into a stage
            experience created to give a clear view in just a few minutes.
          </p>
        </header>
        <div className="app-hero-actions">
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
