import type { Plant } from "../../model/plant";
import { linksOf } from "../../model/plant";
import shared from "../shared.module.css";
import styles from "./PlantScreen.module.css";

function List({
  title,
  items,
  mono = true,
}: {
  readonly title: string;
  readonly items: readonly string[];
  readonly mono?: boolean;
}) {
  return (
    <section className={styles.section} aria-label={title}>
      <h3 className={shared.sectionTitle}>
        {title}
        <span className={styles.count}>{items.length}</span>
      </h3>
      {items.length === 0 ? (
        <p className={styles.muted}>None</p>
      ) : (
        <ul className={styles.items}>
          {items.map((item) => (
            <li key={item} className={mono ? styles.mono : undefined}>
              {item}
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
      <aside className={styles.inspector} aria-label="Inspector">
        <p className={styles.hint}>
          Select an instance to see its kind: states, channels, faults, effectors and links.
        </p>
      </aside>
    );
  }
  const kind = plant.kinds.get(instance.kind);
  const { incoming, outgoing } = linksOf(plant, instance.id);
  const end = (id: string, port: string | null) => (port === null ? id : `${id}.${port}`);
  return (
    <aside className={styles.inspector} aria-label="Inspector">
      <header className={styles.inspectorHead}>
        <h2 className={styles.mono}>{instance.id}</h2>
        <p className={styles.muted}>{[instance.kind, instance.area, instance.component].filter(Boolean).join(" · ")}</p>
      </header>
      <List
        title="Channels"
        items={(kind?.channels ?? []).map((channel) => `${channel.name}${channel.unit ? ` · ${channel.unit}` : ""}`)}
      />
      <List
        title="Links in"
        items={incoming.map((link) => `${end(link.from, link.fromPort)} → ${link.toPort ?? instance.id}`)}
      />
      <List
        title="Links out"
        items={outgoing.map((link) => `${link.fromPort ?? instance.id} → ${end(link.to, link.toPort)}`)}
      />
      <List title="Faults" items={kind?.faults ?? []} />
      <List title="Effectors" items={kind?.effectors ?? []} />
      <List title="States" items={kind?.states ?? []} />
    </aside>
  );
}
