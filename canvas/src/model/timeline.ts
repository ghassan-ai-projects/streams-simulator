import type { Report } from "./report";
import { humanise } from "./report";

export interface EventMark {
  readonly atNs: number;
  readonly label: string;
  readonly raw: string;
  readonly tone: "fault" | "other";
  readonly entity: string;
  readonly untilNs: number | null;
}

export function activeAt(mark: EventMark, atNs: number): boolean {
  return mark.tone === "fault" && mark.atNs <= atNs && (mark.untilNs === null || atNs < mark.untilNs);
}

const endings: Readonly<Record<string, string>> = {
  command: "cleared",
  pulse_end: "ended",
  responder: "cleared by its responder",
};

function endMarks(report: Report): readonly EventMark[] {
  return report.disturbances.flatMap((d): EventMark[] =>
    d.clearedNs === null
      ? []
      : [
          {
            atNs: d.clearedNs,
            label: `${d.target} ${humanise(d.name)} ${endings[d.clearedBy] ?? "cleared"}`,
            raw: `${d.kind} ${d.target} ${d.name} · ${d.clearedBy}`,
            tone: "other",
            entity: d.target,
            untilNs: null,
          },
        ],
  );
}

export function eventMarks(report: Report | null): readonly EventMark[] {
  if (report === null) return [];
  const disturbances = report.disturbances.map((d): EventMark => ({
    atNs: d.onsetNs,
    label: `${d.target} ${humanise(d.name)}`,
    raw: `${d.kind} ${d.action} ${d.target} ${d.name}`,
    tone: d.kind === "fault" ? "fault" : "other",
    entity: d.target,
    untilNs: d.clearedNs,
  }));
  const calls = report.effectorCalls.map((c): EventMark => ({
    atNs: c.atNs,
    label: `${c.entity} ${humanise(c.effector)}`,
    raw: `${c.entity} ${c.effector} · ${c.outcome}`,
    tone: "other",
    entity: c.entity,
    untilNs: null,
  }));
  return [...disturbances, ...endMarks(report), ...calls].sort((a, b) => a.atNs - b.atNs);
}

export function previousEvent(marks: readonly EventMark[], atNs: number): number | null {
  const earlier = marks.filter((mark) => mark.atNs < atNs);
  return earlier.length === 0 ? null : (earlier[earlier.length - 1]?.atNs ?? null);
}

export function nextEvent(marks: readonly EventMark[], atNs: number): number | null {
  return marks.find((mark) => mark.atNs > atNs)?.atNs ?? null;
}

export function faultedAt(report: Report | null, atNs: number): ReadonlyMap<string, readonly string[]> {
  const active = new Map<string, string[]>();
  for (const d of report?.disturbances ?? []) {
    const ended = d.clearedNs !== null && d.clearedNs <= atNs;
    if (d.kind === "fault" && d.onsetNs <= atNs && !ended)
      active.set(d.target, [...(active.get(d.target) ?? []), humanise(d.name)]);
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

const tickSteps = [600e9, 900e9, 1800e9, 3600e9, 7200e9, 10_800e9, 21_600e9, 43_200e9, 86_400e9];

export function timeTicks(startNs: number, endNs: number, most = 9): readonly number[] {
  const span = endNs - startNs;
  if (span <= 0) return [];
  const step = tickSteps.find((candidate) => span / candidate <= most) ?? span;
  const first = Math.ceil(startNs / step) * step;
  return Array.from({ length: Math.floor((endNs - first) / step) + 1 }, (_, index) => first + index * step);
}
