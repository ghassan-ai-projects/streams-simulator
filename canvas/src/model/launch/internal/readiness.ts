import { durationOf } from "../../shared/duration";
import { counted } from "../../shared/words";
import type { LaunchPlan } from "./launch";
import { watchWords } from "./launch";
import { modeWords } from "./words";

export interface ReadinessLine {
  readonly label: string;
  readonly text: string;
}

export function scoringWords(explore: boolean): string {
  const none = "no outside system is scoring this run.";
  if (explore) return `Unscored investigation: ${none}`;
  return `Unscored blind run: ${none} It stays blind until it ends; to test an outside system, set up a connected session instead.`;
}

function lengthWords(plan: LaunchPlan): string {
  const length = durationOf(plan.durationS);
  if (length === "" || !Number.isFinite(plan.seed)) return "Not set";
  return `${length} of simulated time, seed ${String(plan.seed)}`;
}

function paceWords(plan: LaunchPlan): string {
  return plan.pace === "watch" ? `Watch, ${watchWords(plan.durationS)}` : "As fast as possible";
}

function faultWords(plan: LaunchPlan): string {
  const recovering = plan.faults.filter((fault) => fault.untilS !== undefined).length;
  const faults = counted(plan.faults.length, "fault");
  return recovering === 0 ? faults : `${faults} (${String(recovering)} recover${recovering === 1 ? "s" : ""})`;
}

function plannedWords(plan: LaunchPlan): string {
  const parts = [
    plan.faults.length === 0 ? "" : faultWords(plan),
    plan.perturbations.length === 0 ? "" : counted(plan.perturbations.length, "delivery condition"),
    plan.calls.length === 0 ? "" : counted(plan.calls.length, "effector call"),
  ].filter((part) => part !== "");
  return parts.length === 0 ? "Nothing planned: the plant runs undisturbed" : parts.join(" · ");
}

export function readinessOf(plan: LaunchPlan, plantName: string): readonly ReadinessLine[] {
  return [
    { label: "Plant", text: plantName },
    { label: "Run", text: plan.name ?? "Unnamed" },
    { label: "Purpose", text: plan.purpose ?? "None stated: an investigation without a stated question" },
    { label: "Mode", text: `${plan.explore ? "Explore" : "Evaluate"}. ${modeWords(plan.explore)}` },
    { label: "Length", text: lengthWords(plan) },
    { label: "Pace", text: paceWords(plan) },
    { label: "Planned", text: plannedWords(plan) },
    { label: "Scoring", text: scoringWords(plan.explore) },
  ];
}
