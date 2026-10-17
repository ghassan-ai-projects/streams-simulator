import type { ArgValue } from "../../plant";
import { humanise } from "../../shared/words";
import { durationOf } from "../../shared/duration";

interface PlannedFault {
  readonly entity: string;
  readonly fault: string;
  readonly atS: number;
}

export interface PlannedCall {
  readonly entity: string;
  readonly effector: string;
  readonly atS: number;
  readonly args: Readonly<Record<string, ArgValue>>;
}

export interface LaunchPlan {
  readonly seed: number;
  readonly durationS: number;
  readonly explore: boolean;
  readonly faults: readonly PlannedFault[];
  readonly calls: readonly PlannedCall[];
  readonly pace?: "watch" | "fast";
}

function watchRate(durationS: number): number {
  return Math.min(3600, Math.max(60, durationS / 30));
}

export function paceOf(plan: LaunchPlan): number | null {
  return plan.pace === "watch" ? watchRate(plan.durationS) : null;
}

export function watchWords(durationS: number): string {
  const seconds = Math.max(1, Math.round(durationS / watchRate(durationS)));
  return `about ${durationOf(seconds)}`;
}

export type LaunchStep =
  | { readonly kind: "advance"; readonly byNs: number; readonly toS: number }
  | { readonly kind: "inject"; readonly entity: string; readonly fault: string }
  | {
      readonly kind: "invoke";
      readonly entity: string;
      readonly effector: string;
      readonly args: Readonly<Record<string, ArgValue>>;
      readonly atS: number;
    };

function boundaries(plan: LaunchPlan, chunks: number): readonly number[] {
  const step = Math.max(1, Math.ceil(plan.durationS / chunks));
  const marks = new Set<number>([plan.durationS]);
  for (let at = step; at < plan.durationS; at += step) marks.add(at);
  for (const fault of plan.faults) marks.add(fault.atS);
  for (const call of plan.calls) marks.add(call.atS);
  return [...marks].filter((at) => at > 0).sort((a, b) => a - b);
}

function actionsAt(plan: LaunchPlan, atS: number): readonly LaunchStep[] {
  const injections = plan.faults
    .filter((fault) => fault.atS === atS)
    .map((fault): LaunchStep => ({ kind: "inject", entity: fault.entity, fault: fault.fault }));
  const invocations = plan.calls
    .filter((call) => call.atS === atS)
    .map((call): LaunchStep => ({ kind: "invoke", ...call }));
  return [...injections, ...invocations];
}

export function launchSteps(plan: LaunchPlan, chunks: number): readonly LaunchStep[] {
  const steps: LaunchStep[] = [...actionsAt(plan, 0)];
  let previous = 0;
  for (const at of boundaries(plan, chunks)) {
    steps.push({ kind: "advance", byNs: (at - previous) * 1e9, toS: at });
    if (at < plan.durationS) steps.push(...actionsAt(plan, at));
    previous = at;
  }
  return steps;
}

export function stepsFor(plan: LaunchPlan): readonly LaunchStep[] {
  return launchSteps(plan, plan.pace === "watch" ? 30 : 20);
}

export function failureWords(step: LaunchStep | null): string {
  if (step?.kind === "inject")
    return `The fault "${humanise(step.fault)}" could not be added to ${step.entity}. Check that this plant still has it, choose a fault, and start again.`;
  if (step?.kind === "invoke")
    return `${step.entity} refused the effector call ${humanise(step.effector)}. Check the call and start again.`;
  return "The run stopped before it finished. Check the plan and start again.";
}

interface KeptPlan {
  readonly plan: LaunchPlan;
  readonly dropped: readonly string[];
}

function droppedWords(fault: PlannedFault, unknownInstance: boolean): string {
  return unknownInstance
    ? `a fault on ${fault.entity} (no longer in this plant)`
    : `${humanise(fault.fault)} on ${fault.entity} (no longer in this plant)`;
}

export function keepKnownFaults(plan: LaunchPlan, faultsOf: (entity: string) => readonly string[] | null): KeptPlan {
  const dropped: string[] = [];
  const faults = plan.faults.map((fault): PlannedFault => {
    const known = faultsOf(fault.entity);
    if (known?.includes(fault.fault)) return fault;
    dropped.push(droppedWords(fault, known === null));
    return known === null ? { entity: "", fault: "", atS: fault.atS } : { ...fault, fault: "" };
  });
  return { plan: { ...plan, faults }, dropped };
}

export interface CallOutcome {
  readonly entity: string;
  readonly effector: string;
  readonly atS: number;
  readonly accepted: boolean;
  readonly reason: string;
  readonly effectEtaNs: number | null;
  readonly atMs: number;
}

export interface Acknowledgement {
  readonly accepted: boolean;
  readonly reason: string;
  readonly effectEtaNs: number | null;
}

export function outcomeOf(call: PlannedCall, ack: Acknowledgement, atMs: number): CallOutcome {
  const { accepted, effectEtaNs: eta, reason } = ack;
  const { entity, effector, atS } = call;
  return {
    entity,
    effector,
    atS,
    accepted,
    reason: accepted ? "" : reason || "no reason given",
    effectEtaNs: eta,
    atMs,
  };
}

export function rateWords(rate: number | null): string {
  return rate === null ? "" : `${String(Math.max(1, Math.round(rate)))} s per s`;
}
