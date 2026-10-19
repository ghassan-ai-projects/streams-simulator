import { isRecord } from "../../../model/shared/json";

export function choice<T extends string>(value: unknown, choices: readonly T[]): value is T {
  return typeof value === "string" && choices.includes(value as T);
}

export function pick<T extends string>(value: unknown, choices: readonly T[]): T | null {
  return choices.find((candidate) => candidate === value) ?? null;
}

export function rows<T>(value: unknown, decode: (row: Record<string, unknown>) => T | null): readonly T[] | null {
  if (!Array.isArray(value)) return null;
  const result = value.map((row) => (isRecord(row) ? decode(row) : null));
  return result.includes(null) ? null : (result as T[]);
}

export function names(value: unknown): readonly string[] | null {
  if (value === undefined) return [];
  if (!Array.isArray(value) || value.some((entry) => typeof entry !== "string")) return null;
  return value as string[];
}
