import type { Report } from "./report";
import { callWords, faultNames } from "./report";
import { clockOf } from "../../shared/duration";
import { counted, humanise } from "../../shared/words";
import type { BindingNote } from "../../launch";
import { effectorShownS } from "../../launch";

export function runClock(ns: number, startNs: number): string {
  return clockOf((ns - startNs) / 1e9);
}

export type MarkKind =
  "fault" | "cleared" | "disturbance" | "alarm" | "effector" | "binding" | "command" | "connection";

export const markWords: Readonly<Record<MarkKind, string>> = {
  command: "Command evidence",
  connection: "Connection change",
  fault: "Fault starts",
  cleared: "Disturbance ends",
  disturbance: "Other disturbance",
  alarm: "Alarm",
  effector: "Effector call",
  binding: "Gateway binding",
};

export interface TimedBinding {
  readonly atNs: number;
  readonly binding: BindingNote;
}

function bindingMarks(timed: TimedBinding | null): readonly EventMark[] {
  if (timed === null) return [];
  const { matched, refusal } = timed.binding;
  return [
    {
      atNs: timed.atNs,
      label: matched ? "Gateway matched the plan" : "Gateway refused",
      raw: matched ? "binding matched" : `binding refused · ${refusal}`,
      kind: "binding",
      tone: "other",
      entity: "",
      untilNs: null,
    },
  ];
}

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

export function eventMarks(report: Report | null, binding: TimedBinding | null = null): readonly EventMark[] {
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
  const all = [...disturbances, ...endMarks(report), ...calls, ...alarms, ...bindingMarks(binding)];
  return all.sort((a, b) => a.atNs - b.atNs);
}

export function previousEvent(marks: readonly EventMark[], atNs: number): number | null {
  return marks.filter((mark) => mark.atNs < atNs).at(-1)?.atNs ?? null;
}

export function nextEvent(marks: readonly EventMark[], atNs: number): number | null {
  return marks.find((mark) => mark.atNs > atNs)?.atNs ?? null;
}

export function nextAlarm(marks: readonly EventMark[], atNs: number): EventMark | null {
  const alarms = marks.filter((mark) => mark.tone === "alarm");
  return alarms.find((mark) => mark.atNs > atNs) ?? alarms[0] ?? null;
}

export function faultedAt(report: Report | null, atNs: number): ReadonlyMap<string, readonly string[]> {
  if (report === null) return new Map();
  return faultNames(report, (d) => d.onsetNs <= atNs && !(d.clearedNs !== null && d.clearedNs <= atNs));
}

export function effectorAt(report: Report | null, atNs: number): ReadonlyMap<string, string> {
  const latest = new Map<string, { readonly at: number; readonly words: string }>();
  for (const call of report?.effectorCalls ?? []) {
    const known = latest.get(call.entity);
    const live = call.atNs <= atNs && atNs <= call.atNs + effectorShownS * 1e9;
    if (live && (known === undefined || call.atNs >= known.at)) {
      latest.set(call.entity, { at: call.atNs, words: callWords(call) });
    }
  }
  return new Map([...latest].map(([entity, call]) => [entity, call.words] as const));
}

export function kindsOf(marks: readonly EventMark[]): readonly MarkKind[] {
  const present = new Set(marks.map((mark) => mark.kind));
  return (Object.keys(markWords) as MarkKind[]).filter((kind) => present.has(kind));
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
  const tail = calls === 0 || only ? "" : ` · ${counted(calls, "effector call")}`;
  return `First: ${first.words} at ${runClock(first.atNs, startNs)}${tail}`;
}
