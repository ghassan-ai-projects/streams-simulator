import { OctagonAlert, Search } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { areaTree } from "../../model/areaTree";
import type { Plant } from "../../model/plant";
import { areaOf } from "../../model/plant";
import { hallOf, hallsOf } from "../../model/plantView";
import { kindLabel } from "../../model/report";
import { cx } from "../cx";
import shared from "../shared.module.css";
import styles from "./PlantScreen.module.css";

interface OutlineProps {
  readonly flagged?: ReadonlySet<string>;
  readonly plant: Plant;
  readonly hall: string | null;
  readonly selected: string | null;
  readonly onSelect: (id: string) => void;
  readonly onOpenHall?: (hall: string) => void;
  readonly showEmptyAreas?: boolean;
}

interface Item {
  readonly id: string;
  readonly kind: string;
  readonly flagged?: boolean;
}

function matches(item: Item, filter: string): boolean {
  const wanted = filter.trim().toLowerCase();
  return wanted === "" || item.id.toLowerCase().includes(wanted) || item.kind.toLowerCase().includes(wanted);
}

function kindWords(kind: string): string {
  const words = kindLabel(kind);
  const acronym =
    words.charAt(1) === words.charAt(1).toUpperCase() && words.charAt(1) !== words.charAt(1).toLowerCase();
  return acronym ? words : words.charAt(0).toLowerCase() + words.slice(1);
}

interface Group {
  readonly area: string;
  readonly depth: number;
  readonly items: readonly Item[];
}

function itemsByArea(plant: Plant, hall: string | null, filter: string): ReadonlyMap<string, readonly Item[]> {
  const halls = hallsOf(plant);
  const visible = plant.instances.filter(
    (instance) => halls.length === 0 || hallOf(instance.id) === hall || hallOf(instance.id) === null,
  );
  const byArea = new Map<string, Item[]>();
  visible.forEach((instance) => {
    const item = { id: instance.id, kind: kindWords(instance.kind) };
    if (!matches(item, filter)) return;
    const area = areaOf(plant, instance);
    byArea.set(area, [...(byArea.get(area) ?? []), item]);
  });
  return byArea;
}

function groups(plant: Plant, hall: string | null, filter: string, empty: boolean): readonly Group[] {
  const byArea = itemsByArea(plant, hall, filter);
  const tree = areaTree(plant.areas).map((node) => ({ area: node.name, depth: node.depth }));
  const known = new Set(tree.map((node) => node.area));
  const loose = [...byArea.keys()].filter((area) => !known.has(area)).map((area) => ({ area, depth: 0 }));
  const showEmpty = empty && filter.trim() === "";
  const all = [...tree, ...loose].map(({ area, depth }) => ({ area, depth, items: byArea.get(area) ?? [] }));
  return populatedFirst(all).filter((group) => group.items.length > 0 || showEmpty);
}

interface Branch {
  readonly group: Group;
  readonly children: Branch[];
}

function branchesOf(all: readonly Group[]): Branch[] {
  const roots: Branch[] = [];
  const open: Branch[] = [];
  for (const group of all) {
    const branch: Branch = { group, children: [] };
    open.length = Math.min(open.length, group.depth);
    const parent = open[open.length - 1];
    if (group.depth === 0 || parent === undefined) roots.push(branch);
    else parent.children.push(branch);
    open.push(branch);
  }
  return roots;
}

function total(branch: Branch): number {
  return branch.children.reduce((sum, child) => sum + total(child), branch.group.items.length);
}

function splitBranches(branches: readonly Branch[]): { readonly shown: Group[]; readonly empty: Group[] } {
  const ordered = [...branches].sort((a, b) => total(b) - total(a));
  const shown: Group[] = [];
  const empty: Group[] = [];
  for (const branch of ordered) {
    const below = splitBranches(branch.children);
    if (total(branch) === 0) empty.push(branch.group, ...below.empty);
    else {
      shown.push(branch.group, ...below.shown);
      empty.push(...below.empty);
    }
  }
  return { shown, empty };
}

function populatedFirst(all: readonly Group[]): Group[] {
  const { shown, empty } = splitBranches(branchesOf(all));
  return [...shown, ...empty];
}

interface SectionProps {
  readonly title: string;
  readonly depth?: number;
  readonly items: readonly Item[];
  readonly current: string | null;
  readonly marks: "location" | "selection";
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
        aria-current={isCurrent ? (marks === "location" ? "location" : "true") : undefined}
        onClick={() => {
          onPick(item.id);
        }}
      >
        <span className={styles.itemId}>{item.id}</span>
        {item.kind !== "" && <span className={styles.itemKind}>{item.kind}</span>}
        {item.flagged === true && (
          <span className={styles.itemIssue}>
            <OctagonAlert size={12} aria-hidden="true" />
            <span className="sr-only">has issues</span>
          </span>
        )}
      </button>
    </li>
  );
}

function Section({ title, depth = 0, items, current, marks, onPick }: SectionProps) {
  return (
    <details open className={styles.group} style={{ marginLeft: `calc(var(--space-3) * ${String(depth)})` }}>
      <summary className={cx(shared.sectionTitle, items.length === 0 && styles.emptyRow)}>
        <span className={styles.areaName} title={title}>
          {title}
        </span>
        <span className={styles.count} title={items.length === 0 ? "No equipment yet" : undefined}>
          {items.length === 0 ? "0 · empty" : items.length}
        </span>
      </summary>
      {items.length > 0 && (
        <ul className={styles.tree}>
          {items.map((item) => (
            <Entry key={item.id} item={item} current={current} marks={marks} onPick={onPick} />
          ))}
        </ul>
      )}
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

function matchWords(plant: Plant, shown: readonly Group[], filter: string): string {
  if (plant.instances.length === 0) return "No instances yet.";
  const count = shown.reduce((sum, group) => sum + group.items.length, 0);
  if (filter.trim() === "") return "";
  if (count === 0) return `No instance matches "${filter.trim()}".`;
  return `${String(count)} ${count === 1 ? "instance matches" : "instances match"}.`;
}

function marked(shown: readonly Group[], flagged: ReadonlySet<string> | undefined): readonly Group[] {
  if (flagged === undefined || flagged.size === 0) return shown;
  return shown.map((group) => ({
    ...group,
    items: group.items.map((item) => ({ ...item, flagged: flagged.has(item.id) })),
  }));
}

export function Outline({ plant, hall, selected, onSelect, onOpenHall, flagged, showEmptyAreas }: OutlineProps) {
  const [filter, setFilter] = useState("");
  const halls = hallsOf(plant).map((id) => ({ id, kind: "" }));
  const shown = marked(groups(plant, hall, filter, showEmptyAreas === true), flagged);
  return (
    <nav className={styles.outline} aria-label="Outline">
      <Filter value={filter} onChange={setFilter} />
      {halls.length > 0 && onOpenHall !== undefined && filter === "" && (
        <Section title="Halls" items={halls} current={hall} marks="location" onPick={onOpenHall} />
      )}
      {shown.map(({ area, depth, items }) => (
        <Section
          key={area}
          title={area}
          depth={depth}
          items={items}
          current={selected}
          marks="selection"
          onPick={onSelect}
        />
      ))}
      <p className={shown.length === 0 ? styles.noMatch : "sr-only"} role="status">
        {matchWords(plant, shown, filter)}
      </p>
    </nav>
  );
}
