// One small statistic tile. No state, so it works in Server Components.
export function Stat({ label, value }: { label: string; value: string | number }) {
  return (
    <div className="stat">
      <div className="label">{label}</div>
      <div className="stat-value">{value}</div>
    </div>
  );
}