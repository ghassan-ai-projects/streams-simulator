import { CircleCheck, CircleHelp, CircleX, Eye, Loader, Lock, TriangleAlert } from "lucide-react";
import { useEffect, useState } from "react";
import type { ReviewState } from "../../app/reviewStore";
import type { Badge } from "../../model/report";
import { badgeLabels, badgeOf } from "../../model/report";
import styles from "./RunScreen.module.css";

const badgeIcons = {
  verifying: Loader,
  "not-checked": CircleHelp,
  verified: CircleCheck,
  "not-verified": CircleX,
  incomplete: TriangleAlert,
} as const;

function useNow(running: boolean): number {
  const [now, setNow] = useState(() => Date.now());
  useEffect(() => {
    if (!running) return undefined;
    const timer = setInterval(() => {
      setNow(Date.now());
    }, 1000);
    return () => {
      clearInterval(timer);
    };
  }, [running]);
  return now;
}

function Chip({ badge, label }: { readonly badge: Badge; readonly label: string }) {
  const Icon = badgeIcons[badge];
  return (
    <span className={styles[badge]}>
      <Icon size={13} aria-hidden="true" />
      {label}
    </span>
  );
}

export function Status({ state }: { readonly state: ReviewState }) {
  const now = useNow(state.report.phase === "loading" || state.verified.phase === "loading");
  const elapsed = Math.max(0, Math.round((now - state.openedAt) / 1000));
  if (state.report.phase !== "ready") {
    const label =
      state.report.phase === "loading"
        ? `Report · streamsim report reading the run · ${String(elapsed)} s`
        : "No report";
    return (
      <span className={styles.chip} role="status">
        {label}
      </span>
    );
  }
  const report = state.report.value;
  const badge = badgeOf(report, state.verified);
  const label = badge === "verifying" ? `Verifying · replaying the run · ${String(elapsed)} s` : badgeLabels[badge];
  const records = `${state.playback.records.toLocaleString("en-US")} observations${state.playback.phase === "loading" ? " · loading" : ""}`;
  return (
    <>
      <Chip badge={badge} label={label} />
      <span className={report.mode === "explore" ? styles.truth : styles.chip}>
        {report.mode === "explore" ? <Eye size={13} aria-hidden="true" /> : <Lock size={13} aria-hidden="true" />}
        {`${report.mode === "explore" ? "Explore · unblinded" : "Evaluate"} · ${report.ended ? "ended" : "not closed"}`}
      </span>
      <span className={styles.progress} role="status">
        {records}
      </span>
    </>
  );
}

function warnings(state: ReviewState): readonly string[] {
  if (state.report.phase !== "ready") return [];
  const report = state.report.value;
  const delivered = report.counts.delivered ?? 0;
  const list: string[] = [];
  if (report.verification === "files_differ")
    list.push("The run's files differ from what the run wrote: playback shows the files as they are now.");
  if (state.playback.phase === "failed")
    list.push(
      `Playback is incomplete: ${state.playback.records.toLocaleString("en-US")} of ${delivered.toLocaleString("en-US")} delivered observations arrived.`,
    );
  if (state.playback.phase === "ready" && report.ended && state.playback.records !== delivered) {
    list.push(
      `Playback has ${state.playback.records.toLocaleString("en-US")} observations; the report counts ${delivered.toLocaleString("en-US")} delivered.`,
    );
  }
  return list;
}

export function Banners({ state }: { readonly state: ReviewState }) {
  const list = warnings(state);
  if (list.length === 0) return null;
  return (
    <div className={styles.banners} role="alert">
      {list.map((text) => (
        <p key={text} className={styles.banner}>
          <TriangleAlert size={14} aria-hidden="true" />
          {text}
        </p>
      ))}
    </div>
  );
}
