import type { ReviewSession, ReviewState, TwinState } from "../../../app/review";
import { twinFor } from "../../../app/review";
import { twinOffer } from "../../../model/review";
import { failureLine, useFocusOnMount, useStore } from "../../primitives";
import { Section } from "./ReportParts";
import parts from "./ReportParts.module.css";
import styles from "./Twin.module.css";
import { TwinForm } from "./TwinForm";
import { CommandsSection, DivergencesSection, MetricsSection } from "./TwinParts";
import { TwinResult } from "./TwinResult";
import { useTwinSeen } from "./useTwinSeen";

interface TwinPanelProps {
  readonly review: ReviewSession;
  readonly state: ReviewState;
  readonly startNs: number;
}

const intro =
  "Runs this plant's domain as the twin, with the commands the board accepted, and compares it with what the " +
  "board reported. It measures the twin; the board's readings are evidence, not truth.";

function Refused({
  shown,
  review,
}: {
  readonly shown: Extract<TwinState, { phase: "failed" }>;
  readonly review: ReviewSession;
}) {
  const alert = useFocusOnMount<HTMLParagraphElement>({ when: shown.fresh });
  useTwinSeen(review, shown);
  return (
    <p ref={alert} tabIndex={-1} role="alert" className={styles.failure}>
      {`The comparison could not be made: ${failureLine(shown.failure.diagnostic)}`}
    </p>
  );
}

function Outcome({ shown, review }: { readonly shown: TwinState; readonly review: ReviewSession }) {
  if (shown.phase === "failed") return <Refused shown={shown} review={review} />;
  if (shown.phase === "ready") return <TwinResult shown={shown} review={review} />;
  return null;
}

export function TwinPanel({ review, state, startNs }: TwinPanelProps) {
  const stored = useStore(review.twin);
  const offer = twinOffer(state.connection, state.sessionDomain);
  const shown = twinFor(stored, state.worldId);
  if (offer.kind === "absent") return null;
  if (offer.kind === "withheld") {
    return (
      <Section title="Twin vs board">
        <p className={parts.muted}>{offer.why}</p>
      </Section>
    );
  }
  return (
    <>
      <Section title="Twin vs board">
        <p className={parts.muted}>{intro}</p>
        <TwinForm review={review} run={state.worldId ?? ""} domain={offer.domain} busy={shown.phase === "loading"} />
        <p className={styles.status} role="status">
          {shown.phase === "loading" ? "Comparing the board with the twin…" : ""}
        </p>
        <Outcome shown={shown} review={review} />
      </Section>
      {shown.phase === "ready" && <TwinDetails review={review} shown={shown} startNs={startNs} />}
    </>
  );
}

function TwinDetails({
  review,
  shown,
  startNs,
}: {
  readonly review: ReviewSession;
  readonly shown: Extract<TwinState, { phase: "ready" }>;
  readonly startNs: number;
}) {
  return (
    <>
      <CommandsSection report={shown.report} startNs={startNs} />
      <DivergencesSection
        report={shown.report}
        startNs={startNs}
        onSeek={(ns) => {
          review.seek(ns);
        }}
      />
      <MetricsSection report={shown.report} />
    </>
  );
}
