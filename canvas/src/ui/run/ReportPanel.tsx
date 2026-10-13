import type { ReactNode } from "react";
import type { Report } from "../../model/report";
import { humanise, wallClock } from "../../model/report";
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
      <h2 className={shared.sectionTitle}>
        {title}
        {note !== undefined && <span className={styles.note}>{note}</span>}
      </h2>
      {children}
    </section>
  );
}

function Rows({
  rows,
  columns,
  empty = "None",
}: {
  readonly rows: readonly (readonly Cell[])[];
  readonly columns: readonly string[];
  readonly empty?: string;
}) {
  if (rows.length === 0) return <p className={styles.muted}>{empty}</p>;
  return (
    <table className={styles.table}>
      <tbody>
        {rows.map((row, r) => (
          <tr key={r}>
            {row.map((cell, c) => {
              const text = typeof cell === "string" ? cell : cell.text;
              return (
                <td
                  key={c}
                  className={styles[columns[c] ?? "text"]}
                  title={typeof cell === "string" ? undefined : cell.title}
                >
                  {text}
                </td>
              );
            })}
          </tr>
        ))}
      </tbody>
    </table>
  );
}

const fixed = (value: number) =>
  Math.abs(value) >= 100 ? value.toFixed(0) : Math.abs(value) >= 10 ? value.toFixed(1) : value.toFixed(2);
const named = (id: string): Cell => ({ text: humanise(id), title: id });
const count = (report: Report, key: string) =>
  report.ended ? (report.counts[key] ?? 0).toLocaleString("en-US") : "n/a";

function Facts({ report }: { readonly report: Report }) {
  const dropped = Object.entries(report.dropped).map(([reason, n]) => [
    reason.replaceAll("_", " "),
    n.toLocaleString("en-US"),
  ]);
  return (
    <>
      <Section title="Verification" note={report.verification.replaceAll("_", " ") || "not recorded"}>
        <Rows rows={report.checks.map((check) => [check.name, check.status])} columns={["id", "text"]} />
      </Section>
      <Section title="Deliveries" note={report.ended ? undefined : "run not closed"}>
        <Rows
          rows={[...["emitted", "delivered", "dropped"].map((key) => [key, count(report, key)]), ...dropped]}
          columns={["text", "number"]}
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
      <Section title="Disturbances" note="onset · set at">
        <Rows
          rows={report.disturbances.map((d) => [
            at(d.onsetNs),
            d.injectedNs === d.onsetNs ? "" : `set ${at(d.injectedNs)}`,
            d.target,
            named(d.name),
          ])}
          columns={["clock", "clock", "id", "label"]}
        />
      </Section>
      <Section title="Effector calls">
        <Rows
          rows={report.effectorCalls.map((call) => [at(call.atNs), call.entity, named(call.effector), call.outcome])}
          columns={["clock", "id", "label", "text"]}
        />
      </Section>
      <KpiTruth report={report} unit={unit} entity={entity} />
      <Section title="Alarm system" note="ISA-18.2">
        <Rows
          rows={Object.entries(report.alarms).map(([key, n]) => [key.replaceAll("_", " "), String(n)])}
          columns={["text", "number"]}
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
  const kpis = report.kpis.filter((row) => entity === null || row.entity === entity);
  return (
    <Section title="KPI truth" note={entity === null ? "mean · min – max" : `${entity} · mean · min – max`}>
      <Rows
        rows={kpis.map((row) => [
          { text: `${row.entity} · ${humanise(row.state)}`, title: `${row.entity} ${row.state}` },
          `${fixed(row.mean)} ${displayUnit(unit(row.entity, row.state))}`.trim(),
          `${fixed(row.min)} – ${fixed(row.max)}`,
        ])}
        columns={["label", "number", "number"]}
        empty={entity === null ? "None" : `No KPI truth for ${entity}.`}
      />
    </Section>
  );
}

export function ReportPanel({ report, unit, entity, only }: TruthProps) {
  if (only === "kpi") {
    return report.truthUnavailable ? null : <KpiTruth report={report} unit={unit} entity={entity} />;
  }
  return (
    <>
      <Facts report={report} />
      {report.truthUnavailable ? (
        <Section title="Truth">
          <p className={styles.muted}>{`Not available: ${report.truthUnavailable}`}</p>
        </Section>
      ) : (
        <Truth report={report} unit={unit} entity={entity} />
      )}
      <Scoring report={report} />
    </>
  );
}
