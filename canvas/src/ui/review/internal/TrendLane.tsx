import type { OverlayEvent, Span } from "../../../model/review";
import { eventLine, runClock, xOf } from "../../../model/review";
import styles from "./TrendPlot.module.css";

const laneHeight = 18;

function Marker({ event, span, width }: { readonly event: OverlayEvent; readonly span: Span; readonly width: number }) {
  const x = xOf(event.atNs, span, width);
  const middle = laneHeight / 2;
  if (event.kind === "fault") {
    const points = `${String(x)},2 ${String(x + 5)},${String(middle)} ${String(x)},${String(laneHeight - 2)} ${String(x - 5)},${String(middle)}`;
    return <polygon className={styles.fault} points={points} data-kind="fault" />;
  }
  const points = `${String(x - 5)},${String(laneHeight - 3)} ${String(x + 5)},${String(laneHeight - 3)} ${String(x)},3`;
  return <polygon className={event.refused ? styles.refused : styles.call} points={points} data-kind="effector" />;
}

export function EventLane({
  events,
  span,
  width,
}: {
  readonly events: readonly OverlayEvent[];
  readonly span: Span;
  readonly width: number;
}) {
  return (
    <svg
      className={styles.laneSvg}
      viewBox={`0 0 ${String(width)} ${String(laneHeight)}`}
      aria-hidden="true"
      focusable="false"
    >
      {events.map((event) => (
        <Marker key={event.key} event={event} span={span} width={width} />
      ))}
    </svg>
  );
}

export function TimeAxis({ span, startNs }: { readonly span: Span; readonly startNs: number }) {
  const middle = (span.startNs + span.endNs) / 2;
  return (
    <div className={styles.axisTicks} aria-hidden="true">
      {[span.startNs, middle, span.endNs].map((at) => (
        <span key={at}>{runClock(at, startNs)}</span>
      ))}
    </div>
  );
}

export function EventList({
  events,
  startNs,
  onPick,
}: {
  readonly events: readonly OverlayEvent[];
  readonly startNs: number;
  readonly onPick: (event: OverlayEvent) => void;
}) {
  if (events.length === 0) return null;
  return (
    <details className={styles.events}>
      <summary>{`Events in this window (${String(events.length)})`}</summary>
      <ul className={styles.eventList}>
        {events.map((event) => (
          <li key={event.key}>
            <button
              type="button"
              className={styles.eventButton}
              onClick={() => {
                onPick(event);
              }}
            >
              {eventLine(event, startNs)}
            </button>
          </li>
        ))}
      </ul>
    </details>
  );
}
