export function ContextGauge({ used, window: win }: { used: number; window: number }) {
  // No data yet (no turn has reported usage) — hide rather than show 0%.
  if (!win) return null;
  const pct = Math.min(100, Math.round((used / win) * 100));
  return (
    <div className="gauge" title={`${used.toLocaleString()} / ${win.toLocaleString()} context tokens`}>
      <div className="gauge__bar">
        <div className="gauge__fill" style={{ width: `${pct}%` }} />
      </div>
      <span className="gauge__label">{pct}%</span>
    </div>
  );
}
