import type { ArgValue, EffectorSpec } from "./effectors";
import { argProblem } from "./effectors";
import type { ShowingMarks } from "./shared/marks";
import { humanise } from "./shared/words";
import { durationOf } from "./workspace";

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

export function launchSteps(plan: LaunchPlan, chunks = 20): readonly LaunchStep[] {
  const steps: LaunchStep[] = [...actionsAt(plan, 0)];
  let previous = 0;
  for (const at of boundaries(plan, chunks)) {
    steps.push({ kind: "advance", byNs: (at - previous) * 1e9, toS: at });
    if (at < plan.durationS) steps.push(...actionsAt(plan, at));
    previous = at;
  }
  return steps;
}

export type SpecOf = (entity: string, effector: string) => EffectorSpec | undefined;

export function failureWords(step: LaunchStep | null): string {
  if (step?.kind === "inject")
    return `The fault "${humanise(step.fault)}" could not be added to ${step.entity}. Check that this plant still has it, choose a fault, and start again.`;
  if (step?.kind === "invoke")
    return `${step.entity} refused the effector call ${humanise(step.effector)}. Check the call and start again.`;
  return "The run stopped before it finished. Check the plan and start again.";
}

export interface KeptPlan {
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

export interface PlanProblem {
  readonly field: "seed" | "duration" | "call" | number;
  readonly call?: number;
  readonly part?: "instance" | "effector" | "minute" | "args";
  readonly text: string;
}

function faultProblems(plan: LaunchPlan): readonly PlanProblem[] {
  return plan.faults.flatMap((fault, index): PlanProblem[] => {
    const name = `Fault ${String(index + 1)}`;
    if (fault.entity === "" && fault.fault === "")
      return [{ field: index, text: `${name} needs an instance and a fault.` }];
    if (fault.entity === "") return [{ field: index, text: `${name} needs an instance.` }];
    if (fault.fault === "") return [{ field: index, text: `${name} needs a fault.` }];
    if (!Number.isInteger(fault.atS) || fault.atS < 0 || fault.atS >= plan.durationS)
      return [{ field: index, text: `${name} starts inside the run, before its end.` }];
    return [];
  });
}

type CallProblem = Pick<PlanProblem, "part" | "text">;

function callProblem(call: PlannedCall, name: string, plan: LaunchPlan, specOf: SpecOf): CallProblem | null {
  if (call.entity === "" && call.effector === "")
    return { part: "instance", text: `${name} needs an instance and an effector.` };
  if (call.entity === "") return { part: "instance", text: `${name} needs an instance.` };
  if (call.effector === "") return { part: "effector", text: `${name} needs an effector.` };
  if (!Number.isInteger(call.atS) || call.atS < 0 || call.atS >= plan.durationS)
    return { part: "minute", text: `${name} is made inside the run, before its end.` };
  const spec = specOf(call.entity, call.effector);
  if (spec === undefined)
    return { part: "effector", text: `${name}: ${call.entity} has no effector ${humanise(call.effector)}.` };
  const problem = argProblem(spec, call.args);
  return problem === "" ? null : { part: "args", text: `${name}: ${problem}` };
}

function callProblems(plan: LaunchPlan, specOf: SpecOf): readonly PlanProblem[] {
  return plan.calls.flatMap((call, index): PlanProblem[] => {
    const found = callProblem(call, `Effector call ${String(index + 1)}`, plan, specOf);
    return found === null ? [] : [{ field: "call", call: index, ...found }];
  });
}

const anySpec: SpecOf = (_entity, effector) => ({ name: effector, entityArg: null, args: [] });

export function planProblems(plan: LaunchPlan, specOf: SpecOf = anySpec): readonly PlanProblem[] {
  const problems: PlanProblem[] = [];
  if (!Number.isSafeInteger(plan.seed) || plan.seed < 0)
    problems.push({ field: "seed", text: "The seed is a whole number, 0 or more." });
  if (!Number.isInteger(plan.durationS) || plan.durationS < 60)
    problems.push({ field: "duration", text: "The duration is at least one minute." });
  return [...problems, ...faultProblems(plan), ...callProblems(plan, specOf)];
}

function counted(n: number, one: string, many: string): string {
  return `${String(n)} ${n === 1 ? one : many}`;
}

export function disturbanceWords(plan: LaunchPlan): string {
  const faults = plan.faults.length;
  const calls = plan.calls.length;
  const injected = faults === 0 ? "no faults" : counted(faults, "fault", "faults");
  return calls === 0 ? injected : `${injected} and ${counted(calls, "effector call", "effector calls")}`;
}

export function planSummary(plan: LaunchPlan, length: string): string {
  const faults = plan.faults.length;
  const calls = plan.calls.length;
  return [
    `${length} of simulated time`,
    plan.explore ? "Explore" : "Evaluate",
    faults === 0 ? "no faults" : counted(faults, "fault", "faults"),
    calls === 0 ? "" : counted(calls, "effector call", "effector calls"),
  ]
    .filter((part) => part !== "")
    .join(" · ");
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

export function outcomeOf(call: PlannedCall, value: unknown, atMs: number): CallOutcome {
  const ack = typeof value === "object" && value !== null ? (value as Record<string, unknown>) : {};
  const accepted = ack.accepted === true;
  const eta = typeof ack.effect_eta_ns === "number" ? ack.effect_eta_ns : null;
  const reason = typeof ack.reason === "string" ? ack.reason : "";
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

const effectorShownS = 60;
const effectorHeldMs = 3000;

function callWords(call: CallOutcome): string {
  const name = humanise(call.effector);
  return call.accepted ? `${name} · accepted` : `${name} · refused: ${call.reason}`;
}

function faultsAt(plan: LaunchPlan, atS: number): Map<string, readonly string[]> {
  const faulted = new Map<string, readonly string[]>();
  for (const planned of plan.faults) {
    if (planned.entity === "" || planned.fault === "" || planned.atS > atS) continue;
    faulted.set(planned.entity, [...(faulted.get(planned.entity) ?? []), humanise(planned.fault)]);
  }
  return faulted;
}

function callShown(call: CallOutcome, atS: number, atMs: number): boolean {
  if (call.atS > atS) return false;
  return atS < call.atS + effectorShownS || atMs - call.atMs < effectorHeldMs;
}

function callsAt(calls: readonly CallOutcome[], atS: number, atMs: number): Map<string, string> {
  const shown = new Map<string, string>();
  for (const call of calls) {
    if (callShown(call, atS, atMs)) shown.set(call.entity, callWords(call));
  }
  return shown;
}

export function showingMarks(plan: LaunchPlan, calls: readonly CallOutcome[], atS: number, atMs: number): ShowingMarks {
  return { faults: faultsAt(plan, atS), effectors: callsAt(calls, atS, atMs) };
}

export function rateWords(rate: number | null): string {
  return rate === null ? "" : `${String(Math.max(1, Math.round(rate)))} s per s`;
}

export function clockOf(seconds: number): string {
  const whole = Math.max(0, Math.round(seconds));
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${pad(Math.floor(whole / 3600))}:${pad(Math.floor((whole % 3600) / 60))}:${pad(whole % 60)}`;
}
