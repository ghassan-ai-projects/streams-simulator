import type { Issue } from "./bridge";
import type { Composition, NewEquipment } from "./composition";
import { connectionsOf, endpointOf, equipmentOf } from "./composition";

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

export function addEquipment(doc: Composition, added: NewEquipment): Composition {
  const entry: Json = {
    use: added.use,
    ...(added.preset === "" ? {} : { preset: added.preset }),
    names: [added.name],
    area: added.area,
  };
  return { ...doc, equipment: [...entries(doc), entry] };
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

export function moveInstance(doc: Composition, id: string, area: string): Composition {
  const equipment = entries(doc).flatMap((entry) => {
    if (!holds(entry, id)) return [entry];
    if (namesOf(entry).length <= 1) return [{ ...entry, area }];
    return [
      { ...entry, names: namesOf(entry).filter((name) => name !== id) },
      { ...entry, names: [id], area },
    ];
  });
  return { ...doc, equipment };
}

export function connect(doc: Composition, from: string, to: string): Composition {
  if (wires(doc).some((wire) => wire.from === from && wire.to === to)) return doc;
  return { ...doc, connections: [...wires(doc), { from, to }] };
}

export function disconnect(doc: Composition, index: number): Composition {
  return { ...doc, connections: wires(doc).filter((_, at) => at !== index) };
}

export interface IssueTargets {
  readonly instances: ReadonlyMap<string, readonly string[]>;
  readonly listed: readonly { readonly target: string | null; readonly message: string }[];
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
    return { target: targets[0] ?? null, message: issue.message };
  });
  return { instances, listed };
}
