import type { ReviewSession, ReviewState, TwinState } from "../../../app/review";
import { twinFor } from "../../../app/review";
import type { TwinCoverage, TwinReport } from "../../../model/review";
import { coverageOf, isRealBoard, twinOffer } from "../../../model/review";
import { shared, useStore } from "../../primitives";
import { Section } from "./ReportParts";
import parts from "./ReportParts.module.css";
import styles from "./Twin.module.css";
import { TwinChartView } from "./TwinChartView";

interface InspectProps {
  readonly review: ReviewSession;
  readonly state: ReviewState;
  readonly entity: string;
  readonly startNs: number;
}

interface DrawnProps {
  readonly report: TwinReport;
  readonly device: string;
  readonly cursorNs: number;
  readonly startNs: number;
}

interface CoverageProps {
  readonly coverage: TwinCoverage;
  readonly cursorNs: number;
  readonly startNs: number;
}

function Coverage({ coverage, cursorNs, startNs }: CoverageProps) {
  if (coverage.kind === "series") return <TwinChartView series={coverage} startNs={startNs} cursorNs={cursorNs} />;
  return (
    <div className={styles.figure}>
      <h3 className={styles.heading}>{coverage.metric.metric}</h3>
      <p className={styles.meta}>{coverage.why}</p>
    </div>
  );
}

function Drawn({ report, device, cursorNs, startNs }: DrawnProps) {
  const covered = coverageOf(report, device);
  if (covered.length === 0) {
    return (
      <p
        className={parts.muted}
      >{`The twin report does not cover ${device}: the field trace holds none of its metrics.`}</p>
    );
  }
  return (
    <>
      {covered.map((coverage) => (
        <Coverage key={coverage.metric.metric} coverage={coverage} cursorNs={cursorNs} startNs={startNs} />
      ))}
    </>
  );
}

function notDrawn(shown: TwinState): string {
  if (shown.phase === "loading") return "Comparing the board with the twin…";
  if (shown.phase === "failed") return "The comparison could not be made. The Report tab says why.";
  return "Compare the run with the twin to draw its prediction over the board's readings.";
}

function Waiting({ review, shown }: { readonly review: ReviewSession; readonly shown: TwinState }) {
  return (
    <>
      <p className={parts.muted} role="status">
        {notDrawn(shown)}
      </p>
      <p className={parts.muted}>
        <button
          type="button"
          className={shared.secondary}
          onClick={() => {
            review.showTab("report");
          }}
        >
          Open Twin vs board
        </button>
      </p>
    </>
  );
}

export function TwinInspect({ review, state, entity, startNs }: InspectProps) {
  const stored = useStore(review.twin);
  const shown = twinFor(stored, state.worldId);
  const offer = twinOffer(state.connection, state.sessionDomain);
  if (offer.kind !== "offered" || !isRealBoard(state.origins, entity)) return null;
  const device = state.origins.get(entity)?.device ?? entity;
  return (
    <Section title={`Twin prediction · ${entity}`}>
      {shown.phase === "ready" ? (
        <Drawn report={shown.report} device={device} cursorNs={state.positionNs} startNs={startNs} />
      ) : (
        <Waiting review={review} shown={shown} />
      )}
    </Section>
  );
}
