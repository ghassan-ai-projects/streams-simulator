import { text } from "../../shared/json";
import { componentOf, shortKind } from "./composition";

type Json = Readonly<Record<string, unknown>>;

const longestKind = 20;

export function kindOf(entry: Json): string {
  const named = text(entry.as);
  return named === "" ? shortKind(componentOf(String(entry.use))) : named;
}

export function freeKind(base: string, taken: ReadonlySet<string>): string {
  for (let n = 2; ; n += 1) {
    const suffix = `_${String(n)}`;
    const candidate = `${base.slice(0, longestKind - suffix.length).replace(/_+$/, "")}${suffix}`;
    if (!taken.has(candidate)) return candidate;
  }
}

const placement = new Set(["names", "area", "as"]);

function sharedFields(entry: Json): string {
  const shared = Object.entries(entry).filter(([key]) => !placement.has(key));
  return JSON.stringify(shared.sort(([a], [b]) => a.localeCompare(b)));
}

function named(entry: Json): boolean {
  return Array.isArray(entry.names) && entry.names.length > 0 && entry.id_template === undefined;
}

export function joinable(existing: Json, candidate: Json): boolean {
  return named(existing) && existing.area === candidate.area && sharedFields(existing) === sharedFields(candidate);
}
