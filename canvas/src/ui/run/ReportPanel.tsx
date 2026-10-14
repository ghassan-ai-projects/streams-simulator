import type { ReactNode } from "react";
import type { Report, VerifyState } from "../../model/report";
import { badgeOf, humanise, replayLine, wallClock } from "../../model/report";
import { kpisOf } from "../../model/reportViews";
import { displayUnit } from "../../model/units";
import shared from "../shared.module.css";
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
      <thead className="sr-only">
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

const fixed = (value: number) =>
  Math.abs(value) >= 100 ? value.toFixed(0) : Math.abs(value) >= 10 ? value.toFixed(1) : value.toFixed(2);
const named = (id: string): Cell => ({ text: humanise(id), title: id });
const words = (text: string): Cell => ({ text, title: text });
const count = (report: Report, key: string) =>
  report.ended ? (report.counts[key] ?? 0).toLocaleString("en-US") : "n/a";

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

function Facts({ report, verify }: { readonly report: Report; readonly verify: VerifyState }) {
  const dropped = Object.entries(report.dropped).map(([reason, n]) => [
    reason.replaceAll("_", " "),
    n.toLocaleString("en-US"),
  ]);
  return (
    <>
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
      <Section title="Deliveries" note={report.ended ? undefined : "run not closed"}>
        <Rows
          rows={[...["emitted", "delivered", "dropped"].map((key) => [key, count(report, key)]), ...dropped]}
          columns={["text", "number"]}
          headers={["Records", "Count"]}
        />
      </Section>
    </>
  );
}

interface TruthProps {
  readonly report: Report;
  readonly unit: (entity: string, state: string) => string;
  readonly entity: string | null;
  readonly only?: "kpi";
}

function Truth({ report, unit, entity }: TruthProps) {
  const at = (ns: number) => wallClock(ns, report.startNs);
  return (
    <>
      <Section title="Disturbances" note="onset · set">
        <Rows
          rows={report.disturbances.map((d) => [at(d.onsetNs), d.target, named(d.name), at(d.injectedNs)])}
          columns={["clock", "id", "label", "clock"]}
          headers={["Onset", "Target", "Disturbance", "Set at"]}
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
          columns={["clock", "id", "label", "label"]}
          headers={["Time", "Instance", "Effector", "Outcome"]}
        />
      </Section>
      <KpiTruth report={report} unit={unit} entity={entity} />
      <Section title="Alarm system" note="ISA-18.2">
        <Rows
          rows={Object.entries(report.alarms).map(([key, n]) => [key.replaceAll("_", " "), n.toLocaleString("en-US")])}
          columns={["text", "number"]}
          headers={["Measure", "Value"]}
        />
      </Section>
    </>
  );
}

function Scoring({ report }: { readonly report: Report }) {
  const text =
    report.scoring === null
      ? "The report has no scoring section."
      : `${report.scoring.status}${report.scoring.reason ? ` · ${report.scoring.reason}` : ""}`;
  return (
    <Section title="Scoring">
      <p className={styles.muted}>{text}</p>
    </Section>
  );
}

function KpiTruth({ report, unit, entity }: TruthProps) {
  const kpis = kpisOf(report, entity);
  return (
    <Section title="KPI truth" note={entity === null ? "mean · min – max" : `${entity} · mean · min – max`}>
      <Rows
        rows={kpis.map((row) => [
          { text: `${row.entity} · ${humanise(row.state)}`, title: `${row.entity} ${row.state}` },
          `${fixed(row.mean)} ${displayUnit(unit(row.entity, row.state))}`.trim(),
          `${fixed(row.min)} – ${fixed(row.max)}`,
        ])}
        columns={["label", "number", "number"]}
        headers={["KPI", "Mean", "Min – max"]}
        empty={entity === null ? "None" : `No KPI truth for ${entity}.`}
      />
    </Section>
  );
}

const notClosed = "Not available: the run is not closed, so its truth is still sealed.";

const loading: VerifyState = { phase: "loading" };
const unverifiedTruth = "This run did not verify, so the truth below may not describe the files in the run folder.";

export function ReportPanel({
  report,
  unit,
  entity,
  only,
  verify = loading,
}: TruthProps & { readonly verify?: VerifyState }) {
  if (only === "kpi") {
    return report.truthUnavailable || !report.ended ? null : <KpiTruth report={report} unit={unit} entity={entity} />;
  }
  return (
    <>
      <Facts report={report} verify={verify} />
      {badgeOf(report, verify) === "not-verified" && (
        <p className={styles.caution} role="note">
          {unverifiedTruth}
        </p>
      )}
      {report.truthUnavailable || !report.ended ? (
        <Section title="Truth">
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
