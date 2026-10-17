import { CircleCheck, CircleHelp, CircleX, Eye, Info, Loader, Lock, TriangleAlert } from "lucide-react";
import { useState } from "react";
import type { ReviewState } from "../../../app/review";
import { readyReport } from "../../../app/review";
import type { Badge, BannerWarning, Report } from "../../../model/review";
import { badgeLabels, badgeOf, bannerWarnings, modeWords } from "../../../model/review";
import { integrityLine, reportFacts, runIdentity } from "../../../model/review";
import { useInterval } from "../../primitives";
import { RunInfo } from "./RunInfo";
import styles from "./RunStatus.module.css";

const badgeIcons = {
  verifying: Loader,
  "not-checked": CircleHelp,
  verified: CircleCheck,
  "not-verified": CircleX,
  incomplete: TriangleAlert,
} as const;

function useNow(now: () => number, running: boolean): number {
  const [current, setCurrent] = useState(now);
  useInterval(
    () => {
      setCurrent(now());
    },
    running ? 1000 : null,
  );
  return current;
}

function Chip({ badge, label, title }: { readonly badge: Badge; readonly label: string; readonly title: string }) {
  const Icon = badgeIcons[badge];
  return (
    <span className={styles[badge]} title={title}>
      <Icon size={13} aria-hidden="true" />
      {label}
    </span>
  );
}

function Reason({ text }: { readonly text: string }) {
  return (
    <span className={styles.reason} title={text}>
      {text}
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
      <span className={styles.chip}>{loading ? `Reading the run · ${String(elapsed)} s` : "No report"}</span>
      <Announce text={loading ? "Reading the run" : "No report"} />
    </>
  );
}

function Mode({ report }: { readonly report: Report }) {
  const explore = report.mode === "explore";
  return (
    <>
      <span className={explore ? styles.truth : styles.chip} title={explore ? modeWords.explore : modeWords.evaluate}>
        {explore ? <Eye size={13} aria-hidden="true" /> : <Lock size={13} aria-hidden="true" />}
        {explore ? "Explore" : "Evaluate"}
      </span>
      <span className={styles.facts}>{reportFacts(report)}</span>
      <RunInfo lines={[explore ? modeWords.explore : modeWords.evaluate, runIdentity(report)]} />
    </>
  );
}

export function Status({ state, now }: { readonly state: ReviewState; readonly now: () => number }) {
  const current = useNow(now, state.report.phase === "loading" || state.verified.phase === "loading");
  const elapsed = Math.max(0, Math.round((current - state.openedAt) / 1000));
  if (state.report.phase !== "ready")
    return <ReportPending loading={state.report.phase === "loading"} elapsed={elapsed} />;
  const report = state.report.value;
  const badge = badgeOf(report, state.verified);
  const label = badge === "verifying" ? `Verifying · replaying the run · ${String(elapsed)} s` : badgeLabels[badge];
  return (
    <>
      <Mode report={report} />
      <Chip badge={badge} label={label} title={integrityLine(report, state.verified)} />
      {badge === "incomplete" && <Reason text={report.incompleteReason || "the run did not close."} />}
      <Announce text={badge === "verifying" ? "Verifying: replaying the run" : badgeLabels[badge]} />
    </>
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
      <Alerts list={bannerWarnings(readyReport(state), state.verified, state.playback)} />
    </div>
  );
}

function Alerts({ list }: { readonly list: readonly BannerWarning[] }) {
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
