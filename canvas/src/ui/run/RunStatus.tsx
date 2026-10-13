import { CircleCheck, CircleHelp, CircleX, Eye, Loader, Lock, TriangleAlert } from "lucide-react";
import { useEffect, useState } from "react";
import type { ReviewState } from "../../app/reviewStore";
import type { Badge, Report } from "../../model/report";
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

interface Warning {
  readonly text: string;
  readonly details: string;
}

function verifyWarning(state: ReviewState): Warning | null {
  if (state.verified.phase !== "failed") return null;
  const message = state.verified.failure.message;
  const changed = message.includes("domain digest mismatch");
  const text = changed
    ? "Not checked: the plant file changed after this run was recorded. Restore the domain the run used to verify it."
    : "Not checked: streamsim verify could not replay this run.";
  return { text, details: message };
}

function countWarning(state: ReviewState, report: Report): Warning | null {
  const delivered = report.counts.delivered ?? 0;
  const records = state.playback.records.toLocaleString("en-US");
  if (state.playback.phase === "failed")
    return {
      text: `Playback is incomplete: ${records} of ${delivered.toLocaleString("en-US")} delivered observations arrived.`,
      details: "",
    };
  if (state.playback.phase === "ready" && report.ended && state.playback.records !== delivered) {
    return {
      text: `Playback has ${records} observations; the report counts ${delivered.toLocaleString("en-US")} delivered.`,
      details: "",
    };
  }
  return null;
}

function warnings(state: ReviewState): readonly Warning[] {
  if (state.report.phase !== "ready") return [];
  const report = state.report.value;
  const differ =
    report.verification === "files_differ"
      ? {
          text: "The run's files differ from what the run wrote: playback shows the files as they are now.",
          details: "",
        }
      : null;
  return [verifyWarning(state), differ, countWarning(state, report)].filter(
    (warning): warning is Warning => warning !== null,
  );
}

export function Banners({ state }: { readonly state: ReviewState }) {
  const list = warnings(state);
  if (list.length === 0) return null;
  return (
    <div className={styles.banners} role="alert">
      {list.map((warning) => (
        <div key={warning.text} className={styles.banner}>
          <TriangleAlert size={16} aria-hidden="true" />
          <div>
            {warning.text}
            {warning.details !== "" && (
              <details className={styles.details}>
                <summary>Details</summary>
                <pre>{warning.details}</pre>
              </details>
            )}
          </div>
        </div>
      ))}
    </div>
  );
}
