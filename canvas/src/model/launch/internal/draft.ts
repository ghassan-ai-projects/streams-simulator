import type { Plant } from "../../plant";
import { callArgs, effectorOf } from "../../plant";
import type { LaunchPlan } from "./launch";
import { minutesText } from "./minutes";
import type { PerturbationKind } from "./perturbation";
import type { DraftPerturbation } from "./perturbationDraft";
import { endOf, minutesWhole, perturbationDraftOf, plannedPerturbation } from "./perturbationDraft";

export type { DraftPerturbation } from "./perturbationDraft";

export interface DraftFault {
  readonly key: number;
  readonly entity: string;
  readonly fault: string;
  readonly atMinutes: string;
  readonly untilMinutes: string;
}

export interface DraftCall {
  readonly key: number;
  readonly entity: string;
  readonly effector: string;
  readonly atMinutes: string;
  readonly values: Readonly<Record<string, string>>;
}

export interface Draft {
  readonly name: string;
  readonly purpose: string;
  readonly seed: string;
  readonly duration: string;
  readonly unit: "minutes" | "hours";
  readonly explore: boolean;
  readonly pace: "watch" | "fast";
  readonly faults: readonly DraftFault[];
  readonly perturbations: readonly DraftPerturbation[];
  readonly calls: readonly DraftCall[];
  readonly nextKey: number;
}

export const firstDraft: Draft = {
  name: "",
  purpose: "",
  seed: "1",
  duration: "30",
  unit: "minutes",
  explore: true,
  pace: "watch",
  faults: [],
  perturbations: [],
  calls: [],
  nextKey: 1,
};

export function freshPlanOf(explore: boolean): LaunchPlan {
  const { seed, duration, pace } = firstDraft;
  return {
    seed: Number(seed),
    durationS: Number(duration) * 60,
    explore,
    pace,
    faults: [],
    perturbations: [],
    calls: [],
  };
}

const whole = (text: string, scale: number) => Math.round(Number(text.trim() === "" ? Number.NaN : text) * scale);

const authored = (name: "name" | "purpose", text: string) => (text.trim() === "" ? {} : { [name]: text.trim() });

export function planOf(draft: Draft, plant: Plant, kinds: readonly PerturbationKind[] = []): LaunchPlan {
  return {
    ...authored("name", draft.name),
    ...authored("purpose", draft.purpose),
    seed: Number(draft.seed.trim() === "" ? Number.NaN : draft.seed),
    durationS: whole(draft.duration, draft.unit === "hours" ? 3600 : 60),
    explore: draft.explore,
    pace: draft.pace,
    faults: draft.faults.map((fault) => ({
      entity: fault.entity,
      fault: fault.fault,
      atS: minutesWhole(fault.atMinutes),
      ...endOf(fault.untilMinutes),
    })),
    perturbations: draft.perturbations.map((perturbation) => plannedPerturbation(perturbation, kinds)),
    calls: draft.calls.map((call) => ({
      entity: call.entity,
      effector: call.effector,
      atS: minutesWhole(call.atMinutes),
      args: callArgs(effectorOf(plant, call.entity, call.effector), call.entity, call.values),
    })),
  };
}

export function draftOf(plan: LaunchPlan): Draft {
  const hours = plan.durationS % 3600 === 0;
  const faults = plan.faults.length;
  const calls = plan.calls.length;
  return {
    name: plan.name ?? "",
    purpose: plan.purpose ?? "",
    seed: String(plan.seed),
    duration: hours ? String(plan.durationS / 3600) : minutesText(plan.durationS),
    unit: hours ? "hours" : "minutes",
    explore: plan.explore,
    pace: plan.pace ?? "watch",
    faults: plan.faults.map((fault, index) => ({
      key: index + 1,
      entity: fault.entity,
      fault: fault.fault,
      atMinutes: minutesText(fault.atS),
      untilMinutes: fault.untilS === undefined ? "" : minutesText(fault.untilS),
    })),
    calls: plan.calls.map((call, index) => ({
      key: faults + index + 1,
      entity: call.entity,
      effector: call.effector,
      atMinutes: minutesText(call.atS),
      values: Object.fromEntries(Object.entries(call.args).map(([name, value]) => [name, String(value)])),
    })),
    perturbations: plan.perturbations.map((planned, index) => perturbationDraftOf(planned, faults + calls + index + 1)),
    nextKey: faults + calls + plan.perturbations.length + 1,
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
