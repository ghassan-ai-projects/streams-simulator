import { record } from "../../shared/json";
import type { Composition } from "./composition";
import { reshapeInstance } from "./compositionEdits";
import type { ListedIssue } from "./issueTargets";
import type { Row } from "./rows";

function without(row: Row, key: string): Row {
  return Object.fromEntries(Object.entries(row).filter(([name]) => name !== key));
}

export function setPreset(doc: Composition, id: string, preset: string): Composition {
  return reshapeInstance(doc, id, (row) => (preset === "" ? without(row, "preset") : { ...row, preset }));
}

function withParameter(row: Row, name: string, value: number | string | null): Row {
  const kept = Object.entries(record(row.parameters)).filter(([each]) => each !== name);
  const parameters = Object.fromEntries(value === null ? kept : [...kept, [name, value]]);
  return Object.keys(parameters).length === 0 ? without(row, "parameters") : { ...row, parameters };
}

export function setParameter(doc: Composition, id: string, name: string, value: number | string | null): Composition {
  return reshapeInstance(doc, id, (row) => withParameter(row, name, value));
}

export function fieldIssues(listed: readonly ListedIssue[], entry: number, field: string): readonly string[] {
  const pointer =
    field === "preset" ? `/equipment/${String(entry)}/preset` : `/equipment/${String(entry)}/parameters/${field}`;
  return listed.filter((issue) => issue.pointer === pointer).map((issue) => issue.message);
}
