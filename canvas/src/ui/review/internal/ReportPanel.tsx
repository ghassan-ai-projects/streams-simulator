import type { Connection, Report, VerifyState } from "../../../model/review";
import { alarmRow, badgeOf, replayLine } from "../../../model/review";
import { fileLine, integrityLine, kpisOf, unfinishedTruthWords } from "../../../model/review";
import { grouped } from "../../../model/shared/units";
import { spaced } from "../../../model/shared/words";
import { KpiTable } from "./KpiTable";
import { Rows, Section } from "./ReportParts";
import { Events } from "./ReportEvents";
import type { Picked } from "./ReportEvents";
import parts from "./ReportParts.module.css";
import styles from "./ReportPanel.module.css";

const count = (report: Report, key: string) => (report.ended ? grouped(report.counts[key] ?? 0) : "n/a");

function ReplayCheck({ verify }: { readonly verify: VerifyState }) {
  const detail = verify.phase === "ready" ? verify.value.detail : "";
  return (
    <Section title="Replay check">
      <p className={styles.verdict}>{replayLine(verify)}</p>
      <details className={styles.inset}>
        <summary>For developers</summary>
        <p className={parts.muted}>Checked with the streamsim verify command over the run folder.</p>
        {detail !== "" && <pre>{detail}</pre>}
      </details>
    </Section>
  );
}

function Deliveries({ report }: { readonly report: Report }) {
  const dropped = Object.entries(report.dropped).map(([reason, n]) => [spaced(reason), grouped(n)]);
  return (
    <Section title="Deliveries" note={report.ended ? undefined : "run not closed"}>
      <Rows
        rows={[...["emitted", "delivered", "dropped"].map((key) => [key, count(report, key)]), ...dropped]}
        columns={["text", "number"]}
        headers={["Records", "Count"]}
      />
    </Section>
  );
}

function FileCheck({ report }: { readonly report: Report }) {
  return (
    <Section title="File check">
      <p className={parts.muted}>{fileLine(report)}</p>
      <details className={parts.muted}>
        <summary>For developers</summary>
        <p className={parts.muted}>
          The report compares the files in the run folder with the digests the run recorded.
        </p>
        <Rows
          rows={report.checks.map((check) => [{ text: spaced(check.name), title: check.name }, spaced(check.status)])}
          columns={["text", "text"]}
          headers={["File", "Check"]}
        />
      </details>
    </Section>
  );
}

function Integrity({ report, verify }: { readonly report: Report; readonly verify: VerifyState }) {
  return (
    <details className={styles.integrity}>
      <summary>{integrityLine(report, verify)}</summary>
      <ReplayCheck verify={verify} />
      <FileCheck report={report} />
      <Deliveries report={report} />
    </details>
  );
}

interface TruthProps {
  readonly report: Report;
  readonly picked?: Picked | null;
  readonly unit: (entity: string, state: string) => string;
  readonly entity: string | null;
  readonly only?: "kpi";
  readonly connection?: Connection | undefined;
}

const connectedTruth =
  "Connected session: the true results below are the simulator's own record of the plant while it was linked to the " +
  "plant wire host.";

function truthWords(report: Report, connection: TruthProps["connection"]): string {
  if (connection !== undefined && connection !== "none") return connectedTruth;
  return report.mode === "explore"
    ? "Explore run: the true results below were visible while it ran."
    : "Evaluate run: the true results below were hidden from the outside system while it ran and shown when it ended.";
}

function Truth({ report, unit, entity, picked = null, connection }: TruthProps) {
  return (
    <>
      <p className={styles.truthNote}>{truthWords(report, connection)}</p>
      <Events report={report} picked={picked} />
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

function Scoring({ report }: { readonly report: Report }) {
  if (report.scoring !== null) {
    const { status, reason } = report.scoring;
    return (
      <Section title="Scoring">
        <p className={parts.muted}>{`${status}${reason ? ` · ${reason}` : ""}`}</p>
      </Section>
    );
  }
  return (
    <Section title="Scoring">
      <p className={parts.muted}>Not scored: no outside system is connected to this run.</p>
      <details className={parts.muted}>
        <summary>For developers</summary>
        <p>
          Nothing was scored: no consumer verdict was recorded for this run. A consumer attaches through the operator
          endpoint; the canvas does not attach one.
        </p>
      </details>
    </Section>
  );
}

const truthLead =
  "What really happened in the plant. An outside system could not see it while the run was live; it is shown here once the run has ended.";

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

export function ReportPanel({
  report,
  unit,
  entity,
  only,
  picked = null,
  verify = loading,
  connection,
}: TruthProps & { readonly verify?: VerifyState }) {
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
      <Integrity report={report} verify={verify} />
      <TruthCaution report={report} verify={verify} />
      {report.truthUnavailable || !report.ended ? (
        <Section title="True results">
          <p className={parts.muted}>
            {report.truthUnavailable ? `Not available: ${report.truthUnavailable}` : notClosed}
          </p>
        </Section>
      ) : (
        <Truth report={report} unit={unit} entity={entity} picked={picked} connection={connection} />
      )}
      <Scoring report={report} />
    </>
  );
}
