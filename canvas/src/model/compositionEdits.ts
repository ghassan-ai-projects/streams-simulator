import type { Issue } from "./bridge";
import type { Composition, NewEquipment } from "./composition";
import { componentOf, connectionsOf, endpointOf, equipmentOf, shortKind } from "./composition";
import { freeKind, joinable, kindOf } from "./kindNames";

type Json = Readonly<Record<string, unknown>>;

function entries(doc: Composition): readonly Json[] {
  return Array.isArray(doc.equipment) ? (doc.equipment as readonly Json[]) : [];
}

function wires(doc: Composition): readonly Json[] {
  return Array.isArray(doc.connections) ? (doc.connections as readonly Json[]) : [];
}

function namesOf(entry: Json): readonly string[] {
  return Array.isArray(entry.names) ? (entry.names as readonly string[]) : [];
}

function holds(entry: Json, id: string): boolean {
  return namesOf(entry).includes(id) || (namesOf(entry).length === 0 && entry.id_template === id);
}

export function isTemplate(doc: Composition, id: string): boolean {
  return entries(doc).some((entry) => namesOf(entry).length === 0 && entry.id_template === id);
}

function withName(entries_: readonly Json[], at: number, name: string): readonly Json[] {
  return entries_.map((entry, index) => (index === at ? { ...entry, names: [...namesOf(entry), name] } : entry));
}

function distinctKind(all: readonly Json[], entry: Json): Json {
  const kinds = new Set(all.map(kindOf));
  const base = shortKind(componentOf(String(entry.use)));
  if (!kinds.has(kindOf(entry))) return entry;
  return { ...entry, as: freeKind(base, kinds) };
}

function placed(all: readonly Json[], entry: Json): readonly Json[] {
  const at = all.findIndex((existing) => joinable(existing, entry));
  const name = namesOf(entry)[0];
  if (at >= 0 && name !== undefined) return withName(all, at, name);
  return [...all, distinctKind(all, entry)];
}

export function addEquipment(doc: Composition, added: NewEquipment): Composition {
  const entry: Json = {
    use: added.use,
    ...(added.preset === "" ? {} : { preset: added.preset }),
    names: [added.name],
    area: added.area,
  };
  return { ...doc, equipment: placed(entries(doc), entry) };
}

function areasOf(doc: Composition): readonly Json[] {
  return Array.isArray(doc.areas) ? (doc.areas as readonly Json[]) : [];
}

export function areaNameProblem(doc: Composition, name: string, current: string | null): string | null {
  if (name === "") return "Give the area a name.";
  const taken = name !== current && areasOf(doc).some((area) => area.name === name);
  return taken ? `There is already an area named "${name}".` : null;
}

export function addArea(doc: Composition, name: string, parent: string): Composition {
  if (areaNameProblem(doc, name, null) !== null) return doc;
  return { ...doc, areas: [...areasOf(doc), { name, ...(parent === "" ? {} : { parent }) }] };
}

function renamedArea(area: Json, from: string, to: string): Json {
  const name = area.name === from ? to : area.name;
  return area.parent === from ? { ...area, name, parent: to } : { ...area, name };
}

export function renameArea(doc: Composition, from: string, to: string): Composition {
  if (from === to || areaNameProblem(doc, to, from) !== null) return doc;
  return {
    ...doc,
    areas: areasOf(doc).map((area) => renamedArea(area, from, to)),
    equipment: entries(doc).map((entry) => (entry.area === from ? { ...entry, area: to } : entry)),
  };
}

function instanceCount(entry: Json): number {
  const names = namesOf(entry);
  const count = entry.count as { readonly default?: number } | undefined;
  return names.length > 0 ? names.length : (count?.default ?? 1);
}

export function areaInUse(doc: Composition, name: string): string | null {
  const holding = entries(doc).filter((entry) => entry.area === name);
  const inside = areasOf(doc).filter((area) => area.parent === name).length;
  const instances = holding.reduce((sum, entry) => sum + instanceCount(entry), 0);
  if (instances > 0)
    return `Move the ${String(instances)} ${instances === 1 ? "instance" : "instances"} out of it first.`;
  if (inside > 0) return `Delete the ${String(inside)} ${inside === 1 ? "area" : "areas"} inside it first.`;
  return null;
}

export function removeArea(doc: Composition, name: string): Composition {
  if (areaInUse(doc, name) !== null) return doc;
  return { ...doc, areas: areasOf(doc).filter((area) => area.name !== name) };
}

