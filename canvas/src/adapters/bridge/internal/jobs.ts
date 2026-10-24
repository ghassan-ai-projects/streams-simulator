import type { JobRequest, JobStatus, LaunchGateway } from "../../../app/ports";
import { isRecord, text } from "../../../model/shared/json";
import { launchFailure } from "./failures";
import type { Transport } from "./http";
import { action } from "./http";

type JobGateway = Pick<LaunchGateway, "startJob" | "jobStatus" | "stopJob" | "listJobs">;

type JobCall = JobStatus["calls"][number];
type JobState = JobStatus["state"];

const states: readonly JobState[] = ["running", "stopping", "completed", "stopped", "failed"];

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

function jobCall(value: unknown): JobCall | null {
  if (!isRecord(value) || typeof value.accepted !== "boolean") return null;
  const atS = finite(value.at_s);
  const reason = optionalText(value.reason);
  const eta = optionalNumber(value.effect_eta_ns);
  const entity = text(value.entity_id);
  const effector = text(value.effector);
  if (atS === null || reason === null || eta === undefined || entity === "" || effector === "") return null;
  return { entity, effector, atS, accepted: value.accepted, reason, effectEtaNs: eta };
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

export function jobStatus(value: unknown): JobStatus | null {
  if (!isRecord(value)) return null;
  const ids = idsOf(value);
  const state = stateOf(value.state);
  const calls = jobCalls(value.calls);
  const failure = optionalText(value.failure);
  const span = spanOf(value);
  const pace = paceOf(value);
  if (ids === null || state === null || calls === null || failure === null) return null;
  return span === null || pace === null ? null : { ...ids, ...span, ...pace, state, calls, failure };
}

export function jobList(value: unknown): readonly JobStatus[] | null {
  if (!isRecord(value) || !Array.isArray(value.jobs)) return null;
  const jobs = value.jobs.map(jobStatus);
  return jobs.every((job) => job !== null) ? jobs : null;
}

function startArgs(request: JobRequest): Readonly<Record<string, unknown>> {
  const { domain, seed, explore, durationS, rate, faults, calls } = request;
  return {
    domain,
    seed,
    explore,
    duration_s: durationS,
    rate,
    faults: faults.map((fault) => ({ entity_id: fault.entity, fault: fault.fault, at_s: fault.atS })),
    calls: calls.map((call) => ({ entity_id: call.entity, effector: call.effector, at_s: call.atS, args: call.args })),
  };
}

export function jobGateway(transport: Transport): JobGateway {
  return {
    startJob: (request) => action(transport, "job.start", startArgs(request), jobStatus, launchFailure("start")),
    jobStatus: (worldId) => action(transport, "job.status", { world_id: worldId }, jobStatus, launchFailure("status")),
    stopJob: (worldId) => action(transport, "job.stop", { world_id: worldId }, jobStatus, launchFailure("stop")),
    listJobs: () => action(transport, "job.list", {}, jobList, launchFailure("list")),
  };
}
