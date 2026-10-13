import type { ReactNode } from "react";
import type { Report } from "../../model/report";
import { wallClock } from "../../model/report";
import shared from "../shared.module.css";
import styles from "./RunScreen.module.css";

function Section({
  title,
  note,
  children,
}: {
  readonly title: string;
  readonly note?: string;
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

function Rows({ rows }: { readonly rows: readonly (readonly string[])[] }) {
  if (rows.length === 0) {
    return <p className={styles.muted}>None</p>;
  }
  return (
    <table className={styles.table}>
      <tbody>
        {rows.map((row) => (
          <tr key={row.join("|")}>
            {row.map((cell, index) => (
              <td
                key={`${String(index)}:${cell}`}
                className={index > 0 && /^[\d.\-−–, ]+$/.test(cell) ? styles.number : undefined}
              >
                {cell}
              </td>
            ))}
          </tr>
        ))}
      </tbody>
    </table>
  );
}

const fixed = (value: number) =>
  Math.abs(value) >= 100 ? value.toFixed(0) : Math.abs(value) >= 10 ? value.toFixed(1) : value.toFixed(2);

function Facts({ report }: { readonly report: Report }) {
  const counts = ["emitted", "delivered", "dropped"].map((key) => [
    key,
    (report.counts[key] ?? 0).toLocaleString("en-US"),
  ]);
  const dropped = Object.entries(report.dropped).map(([reason, n]) => [
    reason.replaceAll("_", " "),
    n.toLocaleString("en-US"),
  ]);
  return (
    <>
      <Section title="Verification" note={report.verification.replace("_", " ")}>
        <Rows rows={report.checks.map((check) => [check.name, check.status])} />
      </Section>
      <Section title="Deliveries">
        <Rows rows={[...counts, ...dropped]} />
      </Section>
    </>
  );
}

function Truth({ report }: { readonly report: Report }) {
  return (
    <>
      <Section title="Disturbances" note={report.mode === "evaluate" ? "truth, after sealing" : "explore"}>
        <Rows rows={report.disturbances.map((d) => [wallClock(d.atNs), d.what, d.target, d.detail])} />
      </Section>
      <Section title="Effector calls">
        <Rows
          rows={report.effectorCalls.map((call) => [wallClock(call.atNs), call.entity, call.effector, call.outcome])}
        />
      </Section>
      <Section title="KPI truth" note="mean · min – max">
        <Rows
          rows={report.kpis.map((row) => [
            row.entity,
            row.state,
            fixed(row.mean),
            `${fixed(row.min)} – ${fixed(row.max)}`,
          ])}
        />
      </Section>
      <Section title="Alarm system" note="ISA-18.2">
        <Rows rows={Object.entries(report.alarms).map(([key, n]) => [key.replaceAll("_", " "), String(n)])} />
      </Section>
    </>
  );
}

export function ReportPanel({ report }: { readonly report: Report }) {
  const scoring =
    report.scoring === "not scored" ? "Not scored: no consumer verdict was submitted for this run." : report.scoring;
  return (
    <>
      <Facts report={report} />
      <Truth report={report} />
      <Section title="Scoring">
        <p className={styles.muted}>{scoring}</p>
      </Section>
    </>
  );
}