function renamed(endpoint: string, from: string, to: string): string {
  const { id, port } = endpointOf(endpoint);
  return id === from ? `${to}.${port}` : endpoint;
}

export function renameInstance(doc: Composition, from: string, to: string): Composition {
  return {
    ...doc,
    equipment: entries(doc).map((entry) =>
      namesOf(entry).includes(from)
        ? { ...entry, names: namesOf(entry).map((name) => (name === from ? to : name)) }
        : entry,
    ),
    connections: wires(doc).map((wire) => ({
      ...wire,
      from: renamed(String(wire.from), from, to),
      to: renamed(String(wire.to), from, to),
    })),
  };
}

function touches(wire: Json, id: string): boolean {
  return endpointOf(String(wire.from)).id === id || endpointOf(String(wire.to)).id === id;
}

export function removeInstance(doc: Composition, id: string): Composition {
  const equipment = entries(doc).flatMap((entry) => {
    if (!holds(entry, id)) return [entry];
    const names = namesOf(entry).filter((name) => name !== id);
    return names.length === 0 ? [] : [{ ...entry, names }];
  });
  return { ...doc, equipment, connections: wires(doc).filter((wire) => !touches(wire, id)) };
}

function without(all: readonly Json[], source: Json, id: string): readonly Json[] {
  return all.flatMap((entry) => {
    if (entry !== source) return [entry];
    const names = namesOf(entry).filter((name) => name !== id);
    return names.length === 0 ? [] : [{ ...entry, names }];
  });
}

export function moveInstance(doc: Composition, id: string, area: string): Composition {
  const all = entries(doc);
  const source = all.find((entry) => holds(entry, id));
  if (source === undefined) return doc;
  const moved = { ...source, names: [id], area };
  const others = all.filter((entry) => entry !== source);
  if (namesOf(source).length <= 1 && !others.some((entry) => joinable(entry, moved))) {
    return { ...doc, equipment: all.map((entry) => (entry === source ? moved : entry)) };
  }
  return { ...doc, equipment: placed(without(all, source, id), moved) };
}

export function connect(doc: Composition, from: string, to: string): Composition {
  if (wires(doc).some((wire) => wire.from === from && wire.to === to)) return doc;
  return { ...doc, connections: [...wires(doc), { from, to }] };
}

export function disconnect(doc: Composition, index: number): Composition {
  return { ...doc, connections: wires(doc).filter((_, at) => at !== index) };
}

export interface ListedIssue {
  readonly target: string | null;
  readonly message: string;
  readonly pointer: string;
}

export interface IssueTargets {
  readonly instances: ReadonlyMap<string, readonly string[]>;
  readonly listed: readonly ListedIssue[];
}

export interface IssueGroup extends ListedIssue {
  readonly count: number;
}

export function groupIssues(listed: readonly ListedIssue[]): readonly IssueGroup[] {
  const groups = new Map<string, IssueGroup>();
  listed.forEach((issue) => {
    const key = `${issue.target ?? ""}\n${issue.message}\n${issue.target === null ? issue.pointer : ""}`;
    const seen = groups.get(key);
    groups.set(key, { ...issue, count: (seen?.count ?? 0) + 1 });
  });
  return [...groups.values()];
}

function equipmentTargets(doc: Composition, at: number, field: string, index: string): readonly string[] {
  const names = equipmentOf(doc)[at]?.names ?? [];
  if (field !== "names" || !/^\d+$/.test(index)) return names;
  const name = names[Number(index)];
  return name === undefined ? [] : [name];
}

function targetsOf(doc: Composition, pointer: string): readonly string[] {
  const [, collection = "", index = "", field = "", item = ""] = pointer.split("/");
  if (!/^\d+$/.test(index)) return [];
  const at = Number(index);
  if (collection === "equipment") return equipmentTargets(doc, at, field, item);
  if (collection !== "connections") return [];
  const wire = connectionsOf(doc)[at];
  return wire === undefined ? [] : [endpointOf(wire.from).id, endpointOf(wire.to).id];
}

export function issueTargets(doc: Composition, issues: readonly Issue[]): IssueTargets {
  const instances = new Map<string, string[]>();
  const listed = issues.map((issue) => {
    const targets = targetsOf(doc, issue.pointer);
    targets.forEach((id) => instances.set(id, [...(instances.get(id) ?? []), issue.message]));
    return { target: targets[0] ?? null, message: issue.message, pointer: issue.pointer };
  });
  return { instances, listed };
}
