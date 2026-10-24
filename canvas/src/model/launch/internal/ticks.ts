import { clockOf } from "../../shared/duration";
import { humanise } from "../../shared/words";
import type { LaunchPlan } from "./launch";

export interface ProgressTick {
  readonly atS: number;
  readonly call: boolean;
  readonly title: string;
}

function tickOf(entity: string, what: string, atS: number, call: boolean): ProgressTick {
  return { atS, call, title: `${entity} ${humanise(what)} at ${clockOf(atS)}` };
}

export function progressTicks(plan: LaunchPlan): readonly ProgressTick[] {
  const faults = plan.explore ? plan.faults.map((fault) => tickOf(fault.entity, fault.fault, fault.atS, false)) : [];
  const calls = plan.calls.map((call) => tickOf(call.entity, call.effector, call.atS, true));
  return [...faults, ...calls];
}
