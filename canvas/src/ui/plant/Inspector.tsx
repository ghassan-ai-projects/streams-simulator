import type { Plant } from "../../model/plant";
import { linkLines, linksOf } from "../../model/plant";
import { humanise } from "../../model/report";
import shared from "../shared.module.css";
import styles from "./PlantScreen.module.css";

function List({
  title,
  items,
  mono = true,
  show,
  count = items.length,
}: {
  readonly title: string;
  readonly items: readonly string[];
  readonly mono?: boolean;
  readonly show?: (item: string) => string;
  readonly count?: number;
}) {
  return (
    <section className={styles.section} aria-label={title}>
      <h3 className={shared.sectionTitle}>
        {title}
        <span className={styles.count}>{count}</span>
      </h3>
      {items.length === 0 ? (
        <p className={styles.muted}>None</p>
      ) : (
        <ul className={styles.items}>
          {items.map((item) => (
            <li key={item} className={mono ? styles.mono : undefined} title={show && item}>
              {show ? show(item) : item}
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}

export function Inspector({ plant, selected }: { readonly plant: Plant; readonly selected: string | null }) {
  const instance = plant.instances.find((entry) => entry.id === selected);
  if (instance === undefined) {
    return (
      <aside className={styles.inspector} aria-label="Inspector" tabIndex={0}>
        <p className={styles.hint}>
          Select an instance to see its kind: states, channels, faults, effectors and links.
        </p>
      </aside>
    );
  }
  const kind = plant.kinds.get(instance.kind);
  const { incoming, outgoing } = linksOf(plant, instance.id);
  return (
    <aside className={styles.inspector} aria-label="Inspector" tabIndex={0}>
      <header className={styles.inspectorHead}>
        <h2 className={styles.mono}>{instance.id}</h2>
        <p className={styles.muted}>{[instance.kind, instance.area, instance.component].filter(Boolean).join(" · ")}</p>
      </header>
      <List
        title="Channels"
        items={(kind?.channels ?? []).map((channel) => `${channel.name}${channel.unit ? ` · ${channel.unit}` : ""}`)}
      />
      <List title="Links in" items={linkLines(incoming, instance.id, "in")} count={incoming.length} />
      <List title="Links out" items={linkLines(outgoing, instance.id, "out")} count={outgoing.length} />
      <List title="Faults" items={kind?.faults ?? []} mono={false} show={humanise} />
      <List title="Effectors" items={kind?.effectors ?? []} />
      <List title="States" items={kind?.states ?? []} />
    </aside>
  );
}
