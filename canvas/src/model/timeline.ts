import type { Report } from "./report";
import { humanise } from "./report";

export interface EventMark {
  readonly atNs: number;
  readonly label: string;
  readonly raw: string;
  readonly tone: "fault" | "other";
  readonly entity: string;
}

export function eventMarks(report: Report | null): readonly EventMark[] {
  if (report === null) return [];
  const disturbances = report.disturbances.map((d): EventMark => ({
    atNs: d.onsetNs,
    label: `${d.target} ${humanise(d.name)}${d.action === "clear" ? " cleared" : ""}`,
    raw: `${d.kind} ${d.action} ${d.target} ${d.name}`,
    tone: d.kind === "fault" && d.action !== "clear" ? "fault" : "other",
    entity: d.target,
  }));
  const calls = report.effectorCalls.map((c): EventMark => ({
    atNs: c.atNs,
    label: `${c.entity} ${humanise(c.effector)}`,
    raw: `${c.entity} ${c.effector} · ${c.outcome}`,
    tone: "other",
    entity: c.entity,
  }));
  return [...disturbances, ...calls].sort((a, b) => a.atNs - b.atNs);
}

export function previousEvent(marks: readonly EventMark[], atNs: number): number | null {
  const earlier = marks.filter((mark) => mark.atNs < atNs);
  return earlier.length === 0 ? null : (earlier[earlier.length - 1]?.atNs ?? null);
}

export function nextEvent(marks: readonly EventMark[], atNs: number): number | null {
  return marks.find((mark) => mark.atNs > atNs)?.atNs ?? null;
}

export function faultedAt(report: Report | null, atNs: number): ReadonlyMap<string, string> {
  const active = new Map<string, string>();
  for (const d of report?.disturbances ?? []) {
    if (d.kind !== "fault" || d.onsetNs > atNs) continue;
    if (d.action === "clear") active.delete(d.target);
    else active.set(d.target, humanise(d.name));
  }
  return active;
}

export interface ChipPlace {
  readonly left: number;
  readonly row: number;
}

export function placeChips(
  positions: readonly number[],
  widths: readonly number[],
  lane: { readonly width: number; readonly rows: number; readonly gap: number },
): readonly ChipPlace[] {
  const ends: number[] = [];
  return positions.map((position, index) => {
    const width = widths[index] ?? 0;
    const left = Math.max(0, Math.min(position, lane.width - width));
    const free = ends.findIndex((end) => end <= left);
    const row = free >= 0 ? free : ends.length < lane.rows ? ends.length : -1;
    if (row >= 0) ends[row] = left + width + lane.gap;
    return { left, row };
  });
}
