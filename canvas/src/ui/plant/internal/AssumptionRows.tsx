import type { AssumptionRow, KindAssumptions } from "../../../model/workspace";
import { records } from "../../primitives";
import styles from "./Assumptions.module.css";

function Row({ row }: { readonly row: AssumptionRow }) {
  return (
    <li className={records.row}>
      <span className={records.rowTitle}>
        <span>{row.title}</span>
        {row.suggestion !== "" && <span className={records.tag}>{row.suggestion}</span>}
      </span>
      {row.statement !== "" && <p className={records.statement}>{row.statement}</p>}
      <dl className={styles.facts}>
        {row.facts.map((fact) => (
          <FactPair key={`${fact.label}:${fact.value}`} label={fact.label} value={fact.value} />
        ))}
      </dl>
    </li>
  );
}

function FactPair({ label, value }: { readonly label: string; readonly value: string }) {
  return (
    <>
      <dt>{label}</dt>
      <dd>{value}</dd>
    </>
  );
}

function Group({ title, rows }: { readonly title: string; readonly rows: readonly AssumptionRow[] }) {
  if (rows.length === 0) return null;
  return (
    <section className={records.group} aria-label={title}>
      <h4 className={records.groupTitle}>{title}</h4>
      <ul className={records.rows}>
        {rows.map((row) => (
          <Row key={row.key} row={row} />
        ))}
      </ul>
    </section>
  );
}

export function KindBlock({ kind }: { readonly kind: KindAssumptions }) {
  return (
    <section className={records.kind} aria-label={`Assumptions of ${kind.kind}`}>
      <h3 className={records.kindTitle}>
        <span>{kind.kind}</span>
        <span className={records.summary}>{kind.summary}</span>
      </h3>
      <Group title="Readings" rows={kind.readings} />
      <Group title="Dynamics" rows={kind.dynamics} />
      <Group title="Commands" rows={kind.commands} />
      <Group title="Faults" rows={kind.faults} />
    </section>
  );
}
