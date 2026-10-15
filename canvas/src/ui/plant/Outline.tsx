import { OctagonAlert, Search } from "lucide-react";
import { useCallback, useEffect, useRef, useState } from "react";
import type { RefObject } from "react";
import type { Plant } from "../../model/plant";
import { areaOf } from "../../model/plant";
import { areaOrder, hallOf, hallsOf } from "../../model/plantView";
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
  const all = areaOrder(plant).map(([area, depth]) => ({ area, depth, items: byArea.get(area) ?? [] }));
  const showEmpty = empty && filter.trim() === "";
  const ordered = empty ? populatedFirst(all) : all;
  return ordered.filter((group) => group.items.length > 0 || showEmpty);
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
  readonly below?: number;
  readonly current: string | null;
  readonly marks: "location" | "selection";
  readonly onPick: (item: string) => void;
}

function countWords(direct: number, below: number): string {
  if (direct > 0) return String(direct);
  return below > 0 ? `0 · ${String(below)} below` : "0";
}

function countTitle(direct: number, below: number): string | undefined {
  if (direct > 0) return undefined;
  return below > 0 ? `${String(below)} in child areas` : "No equipment yet";
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

function Section({ title, depth = 0, items, below = 0, current, marks, onPick }: SectionProps) {
  return (
    <details open className={styles.group} style={{ marginLeft: `calc(var(--space-3) * ${String(depth)})` }}>
      <summary className={cx(styles.areaRow, shared.disclosure)}>
        <span className={styles.areaName} title={title}>
          {title}
        </span>
        <span className={styles.count} title={countTitle(items.length, below)}>
          {countWords(items.length, below)}
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

function continuesBelow(node: HTMLElement | null): boolean {
  return node !== null && node.scrollTop + node.clientHeight < node.scrollHeight - 1;
}

function useCutBottom(ref: RefObject<HTMLElement | null>, content: unknown): boolean {
  const [cut, setCut] = useState(false);
  const measure = useCallback(() => {
    setCut(continuesBelow(ref.current));
  }, [ref]);
  useEffect(() => {
    const node = ref.current;
    if (node === null) return undefined;
    node.addEventListener("scroll", measure, { passive: true });
    const watcher = typeof ResizeObserver === "undefined" ? null : new ResizeObserver(measure);
    watcher?.observe(node);
    return () => {
      node.removeEventListener("scroll", measure);
      watcher?.disconnect();
    };
  }, [ref, measure]);
  useEffect(measure, [measure, content]);
  return cut;
}

function belowCount(shown: readonly Group[], index: number): number {
  const group = shown[index];
  if (group === undefined) return 0;
  let below = 0;
  for (const child of shown.slice(index + 1)) {
    if (child.depth <= group.depth) break;
    below += child.items.length;
  }
  return below;
}

export function Outline({ plant, hall, selected, onSelect, onOpenHall, flagged, showEmptyAreas }: OutlineProps) {
  const [filter, setFilter] = useState("");
  const navRef = useRef<HTMLElement | null>(null);
  const halls = hallsOf(plant).map((id) => ({ id, kind: "" }));
  const shown = marked(groups(plant, hall, filter, showEmptyAreas === true), flagged);
  const cut = useCutBottom(navRef, [plant, hall, filter, selected, flagged]);
  return (
    <nav ref={navRef} className={styles.outline} aria-label="Outline" data-cut={cut ? "true" : undefined}>
      <Filter value={filter} onChange={setFilter} />
      {halls.length > 0 && onOpenHall !== undefined && filter === "" && (
        <Section title="Halls" items={halls} current={hall} marks="location" onPick={onOpenHall} />
      )}
      {shown.map(({ area, depth, items }, index) => (
        <Section
          key={area}
          title={area}
          depth={depth}
          items={items}
          below={belowCount(shown, index)}
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
