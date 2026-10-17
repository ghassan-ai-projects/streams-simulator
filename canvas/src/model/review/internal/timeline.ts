import type { Report } from "./report";
import { clockOf } from "../../shared/duration";
import { humanise } from "../../shared/words";

export function runClock(ns: number, startNs: number): string {
  return clockOf((ns - startNs) / 1e9);
}

export type MarkKind = "fault" | "cleared" | "disturbance" | "alarm" | "effector";

export const markWords: Readonly<Record<MarkKind, string>> = {
  fault: "Fault starts",
  cleared: "Disturbance ends",
  disturbance: "Other disturbance",
  alarm: "Alarm",
  effector: "Effector call",
};

export interface EventMark {
  readonly atNs: number;
  readonly label: string;
  readonly raw: string;
  readonly kind: MarkKind;
  readonly tone: "fault" | "alarm" | "other";
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
            kind: "cleared",
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
    kind: d.kind === "fault" ? "fault" : "disturbance",
    tone: d.kind === "fault" ? "fault" : "other",
    entity: d.target,
    untilNs: d.clearedNs,
  }));
  const calls = report.effectorCalls.map((c): EventMark => ({
    atNs: c.atNs,
    label: `${c.entity} ${humanise(c.effector)}`,
    raw: `${c.entity} ${c.effector} · ${c.outcome}`,
    kind: "effector",
    tone: "other",
    entity: c.entity,
    untilNs: null,
  }));
  const alarms = report.alarmActivations.map((a): EventMark => ({
    atNs: a.atNs,
    label: `${a.entity} ${humanise(a.state)} activated`,
    raw: `alarm activation ${a.entity} ${a.state} · ${a.priority}`,
    kind: "alarm",
    tone: "alarm",
    entity: a.entity,
    untilNs: null,
  }));
  return [...disturbances, ...endMarks(report), ...calls, ...alarms].sort((a, b) => a.atNs - b.atNs);
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

const effectorShownNs = 60e9;

export function effectorAt(report: Report | null, atNs: number): ReadonlyMap<string, string> {
  const latest = new Map<string, { readonly at: number; readonly words: string }>();
  for (const call of report?.effectorCalls ?? []) {
    const known = latest.get(call.entity);
    const live = call.atNs <= atNs && atNs <= call.atNs + effectorShownNs;
    if (live && (known === undefined || call.atNs >= known.at)) {
      latest.set(call.entity, { at: call.atNs, words: `${humanise(call.effector)} · ${call.outcome}` });
    }
  }
  return new Map([...latest].map(([entity, call]) => [entity, call.words] as const));
}

export function kindsOf(marks: readonly EventMark[]): readonly MarkKind[] {
  const present = new Set(marks.map((mark) => mark.kind));
  return (Object.keys(markWords) as MarkKind[]).filter((kind) => present.has(kind));
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
    const left = Math.max(0, Math.min(position - width / 2, lane.width - width));
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

interface Beginning {
  readonly atNs: number;
  readonly words: string;
}

function beginnings(report: Report): readonly Beginning[] {
  const disturbances = report.disturbances.map((d) => ({
    atNs: d.onsetNs,
    words: `${d.kind === "fault" ? "fault" : "disturbance"} · ${humanise(d.name)} on ${d.target}`,
  }));
  const calls = report.effectorCalls.map((c) => ({
    atNs: c.atNs,
    words: `effector call · ${humanise(c.effector)} on ${c.entity}`,
  }));
  const alarms = report.alarmActivations.map((a) => ({
    atNs: a.atNs,
    words: `alarm · ${humanise(a.state)} on ${a.entity}`,
  }));
  return [...disturbances, ...calls, ...alarms].sort((a, b) => a.atNs - b.atNs);
}

export function firstEventLine(report: Report | null, startNs: number): string | null {
  if (report === null) return null;
  const first = beginnings(report)[0];
  if (first === undefined) return "No faults or effector calls in this run";
  const calls = report.effectorCalls.length;
  const only = calls === 1 && first.words.startsWith("effector call");
  const tail = calls === 0 || only ? "" : ` · ${String(calls)} effector call${calls === 1 ? "" : "s"}`;
  return `First: ${first.words} at ${runClock(first.atNs, startNs)}${tail}`;
}
