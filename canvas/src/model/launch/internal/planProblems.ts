import type { EffectorSpec } from "../../plant";
import { argProblem } from "../../plant";
import { humanise } from "../../shared/words";
import type { LaunchPlan, PlannedCall } from "./launch";

type SpecOf = (entity: string, effector: string) => EffectorSpec | undefined;

type CallPart = "instance" | "effector" | "minute" | "args";

export interface CallProblem {
  readonly field: "call";
  readonly index: number;
  readonly part: CallPart;
  readonly text: string;
}

export type PlanProblem =
  | { readonly field: "seed"; readonly text: string }
  | { readonly field: "duration"; readonly text: string }
  | { readonly field: "fault"; readonly index: number; readonly text: string }
  | CallProblem;

function insideRun(plan: LaunchPlan, atS: number): boolean {
  return Number.isInteger(atS) && atS >= 0 && atS < plan.durationS;
}

function faultText(plan: LaunchPlan, index: number): string | null {
  const fault = plan.faults[index];
  if (fault === undefined) return null;
  const name = `Fault ${String(index + 1)}`;
  if (fault.entity === "" && fault.fault === "") return `${name} needs an instance and a fault.`;
  if (fault.entity === "") return `${name} needs an instance.`;
  if (fault.fault === "") return `${name} needs a fault.`;
  return insideRun(plan, fault.atS) ? null : `${name} starts inside the run, before its end.`;
}

function faultProblems(plan: LaunchPlan): readonly PlanProblem[] {
  return plan.faults.flatMap((_fault, index): PlanProblem[] => {
    const text = faultText(plan, index);
    return text === null ? [] : [{ field: "fault", index, text }];
  });
}

type CallFinding = Pick<CallProblem, "part" | "text">;

type CallTarget = Pick<PlannedCall, "entity" | "effector" | "args">;

export function targetFinding(call: CallTarget, name: string): CallFinding | null {
  if (call.entity === "" && call.effector === "")
    return { part: "instance", text: `${name} needs an instance and an effector.` };
  if (call.entity === "") return { part: "instance", text: `${name} needs an instance.` };
  if (call.effector === "") return { part: "effector", text: `${name} needs an effector.` };
  return null;
}

export function specFinding(call: CallTarget, name: string, specOf: SpecOf): CallFinding | null {
  const spec = specOf(call.entity, call.effector);
  if (spec === undefined)
    return { part: "effector", text: `${name}: ${call.entity} has no effector ${humanise(call.effector)}.` };
  const problem = argProblem(spec, call.args);
  return problem === "" ? null : { part: "args", text: `${name}: ${problem}` };
}

function callFinding(call: PlannedCall, name: string, plan: LaunchPlan, specOf: SpecOf): CallFinding | null {
  const target = targetFinding(call, name);
  if (target !== null) return target;
  if (!insideRun(plan, call.atS)) return { part: "minute", text: `${name} is made inside the run, before its end.` };
  return specFinding(call, name, specOf);
}

function callProblems(plan: LaunchPlan, specOf: SpecOf): readonly PlanProblem[] {
  return plan.calls.flatMap((call, index): PlanProblem[] => {
    const found = callFinding(call, `Effector call ${String(index + 1)}`, plan, specOf);
    return found === null ? [] : [{ field: "call", index, ...found }];
  });
}

function runProblems(plan: LaunchPlan): readonly PlanProblem[] {
  const problems: PlanProblem[] = [];
  if (!Number.isSafeInteger(plan.seed) || plan.seed < 0)
    problems.push({ field: "seed", text: "The seed is a whole number, 0 or more." });
  if (!Number.isInteger(plan.durationS) || plan.durationS < 60)
    problems.push({ field: "duration", text: "The duration is at least one minute." });
  return problems;
}

export function planProblems(plan: LaunchPlan, specOf: SpecOf): readonly PlanProblem[] {
  return [...runProblems(plan), ...faultProblems(plan), ...callProblems(plan, specOf)];
}
