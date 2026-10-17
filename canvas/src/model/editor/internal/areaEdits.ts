import { record } from "../../shared/json";
import { counted } from "../../shared/words";
import type { Composition } from "./composition";
import type { Row } from "./rows";
import { equipmentRows, namesOf, rowsOf } from "./rows";

function areaRows(doc: Composition): readonly Row[] {
  return rowsOf(doc.areas);
}

export function areaNameProblem(doc: Composition, name: string, current: string | null): string | null {
  if (name === "") return "Give the area a name.";
  const taken = name !== current && areaRows(doc).some((area) => area.name === name);
  return taken ? `There is already an area named "${name}".` : null;
}

export function addArea(doc: Composition, name: string, parent: string): Composition {
  if (areaNameProblem(doc, name, null) !== null) return doc;
  return { ...doc, areas: [...areaRows(doc), { name, ...(parent === "" ? {} : { parent }) }] };
}

function renamedArea(area: Row, from: string, to: string): Row {
  const name = area.name === from ? to : area.name;
  return area.parent === from ? { ...area, name, parent: to } : { ...area, name };
}

export function renameArea(doc: Composition, from: string, to: string): Composition {
  if (from === to || areaNameProblem(doc, to, from) !== null) return doc;
  return {
    ...doc,
    areas: areaRows(doc).map((area) => renamedArea(area, from, to)),
    equipment: equipmentRows(doc).map((entry) => (entry.area === from ? { ...entry, area: to } : entry)),
  };
}

function instanceCount(entry: Row): number {
  const names = namesOf(entry);
  const given = record(entry.count).default;
  return names.length > 0 ? names.length : typeof given === "number" ? given : 1;
}

export function areaInUse(doc: Composition, name: string): string | null {
  const holding = equipmentRows(doc).filter((entry) => entry.area === name);
  const inside = areaRows(doc).filter((area) => area.parent === name).length;
  const instances = holding.reduce((sum, entry) => sum + instanceCount(entry), 0);
  if (instances > 0) return `Move the ${counted(instances, "instance")} out of it first.`;
  if (inside > 0) return `Delete the ${counted(inside, "area")} inside it first.`;
  return null;
}

export function removeArea(doc: Composition, name: string): Composition {
  if (areaInUse(doc, name) !== null) return doc;
  return { ...doc, areas: areaRows(doc).filter((area) => area.name !== name) };
}
