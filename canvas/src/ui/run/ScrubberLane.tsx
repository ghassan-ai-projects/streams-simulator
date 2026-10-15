import { useEffect, useRef, useState } from "react";
import type { ChipPlace, EventMark, MarkKind } from "../../model/timeline";
import { activeAt, kindsOf, markWords, placeChips, runClock } from "../../model/timeline";
import { cx } from "../cx";
import styles from "./Scrubber.module.css";

export interface LaneProps {
  readonly marks: readonly EventMark[];
  readonly fraction: (ns: number) => number;
  readonly startNs: number;
  readonly reading: boolean;
  readonly positionNs: number;
  readonly said: string;
  readonly pick: (mark: EventMark) => void;
}

const chipText = (mark: EventMark, startNs: number) => `${runClock(mark.atNs, startNs)} ${mark.label}`;
const chipShort = (mark: EventMark, startNs: number) => `${runClock(mark.atNs, startNs)} ${mark.entity}`;
const chipWidth = (text: string) => Math.min(200, 34 + text.length * 7);
const percent = (value: number) => `${String(value * 100)}%`;

function Swatch({ kind }: { readonly kind: MarkKind }) {
  return <span className={cx(styles.swatch, styles[`mark-${kind}`])} aria-hidden="true" />;
}

export function MarkRail({ marks, fraction, pick }: Pick<LaneProps, "marks" | "fraction" | "pick">) {
  return (
    <div className={styles.rail} aria-hidden="true">
      {marks.map((mark) => (
        <button
          type="button"
          key={`${String(mark.atNs)}${mark.label}`}
          tabIndex={-1}
          className={styles.railMark}
          title={mark.raw}
          style={{ left: percent(fraction(mark.atNs)) }}
          onClick={() => {
            pick(mark);
          }}
        >
          <Swatch kind={mark.kind} />
        </button>
      ))}
    </div>
  );
}

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
  readonly active: boolean;
  readonly pick: (mark: EventMark) => void;
}

function Chip({ mark, chip, text, startNs, active, pick }: ChipProps) {
  return (
    <li className={styles.laneItem} style={{ left: `${String(chip.left)}px`, top: `${String(chip.row * 28)}px` }}>
      <button
        type="button"
        className={active ? styles.chipActive : styles.chip}
        title={`${runClock(mark.atNs, startNs)} · ${mark.raw}`}
        aria-label={chipText(mark, startNs)}
        onClick={() => {
          pick(mark);
        }}
      >
        <Swatch kind={mark.kind} />
        {text}
      </button>
    </li>
  );
}

export function usePick(
  startNs: number,
  onSeek: (ns: number) => void,
  onEvent: ((mark: EventMark) => void) | undefined,
) {
  const [said, setSaid] = useState("");
  const pick = (mark: EventMark) => {
    if (onEvent === undefined) onSeek(mark.atNs);
    else onEvent(mark);
    setSaid(`${chipText(mark, startNs)}: moved to ${runClock(mark.atNs, startNs)}, ${mark.entity} selected`);
  };
  return { said, pick };
}

export function Legend({ marks }: { readonly marks: readonly EventMark[] }) {
  const kinds = kindsOf(marks);
  if (kinds.length === 0) return null;
  return (
    <ul className={styles.legend} aria-label="Event types">
      {kinds.map((kind) => (
        <li key={kind}>
          <Swatch kind={kind} />
          {markWords[kind]}
        </li>
      ))}
    </ul>
  );
}

interface NotesProps {
  readonly said: string;
  readonly hidden: number;
}

function LaneNotes({ said, hidden }: NotesProps) {
  return (
    <div className={styles.laneNotes}>
      <p className="sr-only" aria-live="polite">
        {said}
      </p>
      {hidden > 0 && (
        <p className={styles.laneMore}>
          {`${String(hidden)} more ${hidden === 1 ? "event" : "events"}: use Previous and Next event`}
        </p>
      )}
    </div>
  );
}

function laneHeight(chips: readonly ChipPlace[]): string {
  return `${String((Math.max(0, ...chips.map((chip) => chip.row)) + 1) * 28)}px`;
}

function EmptyLane({ reading }: { readonly reading: boolean }) {
  if (!reading) return null;
  return <p className={styles.laneEmpty}>Events appear once the run is read.</p>;
}

export function Lane({ marks, fraction, startNs, reading, positionNs, said, pick }: LaneProps) {
  const { ref, width } = useWidth();
  const texts = marks.map((mark) => chipShort(mark, startNs));
  const chips = placeChips(
    marks.map((mark) => fraction(mark.atNs) * width),
    texts.map(chipWidth),
    { width, rows: 3, gap: 4 },
  );
  return (
    <>
      {marks.length === 0 && <EmptyLane reading={reading} />}
      <ol ref={ref} className={styles.lane} aria-label="Events" style={{ height: laneHeight(chips) }}>
        {marks.map((mark, index) => {
          const chip = chips[index];
          return chip === undefined || chip.row < 0 ? null : (
            <Chip
              key={`${String(mark.atNs)}${mark.label}`}
              mark={mark}
              chip={chip}
              text={texts[index] ?? ""}
              startNs={startNs}
              active={activeAt(mark, positionNs)}
              pick={pick}
            />
          );
        })}
      </ol>
      <LaneNotes said={said} hidden={chips.filter((chip) => chip.row < 0).length} />
    </>
  );
}
