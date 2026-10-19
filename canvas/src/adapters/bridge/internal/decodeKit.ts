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

export interface Given<T> {
  readonly value: T;
}

export function optional<T extends string>(value: unknown, choices: readonly T[]): Given<T | null> | null {
  if (value === undefined) return { value: null };
  const found = pick(value, choices);
  return found === null ? null : { value: found };
}

export function optionalTexts<K extends string>(
  row: Record<string, unknown>,
  keys: readonly K[],
): Record<K, string> | null {
  if (keys.some((key) => row[key] !== undefined && typeof row[key] !== "string")) return null;
  return Object.fromEntries(keys.map((key) => [key, typeof row[key] === "string" ? row[key] : ""])) as Record<
    K,
    string
  >;
}

export function requiredTexts<K extends string>(
  row: Record<string, unknown>,
  keys: readonly K[],
): Record<K, string> | null {
  const found = optionalTexts(row, keys);
  return found !== null && keys.every((key) => found[key] !== "") ? found : null;
}
