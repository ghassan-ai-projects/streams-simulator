import type { Rect } from "./types";
import { centreOf } from "./viewport";

type Direction = "left" | "right" | "up" | "down";

const directions: Readonly<Record<string, Direction>> = {
  ArrowLeft: "left",
  ArrowRight: "right",
  ArrowUp: "up",
  ArrowDown: "down",
};

export function directionOf(key: string): Direction | null {
  return directions[key] ?? null;
}

function offset(from: Rect, to: Rect, direction: Direction): { readonly along: number; readonly across: number } {
  const a = centreOf(from);
  const b = centreOf(to);
  const dx = b.x - a.x;
  const dy = b.y - a.y;
  switch (direction) {
    case "left":
      return { along: -dx, across: Math.abs(dy) };
    case "right":
      return { along: dx, across: Math.abs(dy) };
    case "up":
      return { along: -dy, across: Math.abs(dx) };
    case "down":
      return { along: dy, across: Math.abs(dx) };
  }
}

export function nearest(fromId: string, direction: Direction, boxes: ReadonlyMap<string, Rect>): string | null {
  const from = boxes.get(fromId);
  if (from === undefined) return null;
  let best: { readonly id: string; readonly score: number } | null = null;
  for (const [id, box] of boxes) {
    const { along, across } = offset(from, box, direction);
    if (id === fromId || along <= 0 || across > along * 2) continue;
    const score = along + across * 2;
    if (best === null || score < best.score) best = { id, score };
  }
  return best?.id ?? null;
}

export function firstOf(boxes: ReadonlyMap<string, Rect>): string | null {
  let first: { readonly id: string; readonly box: Rect } | null = null;
  for (const [id, box] of boxes) {
    if (first === null || box.y < first.box.y || (box.y === first.box.y && box.x < first.box.x)) first = { id, box };
  }
  return first?.id ?? null;
}
