import type { ChangeRow, DiffView, KindGroup } from "../../../model/workspace";
import { records } from "../../primitives";
import styles from "./ModelDiff.module.css";

function Change({ row }: { readonly row: ChangeRow }) {
  return (
    <li className={records.row}>
      <span className={records.rowTitle}>
        <span className={records.tag}>{row.label}</span>
        {!row.sentence.startsWith(row.subject) && <span>{row.subject}</span>}
      </span>
      <p className={records.statement}>{row.sentence}</p>
      <span className={styles.values}>
        <span className={styles.before}>{`Before: ${row.before}`}</span>
        <span>{`After: ${row.after}`}</span>
      </span>
    </li>
  );
}

function Changes({ rows }: { readonly rows: readonly ChangeRow[] }) {
  return (
    <ul className={records.rows}>
      {rows.map((row) => (
        <Change key={row.key} row={row} />
      ))}
    </ul>
  );
}

function Group({ group }: { readonly group: KindGroup }) {
  return (
    <section className={records.kind} aria-label={`Changes to ${group.kind}`}>
      <h3 className={records.kindTitle}>
        <span>{group.kind}</span>
        <span className={styles.status} data-status={group.status}>
          {group.status}
        </span>
        {group.affects !== "" && <span className={records.summary}>{group.affects}</span>}
      </h3>
      {group.relations.length > 0 && <p className={records.summary}>{`Over ${group.relations.join("; ")}`}</p>}
      <Changes rows={group.changes} />
    </section>
  );
}

export function ModelDiffView({ view }: { readonly view: DiffView }) {
  return (
    <div>
      {view.warning !== "" && (
        <p className={styles.warning} role="alert">
          {view.warning}
        </p>
      )}
      <dl className={styles.sides}>
        {view.sides.map(([label, words]) => (
          <SidePair key={label} label={label} words={words} />
        ))}
      </dl>
      <p>{view.headline}</p>
      {view.groups.map((group) => (
        <Group key={group.kind} group={group} />
      ))}
      {view.domain.length > 0 && (
        <section className={records.kind} aria-label="Changes to the plant as a whole">
          <h3 className={records.kindTitle}>The plant as a whole</h3>
          <Changes rows={view.domain} />
        </section>
      )}
      {view.unchanged !== "" && <p className={records.summary}>{view.unchanged}</p>}
    </div>
  );
}

function SidePair({ label, words }: { readonly label: string; readonly words: string }) {
  return (
    <>
      <dt>{label}</dt>
      <dd>{words}</dd>
    </>
  );
}
