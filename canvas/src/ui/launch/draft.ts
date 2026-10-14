import type { LaunchPlan } from "../../model/launch";

export interface DraftFault {
  readonly key: number;
  readonly entity: string;
  readonly fault: string;
  readonly atMinutes: string;
}

export interface Draft {
  readonly seed: string;
  readonly duration: string;
  readonly unit: "minutes" | "hours";
  readonly explore: boolean;
  readonly faults: readonly DraftFault[];
  readonly nextKey: number;
}

export const firstDraft: Draft = { seed: "1", duration: "30", unit: "minutes", explore: true, faults: [], nextKey: 1 };

const whole = (text: string, scale: number) => Math.round(Number(text.trim() === "" ? Number.NaN : text) * scale);

export function planOf(draft: Draft): LaunchPlan {
  return {
    seed: Number(draft.seed.trim() === "" ? Number.NaN : draft.seed),
    durationS: whole(draft.duration, draft.unit === "hours" ? 3600 : 60),
    explore: draft.explore,
    faults: draft.faults.map((fault) => ({
      entity: fault.entity,
      fault: fault.fault,
      atS: whole(fault.atMinutes, 60),
    })),
    calls: [],
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
    nextKey: plan.faults.length + 1,
  };
}

export function pickedOnGraph(draft: Draft, entity: string): Draft {
  const last = draft.faults.at(-1);
  if (last?.entity !== "") return draft;
  return { ...draft, faults: draft.faults.map((fault) => (fault === last ? { ...fault, entity, fault: "" } : fault)) };
}
