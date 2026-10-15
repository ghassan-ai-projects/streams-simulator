import type { CallOutcome, LaunchPlan } from "./launch";
import { hallOf } from "./plantView";
import { humanise } from "./report";

export interface PlanMark {
  readonly status: "fault" | null;
  readonly count: number;
  readonly fault: string;
  readonly effector: string;
}

const effectorShownS = 60;
const effectorHeldMs = 3000;

function callWords(call: CallOutcome): string {
  const name = humanise(call.effector);
  return call.accepted ? `${name} · accepted` : `${name} · refused: ${call.reason}`;
}

function faultsAt(plan: LaunchPlan, atS: number): Map<string, string[]> {
  const faulted = new Map<string, string[]>();
  for (const planned of plan.faults) {
    if (planned.entity === "" || planned.fault === "" || planned.atS > atS) continue;
    faulted.set(planned.entity, [...(faulted.get(planned.entity) ?? []), humanise(planned.fault)]);
  }
  return faulted;
}

function callShown(call: CallOutcome, atS: number, atMs: number): boolean {
  if (call.atS > atS) return false;
  return atS < call.atS + effectorShownS || atMs - call.atMs < effectorHeldMs;
}

function callsAt(calls: readonly CallOutcome[], atS: number, atMs: number): Map<string, string> {
  const shown = new Map<string, string>();
  for (const call of calls) {
    if (callShown(call, atS, atMs)) shown.set(call.entity, callWords(call));
  }
  return shown;
}

function hallsOf(faulted: ReadonlyMap<string, readonly string[]>): Map<string, PlanMark> {
  const halls = new Map<string, number>();
  for (const entity of faulted.keys()) {
    const hall = hallOf(entity);
    if (hall !== null) halls.set(hall, (halls.get(hall) ?? 0) + 1);
  }
  const marks = new Map<string, PlanMark>();
  for (const [hall, count] of halls) marks.set(hall, { status: "fault", count, fault: "", effector: "" });
  return marks;
}

export function planMarks(
  plan: LaunchPlan,
  calls: readonly CallOutcome[],
  atS: number,
  atMs: number,
): ReadonlyMap<string, PlanMark> {
  const faulted = faultsAt(plan, atS);
  const effectors = callsAt(calls, atS, atMs);
  const marks = new Map<string, PlanMark>(hallsOf(faulted));
  for (const entity of new Set([...faulted.keys(), ...effectors.keys()])) {
    const causes = faulted.get(entity) ?? [];
    const status = causes.length > 0 ? "fault" : null;
    marks.set(entity, { status, count: 0, fault: causes.join(", "), effector: effectors.get(entity) ?? "" });
  }
  return marks;
}
