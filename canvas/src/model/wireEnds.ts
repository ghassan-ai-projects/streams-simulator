import type { Rect } from "./viewport";
import type { Point, Side } from "./wireRoute";
import { wireClearance } from "./wireRoute";

export interface End {
  readonly id: string;
  readonly role: "from" | "to";
  readonly point: Point;
  readonly box: Rect;
  readonly side: Side;
  readonly slot: string;
  readonly other: Point;
}

const endGap = 6;

export function runsAcross(end: End): boolean {
  return end.side === "up" || end.side === "down";
}

function axisOf(end: End, point: Point): number {
  return runsAcross(end) ? point.x : point.y;
}

function separated(desired: readonly number[], gap: number): number[] {
  const blocks: { sum: number; count: number }[] = [];
  desired.forEach((value, index) => {
    blocks.push({ sum: value - index * gap, count: 1 });
    for (;;) {
      const last = blocks[blocks.length - 1];
      const before = blocks[blocks.length - 2];
      if (last === undefined || before === undefined || before.sum / before.count <= last.sum / last.count) break;
      blocks.splice(-2, 2, { sum: before.sum + last.sum, count: before.count + last.count });
    }
  });
  return blocks.flatMap((block) => Array<number>(block.count).fill(block.sum / block.count)).map((v, i) => v + i * gap);
}

function within(values: readonly number[], low: number, high: number): number[] {
  const first = values[0] ?? 0;
  const last = values[values.length - 1] ?? first;
  const up = Math.max(0, low - first);
  const down = Math.min(0, high - (last + up));
  return values.map((value) => value + up + down);
}

function places(desired: readonly number[], gap: number, low: number, high: number): number[] {
  const count = desired.length;
  const room = Math.max(0, high - low);
  const fitted = count > 1 ? Math.min(gap, room / (count - 1)) : gap;
  return within(separated(desired, fitted), low, high);
}

function slotsOf(group: readonly End[]): End[][] {
  const slots = new Map<string, End[]>();
  for (const end of group) slots.set(end.slot, [...(slots.get(end.slot) ?? []), end]);
  const first = (ends: readonly End[]) => ends[0] ?? group[0];
  const order = (ends: readonly End[]) => {
    const end = first(ends);
    return end === undefined ? [0, 0] : [axisOf(end, end.point), axisOf(end, end.other)];
  };
  return [...slots.values()].sort((a, b) => {
    const [pa = 0, oa = 0] = order(a);
    const [pb = 0, ob = 0] = order(b);
    return pa - pb || oa - ob || (a[0]?.slot ?? "").localeCompare(b[0]?.slot ?? "");
  });
}

function boxKey(end: End): string {
  const { x, y, width, height } = end.box;
  return `${end.side}:${String(x)}:${String(y)}:${String(width)}:${String(height)}`;
}

function bounds(end: End): readonly [number, number] {
  const { box } = end;
  const start = runsAcross(end) ? box.x : box.y;
  const extent = runsAcross(end) ? box.width : box.height;
  return [start + wireClearance.lane, start + extent - wireClearance.lane];
}

function placeSide(group: readonly End[], pixel: number, placed: Map<string, Point>): void {
  const slots = slotsOf(group);
  const lead = slots[0]?.[0];
  if (lead === undefined) return;
  const [low, high] = bounds(lead);
  const desired = slots.map((slot) => axisOf(lead, slot[0]?.point ?? lead.point));
  const at = places(desired, endGap * pixel, low, high);
  slots.forEach((slot, index) => {
    for (const end of slot) {
      const value = at[index] ?? 0;
      placed.set(
        `${end.id}:${end.role}`,
        runsAcross(end) ? { x: value, y: end.point.y } : { x: end.point.x, y: value },
      );
    }
  });
}

export function spreadEnds(ends: readonly End[], pixel: number): ReadonlyMap<string, Point> {
  const groups = new Map<string, End[]>();
  for (const end of ends) groups.set(boxKey(end), [...(groups.get(boxKey(end)) ?? []), end]);
  const placed = new Map<string, Point>();
  for (const group of groups.values()) placeSide(group, pixel, placed);
  return placed;
}
