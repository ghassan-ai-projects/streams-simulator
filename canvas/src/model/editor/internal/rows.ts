import { list, record, text } from "../../shared/json";
import type { Composition } from "./composition";

export type Row = Readonly<Record<string, unknown>>;

export function rowsOf(value: unknown): readonly Row[] {
  return list(value).map(record);
}

export function equipmentRows(doc: Composition): readonly Row[] {
  return rowsOf(doc.equipment);
}

export function connectionRows(doc: Composition): readonly Row[] {
  return rowsOf(doc.connections);
}

export function namesOf(entry: Row): readonly string[] {
  return list(entry.names).map(text);
}
