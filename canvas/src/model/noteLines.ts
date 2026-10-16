import type { Report } from "./report";
import { humanise } from "./shared/words";

const charsPerLine = 16;

function linesOf(text: string): number {
  return text === "" ? 0 : Math.ceil(text.length / charsPerLine);
}

export function noteLinesOf(fault: string, effector: string): number {
  const cause = fault === "" ? "" : `Fault · ${fault}`;
  return Math.max(1, linesOf(cause) + linesOf(effector));
}

function faultsOf(report: Report): Map<string, string> {
  const named = new Map<string, string[]>();
  for (const disturbance of report.disturbances) {
    if (disturbance.kind !== "fault") continue;
    const names = named.get(disturbance.target) ?? [];
    named.set(disturbance.target, [...names, humanise(disturbance.name)]);
  }
  return new Map([...named].map(([entity, names]) => [entity, names.join(", ")] as const));
}

function callsOf(report: Report): Map<string, string> {
  const longest = new Map<string, string>();
  for (const call of report.effectorCalls) {
    const words = `${humanise(call.effector)} · ${call.outcome}`;
    if (words.length > (longest.get(call.entity) ?? "").length) longest.set(call.entity, words);
  }
  return longest;
}

export function noteLinesByEntity(report: Report | null): ReadonlyMap<string, number> {
  if (report === null) return new Map();
  const faults = faultsOf(report);
  const calls = callsOf(report);
  const entities = new Set([...faults.keys(), ...calls.keys()]);
  return new Map(
    [...entities].map((entity) => [entity, noteLinesOf(faults.get(entity) ?? "", calls.get(entity) ?? "")]),
  );
}
