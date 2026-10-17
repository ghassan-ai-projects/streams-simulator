import { text } from "../../shared/json";
import { endpointOf } from "../../plant";
import type { Composition, NewEquipment } from "./composition";
import { componentOf, shortKind } from "./composition";
import { freeKind, joinable, kindOf } from "./kindNames";
import type { Row } from "./rows";
import { connectionRows, equipmentRows, namesOf } from "./rows";

function holds(entry: Row, id: string): boolean {
  return namesOf(entry).includes(id) || (namesOf(entry).length === 0 && entry.id_template === id);
}

export function isTemplate(doc: Composition, id: string): boolean {
  return equipmentRows(doc).some((entry) => namesOf(entry).length === 0 && entry.id_template === id);
}

function withName(all: readonly Row[], at: number, name: string): readonly Row[] {
  return all.map((entry, index) => (index === at ? { ...entry, names: [...namesOf(entry), name] } : entry));
}

function withoutName(entry: Row, id: string): readonly Row[] {
  const names = namesOf(entry).filter((name) => name !== id);
  return names.length === 0 ? [] : [{ ...entry, names }];
}

function distinctKind(all: readonly Row[], entry: Row): Row {
  const kinds = new Set(all.map(kindOf));
  const base = shortKind(componentOf(text(entry.use)));
  if (!kinds.has(kindOf(entry))) return entry;
  return { ...entry, as: freeKind(base, kinds) };
}

function placed(all: readonly Row[], entry: Row): readonly Row[] {
  const at = all.findIndex((existing) => joinable(existing, entry));
  const name = namesOf(entry)[0];
  if (at >= 0 && name !== undefined) return withName(all, at, name);
  return [...all, distinctKind(all, entry)];
}

export function addEquipment(doc: Composition, added: NewEquipment): Composition {
  const entry: Row = {
    use: added.use,
    ...(added.preset === "" ? {} : { preset: added.preset }),
    names: [added.name],
    area: added.area,
  };
  return { ...doc, equipment: placed(equipmentRows(doc), entry) };
}

function renamed(endpoint: string, from: string, to: string): string {
  const { id, port } = endpointOf(endpoint);
  return id === from ? `${to}.${port}` : endpoint;
}

export function renameInstance(doc: Composition, from: string, to: string): Composition {
  return {
    ...doc,
    equipment: equipmentRows(doc).map((entry) =>
      namesOf(entry).includes(from)
        ? { ...entry, names: namesOf(entry).map((name) => (name === from ? to : name)) }
        : entry,
    ),
    connections: connectionRows(doc).map((wire) => ({
      ...wire,
      from: renamed(text(wire.from), from, to),
      to: renamed(text(wire.to), from, to),
    })),
  };
}

function touches(wire: Row, id: string): boolean {
  return endpointOf(wire.from).id === id || endpointOf(wire.to).id === id;
}

export function removeInstance(doc: Composition, id: string): Composition {
  const equipment = equipmentRows(doc).flatMap((entry) => (holds(entry, id) ? withoutName(entry, id) : [entry]));
  return { ...doc, equipment, connections: connectionRows(doc).filter((wire) => !touches(wire, id)) };
}

export function moveInstance(doc: Composition, id: string, area: string): Composition {
  const all = equipmentRows(doc);
  const source = all.find((entry) => holds(entry, id));
  if (source === undefined) return doc;
  const moved = { ...source, names: [id], area };
  const others = all.filter((entry) => entry !== source);
  if (namesOf(source).length <= 1 && !others.some((entry) => joinable(entry, moved))) {
    return { ...doc, equipment: all.map((entry) => (entry === source ? moved : entry)) };
  }
  const rest = all.flatMap((entry) => (entry === source ? withoutName(entry, id) : [entry]));
  return { ...doc, equipment: placed(rest, moved) };
}

export function connect(doc: Composition, from: string, to: string): Composition {
  if (connectionRows(doc).some((wire) => wire.from === from && wire.to === to)) return doc;
  return { ...doc, connections: [...connectionRows(doc), { from, to }] };
}

export function disconnect(doc: Composition, index: number): Composition {
  return { ...doc, connections: connectionRows(doc).filter((_, at) => at !== index) };
}
