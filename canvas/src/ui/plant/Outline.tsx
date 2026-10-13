import type { NavigationStore } from "../../app/navigationStore";
import type { PlantState, PlantStore } from "../../app/plantStore";
import { areaOf } from "../../model/plant";
import { hallOf, hallsOf } from "../../model/plantView";
import shared from "../shared.module.css";
import styles from "./PlantScreen.module.css";

type Ready = Extract<PlantState, { phase: "ready" }>;

interface OutlineProps {
  readonly state: Ready;
  readonly hall: string | null;
  readonly plants: PlantStore;
  readonly navigation: NavigationStore;
}

function groups(state: Ready, hall: string | null): readonly (readonly [string, readonly string[]])[] {
  const halls = hallsOf(state.plant);
  const visible = state.plant.instances.filter(
    (instance) => halls.length === 0 || hallOf(instance.id) === hall || hallOf(instance.id) === null,
  );
  const byArea = new Map<string, string[]>();
  visible.forEach((instance) => {
    const area = areaOf(state.plant, instance);
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

export function Outline({ state, hall, plants, navigation }: OutlineProps) {
  const halls = hallsOf(state.plant);
  const openHall = (name: string) => {
    navigation.go({ screen: "plant", domain: state.path, hall: name });
  };
  const select = (id: string) => {
    plants.select(id);
  };
  return (
    <nav className={styles.outline} aria-label="Outline">
      {halls.length > 0 && <Section title="Halls" items={halls} current={hall} marks="location" onPick={openHall} />}
      {groups(state, hall).map(([area, ids]) => (
        <Section key={area} title={area} items={ids} current={state.selected} marks="pressed" onPick={select} />
      ))}
    </nav>
  );
}
