import { Cable, CircleCheck, CircleHelp, CircleX, Eye, Loader, Lock, TriangleAlert } from "lucide-react";
import { useState } from "react";
import type { ReviewState } from "../../../app/review";
import type { Badge, Report } from "../../../model/review";
import { badgeLabels, badgeOf, bindingLine, connectionWords, modeWords } from "../../../model/review";
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
    <span className={styles[badge]} title={title} data-part="status-chip">
      <Icon size={13} aria-hidden="true" />
      {label}
    </span>
  );
}

function RunningChip() {
  return (
    <span className={styles.chip} title="The run is still going." data-part="status-chip">
      <Loader size={13} aria-hidden="true" />
      Running
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

const sessionWords =
  "Connected session: the plant ran linked to a plant wire host, and through a gateway to real boards, until it was stopped.";

function Connected({ state }: { readonly state: ReviewState }) {
  if (state.connection === "none") return null;
  const binding = state.binding?.binding ?? null;
  return (
    <>
      <span className={styles.chip} title={sessionWords} data-part="connection">
        <Cable size={13} aria-hidden="true" />
        {connectionWords(state.connection)}
      </span>
      {binding !== null && (
        <span className={binding.matched ? styles.chip : styles.incomplete} title={bindingLine(binding)}>
          {binding.matched ? "Gateway matched" : "Gateway refused · real boards not bound"}
        </span>
      )}
    </>
  );
}

function Mode({ report, state }: { readonly report: Report; readonly state: ReviewState }) {
  if (state.connection !== "none") {
    return (
      <>
        <Connected state={state} />
        <span className={styles.facts}>{reportFacts(report)}</span>
        <RunInfo lines={[sessionWords, runIdentity(report)]} />
      </>
    );
  }
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
  if (state.connection === "running")
    return (
      <>
        <Connected state={state} />
        <Announce text="Running" />
      </>
    );
  const elapsed = Math.max(0, Math.round((current - state.openedAt) / 1000));
  if (state.report.phase !== "ready")
    return <ReportPending loading={state.report.phase === "loading"} elapsed={elapsed} />;
  const report = state.report.value;
  const badge = badgeOf(report, state.verified);
  if (state.running && badge === "incomplete") {
    return (
      <>
        <Mode report={report} state={state} />
        <RunningChip />
        <Announce text="Running" />
      </>
    );
  }
  const shown = badgeLabels[badge];
  const label = badge === "verifying" ? `Verifying · replaying the run · ${String(elapsed)} s` : shown;
  return (
    <>
      <Mode report={report} state={state} />
      <Chip badge={badge} label={label} title={integrityLine(report, state.verified)} />
      {badge === "incomplete" && <Reason text={report.incompleteReason || "the run did not close."} />}
      <Announce text={badge === "verifying" ? "Verifying: replaying the run" : shown} />
    </>
  );
}
