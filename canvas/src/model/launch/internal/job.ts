import type { CallOutcome, LaunchPlan } from "./launch";
import { paceOf } from "./launch";

type JobState = "running" | "stopping" | "completed" | "stopped" | "failed";

interface JobCall {
  readonly entity: string;
  readonly effector: string;
  readonly atS: number;
  readonly accepted: boolean;
  readonly reason: string;
  readonly effectEtaNs: number | null;
}

export interface JobStatus {
  readonly worldId: string;
  readonly runId: string;
  readonly domain: string;
  readonly seed: number;
  readonly state: JobState;
  readonly startNs: number;
  readonly reachedNs: number;
  readonly requestedEndNs: number;
  readonly rate: number;
  readonly explore: boolean;
  readonly calls: readonly JobCall[];
  readonly failure: string;
}

export interface JobRequest {
  readonly domain: string;
  readonly seed: number;
  readonly explore: boolean;
  readonly durationS: number;
  readonly rate: number;
  readonly faults: LaunchPlan["faults"];
  readonly calls: LaunchPlan["calls"];
}

export const unobserved = Number.NEGATIVE_INFINITY;

export function jobRequestOf(domain: string, plan: LaunchPlan): JobRequest {
  const { seed, explore, durationS, faults, calls } = plan;
  return { domain, seed, explore, durationS, rate: paceOf(plan) ?? 0, faults, calls };
}

export function jobActive(status: JobStatus): boolean {
  return status.state === "running" || status.state === "stopping";
}

export function reachedS(status: JobStatus): number {
  return (status.reachedNs - status.startNs) / 1e9;
}

export function plannedS(status: JobStatus): number {
  return (status.requestedEndNs - status.startNs) / 1e9;
}

export function recoveredPlanOf(status: JobStatus): LaunchPlan {
  const { seed, explore } = status;
  return { seed, durationS: plannedS(status), explore, faults: [], calls: [], recovered: true };
}

export function pacedRate(status: JobStatus): number | null {
  return status.rate > 0 ? status.rate : null;
}

function outcomeOf(call: JobCall, atMs: number): CallOutcome {
  const { entity, effector, atS, accepted, reason, effectEtaNs } = call;
  return { entity, effector, atS, accepted, reason: accepted ? "" : reason || "no reason given", effectEtaNs, atMs };
}

export function observedCalls(known: readonly CallOutcome[], status: JobStatus, atMs: number): readonly CallOutcome[] {
  if (status.calls.length <= known.length) return known;
  return [...known, ...status.calls.slice(known.length).map((call) => outcomeOf(call, atMs))];
}

function sentence(message: string): string {
  return `${message.trim().replace(/[.\s]+$/, "")}.`;
}

export function startFailureWords(message: string): string {
  return `${sentence(message)} Check the plan and start again.`;
}

export function jobFailureWords(reason: string): string {
  if (reason.trim() === "") return "The job stopped without saying why. Check the plan and start again.";
  return `${sentence(reason)} Check the plan and start again.`;
}

export const lostWords =
  "The browser lost contact with the run. It may still be going: open the workspace to see where it stands.";
