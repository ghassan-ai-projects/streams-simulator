import type { ReactNode } from "react";
import type { Report, VerifyState } from "../../model/report";
import { badgeOf, humanise, replayLine } from "../../model/report";
import { integrityLine, kpisOf } from "../../model/reportViews";
import { runClock } from "../../model/timeline";
import { grouped } from "../../model/units";
import shared from "../shared.module.css";
import { alarmRow } from "./alarmWords";
import { KpiTable } from "./KpiTable";
import styles from "./RunScreen.module.css";

type Cell = string | { readonly text: string; readonly title: string };

function Section({
  title,
  note,
  children,
}: {
  readonly title: string;
  readonly note?: string | undefined;
  readonly children: ReactNode;
}) {
  return (
    <section className={styles.section} aria-label={title}>
      <div className={shared.sectionTitle}>
        <h2 className={styles.sectionName}>{title}</h2>
        {note !== undefined && <span className={styles.note}>{note}</span>}
      </div>
      {children}
    </section>
  );
}

interface RowsProps {
  readonly rows: readonly (readonly Cell[])[];
  readonly columns: readonly string[];
  readonly headers: readonly string[];
  readonly empty?: string;
}

function Row({ row, columns }: { readonly row: readonly Cell[]; readonly columns: readonly string[] }) {
  return (
    <tr>
      {row.map((cell, c) => {
        const Tag = c === 0 ? "th" : "td";
        return (
          <Tag
            key={c}
            scope={c === 0 ? "row" : undefined}
            className={styles[columns[c] ?? "text"]}
            title={typeof cell === "string" ? undefined : cell.title}
          >
            {typeof cell === "string" ? cell : cell.text}
          </Tag>
        );
      })}
    </tr>
  );
}

function Rows({ rows, columns, headers, empty = "None" }: RowsProps) {
  if (rows.length === 0) return <p className={styles.muted}>{empty}</p>;
  return (
    <table className={styles.table}>
      <thead className={styles.head}>
        <tr>
          {headers.map((header) => (
            <th key={header} scope="col">
              {header}
            </th>
          ))}
        </tr>
      </thead>
      <tbody>
        {rows.map((row, r) => (
          <Row key={r} row={row} columns={columns} />
        ))}
      </tbody>
    </table>
  );
}

const named = (id: string): Cell => ({ text: humanise(id), title: id });
const words = (text: string): Cell => ({ text, title: text });
const count = (report: Report, key: string) => (report.ended ? grouped(report.counts[key] ?? 0) : "n/a");

function ReplayCheck({ verify }: { readonly verify: VerifyState }) {
  const detail = verify.phase === "ready" ? verify.value.detail : "";
  return (
    <Section title="Replay check" note="streamsim verify">
      <p className={styles.verdict}>{replayLine(verify)}</p>
      {detail !== "" && (
        <details className={styles.inset}>
          <summary>{verify.phase === "ready" && verify.value.matches ? "Details" : "What differs"}</summary>
          <pre>{detail}</pre>
        </details>
      )}
    </Section>
  );
}

function Deliveries({ report }: { readonly report: Report }) {
  const dropped = Object.entries(report.dropped).map(([reason, n]) => [reason.replaceAll("_", " "), grouped(n)]);
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

function Integrity({ report, verify }: { readonly report: Report; readonly verify: VerifyState }) {
  return (
    <details className={styles.integrity}>
      <summary>{integrityLine(report, verify)}</summary>
      <ReplayCheck verify={verify} />
      <Section title="File check" note={report.verification.replaceAll("_", " ") || "not recorded"}>
        <p className={styles.muted}>
          The report compares the files in the run folder with the digests the run recorded.
        </p>
        <Rows
          rows={report.checks.map((check) => [
            { text: check.name.replaceAll("_", " "), title: check.name },
            check.status.replaceAll("_", " "),
          ])}
          columns={["text", "text"]}
          headers={["File", "Check"]}
        />
      </Section>
      <Deliveries report={report} />
    </details>
  );
}

interface TruthProps {
  readonly report: Report;
  readonly unit: (entity: string, state: string) => string;
  readonly entity: string | null;
  readonly only?: "kpi";
}

function TruthNote({ report }: { readonly report: Report }) {
  const text =
    report.mode === "explore"
      ? "Explore run: the true results below were visible while it ran."
      : "Evaluate run: the true results below were hidden from the outside system while it ran and shown when it ended.";
  return <p className={styles.truthNote}>{text}</p>;
}

function Events({ report }: { readonly report: Report }) {
  const at = (ns: number) => runClock(ns, report.startNs);
  if (report.disturbances.length === 0 && report.effectorCalls.length === 0) {
    return (
      <Section title="Disturbances and effector calls">
        <p className={styles.muted}>None in this run.</p>
      </Section>
    );
  }
  return (
    <>
      <Section title="Disturbances">
        <Rows
          rows={report.disturbances.map((d) => [
            { text: at(d.onsetNs), title: `set at ${at(d.injectedNs)}` },
            d.target,
            named(d.name),
          ])}
          columns={["clock", "id", "label"]}
          headers={["Starts", "Instance", "Disturbance"]}
        />
      </Section>
      <Section title="Effector calls">
        <Rows
          rows={report.effectorCalls.map((call) => [
            at(call.atNs),
            call.entity,
            named(call.effector),
            words(call.outcome),
          ])}
          columns={["clock", "id", "effector", "outcome"]}
          headers={["Time", "Instance", "Effector", "Outcome"]}
        />
      </Section>
    </>
  );
}

function Truth({ report, unit, entity }: TruthProps) {
  return (
    <>
      <TruthNote report={report} />
      <Events report={report} />
      <KpiTruth report={report} unit={unit} entity={entity} />
      <Section title="Alarms">
        <p className={styles.muted}>How the alarm system behaved during the run.</p>
        <Rows
          rows={Object.entries(report.alarms).map(([key, n]) => alarmRow(key, grouped(n)))}
          columns={["text", "number", "text"]}
          headers={["Measure", "Count", "What it means"]}
        />
        <details className={styles.muted}>
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
        <p className={styles.muted}>{`${status}${reason ? ` · ${reason}` : ""}`}</p>
      </Section>
    );
  }
  return (
    <Section title="Scoring">
      <p className={styles.muted}>Not scored: no outside system is connected to this run.</p>
      <details className={styles.muted}>
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
    <Section title="True results">
      {lead && <p className={styles.muted}>{truthLead}</p>}
      {kpis.length === 0 ? (
        <p className={styles.muted}>{entity === null ? "None" : `No true results for ${entity}.`}</p>
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
  verify = loading,
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
          <p className={styles.muted}>
            {report.truthUnavailable ? `Not available: ${report.truthUnavailable}` : notClosed}
          </p>
        </Section>
      ) : (
        <Truth report={report} unit={unit} entity={entity} />
      )}
      <Scoring report={report} />
    </>
  );
}
