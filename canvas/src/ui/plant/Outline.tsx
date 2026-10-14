import type { Plant } from "../../model/plant";
import { areaOf } from "../../model/plant";
import { hallOf, hallsOf } from "../../model/plantView";
import shared from "../shared.module.css";
import styles from "./PlantScreen.module.css";

interface OutlineProps {
  readonly plant: Plant;
  readonly hall: string | null;
  readonly selected: string | null;
  readonly onSelect: (id: string) => void;
  readonly onOpenHall?: (hall: string) => void;
}

function groups(plant: Plant, hall: string | null): readonly (readonly [string, readonly string[]])[] {
  const halls = hallsOf(plant);
  const visible = plant.instances.filter(
    (instance) => halls.length === 0 || hallOf(instance.id) === hall || hallOf(instance.id) === null,
  );
  const byArea = new Map<string, string[]>();
  visible.forEach((instance) => {
    const area = areaOf(plant, instance);
    byArea.set(area, [...(byArea.get(area) ?? []), instance.id]);
  });
  return [...byArea.entries()];
}

interface SectionProps {
  readonly title: string;
  readonly items: readonly string[];
  readonly current: string | null;
  readonly marks: "location" | "pressed";
  readonly onPick: (item: string) => void;
}

function Section({ title, items, current, marks, onPick }: SectionProps) {
  return (
    <details open className={styles.group}>
      <summary className={shared.sectionTitle}>
        {title}
        <span className={styles.count}>{items.length}</span>
      </summary>
      <ul className={styles.tree}>
        {items.map((item) => (
          <li key={item}>
            <button
              type="button"
              className={item === current ? styles.current : styles.item}
              aria-current={marks === "location" && item === current ? "location" : undefined}
              aria-pressed={marks === "pressed" ? item === current : undefined}
              onClick={() => {
                onPick(item);
              }}
            >
              {item}
            </button>
          </li>
        ))}
      </ul>
    </details>
  );
}

export function Outline({ plant, hall, selected, onSelect, onOpenHall }: OutlineProps) {
  const halls = hallsOf(plant);
  return (
    <nav className={styles.outline} aria-label="Outline">
      {halls.length > 0 && onOpenHall !== undefined && (
        <Section title="Halls" items={halls} current={hall} marks="location" onPick={onOpenHall} />
      )}
      {groups(plant, hall).map(([area, ids]) => (
        <Section key={area} title={area} items={ids} current={selected} marks="pressed" onPick={onSelect} />
      ))}
    </nav>
  );
}
