import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight } from "lucide-react";
import type { KeyboardEvent, ReactNode } from "react";
import { useEffect, useMemo, useRef, useState } from "react";
import type { Report } from "../../model/report";
import { wallClock } from "../../model/report";
import type { ChipPlace, EventMark } from "../../model/timeline";
import { eventMarks, nextEvent, placeChips, previousEvent } from "../../model/timeline";
import styles from "./RunScreen.module.css";

interface ScrubberProps {
  readonly startNs: number;
  readonly endNs: number;
  readonly positionNs: number;
  readonly report: Report | null;
  readonly onSeek: (ns: number) => void;
  readonly onEvent?: (mark: EventMark) => void;
}

interface LaneProps {
  readonly marks: readonly EventMark[];
  readonly fraction: (ns: number) => number;
  readonly startNs: number;
  readonly onSeek: (ns: number) => void;
  readonly onEvent: ((mark: EventMark) => void) | undefined;
}

const chipText = (mark: EventMark, startNs: number) => `${wallClock(mark.atNs, startNs)} ${mark.label}`;
const chipWidth = (text: string) => Math.min(320, 20 + text.length * 7);

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

interface ChipProps {
  readonly mark: EventMark;
  readonly chip: ChipPlace;
  readonly text: string;
  readonly startNs: number;
  readonly pick: (mark: EventMark) => void;
}

function Chip({ mark, chip, text, startNs, pick }: ChipProps) {
  return (
    <li className={styles.laneItem} style={{ left: `${String(chip.left)}px`, top: `${String(chip.row * 28)}px` }}>
      <button
        type="button"
        className={mark.tone === "fault" ? styles.chipFault : styles.chipEvent}
        title={`${wallClock(mark.atNs, startNs)} · ${mark.raw}`}
        onClick={() => {
          pick(mark);
        }}
      >
        {text}
      </button>
    </li>
  );
}

function Lane({ marks, fraction, startNs, onSeek, onEvent }: LaneProps) {
  const { ref, width } = useWidth();
  const texts = marks.map((mark) => chipText(mark, startNs));
  const chips = placeChips(
    marks.map((mark) => fraction(mark.atNs) * width),
    texts.map(chipWidth),
    { width, rows: 2, gap: 4 },
  );
  const hidden = chips.filter((chip) => chip.row < 0).length;
  const pick = (mark: EventMark) => {
    if (onEvent === undefined) onSeek(mark.atNs);
    else onEvent(mark);
  };
  return (
    <>
      <ol
        ref={ref}
        className={styles.lane}
        aria-label="Events"
        style={{ height: `${String((Math.max(0, ...chips.map((chip) => chip.row)) + 1) * 28)}px` }}
      >
        {marks.map((mark, index) => {
          const chip = chips[index];
          return chip === undefined || chip.row < 0 ? null : (
            <Chip
              key={`${String(mark.atNs)}${mark.label}`}
              mark={mark}
              chip={chip}
              text={texts[index] ?? ""}
              startNs={startNs}
              pick={pick}
            />
          );
        })}
      </ol>
      {hidden > 0 && <p className={styles.laneMore}>{`${String(hidden)} more events: use Previous and Next event`}</p>}
    </>
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
      aria-label={target === null ? `${label}: none` : label}
      title={target === null ? `${label}: none` : label}
      aria-disabled={target === null}
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

export function Scrubber({ startNs, endNs, positionNs, report, onSeek, onEvent }: ScrubberProps) {
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
        <PlaybackRange startNs={startNs} span={span} positionNs={positionNs} onSeek={onSeek} onKeyDown={onKeyDown} />
      </div>
      <p id="scrub-keys" className="sr-only">
        Arrow keys move one second, Shift and arrow keys one minute, Page Up and Page Down further.
      </p>
      <Lane marks={marks} fraction={fraction} startNs={startNs} onSeek={onSeek} onEvent={onEvent} />
    </div>
  );
}
