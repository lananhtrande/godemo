"use client";

import { useEffect, useState } from "react";
import { BROWSER_API_URL, type Site } from "@/lib/api";
import { SocBar, StatusBadge, formatPower } from "@/components/SocBar";
import { Sparkline } from "@/components/Sparkline";

type Connection = "connecting" | "live" | "reconnecting";

const MAX_POINTS = 120; // keep the last 2 minutes in the browser
const STALE_AFTER_MS = 5000;

export function LiveSite({ initial }: { initial: Site }) {
  const [site, setSite] = useState<Site>(initial);
  const [history, setHistory] = useState<number[]>([initial.soc]);
  const [connection, setConnection] = useState<Connection>("connecting");
  const [now, setNow] = useState(() => Date.now());

  // Open one SSE connection per page; the browser reconnects automatically on errors.
  useEffect(() => {
    const source = new EventSource(
      `${BROWSER_API_URL}/api/v1/sites/${initial.id}/stream`,
    );
    source.onopen = () => setConnection("live");
    source.onerror = () => setConnection("reconnecting");
    source.onmessage = (event) => {
      const update: Site = JSON.parse(event.data);
      setSite(update);
      // Immutable update: create a new array instead of pushing into the old one.
      setHistory((h) => [...h, update.soc].slice(-MAX_POINTS));
    };
    return () => source.close(); // cleanup: no open connection after leaving the page
  }, [initial.id]);

  // Tick once per second so the "last update" age stays correct.
  useEffect(() => {
    const t = setInterval(() => setNow(Date.now()), 1000);
    return () => clearInterval(t);
  }, []);

  const ageMs = now - new Date(site.updatedAt).getTime();
  const stale = ageMs > STALE_AFTER_MS;

  return (
    <section>
      <div className="page-head">
        <div>
          <h1>{site.name}</h1>
          <p className="muted">
            {site.capacityMWh} MWh · max {site.maxPowerMW} MW
          </p>
        </div>
        <span className={`conn conn-${stale ? "stale" : connection}`}>
          {stale
            ? `stale · last update ${Math.round(ageMs / 1000)}s ago`
            : connection === "live"
              ? "● live"
              : connection}
        </span>
      </div>

      <div className="detail">
        <div className="panel">
          <div className="label">State of charge</div>
          <div className="soc-big">{site.soc.toFixed(1)} %</div>
          <SocBar soc={site.soc} />
        </div>
        <div className="panel">
          <div className="label">Power</div>
          <div className="soc-big">{formatPower(site.powerMW)}</div>
          <StatusBadge status={site.status} />
        </div>
      </div>

      <div className="panel">
        <div className="label">SoC, last {history.length} updates</div>
        <Sparkline values={history} min={0} max={100} />
      </div>

      <p className="muted small">
        Updated {new Date(site.updatedAt).toLocaleTimeString()} (UTC from the
        server, shown in your local time)
      </p>
    </section>
  );
}
