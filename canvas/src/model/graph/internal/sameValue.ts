type Kind = "set" | "list" | "map" | "fields";

function kindOf(value: object): Kind {
  if (value instanceof Set) return "set";
  if (Array.isArray(value)) return "list";
  return value instanceof Map ? "map" : "fields";
}

function sameSet(a: ReadonlySet<unknown>, b: ReadonlySet<unknown>): boolean {
  return a.size === b.size && [...a].every((item) => b.has(item));
}

function sameList(a: readonly unknown[], b: readonly unknown[]): boolean {
  return a.length === b.length && a.every((item, at) => sameValue(item, b[at]));
}

function sameFields(a: Readonly<Record<string, unknown>>, b: Readonly<Record<string, unknown>>): boolean {
  const keys = Object.keys(a);
  return keys.length === Object.keys(b).length && keys.every((key) => key in b && sameValue(a[key], b[key]));
}

function sameObjects(kind: Kind, a: object, b: object): boolean {
  if (kind === "set") return sameSet(a as ReadonlySet<unknown>, b as ReadonlySet<unknown>);
  if (kind === "list") return sameList(a as readonly unknown[], b as readonly unknown[]);
  return kind === "fields" && sameFields(a as Record<string, unknown>, b as Record<string, unknown>);
}

export function sameValue(a: unknown, b: unknown): boolean {
  if (Object.is(a, b)) return true;
  if (typeof a !== "object" || typeof b !== "object" || a === null || b === null) return false;
  const kind = kindOf(a);
  return kind === kindOf(b) && sameObjects(kind, a, b);
}
