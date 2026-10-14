interface PlannedFault {
  readonly entity: string;
  readonly fault: string;
  readonly atS: number;
}

export interface LaunchPlan {
  readonly seed: number;
  readonly durationS: number;
  readonly explore: boolean;
  readonly faults: readonly PlannedFault[];
}

export type LaunchStep =
  | { readonly kind: "advance"; readonly byNs: number; readonly toS: number }
  | { readonly kind: "inject"; readonly entity: string; readonly fault: string };

function boundaries(plan: LaunchPlan, chunks: number): readonly number[] {
  const step = Math.max(1, Math.ceil(plan.durationS / chunks));
  const marks = new Set<number>([plan.durationS]);
  for (let at = step; at < plan.durationS; at += step) marks.add(at);
  for (const fault of plan.faults) marks.add(fault.atS);
  return [...marks].filter((at) => at > 0).sort((a, b) => a - b);
}

function injectionsAt(plan: LaunchPlan, atS: number): readonly LaunchStep[] {
  return plan.faults
    .filter((fault) => fault.atS === atS)
    .map((fault): LaunchStep => ({ kind: "inject", entity: fault.entity, fault: fault.fault }));
}

export function launchSteps(plan: LaunchPlan, chunks = 20): readonly LaunchStep[] {
  const steps: LaunchStep[] = [...injectionsAt(plan, 0)];
  let previous = 0;
  for (const at of boundaries(plan, chunks)) {
    steps.push({ kind: "advance", byNs: (at - previous) * 1e9, toS: at });
    if (at < plan.durationS) steps.push(...injectionsAt(plan, at));
    previous = at;
  }
  return steps;
}

export interface PlanProblem {
  readonly field: "seed" | "duration" | number;
  readonly text: string;
}

function faultProblems(plan: LaunchPlan): readonly PlanProblem[] {
  return plan.faults.flatMap((fault, index): PlanProblem[] => {
    const name = `Fault ${String(index + 1)}`;
    if (fault.entity === "" || fault.fault === "")
      return [{ field: index, text: `${name} needs an instance and a fault.` }];
    if (!Number.isInteger(fault.atS) || fault.atS < 0 || fault.atS >= plan.durationS)
      return [{ field: index, text: `${name} starts inside the run, before its end.` }];
    return [];
  });
}

export function planProblems(plan: LaunchPlan): readonly PlanProblem[] {
  const problems: PlanProblem[] = [];
  if (!Number.isSafeInteger(plan.seed) || plan.seed < 0)
    problems.push({ field: "seed", text: "The seed is a whole number, 0 or more." });
  if (!Number.isInteger(plan.durationS) || plan.durationS < 60)
    problems.push({ field: "duration", text: "The duration is at least one minute." });
  return [...problems, ...faultProblems(plan)];
}

export function planSummary(plan: LaunchPlan, length: string): string {
  const faults = plan.faults.length;
  return [
    `${length} of simulated time`,
    plan.explore ? "Explore" : "Evaluate",
    faults === 0 ? "no faults" : `${String(faults)} ${faults === 1 ? "fault" : "faults"}`,
  ].join(" · ");
}

export function speedOf(atS: number, wallMs: number): string {
  if (wallMs < 1000 || atS <= 0) return "";
  const speed = atS / (wallMs / 1000);
  return speed >= 10 ? `about ${String(Math.round(speed))}× real time` : `about ${speed.toFixed(1)}× real time`;
}

export function clockOf(seconds: number): string {
  const whole = Math.max(0, Math.round(seconds));
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${pad(Math.floor(whole / 3600))}:${pad(Math.floor((whole % 3600) / 60))}:${pad(whole % 60)}`;
}
