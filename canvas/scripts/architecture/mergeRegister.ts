export function mergeRegister(base: unknown, patch: unknown): unknown {
  if (!isObject(patch)) return patch;
  const kept = isObject(base) ? Object.entries(base).filter(([key]) => patch[key] !== null) : [];
  const merged = new Map<string, unknown>(kept);
  for (const [key, value] of Object.entries(patch)) {
    if (value !== null) merged.set(key, mergeRegister(merged.get(key), value));
  }
  return Object.fromEntries(merged);
}

function isObject(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}
