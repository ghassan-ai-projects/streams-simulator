import { useRef, useState } from "react";
import type { ChipPlace, EventMark, InstanceOrigin, MarkKind } from "../../../model/review";
import { activeAt, kindsOf, markWords, placeChips, runClock } from "../../../model/review";
import { OriginMark } from "../../graph";
import { cx, useResizeObserver } from "../../primitives";
import styles from "./ScrubberLane.module.css";

export interface LaneProps {
  readonly marks: readonly EventMark[];
  readonly fraction: (ns: number) => number;
  readonly startNs: number;
  readonly reading: boolean;
  readonly positionNs: number;
  readonly said: string;
  readonly pick: (mark: EventMark) => void;
  readonly origins: ReadonlyMap<string, InstanceOrigin>;
}

const chipText = (mark: EventMark, startNs: number) => `${runClock(mark.atNs, startNs)} ${mark.label}`;
const eventWords = (mark: EventMark) =>
  mark.kind === "command" || mark.kind === "connection" || mark.entity === "" ? mark.label : mark.entity;
const chipShort = (mark: EventMark, startNs: number) => `${runClock(mark.atNs, startNs)} ${eventWords(mark)}`;
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
  useResizeObserver(ref, () => {
    const measured = ref.current?.clientWidth ?? 0;
    setWidth(measured || 800);
  });
  return { ref, width };
}

interface ChipProps {
  readonly mark: EventMark;
  readonly chip: ChipPlace;
  readonly text: string;
  readonly startNs: number;
  readonly active: boolean;
  readonly pick: (mark: EventMark) => void;
  readonly origin: InstanceOrigin | null;
}

function Chip({ mark, chip, text, startNs, active, pick, origin }: ChipProps) {
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
        <OriginMark origin={origin} variant="glyph" voice="silent" />
        <span className={styles.chipText}>{text}</span>
      </button>
    </li>
  );
}

export function usePick(startNs: number, onEvent: (mark: EventMark) => void) {
  const [said, setSaid] = useState("");
  const pick = (mark: EventMark) => {
    onEvent(mark);
    const selected = mark.entity === "" ? "" : `, ${mark.entity} selected`;
    setSaid(`${chipText(mark, startNs)}: moved to ${runClock(mark.atNs, startNs)}${selected}`);
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

export function Lane({ marks, fraction, startNs, reading, positionNs, said, pick, origins }: LaneProps) {
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
              origin={origins.get(mark.entity) ?? null}
            />
          );
        })}
      </ol>
      <LaneNotes said={said} hidden={chips.filter((chip) => chip.row < 0).length} />
    </>
  );
}
