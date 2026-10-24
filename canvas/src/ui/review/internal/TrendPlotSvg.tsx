import type { KeyboardEvent, MouseEvent } from "react";
import type { Plot, Readout, Span } from "../../../model/review";
import { nsAtX, xOf } from "../../../model/review";
import { plotFrame } from "../../../model/review";
import { cx } from "../../primitives";
import styles from "./TrendPlot.module.css";

export interface PlotCursor {
  readonly span: Span;
  readonly width: number;
  readonly cursorNs: number;
  readonly probeNs: number | null;
  readonly marks: readonly number[];
}

export interface PlotHandlers {
  readonly onKey: (event: KeyboardEvent) => void;
  readonly onPoint: (ns: number) => void;
  readonly onLeave: () => void;
  readonly onPick: (ns: number) => void;
  readonly onBlur: () => void;
}

interface Props {
  readonly plot: Plot;
  readonly cursor: PlotCursor;
  readonly readout: Readout | null;
  readonly handlers: PlotHandlers;
}

const { height } = plotFrame;
const denseDivisor = 3;

function pointsOf(points: readonly { readonly x: number; readonly y: number }[]): string {
  return points.map((point) => `${point.x.toFixed(1)},${point.y.toFixed(1)}`).join(" ");
}

function VerticalLine({
  at,
  cursor,
  className,
}: {
  readonly at: number;
  readonly cursor: PlotCursor;
  readonly className: string | undefined;
}) {
  const x = xOf(at, cursor.span, cursor.width);
  return <line className={className} x1={x} x2={x} y1={0} y2={height} />;
}

function Backdrop({ plot, cursor }: Pick<Props, "plot" | "cursor">) {
  return (
    <>
      {cursor.marks.map((at) => (
        <VerticalLine key={at} at={at} cursor={cursor} className={styles.eventLine} />
      ))}
      {plot.bands.map((band) => (
        <rect
          key={`${String(band.x0)}-${band.quality}`}
          className={styles.band}
          data-quality={band.quality}
          x={band.x0}
          y={height - 4}
          width={Math.max(3, band.x1 - band.x0)}
          height={4}
        />
      ))}
    </>
  );
}

function Lines({ plot, width }: { readonly plot: Plot; readonly width: number }) {
  return (
    <>
      {plot.segments.map((points) => (
        <polyline
          key={String(points[0]?.x)}
          className={cx(styles.line, points.length > width / denseDivisor && styles.dense)}
          points={pointsOf(points)}
        />
      ))}
      {plot.dots.map((point) => (
        <line key={point.x} className={styles.dot} x1={point.x} x2={point.x} y1={point.y} y2={point.y} />
      ))}
    </>
  );
}

function Extreme({ tick, y, width }: { readonly tick: Plot["top"]; readonly y: number; readonly width: number }) {
  if (tick === null) return null;
  const atEnd = tick.side === "end";
  return (
    <text className={styles.tick} x={atEnd ? width - 3 : 3} y={y} textAnchor={atEnd ? "end" : "start"}>
      {tick.text}
    </text>
  );
}

function Extremes({ plot, width }: { readonly plot: Plot; readonly width: number }) {
  const { top, bottom } = plot;
  return (
    <>
      <Extreme tick={top} y={Math.max(9, (top?.y ?? 0) - 2)} width={width} />
      <Extreme
        tick={bottom !== null && bottom.text !== top?.text ? bottom : null}
        y={Math.min(height - 6, (bottom?.y ?? 0) + 10)}
        width={width}
      />
    </>
  );
}

function Drawing({ plot, cursor }: Pick<Props, "plot" | "cursor">) {
  return (
    <svg
      className={styles.plot}
      viewBox={`0 0 ${String(cursor.width)} ${String(height)}`}
      aria-hidden="true"
      focusable="false"
    >
      <Backdrop plot={plot} cursor={cursor} />
      <Lines plot={plot} width={cursor.width} />
      <Extremes plot={plot} width={cursor.width} />
      <VerticalLine at={cursor.cursorNs} cursor={cursor} className={styles.cursor} />
      {cursor.probeNs !== null && <VerticalLine at={cursor.probeNs} cursor={cursor} className={styles.probe} />}
    </svg>
  );
}

const unplottedWords = {
  empty: "No readings in this window",
  words: "States are not plotted; the reading at the cursor is shown beside it",
  plotted: "",
} as const;

function valueText(plot: Plot, readout: Readout | null): string {
  if (readout === null) return `${plot.label}: no reading yet at this time`;
  return `${plot.label}: ${readout.words}`;
}

function timeOf(event: MouseEvent<HTMLDivElement>, cursor: PlotCursor): number {
  const box = event.currentTarget.getBoundingClientRect();
  return nsAtX(event.clientX - box.left, cursor.span, box.width);
}

export function TrendPlotSvg({ plot, cursor, readout, handlers }: Props) {
  const { span } = cursor;
  const at = cursor.probeNs ?? cursor.cursorNs;
  return (
    <div
      className={styles.track}
      role="slider"
      tabIndex={0}
      aria-label={`${plot.label} trend${plot.unit === "" ? "" : `, ${plot.unit}`}`}
      aria-valuemin={0}
      aria-valuemax={Math.round((span.endNs - span.startNs) / 1e9)}
      aria-valuenow={Math.round((at - span.startNs) / 1e9)}
      aria-valuetext={valueText(plot, readout)}
      data-plot={plot.key}
      onKeyDown={handlers.onKey}
      onPointerMove={(event) => {
        handlers.onPoint(timeOf(event, cursor));
      }}
      onPointerLeave={handlers.onLeave}
      onClick={(event) => {
        handlers.onPick(timeOf(event, cursor));
      }}
      onBlur={handlers.onBlur}
    >
      {plot.state === "plotted" ? (
        <Drawing plot={plot} cursor={cursor} />
      ) : (
        <p className={styles.unplotted}>{unplottedWords[plot.state]}</p>
      )}
    </div>
  );
}
