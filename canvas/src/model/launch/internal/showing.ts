import type { ShowingMarks } from "../../shared/marks";
import { humanise } from "../../shared/words";
import type { CallOutcome, LaunchPlan, PlannedFault } from "./launch";

export const effectorShownS = 60;
const effectorHeldMs = 3000;

function callWords(call: CallOutcome): string {
  const name = humanise(call.effector);
  return call.accepted ? `${name} · accepted` : `${name} · refused: ${call.reason}`;
}

const noFaults: readonly PlannedFault[] = [];

function faultsAt(faults: readonly PlannedFault[], atS: number): Map<string, readonly string[]> {
  const faulted = new Map<string, readonly string[]>();
  for (const planned of faults) {
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

export function shownFaults(plan: LaunchPlan, injected: readonly PlannedFault[] = noFaults): readonly PlannedFault[] {
  if (!plan.explore) return noFaults;
  return [...plan.faults, ...injected].filter((planned) => planned.entity !== "" && planned.fault !== "");
}

export function showingMarks(
  plan: LaunchPlan,
  calls: readonly CallOutcome[],
  atS: number,
  atMs: number,
  injected: readonly PlannedFault[] = noFaults,
): ShowingMarks {
  return { faults: faultsAt(shownFaults(plan, injected), atS), effectors: callsAt(calls, atS, atMs) };
}
