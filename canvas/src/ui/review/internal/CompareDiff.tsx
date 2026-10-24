import type { Comparison, PlanLane, PlanRow } from "../../../model/review";
import { planLanes } from "../../../model/review";
import styles from "./Compare.module.css";

function Row({ row }: { readonly row: PlanRow }) {
  return (
    <li className={styles.change} data-change={row.change}>
      <span className={styles.tag}>{row.label}</span>
      {row.change === "changed" ? (
        <span className={styles.versions}>
          <span>{row.text}</span>
          <span>{`Reference: ${row.reference ?? ""}`}</span>
          <span>{`This run: ${row.thisRun ?? ""}`}</span>
          <span className={styles.fields}>{`Differs in: ${row.differences.join(", ")}`}</span>
        </span>
      ) : (
        <span>{row.text}</span>
      )}
    </li>
  );
}

function Lane({ lane }: { readonly lane: PlanLane }) {
  const same = lane.rows.length === 0 && lane.identical === "";
  return (
    <section className={styles.lane} aria-label={lane.title}>
      <h4 className={styles.laneTitle}>{lane.title}</h4>
      {same && <p className={styles.muted}>None in either run.</p>}
      {lane.rows.length > 0 && (
        <ul className={styles.changes}>
          {lane.rows.map((row) => (
            <Row key={row.key} row={row} />
          ))}
        </ul>
      )}
      {lane.identical !== "" && <p className={styles.muted}>{lane.identical}</p>}
    </section>
  );
}

export function CompareDiff({ comparison }: { readonly comparison: Comparison }) {
  return (
    <section className={styles.block} aria-label="What differs">
      <h3 className={styles.blockTitle}>What differs</h3>
      {planLanes(comparison).map((lane) => (
        <Lane key={lane.key} lane={lane} />
      ))}
    </section>
  );
}
