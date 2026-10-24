import type { Trend } from "../../../model/review";
import { trendFrame } from "../../../model/review";
import styles from "./ObservedTable.module.css";

const { width, height } = trendFrame;

function pointsOf(points: Trend["segments"][number]): string {
  return points.map((point) => `${point.x.toFixed(1)},${point.y.toFixed(1)}`).join(" ");
}

export function MiniTrend({
  trend,
  label,
  note,
}: {
  readonly trend: Trend;
  readonly label: string;
  readonly note: string;
}) {
  return (
    <svg
      className={styles.trend}
      data-part="trend"
      data-gaps={trend.gaps}
      viewBox={`0 0 ${String(width)} ${String(height)}`}
      role="img"
      aria-label={`${label} trend. ${note}`}
      focusable="false"
    >
      <title>{note}</title>
      {trend.segments.map((points) => (
        <polyline key={String(points[0]?.x)} className={styles.line} points={pointsOf(points)} />
      ))}
      {trend.dots.map((point) => (
        <circle key={String(point.x)} className={styles.dot} cx={point.x} cy={point.y} r={1.6} />
      ))}
    </svg>
  );
}
