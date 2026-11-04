import type { ParameterSpec, PresetOption } from "../../../model/plant";
import { isRecord, list, record, text } from "../../../model/shared/json";
import { finiteOrNull } from "./transport/decodeKit";

function defaultOf(value: unknown): number | string | null {
  return typeof value === "number" || typeof value === "string" ? value : null;
}

function parameterOf(raw: unknown): ParameterSpec | null {
  const row = record(raw);
  const type = row.type === "enum" || row.type === "integer" || row.type === "number" ? row.type : null;
  if (type === null || text(row.name) === "") return null;
  return {
    name: text(row.name),
    title: text(row.title) || text(row.name),
    unit: text(row.unit),
    type,
    min: finiteOrNull(row.min),
    max: finiteOrNull(row.max),
    values: list(row.values).map(text),
    default: defaultOf(row.default),
  };
}

export function parametersOf(document: Record<string, unknown>): readonly ParameterSpec[] {
  return list(document.parameters).flatMap((raw) => {
    const parameter = parameterOf(raw);
    return parameter === null ? [] : [parameter];
  });
}

export function presetsOf(document: Record<string, unknown>): readonly PresetOption[] {
  if (!isRecord(document.presets)) return [];
  return Object.entries(document.presets).map(([name, raw]) => ({ name, title: text(record(raw).title) || name }));
}
