import type { Plant } from "../../model/plant";
import { areaOf } from "../../model/plant";
import { hallOf } from "../../model/plantView";
import { kindLabel } from "../../model/report";
import type { RunRow } from "../../model/workspace";
import { runFacts, whenOf } from "../../model/workspace";
import shared from "../shared.module.css";
import styles from "./PlantScreen.module.css";

function counts(values: readonly string[]): readonly (readonly [string, number])[] {
  const tally = new Map<string, number>();
  values.forEach((value) => tally.set(value, (tally.get(value) ?? 0) + 1));
  return [...tally.entries()].sort((a, b) => b[1] - a[1] || a[0].localeCompare(b[0]));
}

function Tally({ title, rows }: { readonly title: string; readonly rows: readonly (readonly [string, number])[] }) {
  return (
    <section className={styles.section} aria-label={title}>
      <h3 className={shared.sectionTitle}>
        {title}
        <span className={styles.count}>{rows.length}</span>
      </h3>
      <ul className={styles.tally}>
        {rows.map(([name, n]) => (
          <li key={name}>
            <span>{name}</span>
            <span className={styles.mono}>{n}</span>
          </li>
        ))}
      </ul>
    </section>
  );
}

type Review = ((worldId: string) => void) | undefined;

function RunItem({ row, onReview }: { readonly row: RunRow; readonly onReview: Review }) {
  const words = [
    row.run.worldId,
    row.run.explore ? "Explore" : "Evaluate",
    ...runFacts(row.run),
    whenOf(row.run.recordedAt),
  ];
  return (
    <li>
      <span className={styles.figures}>{words.filter(Boolean).join(" · ")}</span>
      {onReview !== undefined && (
        <button
          type="button"
          className={styles.linkButton}
          aria-label={`Review ${row.run.worldId}`}
          onClick={() => {
            onReview(row.run.worldId);
          }}
        >
          Review
        </button>
      )}
    </li>
  );
}

function Runs({ runs, onReview }: { readonly runs: readonly RunRow[]; readonly onReview: Review }) {
  return (
    <section className={styles.section} aria-label="Recent runs">
      <h3 className={shared.sectionTitle}>
        Recent runs
        <span className={styles.count}>{runs.length}</span>
      </h3>
      {runs.length === 0 && <p className={styles.muted}>No runs yet: use Run plant.</p>}
      <ul className={styles.tally}>
        {runs.slice(0, 5).map((row) => (
          <RunItem key={row.run.worldId} row={row} onReview={onReview} />
        ))}
      </ul>
    </section>
  );
}

export function PlantOverview({
  plant,
  runs,
  onReview,
}: {
  readonly plant: Plant;
  readonly runs: readonly RunRow[];
  readonly onReview?: ((worldId: string) => void) | undefined;
}) {
  const areas = counts(plant.instances.map((instance) => areaOf(plant, instance)));
  const halls = counts(plant.instances.flatMap((instance) => hallOf(instance.id) ?? []));
  const where = halls.length > 0 ? `${String(halls.length)} halls and the site` : `${String(areas.length)} areas`;
  return (
    <aside className={styles.inspector} aria-label="Inspector" tabIndex={0}>
      <header className={styles.inspectorHead}>
        <h2>{plant.title || plant.id}</h2>
        <p
          className={styles.kindLine}
        >{`${String(plant.instances.length)} instances in ${where} · ${String(plant.links.length)} links`}</p>
        <p className={styles.hint}>
          Select an instance to see its channels (the values measured on it, one for each port and state), links, faults
          and effectors.
        </p>
      </header>
      <Runs runs={runs} onReview={onReview} />
      {halls.length > 0 && <Tally title="Halls" rows={halls} />}
      <Tally title={halls.length > 0 ? "Areas across the halls" : "Areas"} rows={areas} />
      <Tally title="Kinds" rows={counts(plant.instances.map((instance) => kindLabel(instance.kind)))} />
    </aside>
  );
}
