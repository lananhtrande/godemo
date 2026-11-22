// A tiny SVG line chart, no chart library needed.
export function Sparkline({
  values,
  min,
  max,
  width = 800,
  height = 160,
}: {
  values: number[];
  min: number;
  max: number;
  width?: number;
  height?: number;
}) {
  if (values.length < 2) {
    return <div className="muted small">Waiting for data…</div>;
  }
  const step = width / (values.length - 1);
  const y = (v: number) => height - ((v - min) / (max - min)) * height;
  const points = values.map((v, i) => `${(i * step).toFixed(1)},${y(v).toFixed(1)}`);

  return (
    <svg
      viewBox={`0 0 ${width} ${height}`}
      className="sparkline"
      preserveAspectRatio="none"
      role="img"
      aria-label="State of charge over time"
    >
      {[25, 50, 75].map((g) => (
        <line key={g} x1={0} x2={width} y1={y(g)} y2={y(g)} className="grid-line" />
      ))}
      <polyline points={points.join(" ")} className="line" />
    </svg>
  );
}
