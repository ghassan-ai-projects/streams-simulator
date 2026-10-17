import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight } from "lucide-react";
import type { CSSProperties, KeyboardEvent, ReactNode } from "react";
import { useMemo, useState } from "react";
import type { Report } from "../../../model/review";
import type { EventMark } from "../../../model/review";
import { eventMarks, firstEventLine, nextEvent, previousEvent, runClock, timeTicks } from "../../../model/review";
import runStyles from "./PlayControls.module.css";
import styles from "./Scrubber.module.css";
import { Lane, Legend, MarkRail, usePick } from "./ScrubberLane";
import { useEventKeys } from "./useEventKeys";
import { AlarmJump } from "./AlarmJump";
import type { PlayProps } from "./PlayControls";
import { PlayControls, StepButton } from "./PlayControls";

interface ScrubberProps extends PlayProps {
  readonly startNs: number;
  readonly endNs: number;
  readonly positionNs: number;
  readonly report: Report | null;
  readonly onSeek: (ns: number) => void;
  readonly onEvent: (mark: EventMark) => void;
}

interface TicksProps {
  readonly startNs: number;
  readonly endNs: number;
  readonly fraction: (ns: number) => number;
}

function Ticks({ startNs, endNs, fraction }: TicksProps) {
  return (
    <div className={styles.ticks} aria-hidden="true">
      {timeTicks(0, endNs - startNs).map((offset) => (
        <span key={offset} style={{ left: `${String(fraction(startNs + offset) * 100)}%` }}>
          {runClock(startNs + offset, startNs)}
        </span>
      ))}
    </div>
  );
}

interface StepsProps {
  readonly marks: readonly EventMark[];
  readonly positionNs: number;
  readonly clamp: (ns: number) => number;
  readonly onSeek: (ns: number) => void;
}

const keyOf: Readonly<Record<string, string | undefined>> = { "Previous event": ",", "Next event": "." };

function Steps({ marks, positionNs, clamp, onSeek }: StepsProps) {
  const previous = previousEvent(marks, positionNs);
  const next = nextEvent(marks, positionNs);
  const button = (label: string, icon: ReactNode, target: number | null) => (
    <StepButton
      label={target === null ? `${label}: none` : label}
      title={target === null ? `${label}: none` : `${label}${keyOf[label] === undefined ? "" : ` (${keyOf[label]})`}`}
      shortcut={keyOf[label]}
      disabled={target === null}
      onActivate={() => {
        if (target !== null) onSeek(target);
      }}
    >
      {icon}
    </StepButton>
  );
  return (
    <div className={runStyles.steps} role="group" aria-label="Playback steps">
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
  const played = `${String(((positionNs - startNs) / span) * 100)}%`;
  return (
    <input
      type="range"
      className={styles.range}
      style={{ "--played": played } as CSSProperties}
      min={0}
      max={Math.round(span / 1e9)}
      value={Math.round((positionNs - startNs) / 1e9)}
      onChange={(event) => {
        onSeek(startNs + Number(event.target.value) * 1e9);
      }}
      onKeyDown={onKeyDown}
      aria-label="Playback time"
      aria-valuetext={runClock(positionNs, startNs)}
      aria-describedby="scrub-keys"
    />
  );
}

function FirstEvent({ report, startNs }: { readonly report: Report | null; readonly startNs: number }) {
  const line = firstEventLine(report, startNs);
  if (line === null) return null;
  return (
    <p className={styles.first} title={line}>
      {line}
    </p>
  );
}

const unknownClock = "--:--:--";

interface RowProps extends ScrubberProps {
  readonly marks: readonly EventMark[];
  readonly clamp: (ns: number) => number;
}

function TimeRow(props: RowProps) {
  const { startNs, endNs, positionNs, marks, clamp } = props;
  const known = props.report !== null;
  const [said, setSaid] = useState("");
  const onSeek = (ns: number) => {
    setSaid(`Moved to ${runClock(ns, startNs)}`);
    props.onSeek(ns);
  };
  useEventKeys(marks, positionNs, onSeek);
  const onEvent = (mark: EventMark) => {
    setSaid(`Moved to ${runClock(mark.atNs, startNs)}, ${mark.label}`);
    props.onEvent(mark);
  };
  return (
    <div className={styles.timeRow}>
      <p className="sr-only" aria-live="polite">
        {said}
      </p>
      <PlayControls {...props} />
      <span className={styles.time}>{known ? runClock(positionNs, startNs) : unknownClock}</span>
      <span className={styles.span}>{`of ${known ? runClock(endNs, startNs) : unknownClock}`}</span>
      <span className={styles.controls}>
        <Legend marks={marks} />
        <Steps marks={marks} positionNs={positionNs} clamp={clamp} onSeek={onSeek} />
        <AlarmJump marks={marks} positionNs={positionNs} onEvent={onEvent} />
      </span>
    </div>
  );
}

export function Scrubber(props: ScrubberProps) {
  const { startNs, endNs, positionNs, report, onSeek, onEvent } = props;
  const span = Math.max(1, endNs - startNs);
  const fraction = (ns: number) => Math.min(1, Math.max(0, (ns - startNs) / span));
  const clamp = (ns: number) => Math.min(endNs, Math.max(startNs, ns));
  const marks = useMemo(() => eventMarks(report), [report]);
  const { said, pick } = usePick(startNs, onEvent);
  const onKeyDown = (event: KeyboardEvent<HTMLInputElement>) => {
    if (!event.shiftKey || (event.key !== "ArrowLeft" && event.key !== "ArrowRight")) return;
    event.preventDefault();
    onSeek(clamp(positionNs + (event.key === "ArrowLeft" ? -60e9 : 60e9)));
  };
  return (
    <div className={styles.scrubber}>
      <TimeRow {...props} marks={marks} clamp={clamp} />
      <div className={styles.axis}>
        <Ticks startNs={startNs} endNs={endNs} fraction={fraction} />
      </div>
      <div className={styles.trackWrap}>
        <PlaybackRange startNs={startNs} span={span} positionNs={positionNs} onSeek={onSeek} onKeyDown={onKeyDown} />
        <MarkRail marks={marks} fraction={fraction} pick={pick} />
      </div>
      <p id="scrub-keys" className="sr-only">
        Arrow keys move one second, Shift and arrow keys one minute, Page Up and Page Down further.
      </p>
      <Lane
        marks={marks}
        fraction={fraction}
        startNs={startNs}
        reading={report === null}
        positionNs={positionNs}
        said={said}
        pick={pick}
      />
      <FirstEvent report={report} startNs={startNs} />
    </div>
  );
}
