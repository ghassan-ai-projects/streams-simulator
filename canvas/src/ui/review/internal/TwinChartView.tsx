import type { ReactNode } from "react";
import type { TwinChart, TwinCoverage } from "../../../model/review";
import { chartFrame, chartOf, metricLine, truncationNote } from "../../../model/review";
import { codeTitle, displayUnit } from "../../../model/shared/units";
import styles from "./Twin.module.css";

type Series = Extract<TwinCoverage, { kind: "series" }>;

const { width, height, left, right, top, bottom } = chartFrame;
const plotWidth = width - left - right;
const plotHeight = height - top - bottom;

function Swatch({ children }: { readonly children: ReactNode }) {
  return (
    <svg className={styles.swatch} viewBox="0 0 26 12" aria-hidden="true" focusable="false">
      {children}
    </svg>
  );
}

function Legend() {
  return (
    <ul className={styles.legend} aria-label="Legend">
      <li>
        <Swatch>
          <line className={styles.boardLine} x1="1" y1="6" x2="25" y2="6" />
          <circle className={styles.boardMark} cx="13" cy="6" r="3" />
        </Swatch>
        Board reading
      </li>
      <li>
        <Swatch>
          <line className={styles.twinLine} x1="1" y1="6" x2="25" y2="6" />
          <rect className={styles.twinMark} x="10" y="3" width="6" height="6" />
        </Swatch>
        Twin prediction
      </li>
      <li>
        <Swatch>
          <rect className={styles.shade} x="2" y="1" width="22" height="10" />
        </Swatch>
        Where they part ways
      </li>
    </ul>
  );
}

function Axes({ chart }: { readonly chart: TwinChart }) {
  return (
    <>
      <rect className={styles.frame} x={left} y={top} width={plotWidth} height={plotHeight} />
      {chart.yTicks.map((tick) => (
        <g key={`y${String(tick.at)}`}>
          <line className={styles.grid} x1={left} x2={left + plotWidth} y1={tick.at} y2={tick.at} />
          <text className={styles.tick} x={left - 4} y={tick.at + 4} textAnchor="end">
            {tick.label}
          </text>
        </g>
      ))}
      {chart.xTicks.map((tick, index) => (
        <text key={index} className={styles.tick} x={tick.at} y={height - 6} textAnchor={index === 0 ? "start" : "end"}>
          {tick.label}
        </text>
      ))}
    </>
  );
}

function Lines({ chart }: { readonly chart: TwinChart }) {
  return (
    <>
      {chart.shades.map((shade) => (
        <rect key={shade.name} className={styles.shade} x={shade.x} y={top} width={shade.width} height={plotHeight} />
      ))}
      <polyline className={styles.boardLine} points={chart.board} />
      <polyline className={styles.twinLine} points={chart.twin} />
      {chart.boardMarks.map((mark, index) => (
        <circle key={index} className={styles.boardMark} cx={mark.x} cy={mark.y} r={2.5} />
      ))}
      {chart.twinMarks.map((mark, index) => (
        <rect key={index} className={styles.twinMark} x={mark.x - 2.5} y={mark.y - 2.5} width={5} height={5} />
      ))}
      {chart.cursor !== null && (
        <line className={styles.cursor} x1={chart.cursor} x2={chart.cursor} y1={top} y2={top + plotHeight} />
      )}
    </>
  );
}

function describe(series: Series): string {
  const { metric, divergences } = series;
  const stretches = `${String(divergences.length)} ${divergences.length === 1 ? "stretch" : "stretches"} where they part ways`;
  return `${metric.metric}: the board's readings as a solid line, the twin's prediction as a dashed line, ${stretches}.`;
}

interface ChartViewProps {
  readonly series: Series;
  readonly startNs: number;
  readonly cursorNs: number;
}

export function TwinChartView({ series, startNs, cursorNs }: ChartViewProps) {
  const { metric, divergences, undrawn } = series;
  const chart = chartOf(metric, divergences, startNs, cursorNs);
  const symbol = displayUnit(metric.unit);
  const unit = symbol === "" ? "" : ` (${symbol})`;
  const truncated = truncationNote(metric.series, undrawn);
  return (
    <figure className={styles.figure}>
      <figcaption
        className={styles.heading}
        title={codeTitle(metric.unit) || undefined}
      >{`${metric.metric}${unit}`}</figcaption>
      <svg
        className={styles.chart}
        viewBox={`0 0 ${String(width)} ${String(height)}`}
        role="img"
        aria-label={describe(series)}
      >
        <Axes chart={chart} />
        <Lines chart={chart} />
      </svg>
      <Legend />
      <p className={styles.meta}>{metricLine(metric)}</p>
      {truncated !== "" && <p className={styles.meta}>{truncated}</p>}
    </figure>
  );
}
