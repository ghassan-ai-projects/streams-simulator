import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight } from "lucide-react";
import type { KeyboardEvent, ReactNode } from "react";
import { useEffect, useMemo, useRef, useState } from "react";
import type { Report } from "../../model/report";
import { wallClock } from "../../model/report";
import type { EventMark } from "../../model/timeline";
import { eventMarks, laneRows, nextEvent, previousEvent } from "../../model/timeline";
import styles from "./RunScreen.module.css";

interface ScrubberProps {
  readonly startNs: number;
  readonly endNs: number;
  readonly positionNs: number;
  readonly report: Report | null;
  readonly density: readonly number[];
  readonly onSeek: (ns: number) => void;
}

function Density({ density }: { readonly density: readonly number[] }) {
  const peak = Math.max(1, ...density);
  return (
    <div className={styles.density} aria-hidden="true">
      {density.map((count, index) => (
        <span key={index} style={{ height: `${String(Math.round((count / peak) * 100))}%` }} />
      ))}
    </div>
  );
}

interface LaneProps {
  readonly marks: readonly EventMark[];
  readonly fraction: (ns: number) => number;
  readonly startNs: number;
  readonly onSeek: (ns: number) => void;
}

const chipWidth = 168;

function useWidth() {
  const [width, setWidth] = useState(800);
  const ref = useRef<HTMLOListElement>(null);
  useEffect(() => {
    const node = ref.current;
    if (node === null || typeof ResizeObserver === "undefined") return undefined;
    const observer = new ResizeObserver(() => {
      setWidth(node.clientWidth || 800);
    });
    observer.observe(node);
    return () => {
      observer.disconnect();
    };
  }, []);
  return { ref, width };
}

function Lane({ marks, fraction, startNs, onSeek }: LaneProps) {
  const { ref, width } = useWidth();
  const rows = laneRows(
    marks.map((mark) => fraction(mark.atNs) * width),
    chipWidth + 4,
    2,
  );
  return (
    <ol
      ref={ref}
      className={styles.lane}
      aria-label="Events"
      style={{ height: `${String((Math.max(0, ...rows) + 1) * 28)}px` }}
    >
      {marks.map((mark, index) => (
        <li
          key={`${String(mark.atNs)}${mark.label}`}
          style={{ left: `${String(fraction(mark.atNs) * 100)}%`, top: `${String((rows[index] ?? 0) * 28)}px` }}
          className={styles.laneItem}
        >
          <button
            type="button"
            className={mark.tone === "fault" ? styles.chipFault : styles.chipEvent}
            title={`${wallClock(mark.atNs, startNs)} · ${mark.raw}`}
            onClick={() => {
              onSeek(mark.atNs);
            }}
          >
            {`${wallClock(mark.atNs, startNs)} ${mark.label}`}
          </button>
        </li>
      ))}
    </ol>
  );
}

interface StepsProps {
  readonly marks: readonly EventMark[];
  readonly positionNs: number;
  readonly clamp: (ns: number) => number;
  readonly onSeek: (ns: number) => void;
}

function Steps({ marks, positionNs, clamp, onSeek }: StepsProps) {
  const previous = previousEvent(marks, positionNs);
  const next = nextEvent(marks, positionNs);
  const button = (label: string, icon: ReactNode, target: number | null) => (
    <button
      type="button"
      className={styles.step}
      aria-label={label}
      title={label}
      disabled={target === null}
      onClick={() => {
        if (target !== null) onSeek(target);
      }}
    >
      {icon}
    </button>
  );
  return (
    <div className={styles.steps} role="group" aria-label="Playback steps">
      {button("Previous event", <ChevronsLeft size={14} aria-hidden="true" />, previous)}
      {button("Back 10 minutes", <ChevronLeft size={14} aria-hidden="true" />, clamp(positionNs - 600e9))}
      {button("Forward 10 minutes", <ChevronRight size={14} aria-hidden="true" />, clamp(positionNs + 600e9))}
      {button("Next event", <ChevronsRight size={14} aria-hidden="true" />, next)}
    </div>
  );
}

interface RangeProps {
  readonly startNs: number;
  readonly span: number;
  readonly positionNs: number;
  readonly onSeek: (ns: number) => void;
  readonly onKeyDown: (event: KeyboardEvent<HTMLInputElement>) => void;
}

function PlaybackRange({ startNs, span, positionNs, onSeek, onKeyDown }: RangeProps) {
  return (
    <input
      type="range"
      className={styles.range}
      min={0}
      max={Math.round(span / 1e9)}
      value={Math.round((positionNs - startNs) / 1e9)}
      onChange={(event) => {
        onSeek(startNs + Number(event.target.value) * 1e9);
      }}
      onKeyDown={onKeyDown}
      aria-label="Playback time"
      aria-valuetext={wallClock(positionNs, startNs)}
      aria-describedby="scrub-keys"
    />
  );
}

export function Scrubber({ startNs, endNs, positionNs, report, density, onSeek }: ScrubberProps) {
  const span = Math.max(1, endNs - startNs);
  const fraction = (ns: number) => Math.min(1, Math.max(0, (ns - startNs) / span));
  const clamp = (ns: number) => Math.min(endNs, Math.max(startNs, ns));
  const marks = useMemo(() => eventMarks(report), [report]);
  const onKeyDown = (event: KeyboardEvent<HTMLInputElement>) => {
    if (!event.shiftKey || (event.key !== "ArrowLeft" && event.key !== "ArrowRight")) return;
    event.preventDefault();
    onSeek(clamp(positionNs + (event.key === "ArrowLeft" ? -60e9 : 60e9)));
  };
  return (
    <div className={styles.scrubber}>
      <div className={styles.timeRow}>
        <span className={styles.time}>{wallClock(positionNs, startNs)}</span>
        <span className={styles.muted}>{`of ${wallClock(startNs)} – ${wallClock(endNs, startNs)} UTC`}</span>
        <Steps marks={marks} positionNs={positionNs} clamp={clamp} onSeek={onSeek} />
      </div>
      <div className={styles.track}>
        <Density density={density} />
        <PlaybackRange startNs={startNs} span={span} positionNs={positionNs} onSeek={onSeek} onKeyDown={onKeyDown} />
      </div>
      <p id="scrub-keys" className="sr-only">
        Arrow keys move one second, Shift and arrow keys one minute, Page Up and Page Down further.
      </p>
      <Lane marks={marks} fraction={fraction} startNs={startNs} onSeek={onSeek} />
    </div>
  );
}
