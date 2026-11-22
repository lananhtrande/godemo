import type { SiteStatus } from "@/lib/api";

const STATUS_LABEL: Record<SiteStatus, string> = {
  charging: "Charging",
  discharging: "Discharging",
  idle: "Idle",
};

export function StatusBadge({ status }: { status: SiteStatus }) {
  return <span className={`badge badge-${status}`}>{STATUS_LABEL[status]}</span>;
}

// Works in both server and client components (no state, no effects).
export function SocBar({ soc }: { soc: number }) {
  const level = soc < 20 ? "low" : soc < 50 ? "mid" : "high";
  return (
    <div className="soc-bar" aria-label={`State of charge ${soc.toFixed(1)} %`}>
      <div className={`soc-fill soc-${level}`} style={{ width: `${soc}%` }} />
    </div>
  );
}

export function formatPower(powerMW: number) {
  const sign = powerMW > 0 ? "+" : "";
  return `${sign}${powerMW.toFixed(2)} MW`;
}
