import { useMemo } from "react";
import type { Report } from "../../model/report";
import { wallClock } from "../../model/report";
import styles from "./RunScreen.module.css";

interface ScrubberProps {
  readonly startNs: number;
  readonly endNs: number;
  readonly positionNs: number;
  readonly report: Report | null;
  readonly density: readonly number[];
  readonly onSeek: (ns: number) => void;
}

interface Mark {
  readonly atNs: number;
  readonly label: string;
  readonly tone: "fault" | "other";
}

function marks(report: Report | null): readonly Mark[] {
  if (report === null) return [];
  const disturbances = report.disturbances.map((d): Mark => ({
    atNs: d.atNs,
    label: `${d.target} ${d.detail}`,
    tone: d.what.startsWith("fault") ? "fault" : "other",
  }));
  const calls = report.effectorCalls.map((c): Mark => ({
    atNs: c.atNs,
    label: `${c.entity} ${c.effector}`,
    tone: "other",
  }));
  return [...disturbances, ...calls].sort((a, b) => a.atNs - b.atNs);
}

function Marks({
  list,
  percent,
  onSeek,
}: {
  readonly list: readonly Mark[];
  readonly percent: (ns: number) => string;
  readonly onSeek: (ns: number) => void;
}) {
  return (
    <>
      {list.map((mark) => (
        <button
          key={`${String(mark.atNs)}${mark.label}`}
          type="button"
          className={mark.tone === "fault" ? styles.markFault : styles.mark}
          style={{ left: percent(mark.atNs) }}
          title={`${wallClock(mark.atNs)} · ${mark.label}`}
          aria-label={`Jump to ${wallClock(mark.atNs)}, ${mark.label}`}
          onClick={() => {
            onSeek(mark.atNs);
          }}
        />
      ))}
    </>
  );
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

export function Scrubber({ startNs, endNs, positionNs, report, density, onSeek }: ScrubberProps) {
  const span = Math.max(1, endNs - startNs);
  const percent = (ns: number) => `${String(Math.min(100, Math.max(0, ((ns - startNs) / span) * 100)))}%`;
  const list = useMemo(() => marks(report), [report]);
  return (
    <div className={styles.scrubber}>
      <div className={styles.time} aria-live="polite">
        {wallClock(positionNs)}
        <span className={styles.muted}>{` of ${wallClock(startNs)} – ${wallClock(endNs)} UTC`}</span>
      </div>
      <div className={styles.track}>
        <Density density={density} />
        <input
          type="range"
          className={styles.range}
          min={0}
          max={Math.round(span / 1e9)}
          value={Math.round((positionNs - startNs) / 1e9)}
          onChange={(event) => {
            onSeek(startNs + Number(event.target.value) * 1e9);
          }}
          aria-label="Playback time"
          aria-valuetext={wallClock(positionNs)}
        />
        <Marks list={list} percent={percent} onSeek={onSeek} />
      </div>
    </div>
  );
}
