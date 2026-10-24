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
  readonly recovered?: boolean;
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

export function rateWords(rate: number | null): string {
  return rate === null ? "" : `${String(Math.max(1, Math.round(rate)))} s per s`;
}
