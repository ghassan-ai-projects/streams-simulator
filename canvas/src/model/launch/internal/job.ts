import type { CallOutcome, LaunchPlan, PlannedCall, PlannedFault } from "./launch";
import { paceOf } from "./launch";
import type { PlannedPerturbation } from "./perturbation";

type JobState = "running" | "paused" | "stopping" | "completed" | "stopped" | "failed";

interface JobCall {
  readonly entity: string;
  readonly effector: string;
  readonly atS: number;
  readonly accepted: boolean;
  readonly reason: string;
  readonly effectEtaNs: number | null;
  readonly immediate: boolean;
}

export interface JobStatus {
  readonly name: string;
  readonly purpose: string;
  readonly faults: LaunchPlan["faults"];
  readonly perturbations: readonly PlannedPerturbation[];
  readonly plannedCalls: readonly PlannedCall[];
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
  readonly name?: string;
  readonly purpose?: string;
  readonly domain: string;
  readonly seed: number;
  readonly explore: boolean;
  readonly durationS: number;
  readonly rate: number;
  readonly faults: LaunchPlan["faults"];
  readonly perturbations: LaunchPlan["perturbations"];
  readonly calls: LaunchPlan["calls"];
}

export const unobserved = Number.NEGATIVE_INFINITY;

export function jobRequestOf(domain: string, plan: LaunchPlan): JobRequest {
  const { name, purpose, seed, explore, durationS, faults, perturbations, calls } = plan;
  return {
    ...(name === undefined ? {} : { name }),
    ...(purpose === undefined ? {} : { purpose }),
    domain,
    seed,
    explore,
    durationS,
    rate: paceOf(plan) ?? 0,
    faults,
    perturbations,
    calls,
  };
}

export interface InjectedFault {
  readonly status: JobStatus;
  readonly faultId: string;
}

export interface ImmediateCall {
  readonly status: JobStatus;
  readonly call: JobCall;
}

export function jobActive(status: JobStatus): boolean {
  return status.state === "running" || status.state === "paused" || status.state === "stopping";
}

export function jobPaused(status: JobStatus): boolean {
  return status.state === "paused";
}

export function reachedS(status: JobStatus): number {
  return (status.reachedNs - status.startNs) / 1e9;
}

export function plannedS(status: JobStatus): number {
  return (status.requestedEndNs - status.startNs) / 1e9;
}

export function recoveredPlanOf(status: JobStatus): LaunchPlan {
  const { seed, explore, name, purpose, perturbations, plannedCalls } = status;
  const faults = status.faults.filter((fault) => fault.immediate !== true);
  return {
    ...(name === "" ? {} : { name }),
    ...(purpose === "" ? {} : { purpose }),
    seed,
    durationS: plannedS(status),
    explore,
    faults,
    perturbations,
    calls: plannedCalls,
    recovered: true,
  };
}

function sameFault(known: PlannedFault, fault: PlannedFault): boolean {
  if (known.faultId !== undefined && fault.faultId !== undefined) return known.faultId === fault.faultId;
  return (
    known.entity === fault.entity && known.fault === fault.fault && Math.floor(known.atS) === Math.floor(fault.atS)
  );
}

export function observedInjected(known: readonly PlannedFault[], status: JobStatus): readonly PlannedFault[] {
  const fresh = status.faults
    .filter((fault) => fault.immediate === true)
    .filter((fault) => !known.some((each) => sameFault(each, fault)))
    .map(({ entity, fault, atS, faultId }) => ({ entity, fault, atS, ...(faultId === undefined ? {} : { faultId }) }));
  return fresh.length === 0 ? known : [...known, ...fresh];
}

export function pacedRate(status: JobStatus): number | null {
  return status.rate > 0 ? status.rate : null;
}

function outcomeOf(call: JobCall, atMs: number): CallOutcome {
  const { entity, effector, atS, accepted, reason, effectEtaNs, immediate } = call;
  const why = accepted ? "" : reason || "no reason given";
  return { entity, effector, atS, accepted, reason: why, effectEtaNs, immediate, atMs };
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
