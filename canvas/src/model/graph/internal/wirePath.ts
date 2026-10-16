import type { Point } from "./wireRoute";

function towards(from: Point, to: Point, distance: number): Point {
  const length = Math.hypot(to.x - from.x, to.y - from.y);
  if (length === 0) return from;
  return { x: from.x + ((to.x - from.x) / length) * distance, y: from.y + ((to.y - from.y) / length) * distance };
}

function corner(before: Point, at: Point, after: Point, radius: number): string {
  const room = Math.hypot(at.x - before.x, at.y - before.y) / 2;
  const out = Math.hypot(after.x - at.x, after.y - at.y) / 2;
  const entry = towards(at, before, Math.min(radius, room, out));
  const exit = towards(at, after, Math.min(radius, room, out));
  return ` L ${String(entry.x)} ${String(entry.y)} Q ${String(at.x)} ${String(at.y)} ${String(exit.x)} ${String(exit.y)}`;
}

export function wirePath(points: readonly Point[], radius = 6): string {
  const [first] = points;
  if (first === undefined) return "";
  let path = `M ${String(first.x)} ${String(first.y)}`;
  for (let index = 1; index < points.length; index += 1) {
    const at = points[index] ?? first;
    const before = points[index - 1] ?? first;
    const after = points[index + 1];
    path += after === undefined ? ` L ${String(at.x)} ${String(at.y)}` : corner(before, at, after, radius);
  }
  return path;
}

export function labelPoint(points: readonly Point[]): Point {
  let best: { readonly length: number; readonly from: Point; readonly to: Point } | null = null;
  points.slice(1).forEach((to, index) => {
    const from = points[index] ?? to;
    const length = Math.hypot(to.x - from.x, to.y - from.y);
    if (best === null || length > best.length) best = { length, from, to };
  });
  const found = best as { readonly from: Point; readonly to: Point } | null;
  if (found === null) return points[0] ?? { x: 0, y: 0 };
  return { x: (found.from.x + found.to.x) / 2, y: (found.from.y + found.to.y) / 2 };
}
