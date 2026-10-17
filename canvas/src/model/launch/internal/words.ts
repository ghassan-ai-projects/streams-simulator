import { durationOf } from "../../shared/duration";
import { counted, evaluateRule } from "../../shared/words";
import type { LaunchPlan } from "./launch";
import { watchWords } from "./launch";
import type { PlanProblem } from "./planProblems";

function faultCountWords(plan: LaunchPlan): string {
  return plan.faults.length === 0 ? "no faults" : counted(plan.faults.length, "fault");
}

function callCountWords(plan: LaunchPlan): string {
  return plan.calls.length === 0 ? "" : counted(plan.calls.length, "effector call");
}

export function disturbanceWords(plan: LaunchPlan): string {
  const faults = faultCountWords(plan);
  const calls = callCountWords(plan);
  return calls === "" ? faults : `${faults} and ${calls}`;
}

export function planSummary(plan: LaunchPlan, length: string): string {
  return [
    `${length} of simulated time`,
    plan.explore ? "Explore" : "Evaluate",
    faultCountWords(plan),
    callCountWords(plan),
  ]
    .filter((part) => part !== "")
    .join(" · ");
}

export function modeWords(explore: boolean): string {
  return explore ? "The truth is visible after the run; not for scoring." : evaluateRule;
}

export function problemWords(problems: readonly PlanProblem[]): string {
  if (problems.length === 0) return "";
  return `${String(problems.length)} to fix before starting: ${problems.map((problem) => problem.text).join(" ")}`;
}

function startWords(plan: LaunchPlan): string {
  const length = durationOf(plan.durationS) || "the chosen time";
  const pace = plan.pace === "watch" ? `paced to take ${watchWords(plan.durationS)}` : "as fast as it can";
  return `Start simulates ${length} of plant time in ${plan.explore ? "Explore" : "Evaluate"} mode with ${disturbanceWords(plan)}, ${pace}, then saves the run and opens its review.`;
}

function fieldName(problem: PlanProblem): string {
  switch (problem.field) {
    case "seed":
      return "Seed";
    case "duration":
      return "Duration";
    case "call":
      return `Effector call ${String(problem.index + 1)}`;
    case "fault":
      return `Fault ${String(problem.index + 1)}`;
  }
}

export function planCue(shown: readonly PlanProblem[], plan: LaunchPlan): string {
  if (shown.length === 0) return startWords(plan);
  const names = shown.map(fieldName).join(", ");
  return shown.length === 1 ? `${names} needs attention before starting.` : `${names} need attention before starting.`;
}
