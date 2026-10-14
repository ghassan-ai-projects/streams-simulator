import { CircleCheck, CircleHelp, CircleX, Eye, Info, Loader, Lock, TriangleAlert } from "lucide-react";
import { useEffect, useState } from "react";
import type { ReviewState } from "../../app/reviewStore";
import type { Badge, Report } from "../../model/report";
import { badgeLabels, badgeOf } from "../../model/report";
import { runFacts, runIdentity } from "../../model/reportViews";
import { grouped } from "../../model/units";
import { RunInfo } from "./RunInfo";
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

function Announce({ text }: { readonly text: string }) {
  return (
    <span className="sr-only" role="status">
      {text}
    </span>
  );
}

function ReportPending({ loading, elapsed }: { readonly loading: boolean; readonly elapsed: number }) {
  return (
    <>
      <span className={styles.chip}>
        {loading ? `Report · streamsim report reading the run · ${String(elapsed)} s` : "No report"}
      </span>
      <Announce text={loading ? "streamsim report is reading the run" : "No report"} />
    </>
  );
}

const modeWords = {
  explore: "Explore: the truth is visible; not for scoring.",
  evaluate:
    "Evaluate: the truth stays sealed while it runs, so a consumer can be scored; it is revealed when the run ends.",
} as const;

function Mode({ report }: { readonly report: Report }) {
  const explore = report.mode === "explore";
  return (
    <>
      <span className={explore ? styles.truth : styles.chip} title={explore ? modeWords.explore : modeWords.evaluate}>
        {explore ? <Eye size={13} aria-hidden="true" /> : <Lock size={13} aria-hidden="true" />}
        {explore ? "Explore" : "Evaluate"}
      </span>
      {!report.ended && <span className={styles.chip}>not finished</span>}
      <span className={styles.facts}>{runFacts(report)}</span>
      <RunInfo lines={[explore ? modeWords.explore : modeWords.evaluate, runIdentity(report)]} />
    </>
  );
}

export function Status({ state }: { readonly state: ReviewState }) {
  const now = useNow(state.report.phase === "loading" || state.verified.phase === "loading");
  const elapsed = Math.max(0, Math.round((now - state.openedAt) / 1000));
  if (state.report.phase !== "ready")
    return <ReportPending loading={state.report.phase === "loading"} elapsed={elapsed} />;
  const report = state.report.value;
  const badge = badgeOf(report, state.verified);
  const label = badge === "verifying" ? `Verifying · replaying the run · ${String(elapsed)} s` : badgeLabels[badge];
  return (
    <>
      <Mode report={report} />
      <Chip badge={badge} label={label} />
      <Announce text={badge === "verifying" ? "Verifying: replaying the run" : badgeLabels[badge]} />
    </>
  );
}

interface Warning {
  readonly text: string;
  readonly details: string;
}

function verifyWarning(state: ReviewState): Warning | null {
  if (state.verified.phase !== "failed") return null;
  return { text: "Not checked: streamsim verify could not replay this run.", details: state.verified.failure.message };
}

function countWarning(state: ReviewState, report: Report): Warning | null {
  const delivered = report.counts.delivered ?? 0;
  const records = grouped(state.playback.records);
  if (state.playback.phase === "failed")
    return {
      text: `Playback is incomplete: ${records} of ${grouped(delivered)} delivered observations arrived.`,
      details: "",
    };
  if (state.playback.phase === "ready" && report.ended && state.playback.records !== delivered) {
    return {
      text: `Playback has ${records} observations; the report counts ${grouped(delivered)} delivered.`,
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

function DomainNote({ state }: { readonly state: ReviewState }) {
  if (state.verified.phase !== "ready" || !state.verified.value.domainChanged) return null;
  return (
    <p className={styles.domainNote} role="status">
      <Info size={14} aria-hidden="true" />
      The plant file changed after this run. The review and the replay use the plant the run recorded.
    </p>
  );
}

export function Banners({ state }: { readonly state: ReviewState }) {
  return (
    <div>
      <DomainNote state={state} />
      <Alerts list={warnings(state)} />
    </div>
  );
}

function Alerts({ list }: { readonly list: readonly Warning[] }) {
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
