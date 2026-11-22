import Link from "next/link";
import { getSites, getSummary } from "@/lib/api";
import { Stat } from "@/components/Stat";
import { SocBar, StatusBadge, formatPower } from "@/components/SocBar";
import { AutoRefresh } from "@/components/AutoRefresh";

// Server Component: data is fetched on the server, the browser receives ready HTML.
export const dynamic = "force-dynamic";

export default async function OverviewPage() {
    // Both requests run in parallel, not one after the other.
  const [sites, summary] = await Promise.all([getSites(), getSummary()]);

  return (
    <>
      <div className="page-head">
        <div>
          <h1>Sites</h1>
            <p className="muted">{summary.siteCount} sites</p>
        </div>
        {/* Small client component that asks the server to re-render every 5 s */}
        <AutoRefresh seconds={5} />
      </div>

      <div className="stats">
        <Stat label="Average SoC" value={`${summary.averageSoC.toFixed(1)} %`} />
        <Stat label="Total power" value={formatPower(summary.totalPowerMW)} />
        <Stat label="Charging" value={summary.chargingCount} />
        <Stat label="Discharging" value={summary.dischargingCount} />
        <Stat label="Idle" value={summary.idleCount} />
      </div>

      <div className="grid">
        {sites.map((site) => (
          <Link key={site.id} href={`/sites/${site.id}`} className="card">
            <div className="card-head">
              <h2>{site.name}</h2>
              <StatusBadge status={site.status} />
            </div>
            <div className="soc-value">{site.soc.toFixed(1)} %</div>
            <SocBar soc={site.soc} />
            <div className="card-foot muted">
              <span>{formatPower(site.powerMW)}</span>
              <span>
                {site.capacityMWh} MWh / {site.maxPowerMW} MW
              </span>
            </div>
          </Link>
        ))}
      </div>
    </>
  );
}
