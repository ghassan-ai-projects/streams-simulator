import type { HandoverState } from "../../../app/review";
import type { Connection, Report, VerifyState } from "../../../model/review";
import { alarmRow, badgeOf, kpisOf, unfinishedTruthWords } from "../../../model/review";
import { grouped } from "../../../model/shared/units";
import { KpiTable } from "./KpiTable";
import { noReplayControl } from "./OpenQuestions";
import type { ReplayControl } from "./OpenQuestions";
import { Rows, Section } from "./ReportParts";
import { Events } from "./ReportEvents";
import type { EventActions, Picked } from "./ReportEvents";
import { ThisRun } from "./ThisRun";
import parts from "./ReportParts.module.css";
import styles from "./ReportPanel.module.css";

interface TruthProps {
  readonly report: Report;
  readonly picked?: Picked | null;
  readonly unit: (entity: string, state: string) => string;
  readonly entity: string | null;
  readonly only?: "kpi";
  readonly connection?: Connection | undefined;
  readonly actions?: EventActions | undefined;
}

function Truth({ report, unit, entity, picked = null, actions }: TruthProps) {
  return (
    <>
      <Events report={report} picked={picked} {...(actions === undefined ? {} : { actions })} />
      <KpiTruth report={report} unit={unit} entity={entity} />
      <Section title="Alarms">
        <p className={parts.muted}>How the alarm system behaved during the run.</p>
        <Rows
          rows={Object.entries(report.alarms).map(([key, n]) => alarmRow(key, grouped(n)))}
          columns={["text", "number", "text"]}
          headers={["Measure", "Count", "What it means"]}
        />
        <details className={parts.muted}>
          <summary>For developers</summary>
          <p>The alarm measures follow ISA-18.2: floods, chattering, stale and standing alarms.</p>
        </details>
      </Section>
    </>
  );
}

const truthLead = "What really happened in the plant.";

function KpiTruth({ report, unit, entity, lead = true }: TruthProps & { readonly lead?: boolean }) {
  const kpis = kpisOf(report, entity);
  return (
    <Section title="True results" note={entity ?? undefined}>
      {lead && <p className={parts.muted}>{truthLead}</p>}
      {!report.complete && <p className={parts.muted}>{unfinishedTruthWords}</p>}
      {kpis.length === 0 ? (
        <p className={parts.muted}>{entity === null ? "None" : `No true results for ${entity}.`}</p>
      ) : (
        <KpiTable rows={kpis} grouped={entity === null} unit={unit} />
      )}
    </Section>
  );
}

const notClosed = "Not available yet: the run has not ended, so its true results are still hidden.";

const loading: VerifyState = { phase: "loading" };
const unverifiedTruth = "This run did not verify, so the truth below may not describe the files in the run folder.";

function TruthCaution({ report, verify }: { readonly report: Report; readonly verify: VerifyState }) {
  const rejected = verify.phase === "ready" && !verify.value.matches;
  if (!rejected && badgeOf(report, verify) !== "not-verified") return null;
  return (
    <p className={styles.caution} role="note">
      {unverifiedTruth}
    </p>
  );
}

const idleHandover: HandoverState = { phase: "idle" };

interface PanelProps extends TruthProps {
  readonly verify?: VerifyState;
  readonly replay?: ReplayControl;
  readonly handover?: HandoverState;
}

function Truthful({ report, unit, entity, picked, actions }: Omit<TruthProps, "only">) {
  if (report.truthUnavailable || !report.ended)
    return (
      <Section title="True results">
        <p className={parts.muted}>
          {report.truthUnavailable ? `Not available: ${report.truthUnavailable}` : notClosed}
        </p>
      </Section>
    );
  return <Truth report={report} unit={unit} entity={entity} picked={picked ?? null} actions={actions} />;
}

export function ReportPanel({
  report,
  unit,
  entity,
  only,
  picked = null,
  verify = loading,
  connection = "none",
  actions,
  replay = noReplayControl,
  handover = idleHandover,
}: PanelProps) {
  if (only === "kpi") {
    return report.truthUnavailable || !report.ended ? null : (
      <>
        <TruthCaution report={report} verify={verify} />
        <KpiTruth report={report} unit={unit} entity={entity} lead={false} />
      </>
    );
  }
  return (
    <>
      <ThisRun report={report} verify={verify} connection={connection} replay={replay} handover={handover} />
      <TruthCaution report={report} verify={verify} />
      <Truthful report={report} unit={unit} entity={entity} picked={picked} actions={actions} />
    </>
  );
}
