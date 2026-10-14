import { Search } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import type { Plant } from "../../model/plant";
import { areaOf } from "../../model/plant";
import { hallOf, hallsOf } from "../../model/plantView";
import { humanise } from "../../model/report";
import shared from "../shared.module.css";
import styles from "./PlantScreen.module.css";

interface OutlineProps {
  readonly plant: Plant;
  readonly hall: string | null;
  readonly selected: string | null;
  readonly onSelect: (id: string) => void;
  readonly onOpenHall?: (hall: string) => void;
}

interface Item {
  readonly id: string;
  readonly kind: string;
}

function matches(item: Item, filter: string): boolean {
  const wanted = filter.trim().toLowerCase();
  return wanted === "" || item.id.toLowerCase().includes(wanted) || item.kind.toLowerCase().includes(wanted);
}

function groups(plant: Plant, hall: string | null, filter: string): readonly (readonly [string, readonly Item[]])[] {
  const halls = hallsOf(plant);
  const visible = plant.instances.filter(
    (instance) => halls.length === 0 || hallOf(instance.id) === hall || hallOf(instance.id) === null,
  );
  const byArea = new Map<string, Item[]>();
  visible.forEach((instance) => {
    const item = { id: instance.id, kind: humanise(instance.kind) };
    if (!matches(item, filter)) return;
    const area = areaOf(plant, instance);
    byArea.set(area, [...(byArea.get(area) ?? []), item]);
  });
  return [...byArea.entries()];
}

interface SectionProps {
  readonly title: string;
  readonly items: readonly Item[];
  readonly current: string | null;
  readonly marks: "location" | "pressed";
  readonly onPick: (item: string) => void;
}

function Entry({ item, current, marks, onPick }: Omit<SectionProps, "title" | "items"> & { readonly item: Item }) {
  const button = useRef<HTMLButtonElement>(null);
  const isCurrent = item.id === current;
  useEffect(() => {
    if (isCurrent) button.current?.scrollIntoView({ block: "nearest" });
  }, [isCurrent]);
  return (
    <li>
      <button
        ref={button}
        type="button"
        className={isCurrent ? styles.current : styles.item}
        aria-current={marks === "location" && isCurrent ? "location" : undefined}
        aria-pressed={marks === "pressed" ? isCurrent : undefined}
        onClick={() => {
          onPick(item.id);
        }}
      >
        <span className={styles.itemId}>{item.id}</span>
        {item.kind !== "" && <span className={styles.itemKind}>{item.kind}</span>}
      </button>
    </li>
  );
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
          <Entry key={item.id} item={item} current={current} marks={marks} onPick={onPick} />
        ))}
      </ul>
    </details>
  );
}

function Filter({ value, onChange }: { readonly value: string; readonly onChange: (value: string) => void }) {
  return (
    <label className={styles.filter}>
      <Search size={14} aria-hidden="true" />
      <input
        type="search"
        placeholder="Filter by id or kind"
        aria-label="Filter instances"
        value={value}
        onChange={(event) => {
          onChange(event.target.value);
        }}
      />
    </label>
  );
}

export function Outline({ plant, hall, selected, onSelect, onOpenHall }: OutlineProps) {
  const [filter, setFilter] = useState("");
  const halls = hallsOf(plant).map((id) => ({ id, kind: "" }));
  const shown = groups(plant, hall, filter);
  return (
    <nav className={styles.outline} aria-label="Outline">
      <Filter value={filter} onChange={setFilter} />
      {halls.length > 0 && onOpenHall !== undefined && filter === "" && (
        <Section title="Halls" items={halls} current={hall} marks="location" onPick={onOpenHall} />
      )}
      {shown.map(([area, items]) => (
        <Section key={area} title={area} items={items} current={selected} marks="pressed" onPick={onSelect} />
      ))}
      {shown.length === 0 && <p className={styles.noMatch}>{`No instance matches "${filter.trim()}".`}</p>}
    </nav>
  );
}
