import type { Report } from "./report";
import { immediateWords } from "./report";
import { runClock } from "./timeline";

export interface ActionStory {
  readonly atNs: number;
  readonly entity: string;
  readonly effector: string;
  readonly requestedAt: string;
  readonly args: readonly { readonly name: string; readonly value: string }[];
  readonly outcome: string;
  readonly applied: boolean;
  readonly source: string | null;
}

interface Picked {
  readonly atNs: number;
  readonly entity: string;
}

export function storyOf(report: Report, picked: Picked | null): ActionStory | null {
  const call = report.effectorCalls.find((each) => each.atNs === picked?.atNs && each.entity === picked.entity);
  if (call === undefined) return null;
  return {
    atNs: call.atNs,
    entity: call.entity,
    effector: call.effector,
    requestedAt: runClock(call.atNs, report.startNs),
    args: Object.entries(call.args).map(([name, value]) => ({ name, value: String(value) })),
    outcome: call.outcome,
    applied: call.applied,
    source: call.immediate ? immediateWords : null,
  };
}

const minuteNs = 60e9;

export function storyWindow(report: Pick<Report, "startNs" | "endNs">, atNs: number, side: "before" | "after"): number {
  if (side === "before") return Math.max(report.startNs, atNs - minuteNs);
  const later = atNs + minuteNs;
  return report.endNs > report.startNs ? Math.min(report.endNs, later) : later;
}
