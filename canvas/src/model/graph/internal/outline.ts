import type { OutlineItem, OutlineGroup } from "./types";
import type { Plant } from "../../plant";
import { areaOf } from "../../plant";
import { areaOrder, hallOf, hallsOf } from "./plantView";
import { kindLabel } from "../../shared/words";

function matches(item: OutlineItem, filter: string): boolean {
  const wanted = filter.trim().toLowerCase();
  return wanted === "" || item.id.toLowerCase().includes(wanted) || item.kind.toLowerCase().includes(wanted);
}

function kindWords(kind: string): string {
  const words = kindLabel(kind);
  const acronym =
    words.charAt(1) === words.charAt(1).toUpperCase() && words.charAt(1) !== words.charAt(1).toLowerCase();
  return acronym ? words : words.charAt(0).toLowerCase() + words.slice(1);
}

function itemsByArea(plant: Plant, hall: string | null, filter: string): ReadonlyMap<string, readonly OutlineItem[]> {
  const halls = hallsOf(plant);
  const visible = plant.instances.filter(
    (instance) => halls.length === 0 || hallOf(instance.id) === hall || hallOf(instance.id) === null,
  );
  const byArea = new Map<string, OutlineItem[]>();
  visible.forEach((instance) => {
    const item = { id: instance.id, kind: kindWords(instance.kind) };
    if (!matches(item, filter)) return;
    const area = areaOf(plant, instance);
    byArea.set(area, [...(byArea.get(area) ?? []), item]);
  });
  return byArea;
}

export function groups(plant: Plant, hall: string | null, filter: string, empty: boolean): readonly OutlineGroup[] {
  const byArea = itemsByArea(plant, hall, filter);
  const all = areaOrder(plant).map(([area, depth]) => ({ area, depth, items: byArea.get(area) ?? [] }));
  const showEmpty = empty && filter.trim() === "";
  const ordered = empty ? populatedFirst(all) : all;
  return ordered.filter((group) => group.items.length > 0 || showEmpty);
}

interface Branch {
  readonly group: OutlineGroup;
  readonly children: Branch[];
}

function branchesOf(all: readonly OutlineGroup[]): Branch[] {
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

function splitBranches(branches: readonly Branch[]): {
  readonly shown: OutlineGroup[];
  readonly empty: OutlineGroup[];
} {
  const ordered = [...branches].sort((a, b) => total(b) - total(a));
  const shown: OutlineGroup[] = [];
  const empty: OutlineGroup[] = [];
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

function populatedFirst(all: readonly OutlineGroup[]): OutlineGroup[] {
  const { shown, empty } = splitBranches(branchesOf(all));
  return [...shown, ...empty];
}

export function matchWords(plant: Plant, shown: readonly OutlineGroup[], filter: string): string {
  if (plant.instances.length === 0) return "No instances yet.";
  const count = shown.reduce((sum, group) => sum + group.items.length, 0);
  if (filter.trim() === "") return "";
  if (count === 0) return `No instance matches "${filter.trim()}".`;
  return `${String(count)} ${count === 1 ? "instance matches" : "instances match"}.`;
}

export function marked(
  shown: readonly OutlineGroup[],
  flagged: ReadonlySet<string> | undefined,
): readonly OutlineGroup[] {
  if (flagged === undefined || flagged.size === 0) return shown;
  return shown.map((group) => ({
    ...group,
    items: group.items.map((item) => ({ ...item, flagged: flagged.has(item.id) })),
  }));
}

export function belowCount(shown: readonly OutlineGroup[], index: number): number {
  const group = shown[index];
  if (group === undefined) return 0;
  let below = 0;
  for (const child of shown.slice(index + 1)) {
    if (child.depth <= group.depth) break;
    below += child.items.length;
  }
  return below;
}
