import { hallOf } from "./graph";
import type { Report } from "./report";

interface Event {
  readonly atNs: number;
  readonly entity: string;
}

function eventsOf(report: Report): Event[] {
  const faults = report.disturbances
    .filter((disturbance) => disturbance.kind === "fault")
    .map((disturbance) => ({ atNs: disturbance.onsetNs, entity: disturbance.target }));
  const calls = report.effectorCalls.map((call) => ({ atNs: call.atNs, entity: call.entity }));
  return [...faults, ...calls].sort((a, b) => a.atNs - b.atNs);
}

function shownAs(entity: string, shown: ReadonlySet<string>): string | null {
  if (shown.has(entity)) return entity;
  const hall = hallOf(entity);
  return hall !== null && shown.has(hall) ? hall : null;
}

export function firstEventNode(report: Report | null, shown: ReadonlySet<string>): string | null {
  if (report === null) return null;
  for (const event of eventsOf(report)) {
    const node = shownAs(event.entity, shown);
    if (node !== null) return node;
  }
  return null;
}
