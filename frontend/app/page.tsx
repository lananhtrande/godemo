import Link from "next/link";
import { getSites } from "@/lib/api";
import { SocBar, StatusBadge, formatPower } from "@/components/SocBar";
import { AutoRefresh } from "@/components/AutoRefresh";

// Server Component: data is fetched on the server, the browser receives ready HTML.
export const dynamic = "force-dynamic";

export default async function OverviewPage() {
  const sites = await getSites();
  const totalPower = sites.reduce((sum, s) => sum + s.powerMW, 0);

  return (
    <>
      <div className="page-head">
        <div>
          <h1>Sites</h1>
          <p className="muted">
            {sites.length} sites · total power {formatPower(totalPower)}
          </p>
        </div>
        {/* Small client component that asks the server to re-render every 5 s */}
        <AutoRefresh seconds={5} />
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
