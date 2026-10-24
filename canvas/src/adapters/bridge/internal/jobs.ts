import type { ImmediateCall, InjectedFault, JobRequest, JobStatus, LaunchGateway } from "../../../app/ports";
import { isRecord, text } from "../../../model/shared/json";
import { launchFailure } from "./failures";
import { faultArgs, jobFaults, jobPerturbations, perturbationArgs } from "./jobPlan";
import type { Transport } from "./http";
import { action } from "./http";

type JobGateway = Pick<
  LaunchGateway,
  | "startJob"
  | "jobStatus"
  | "stopJob"
  | "listJobs"
  | "pauseJob"
  | "resumeJob"
  | "stepJob"
  | "setJobRate"
  | "injectFaultNow"
  | "invokeEffectorNow"
>;

type JobCall = JobStatus["calls"][number];
type JobState = JobStatus["state"];

const states: readonly JobState[] = ["running", "paused", "stopping", "completed", "stopped", "failed"];

function finite(value: unknown): number | null {
  return typeof value === "number" && Number.isFinite(value) ? value : null;
}

function optionalText(value: unknown): string | null {
  if (value === undefined) return "";
  return typeof value === "string" ? value : null;
}

function optionalNumber(value: unknown): number | null | undefined {
  if (value === undefined) return null;
  return finite(value) ?? undefined;
}

function optionalFlag(value: unknown): boolean | null {
  if (value === undefined) return false;
  return typeof value === "boolean" ? value : null;
}

function jobCall(value: unknown): JobCall | null {
  if (!isRecord(value) || typeof value.accepted !== "boolean") return null;
  const atS = finite(value.at_s);
  const reason = optionalText(value.reason);
  const eta = optionalNumber(value.effect_eta_ns);
  const immediate = optionalFlag(value.immediate);
  const entity = text(value.entity_id);
  const effector = text(value.effector);
  if (atS === null || reason === null || eta === undefined || immediate === null) return null;
  if (entity === "" || effector === "") return null;
  return { entity, effector, atS, accepted: value.accepted, reason, effectEtaNs: eta, immediate };
}

function jobCalls(value: unknown): readonly JobCall[] | null {
  if (!Array.isArray(value)) return null;
  const calls = value.map(jobCall);
  return calls.every((call) => call !== null) ? calls : null;
}

function stateOf(value: unknown): JobState | null {
  return states.find((state) => state === value) ?? null;
}

interface Ids {
  readonly worldId: string;
  readonly runId: string;
  readonly domain: string;
}

function idsOf(value: Record<string, unknown>): Ids | null {
  const ids = { worldId: text(value.world_id), runId: text(value.run_id), domain: text(value.domain) };
  return Object.values(ids).some((id) => id === "") ? null : ids;
}

interface Span {
  readonly startNs: number;
  readonly reachedNs: number;
  readonly requestedEndNs: number;
}

function spanOf(value: Record<string, unknown>): Span | null {
  const startNs = finite(value.start_ns);
  const reachedNs = finite(value.reached_ns);
  const requestedEndNs = finite(value.requested_end_ns);
  if (startNs === null || reachedNs === null || requestedEndNs === null) return null;
  return { startNs, reachedNs, requestedEndNs };
}

interface Pace {
  readonly seed: number;
  readonly rate: number;
  readonly explore: boolean;
}

function paceOf(value: Record<string, unknown>): Pace | null {
  const seed = finite(value.seed);
  const rate = finite(value.rate);
  if (seed === null || rate === null || rate < 0 || typeof value.explore !== "boolean") return null;
  return { seed, rate, explore: value.explore };
}

interface Authored {
  readonly name: string;
  readonly purpose: string;
  readonly faults: JobStatus["faults"];
  readonly perturbations: JobStatus["perturbations"];
}

function authoredOf(value: Record<string, unknown>): Authored | null {
  const name = optionalText(value.name);
  const purpose = optionalText(value.purpose);
  const faults = jobFaults(value.faults);
  const perturbations = jobPerturbations(value.perturbations);
  if (name === null || purpose === null || faults === null || perturbations === null) return null;
  return { name, purpose, faults, perturbations };
}

export function jobStatus(value: unknown): JobStatus | null {
  if (!isRecord(value)) return null;
  const ids = idsOf(value);
  const state = stateOf(value.state);
  const calls = jobCalls(value.calls);
  const failure = optionalText(value.failure);
  const span = spanOf(value);
  const pace = paceOf(value);
  const authored = authoredOf(value);
  if (ids === null || state === null || calls === null || failure === null || authored === null) return null;
  return span === null || pace === null ? null : { ...ids, ...span, ...pace, ...authored, state, calls, failure };
}

export function jobList(value: unknown): readonly JobStatus[] | null {
  if (!isRecord(value) || !Array.isArray(value.jobs)) return null;
  const jobs = value.jobs.map(jobStatus);
  return jobs.every((job) => job !== null) ? jobs : null;
}

export function injectedFault(value: unknown): InjectedFault | null {
  const status = jobStatus(value);
  const faultId = isRecord(value) ? text(value.fault_id) : "";
  return status === null || faultId === "" ? null : { status, faultId };
}

export function immediateCall(value: unknown): ImmediateCall | null {
  const status = jobStatus(value);
  const call = isRecord(value) ? jobCall(value.call) : null;
  return status === null || call === null ? null : { status, call };
}

function startArgs(request: JobRequest): Readonly<Record<string, unknown>> {
  const { name, purpose, domain, seed, explore, durationS, rate, faults, perturbations, calls } = request;
  return {
    ...(name === undefined ? {} : { name }),
    ...(purpose === undefined ? {} : { purpose }),
    domain,
    seed,
    explore,
    duration_s: durationS,
    rate,
    faults: faults.map(faultArgs),
    perturbations: perturbations.map(perturbationArgs),
    calls: calls.map((call) => ({ entity_id: call.entity, effector: call.effector, at_s: call.atS, args: call.args })),
  };
}

export function jobGateway(transport: Transport): JobGateway {
  return {
    startJob: (request) => action(transport, "job.start", startArgs(request), jobStatus, launchFailure("start")),
    jobStatus: (worldId) => action(transport, "job.status", { world_id: worldId }, jobStatus, launchFailure("status")),
    stopJob: (worldId) => action(transport, "job.stop", { world_id: worldId }, jobStatus, launchFailure("stop")),
    listJobs: () => action(transport, "job.list", {}, jobList, launchFailure("list")),
    ...controlGateway(transport),
  };
}

function controlGateway(transport: Transport) {
  const control = (name: string, args: Readonly<Record<string, unknown>>) =>
    action(transport, name, args, jobStatus, launchFailure("control"));
  return {
    pauseJob: (worldId: string) => control("job.pause", { world_id: worldId }),
    resumeJob: (worldId: string) => control("job.resume", { world_id: worldId }),
    stepJob: (worldId: string) => control("job.step", { world_id: worldId }),
    setJobRate: (worldId: string, rate: number) => control("job.rate", { world_id: worldId, rate }),
    injectFaultNow: (worldId: string, entity: string, fault: string) =>
      action(
        transport,
        "job.inject_fault",
        { world_id: worldId, entity_id: entity, fault },
        injectedFault,
        launchFailure("control"),
      ),
    invokeEffectorNow: (worldId: string, entity: string, effector: string, args: Readonly<Record<string, unknown>>) =>
      action(
        transport,
        "job.invoke_effector",
        { world_id: worldId, entity_id: entity, effector, args },
        immediateCall,
        launchFailure("control"),
      ),
  };
}
