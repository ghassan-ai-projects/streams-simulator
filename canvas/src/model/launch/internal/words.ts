import { durationOf } from "../../shared/duration";
import { evaluateRule } from "../../shared/words";
import type { LaunchPlan, PlanProblem } from "./launch";
import { disturbanceWords, watchWords } from "./launch";

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
  if (problem.field === "seed") return "Seed";
  if (problem.field === "duration") return "Duration";
  if (problem.field === "call") return `Effector call ${String((problem.call ?? 0) + 1)}`;
  return `Fault ${String(problem.field + 1)}`;
}

export function planCue(shown: readonly PlanProblem[], plan: LaunchPlan): string {
  if (shown.length === 0) return startWords(plan);
  const names = shown.map(fieldName).join(", ");
  return shown.length === 1 ? `${names} needs attention before starting.` : `${names} need attention before starting.`;
}
