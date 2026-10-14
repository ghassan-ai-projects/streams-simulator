import { callArgs } from "../../model/effectors";
import type { LaunchPlan } from "../../model/launch";
import type { Plant } from "../../model/plant";
import { effectorOf } from "../../model/plant";

export interface DraftFault {
  readonly key: number;
  readonly entity: string;
  readonly fault: string;
  readonly atMinutes: string;
}

export interface DraftCall {
  readonly key: number;
  readonly entity: string;
  readonly effector: string;
  readonly atMinutes: string;
  readonly values: Readonly<Record<string, string>>;
}

export interface Draft {
  readonly seed: string;
  readonly duration: string;
  readonly unit: "minutes" | "hours";
  readonly explore: boolean;
  readonly faults: readonly DraftFault[];
  readonly calls: readonly DraftCall[];
  readonly nextKey: number;
}

export const firstDraft: Draft = {
  seed: "1",
  duration: "30",
  unit: "minutes",
  explore: true,
  faults: [],
  calls: [],
  nextKey: 1,
};

const whole = (text: string, scale: number) => Math.round(Number(text.trim() === "" ? Number.NaN : text) * scale);

export function planOf(draft: Draft, plant: Plant): LaunchPlan {
  return {
    seed: Number(draft.seed.trim() === "" ? Number.NaN : draft.seed),
    durationS: whole(draft.duration, draft.unit === "hours" ? 3600 : 60),
    explore: draft.explore,
    faults: draft.faults.map((fault) => ({
      entity: fault.entity,
      fault: fault.fault,
      atS: whole(fault.atMinutes, 60),
    })),
    calls: draft.calls.map((call) => ({
      entity: call.entity,
      effector: call.effector,
      atS: whole(call.atMinutes, 60),
      args: callArgs(effectorOf(plant, call.entity, call.effector), call.entity, call.values),
    })),
  };
}

const minutesOf = (seconds: number) => String(Number((seconds / 60).toFixed(2)));

export function draftOf(plan: LaunchPlan): Draft {
  const hours = plan.durationS % 3600 === 0;
  return {
    seed: String(plan.seed),
    duration: hours ? String(plan.durationS / 3600) : minutesOf(plan.durationS),
    unit: hours ? "hours" : "minutes",
    explore: plan.explore,
    faults: plan.faults.map((fault, index) => ({
      key: index + 1,
      entity: fault.entity,
      fault: fault.fault,
      atMinutes: minutesOf(fault.atS),
    })),
    calls: plan.calls.map((call, index) => ({
      key: plan.faults.length + index + 1,
      entity: call.entity,
      effector: call.effector,
      atMinutes: minutesOf(call.atS),
      values: Object.fromEntries(Object.entries(call.args).map(([name, value]) => [name, String(value)])),
    })),
    nextKey: plan.faults.length + plan.calls.length + 1,
  };
}

export function pickedOnGraph(draft: Draft, entity: string): Draft {
  const fault = draft.faults.at(-1);
  if (fault?.entity === "")
    return { ...draft, faults: draft.faults.map((f) => (f === fault ? { ...f, entity, fault: "" } : f)) };
  const call = draft.calls.at(-1);
  if (call?.entity !== "") return draft;
  return { ...draft, calls: draft.calls.map((c) => (c === call ? { ...c, entity, effector: "", values: {} } : c)) };
}
